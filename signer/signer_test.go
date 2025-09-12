package signer

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/bluecatengineering/traefik-aws-plugin/ecs"
)

func TestCanonRequestV4(t *testing.T) {
	testCases := []struct {
		cr                 *CanonRequest
		expectedSig        string
		expectedAuthHeader string
		name               string
	}{
		{
			name:               "get versioned object request",
			expectedSig:        "507d483a5b2823c61593ed3b1db476ba3568e4b82c1c7005441ba406e83699cd",
			expectedAuthHeader: "AWS4-HMAC-SHA256 Credential=KEY/20130524/us-east-1/s3/aws4_request,SignedHeaders=host;range;x-amz-content-sha256;x-amz-date,Signature=507d483a5b2823c61593ed3b1db476ba3568e4b82c1c7005441ba406e83699cd",
			cr: &CanonRequest{
				Creds: &ecs.Credentials{
					AccessSecretKey: "SECRET",
					AccessKeyId:     "KEY",
				},
				httpMethod: "GET",
				date:       "20130524T000000Z",
				Region:     "us-east-1",
				Service:    "s3",
				canonUri:   "/test.txt",
				amzHeaders: map[string]string{
					"host":                 "examplebucket.s3.amazonaws.com",
					"range":                "bytes=0-9",
					"x-amz-content-sha256": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
					"x-amz-date":           "20130524T000000Z",
				},
			},
		},
		{
			name:               "get object request",
			expectedSig:        "5e00930a4878798235e8c6527ca3cfd780b87b472de204b22a07fbc10841e751",
			expectedAuthHeader: "AWS4-HMAC-SHA256 Credential=KEY/20130524/us-east-1/s3/aws4_request,SignedHeaders=host;range;x-amz-content-sha256;x-amz-date,Signature=5e00930a4878798235e8c6527ca3cfd780b87b472de204b22a07fbc10841e751",
			cr: &CanonRequest{
				Creds: &ecs.Credentials{
					AccessSecretKey: "SECRET",
					AccessKeyId:     "KEY",
				},
				httpMethod: "GET",
				date:       "20130524T000000Z",
				Region:     "us-east-1",
				Service:    "s3",
				canonUri:   "/test.txt",
				amzHeaders: map[string]string{
					"host":                 "examplebucket.s3.amazonaws.com",
					"range":                "bytes=0-9",
					"x-amz-content-sha256": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
					"x-amz-date":           "20130524T000000Z",
				},
			},
		},
		{
			name:               "put object request",
			expectedSig:        "7de0355f21977e2d7defda79bd7e0b671008cbe91c5cb1bdde815295d54e17fb",
			expectedAuthHeader: "AWS4-HMAC-SHA256 Credential=KEY/20130524/us-east-1/s3/aws4_request,SignedHeaders=date;host;x-amz-content-sha256;x-amz-date;x-amz-storage-class,Signature=7de0355f21977e2d7defda79bd7e0b671008cbe91c5cb1bdde815295d54e17fb",
			cr: &CanonRequest{
				Creds: &ecs.Credentials{
					AccessSecretKey: "SECRET",
					AccessKeyId:     "KEY",
				},
				httpMethod: "PUT",
				date:       "Fri, 24 May 2013 00:00:00 GMT",
				Region:     "us-east-1",
				Service:    "s3",
				canonUri:   "/test%24file.text",
				amzHeaders: map[string]string{
					"host":                 "examplebucket.s3.amazonaws.com",
					"date":                 "Fri, 24 May 2013 00:00:00 GMT",
					"x-amz-content-sha256": "44ce7dd67c959e0d3524ffac1771dfbba87d2b6b4b4e99e42034a8b803f8b072",
					"x-amz-date":           "20130524T000000Z",
					"x-amz-storage-class":  "REDUCED_REDUNDANCY",
				},
			},
		},
	}

	values := make(url.Values)
	values.Add("versionId", "abcdef")
	testCases[0].cr.queryParams = values

	for _, tt := range testCases {
		signature := tt.cr.SignatureV4()
		authHeader := tt.cr.AuthHeader()
		if signature != tt.expectedSig {
			t.Errorf("found an error while testing %s: the signature didn't match the expected string\nexpected: %s\nfound: %s\nStringToSign:\n%s\nRequestString: \n%s\n", tt.name, tt.expectedSig, signature, tt.cr.StringToSignV4(), tt.cr.RequestString())
		}
		if authHeader != tt.expectedAuthHeader {
			t.Errorf("found unexpected authorization header; expected:\n%s,\nfound:\n%s", tt.expectedAuthHeader, authHeader)
		}
	}
}

func TestSignerV4(t *testing.T) {
	crTemplate := &CanonRequest{
		Creds: &ecs.Credentials{
			AccessSecretKey: "SECRET",
			AccessKeyId:     "KEY",
		},
		Region:  "us-east-1",
		Service: "s3",
	}

	r1, _ := http.NewRequest(http.MethodPut, "https://examplebucket.s3.amazonaws.com/text/test/file.text", bytes.NewReader([]byte("")))
	r2, _ := http.NewRequest(http.MethodGet, "https://examplebucket.s3.amazonaws.com/text/test/file.text?versionId=abcdef", nil)

	testCases := []struct {
		name     string
		expected string
		request  *http.Request
	}{
		{
			name:     "basic",
			expected: "hello",
			request:  r1,
		},
		{
			name:     "basic",
			expected: "hello",
			request:  r2,
		},
	}

	for _, tt := range testCases {
		tt.request.Header.Set("Content-Type", "application/json")
		tt.request.Header.Set("Host", tt.request.URL.Host)
		cr := CreateCanonRequest(tt.request, make([]byte, 0), *crTemplate)
		fmt.Printf("%v\n", cr.AuthHeader())
		fmt.Printf("%v\n", cr.StringToSignV4())
		fmt.Printf("%v\n", cr.RequestString())
	}
}
