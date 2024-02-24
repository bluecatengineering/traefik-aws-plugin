package traefik_aws_plugin

import (
	"context"
	"fmt"
	"github.com/bluecatengineering/traefik-aws-plugin/ecs"
	"github.com/bluecatengineering/traefik-aws-plugin/local"
	"github.com/bluecatengineering/traefik-aws-plugin/log"
	"github.com/bluecatengineering/traefik-aws-plugin/s3"
	"io"
	"net/http"
)

type Service interface {
	Put(name string, payload []byte, contentType string, rw http.ResponseWriter) ([]byte, error)
	Post(name string, payload []byte, contentType string, rw http.ResponseWriter) ([]byte, error)
	Get(name string, rw http.ResponseWriter) ([]byte, error)
}

type Config struct {
	TimeoutSeconds int
	Service        string

	// S3
	Bucket string
	Prefix string
	Region string

	// Local Directory
	Directory string
}

func CreateConfig() *Config {
	return &Config{TimeoutSeconds: 5}
}

type AwsPlugin struct {
	next    http.Handler
	name    string
	service Service
}

func (plugin AwsPlugin) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPut:
		plugin.put(rw, req)
	case http.MethodPost:
		plugin.post(rw, req)
	case http.MethodGet:
		plugin.get(rw, req)
	default:
		http.Error(rw, fmt.Sprintf("Method %s not implemented", req.Method), http.StatusNotImplemented)
	}
	plugin.next.ServeHTTP(rw, req)
}

func (plugin *AwsPlugin) put(rw http.ResponseWriter, req *http.Request) {
	payload, err := io.ReadAll(req.Body)
	if err != nil {
		rw.WriteHeader(http.StatusNotAcceptable)
		log.Error(fmt.Sprintf("Reading body failed: %s", err.Error()))
		return
	}
	resp, err := plugin.service.Put(req.URL.Path[1:], payload, req.Header.Get("Content-Type"), rw)
	handleResponse(resp, err, rw)
}

func (plugin *AwsPlugin) post(rw http.ResponseWriter, req *http.Request) {
	payload, err := io.ReadAll(req.Body)
	if err != nil {
		rw.WriteHeader(http.StatusNotAcceptable)
		log.Error(fmt.Sprintf("Reading body failed: %s", err.Error()))
		return
	}
	resp, err := plugin.service.Post(req.URL.Path[1:], payload, req.Header.Get("Content-Type"), rw)
	handleResponse(resp, err, rw)
}

func (plugin *AwsPlugin) get(rw http.ResponseWriter, req *http.Request) {
	resp, err := plugin.service.Get(req.URL.Path[1:], rw)
	handleResponse(resp, err, rw)
}

func handleResponse(resp []byte, reqErr error, rw http.ResponseWriter) {
	if reqErr != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		http.Error(rw, fmt.Sprintf("Put error: %s", reqErr.Error()), http.StatusInternalServerError)
		log.Error(reqErr.Error())
		return
	}
	rw.WriteHeader(http.StatusOK)
	_, err := rw.Write(resp)
	if err != nil {
		http.Error(rw, string(resp)+err.Error(), http.StatusBadGateway)
	}
}

func New(_ context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	plugin := &AwsPlugin{next: next, name: name}
	switch config.Service {
	case "s3":
		plugin.service = s3.New(config.Bucket, config.Prefix, config.Region, config.TimeoutSeconds, ecs.GetCredentials())
		return plugin, nil
	case "local":
		plugin.service = local.New(config.Directory)
		return plugin, nil
	default:
		log.Error(fmt.Sprintf("unknown service: %s", config.Service))
	}
	return next, fmt.Errorf("invalid config: %v", config)
}
