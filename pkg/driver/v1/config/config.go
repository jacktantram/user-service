package config

import (
	"fmt"
	"io"
	"os"

	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v3"
)

var (
	FilePath = os.Getenv("CONFIG_FILE")
)

func LoadConfig(config interface{}) error {
	path := FilePath
	if FilePath == "" {
		path = "config.yaml"
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Init new YAML decode
	d := yaml.NewDecoder(file)

	// Start YAML decoding from file
	if err = d.Decode(config); err != nil {
		if err != io.EOF {
			return err
		}
	}

	if err = envconfig.Process("", config); err != nil {
		return fmt.Errorf("cannot process env config: %w", err)
	}
	return nil
}
