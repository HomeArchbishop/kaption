package handler

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	vosk "github.com/HomeArchbishop/kaption/third_party/vosk/go"

	"github.com/gorilla/websocket"
)

var (
	fileIndexMap             = map[string]int{}
	timePrefixSumSecSliceMap = map[string][]float64{} // sum of durations of handled files before current index

	mutex = sync.Mutex{}
	cond  = sync.NewCond(&mutex)

	// prefixMapMutex = sync.Mutex{}
)

type VoiceRecognitionResult struct {
	Result []struct {
		Conf, End, Start float64
		Word             string
	}
	Text string
}

type CaptionResult struct {
	End, Start struct {
		Index int
		Time  float64
	}
	Text string
}

func BinaryMessageHandler(msg []byte, tempDirPath string, ffmpegPath string, rec *vosk.VoskRecognizer, conn *websocket.Conn, wg *sync.WaitGroup) {
	videoHash := string(msg[:16])
	fileHash := string(msg[16:32])
	fileIndex, _ := strconv.Atoi(string(msg[32:48]))
	fileBinary := msg[48:]

	tempFileName := videoHash + "_" + fileHash + "_" + strconv.Itoa(fileIndex)
	tempWavName := videoHash + "_" + fileHash + "_" + strconv.Itoa(fileIndex) + ".wav"
	tempFilePath := filepath.Join(tempDirPath, tempFileName)
	tempWavPath := filepath.Join(tempDirPath, tempWavName)

	// save original file
	if err := os.WriteFile(tempFilePath, fileBinary, 0644); err != nil {
		log.Printf("Error during file writing: %v", err)
		return
	}

	// ffmpeg convert original (.ts or others) to .wav
	convertCmd := exec.Command("cmd", "/C", ffmpegPath, "-y", "-i", tempFilePath, tempWavPath)
	if _, err := convertCmd.CombinedOutput(); err != nil {
		log.Printf("Error executing ffmpeg (convert) (%s): %v", tempFilePath, err)
		return
	}
	log.Printf("ffmpeg converted: %s", tempWavPath)

	// ffmpeg get .wav length
	var tempWavDurationSec float64
	tempNilWavPath := filepath.Join(tempDirPath, "temp.wav")
	durationCmd := exec.Command("cmd", "/C", ffmpegPath, "-i", tempWavPath, tempNilWavPath)
	if durationCmdOutBytes, err := durationCmd.CombinedOutput(); err != nil {
		log.Printf("Error executing ffmpeg (duration) (%s): %v", tempWavPath, err)
		return
	} else {
		durationCmdOut := string(durationCmdOutBytes)
		for _, line := range strings.Split(durationCmdOut, "\n") {
			if strings.Contains(line, "Duration") {
				duration := strings.Fields(line)[1]
				duration = strings.TrimRight(duration, ",")
				duration = strings.TrimSuffix(duration, ".")
				durationSlice := strings.Split(duration, ":")
				hours, _ := strconv.ParseFloat(durationSlice[0], 64)
				minutes, _ := strconv.ParseFloat(durationSlice[1], 64)
				seconds, _ := strconv.ParseFloat(durationSlice[2], 64)
				tempWavDurationSec = float64(hours*3600 + minutes*60 + seconds)
				break
			}
		}
		log.Printf("ffmpeg get duration: %f", tempWavDurationSec)
	}

	// delete original file
	if err := os.Remove(tempFilePath); err != nil {
		log.Printf("Error during file delete: %v", err)
		return
	}

	// Read .wav files in order according to fileIndex and perform voice recognition on them
	mutex.Lock()
	defer mutex.Unlock()
	defer cond.Broadcast()

	if _, isExist := fileIndexMap[videoHash]; !isExist {
		fileIndexMap[videoHash] = -1
	}
	for fileIndexMap[videoHash] != fileIndex-1 {
		cond.Wait()
	}

	log.Printf("handling start, [videoHash:] %s [fileHash:] %s [fileIndex:] %d [rec:] %p", videoHash, fileHash, fileIndex, rec)

	if len(timePrefixSumSecSliceMap[videoHash]) == 0 {
		timePrefixSumSecSliceMap[videoHash] = append(timePrefixSumSecSliceMap[videoHash], 0.0)
	}
	timePrefixSumSecSliceMap[videoHash] = append(timePrefixSumSecSliceMap[videoHash], timePrefixSumSecSliceMap[videoHash][len(timePrefixSumSecSliceMap[videoHash])-1]+tempWavDurationSec)

	file, fileOpenErr := os.Open(tempWavPath)
	if fileOpenErr != nil {
		log.Printf("Open file error (%s): %v", tempWavPath, fileOpenErr)
	}
	defer file.Close()

	buf := make([]byte, 4096)
	for {
		if _, err := file.Read(buf); err != nil {
			if err != io.EOF {
				log.Printf("Read file into buf error %v", err)
				return
			}
			break
		}

		if rec.AcceptWaveform(buf) != 0 {
			result := rec.Result()
			dec := json.NewDecoder(strings.NewReader(result))
			voiceRecognitionResult := VoiceRecognitionResult{}
			if err := dec.Decode(&voiceRecognitionResult); err != io.EOF && err != nil {
				log.Printf("JSON decode result error %v", err)
			}
			log.Printf("result: %v", voiceRecognitionResult)
			if voiceRecognitionResult.Text != "" {
				wg.Add(1)
				handleOneVoiceRecognitionResult(voiceRecognitionResult, videoHash, conn, wg)
			}
		}
	}

	fileIndexMap[videoHash] = fileIndex

	wg.Done()
}

