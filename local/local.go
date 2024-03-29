package local

import (
	"fmt"
	"net/http"
	"os"

	"github.com/bluecatengineering/traefik-aws-plugin/log"
	"github.com/google/uuid"
)

type Local struct {
	directory string
}

func New(directory string) *Local {
	return &Local{
		directory: directory,
	}
}

func (local *Local) Put(name string, payload []byte, _ string, rw http.ResponseWriter) ([]byte, error) {
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
	rw.Header().Add("Location", name)
	return []byte(fmt.Sprintf("%q written", filePath)), nil
}

func (local *Local) Post(path string, payload []byte, contentType string, rw http.ResponseWriter) ([]byte, error) {
	return local.Put(path+"/"+uuid.NewString(), payload, contentType, rw)
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
