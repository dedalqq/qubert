package config

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"path"

	"qubert/application"
)

func Save(fileName string, config *application.Config) error {
	buf := bytes.NewBuffer([]byte{})
	jsonEncoder := json.NewEncoder(buf)
	jsonEncoder.SetIndent("", "    ")
	err := jsonEncoder.Encode(config)
	if err != nil {
		return err
	}

	err = os.MkdirAll(path.Dir(fileName), 0744)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(fileName, buf.Bytes(), 0644)
	if err != nil {
		return err
	}

	return nil
}

func Load(fileName string) (*application.Config, error) {
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		cfg := application.DefaultConfig()

		err = Save(fileName, cfg)
		if err != nil {
			return nil, err
		}

		return cfg, nil
	}

	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}

	defer func() {
		err = f.Close()
		if err != nil {
			panic(err)
		}
	}()

	cfg := &application.Config{}

	err = json.NewDecoder(f).Decode(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
