package recognizer

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	vosk "github.com/HomeArchbishop/kaption/third_party/vosk/go"
)

var model *vosk.VoskModel = nil

func InitModel() error {
	executablePath, _ := os.Executable()
	modelPath := filepath.Join(filepath.Dir(executablePath), "model")

	fmt.Print("\n")
	_model, newModelErr := vosk.NewModel(modelPath)
	fmt.Print("\n")
	if newModelErr != nil {
		log.Print(newModelErr)
		return newModelErr
	}
	model = _model
	return nil
}

func CreateNewRecognizer() (*vosk.VoskRecognizer, error) {
	sampleRate := 96000.0
	rec, newRecErr := vosk.NewRecognizer(model, sampleRate)
	if newRecErr != nil {
		log.Print(newRecErr)
		return nil, newRecErr
	}
	rec.SetWords(1)

	return rec, nil
}
