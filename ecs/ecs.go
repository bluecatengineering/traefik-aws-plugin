package ecs

import (
	"encoding/json"
	"github.com/bluecatengineering/traefik-aws-plugin/log"
	"io"
	"net/http"
	"os"
	"time"
)

const onErrorRotationInterval = 10 * time.Second

type Credentials struct {
	AccessKeyId     string    `json:"AccessKeyId"`
	AccessSecretKey string    `json:"SecretAccessKey"`
	SecurityToken   string    `json:"Token"`
	RoleArn         string    `json:"RoleArn"`
	Expiration      time.Time `json:"Expiration"`
}

func GetCredentials() *Credentials {
	creds := &Credentials{}
	go getCredentials(creds)
	return creds
}

// https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-iam-roles.html
func getCredentials(creds *Credentials) {
	client := &http.Client{}
	credsUri := "http://169.254.170.2" + os.Getenv("AWS_CONTAINER_CREDENTIALS_RELATIVE_URI")
	for {
		err := refreshCredentials(creds, client, credsUri)
		if err != nil {
			log.Error(err.Error())
			onErrorTimer := time.NewTimer(onErrorRotationInterval)
			<-onErrorTimer.C
		} else {
			// Renew when half the lifetime is reached
			duration := time.Until(creds.Expiration) / 2
			renewalTimer := time.NewTimer(duration)
			<-renewalTimer.C
		}
	}
}

func refreshCredentials(creds *Credentials, client *http.Client, credsUri string) error {
	req, err := http.NewRequest(http.MethodGet, credsUri, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, creds)
	if err != nil {
		return err
	}
	return nil
}
