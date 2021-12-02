package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"

	"qubert/application"
)

func getConfig(fileName string) (*application.Config, error) {
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		cfg := application.DefaultConfig()

		buf := bytes.NewBuffer([]byte{})
		jsonEncoder := json.NewEncoder(buf)
		jsonEncoder.SetIndent("", "    ")
		err = jsonEncoder.Encode(cfg)
		if err != nil {
			return nil, err
		}

		err = ioutil.WriteFile(fileName, buf.Bytes(), 0644)
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
