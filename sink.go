package aws_sink

import (
	"context"
	"fmt"
	"github.com/yoeluk/aws-sink/aws"
	"github.com/yoeluk/aws-sink/local"
	"github.com/yoeluk/aws-sink/log"
	"github.com/yoeluk/aws-sink/s3"
	"io"
	"net/http"
)

type Sink interface {
	Put(name string, payload []byte, contentType string, rw http.ResponseWriter) ([]byte, error)
}

type Config struct {
	Timeout  int
	SinkType string
	// s3 sink
	Bucket string
	Prefix string // including the leading slash
	Region string
	// local sink
	LocalDirectory string
}

func CreateConfig() *Config {
	return &Config{Timeout: 5}
}

type AwsSink struct {
	next http.Handler
	name string
	sink Sink
}

var demoCreds = aws.Credentials{
	AccessKeyId:     "[YOUR AWS KEY HERE]",
	AccessSecretKey: "[YOUR AWS SECRET HERE]",
	SecurityToken:   "[YOUR AWS SECRET SECRET TOKEN]",
}

func (s AwsSink) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusNotAcceptable)
		log.Error(fmt.Sprintf("found an error reading the body %s", err.Error()))
		return
	}
	resp, err := s.sink.Put(r.URL.Path[1:], payload, r.Header.Get("Content-Type"), w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, fmt.Sprintf("encountered an error putting the object, error: %q", err.Error()), http.StatusInternalServerError)
		log.Error(err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resp)
	if err != nil {
		http.Error(w, string(resp)+err.Error(), http.StatusBadGateway)
	}
	s.next.ServeHTTP(w, r)
}

func New(_ context.Context, next http.Handler, cfg *Config, name string) (http.Handler, error) {
	awsSink := &AwsSink{next: next, name: name}
	switch cfg.SinkType {
	case "s3":
		awsSink.sink = s3.New(cfg.Bucket, cfg.Prefix, cfg.Region, cfg.Timeout, aws.EcsCredentials())
		return awsSink, nil
	case "local":
		awsSink.sink = local.New("my-local-region", cfg.LocalDirectory, &demoCreds)
		return awsSink, nil
	default:
		log.Error(fmt.Sprintf("unknown sinkType: %s", cfg.SinkType))
	}
	return next, fmt.Errorf("couldn't start the sink plugin with config: %v", cfg)
}