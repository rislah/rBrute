package config

import (
	"io"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

func NewConfig(path string) *Config {
	b := read(path)
	config := new(Config)
	err := yaml.Unmarshal(b, config)
	if err != nil {
		log.Fatal(err)
	}
	return config
}

func read(path string) []byte {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	buffer := make([]byte, 4096)
	var read int
	for {
		readCount, err := file.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(err)
		}
		read += readCount
	}
	return buffer[:read]
}
