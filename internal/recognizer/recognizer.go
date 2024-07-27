package recognizer

import (
	"fmt"
	"log"

	vosk "github.com/HomeArchbishop/kaption/third_party/vosk/go"
)

var model *vosk.VoskModel = nil
var sampleRate float64 = 96000.0

func SetSampleRate(rate float64) {
	sampleRate = rate
}

func InitModel() error {
	fmt.Print("\n")
	_model, newModelErr := vosk.NewModel("model")
	fmt.Print("\n")
	if newModelErr != nil {
		log.Print(newModelErr)
		return newModelErr
	}
	model = _model
	return nil
}

func CreateNewRecognizer() (*vosk.VoskRecognizer, error) {
	rec, newRecErr := vosk.NewRecognizer(model, sampleRate)
	if newRecErr != nil {
		log.Print(newRecErr)
		return nil, newRecErr
	}
	rec.SetWords(1)

	return rec, nil
}
