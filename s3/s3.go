package s3

import (
	"bytes"
	"context"
	"fmt"
	"github.com/bluecatengineering/traefik-aws-plugin/ecs"
	"github.com/bluecatengineering/traefik-aws-plugin/log"
	"github.com/bluecatengineering/traefik-aws-plugin/signer"
	"github.com/google/uuid"
	"io"
	"net/http"
	"time"
)

type S3 struct {
	client         *http.Client
	crTemplate     *signer.CanonRequest
	bucketUri      string
	prefix         string
	timeoutSeconds int
}

func New(bucket, prefix, region string, timeoutSeconds int, creds *ecs.Credentials) *S3 {
	crTemplate := &signer.CanonRequest{
		Creds:   creds,
		Region:  region,
		Service: "s3",
	}
	return &S3{
		client:         &http.Client{},
		crTemplate:     crTemplate,
		bucketUri:      fmt.Sprintf("https://%s.s3.amazonaws.com", bucket),
		prefix:         prefix,
		timeoutSeconds: timeoutSeconds,
	}
}

func (s3 *S3) request(httpMethod string, name string, payload []byte, contentType string, rw http.ResponseWriter) ([]byte, error) {
	uri := s3.bucketUri + s3.prefix + "/" + name
	var payloadReader io.Reader = nil
	if payload != nil {
		payloadReader = bytes.NewReader(payload)
	}
	req, err := http.NewRequest(httpMethod, uri, payloadReader)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s3.timeoutSeconds)*time.Second)
	if cancel != nil {
		defer cancel()
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	req.Header.Set("Host", req.URL.Host)
	cr := signer.CreateCanonRequest(req, payload, *s3.crTemplate)
	req.Header.Set("Authorization", cr.AuthHeader())
	resp, err := s3.client.Do(req.WithContext(ctx))
	if err != nil {
		log.Error(fmt.Sprintf("%s %q failed, status: %q, error: %s", httpMethod, uri, resp.Status, err.Error()))
		return nil, err
	}
	if resp.StatusCode > 299 {
		return nil, fmt.Errorf(cr.RequestString())
	}
	response, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error(fmt.Sprintf("Reading S3 response body failed: %q", err.Error()))
	}
	copyHeader(rw.Header(), resp.Header)
	return response, nil
}

func (s3 *S3) Put(name string, payload []byte, contentType string, rw http.ResponseWriter) ([]byte, error) {
	return s3.request(http.MethodPut, name, payload, contentType, rw)
}

func (s3 *S3) Post(path string, payload []byte, contentType string, rw http.ResponseWriter) ([]byte, error) {
	return s3.Put(path+"/"+uuid.NewString(), payload, contentType, rw)
}


func (s3 *S3) Get(name string, rw http.ResponseWriter) ([]byte, error) {
	return s3.request(http.MethodGet, name, nil, "", rw)
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}
