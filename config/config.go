package config

import (
	"encoding/json"
	"github.com/kfur/subtitler/app/shared/recaptcha"
	"github.com/kfur/subtitler/app/shared/server"
	"github.com/kfur/subtitler/app/shared/view"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

// *****************************************************************************
// Application Settings
// *****************************************************************************

// config the settings variable
var Config = &configuration{}

// SpeechToTextV1Options : Service options
type SpeechToTextOptions struct {
	URL            string `json:"URL"`
	IAMApiKey      string `json:"IAMApiKey"`
}

// configuration contains the application settings
type configuration struct {
	Recaptcha recaptcha.Info       `json:"Recaptcha"`
	Server    server.Server        `json:"Server"`
	//Session   session.Session      `json:"Session"`
	Template  view.Template        `json:"Template"`
	View      view.View            `json:"View"`
	STTOptions SpeechToTextOptions  `json:"SpeechToTextOptions"`
}

// ParseJSON unmarshals bytes to structs
func (c *configuration) ParseJSON(b []byte) error {
	return json.Unmarshal(b, &c)
}

// Parser must implement ParseJSON
type Parser interface {
	ParseJSON([]byte) error
}

// Load the JSON config file
func Load(configFile string, p Parser) {
	var err error
	var absPath string
	var input = io.ReadCloser(os.Stdin)
	if absPath, err = filepath.Abs(configFile); err != nil {
		log.Fatalln(err)
	}

	if input, err = os.Open(absPath); err != nil {
		log.Fatalln(err)
	}

	// Read the config file
	jsonBytes, err := ioutil.ReadAll(input)
	input.Close()
	if err != nil {
		log.Fatalln(err)
	}

	// Parse the config
	if err := p.ParseJSON(jsonBytes); err != nil {
		log.Fatalln("Could not parse %q: %v", configFile, err)
	}
}
