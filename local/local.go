package local

import (
	"fmt"
	"github.com/bluecatengineering/traefik-aws-plugin/log"
	"net/http"
	"os"
)

type Local struct {
	directory string
}

func New(directory string) *Local {
	return &Local{
		directory: directory,
	}
}

func (local *Local) Put(name string, payload []byte, _ string, _ http.ResponseWriter) ([]byte, error) {
	filePath := fmt.Sprintf("%s/%s", local.directory, name)
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	_, err = file.Write(payload)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	log.Debug(fmt.Sprintf("%q written", filePath))
	return []byte(fmt.Sprintf("%q written", filePath)), nil
}

func (local *Local) Get(name string, _ http.ResponseWriter) ([]byte, error) {
	filePath := fmt.Sprintf("%s/%s", local.directory, name)
	payload, err := os.ReadFile(filePath)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	log.Debug(fmt.Sprintf("%q read", filePath))
	return payload, nil
}