func handleOneVoiceRecognitionResult(voiceRecognitionResult VoiceRecognitionResult, videoHash string, conn *websocket.Conn, wg *sync.WaitGroup) {
	captionResult := createCaptionResult(voiceRecognitionResult, videoHash)

	// TODO: translation
	translatedCaptionResult := captionResult

	respBodyBytes, _ := json.Marshal(translatedCaptionResult)
	nextWriter, nwErr := conn.NextWriter(websocket.TextMessage)
	if nwErr != nil {
		log.Printf("Error during message writer creation: %v", nwErr)
		return
	}
	if _, err := nextWriter.Write(respBodyBytes); err != nil {
		log.Printf("Error during message writing: %v", err)
		return
	}
	if err := nextWriter.Close(); err != nil {
		log.Printf("Error during message sending: %v", err)
	}

	wg.Done()
}

func createCaptionResult(voiceRecognitionResult VoiceRecognitionResult, videoHash string) CaptionResult {
	startTime := voiceRecognitionResult.Result[0].Start
	endTime := voiceRecognitionResult.Result[len(voiceRecognitionResult.Result)-1].End

	startSliceIndex := findTimeLocation(videoHash, startTime)
	endSliceIndex := findTimeLocation(videoHash, endTime)

	startRelativeTime := startTime - timePrefixSumSecSliceMap[videoHash][startSliceIndex]
	endRelativeTime := endTime - timePrefixSumSecSliceMap[videoHash][endSliceIndex]

	captionResult := CaptionResult{}
	captionResult.Start.Index = startSliceIndex
	captionResult.Start.Time = startRelativeTime
	captionResult.End.Index = endSliceIndex
	captionResult.End.Time = endRelativeTime
	captionResult.Text = voiceRecognitionResult.Text

	return captionResult
}

func findTimeLocation(videoHash string, time float64) int {
	left := 0
	right := len(timePrefixSumSecSliceMap[videoHash]) - 1
	for left < right {
		mid := (left + right + 1) / 2
		if timePrefixSumSecSliceMap[videoHash][mid] >= time {
			right = mid - 1
		} else {
			left = mid
		}
	}
	return left
}
