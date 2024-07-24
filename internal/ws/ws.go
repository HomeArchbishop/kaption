package ws

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/HomeArchbishop/kaption/internal/handler"
	"github.com/HomeArchbishop/kaption/internal/recognizer"
	vosk "github.com/HomeArchbishop/kaption/third_party/vosk/go"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

var recMap = map[string]*vosk.VoskRecognizer{}

func generateRandomHash() (string, error) {
	bytes := make([]byte, 16)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func removeTempDir(tempDirPath string) {
	if err := os.RemoveAll(tempDirPath); err != nil {
		log.Printf("Error removing directory %s: %v", tempDirPath, err)
	}
}

func SocketHandler(w http.ResponseWriter, r *http.Request) {
	var wg sync.WaitGroup

	executablePath, _ := os.Executable()
	tempPath := filepath.Join(filepath.Dir(executablePath), "temp")
	if _, err := os.Stat(tempPath); os.IsNotExist(err) {
		if err := os.Mkdir(tempPath, 0755); err != nil {
			log.Printf("Error creating directory %s: %v", tempPath, err)
			return
		}
	}

	ffmpegPath := filepath.Join(filepath.Dir(executablePath), "ffmpeg.exe")

	hash, _ := generateRandomHash()
	tempDirPath := filepath.Join(tempPath, hash)
	if err := os.Mkdir(tempDirPath, 0755); err != nil {
		log.Printf("Error creating directory %s: %v", tempDirPath, err)
		return
	}
	defer removeTempDir(tempDirPath)

	conn, upgradeErr := upgrader.Upgrade(w, r, nil)
	if upgradeErr != nil {
		log.Print("Error during connection upgradation:", upgradeErr)
		return
	}
	defer conn.Close()

	var videoHash string
	for {
		messageType, message, readErr := conn.ReadMessage()
		if readErr != nil {
			if !websocket.IsCloseError(readErr, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				log.Println("Error during message reading:", readErr)
			}
			break
		}

		if messageType == 1 /* string */ {
			log.Printf("Received: %d %s", messageType, message)
		} else /* bytes */ {
			log.Printf("Received: %d %d %s %s %s", messageType, len(message), string(message[:16]), string(message[16:32]), string(message[32:48]))
			videoHash = string(message[:16])
			_, existsRec := recMap[videoHash]
			if !existsRec {
				rec, _ := recognizer.CreateNewRecognizer()
				recMap[videoHash] = rec
				defer rec.Free()
			}
			wg.Add(1)
			go handler.BinaryMessageHandler(message, tempDirPath, ffmpegPath, recMap[videoHash], conn, &wg)
		}
	}

	wg.Wait()
}
