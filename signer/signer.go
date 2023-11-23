package signer

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/bluecatengineering/traefik-aws-plugin/ecs"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

// https://docs.aws.amazon.com/IAM/latest/UserGuide/create-signed-request.html
type CanonRequest struct {
	// ecs
	Creds   *ecs.Credentials
	Region  string
	Service string

	// V4 data
	httpMethod  string
	date        string
	queryParams map[string]string
	amzHeaders  map[string]string
	canonUri    string
}

// https://docs.aws.amazon.com/IAM/latest/UserGuide/create-signed-request.html#create-canonical-request
func (cr *CanonRequest) RequestString() string {
	queryString := canonString(cr.queryParams, "=", "&", true)
	headers := canonString(cr.amzHeaders, ":", "\n", false)
	signedHeaders := strings.Join(sortedKeys(cr.amzHeaders), ";")
	hashedPayload := cr.amzHeaders["x-amz-content-sha256"]
	return fmt.Sprintf("%s\n%s\n%s\n%s\n\n%s\n%s", cr.httpMethod, cr.canonUri, queryString, headers, signedHeaders, hashedPayload)
}

// https://docs.aws.amazon.com/IAM/latest/UserGuide/create-signed-request.html#create-string-to-sign
func (cr *CanonRequest) StringToSignV4() string {
	algorithm := "AWS4-HMAC-SHA256"

	requestDateTime := cr.date
	if amzDate, ok := cr.amzHeaders["x-amz-date"]; ok {
		requestDateTime = amzDate
	}

	credentialScope := requestDateTime[:8] + "/" + cr.Region + "/" + cr.Service + "/aws4_request"

	sha := sha256.New()
	sha.Write([]byte(cr.RequestString()))
	canonRequestSha := sha.Sum(nil)
	hashedCanonRequest := hex.EncodeToString(canonRequestSha)

	return fmt.Sprintf("%s\n%s\n%s\n%s", algorithm, requestDateTime, credentialScope, hashedCanonRequest)
}

// https://docs.aws.amazon.com/IAM/latest/UserGuide/create-signed-request.html#calculate-signature
func (cr *CanonRequest) SignatureV4() string {
	date := cr.date
	if amzDate, ok := cr.amzHeaders["x-amz-date"]; ok {
		date = amzDate
	}

	dateKey := hmac.New(sha256.New, []byte("AWS4"+cr.Creds.AccessSecretKey))
	dateKey.Write([]byte(date[:8]))

	dateRegionKey := hmac.New(sha256.New, dateKey.Sum(nil))
	dateRegionKey.Write([]byte(cr.Region))

	dateRegionServiceKey := hmac.New(sha256.New, dateRegionKey.Sum(nil))
	dateRegionServiceKey.Write([]byte(cr.Service))

	signingKey := hmac.New(sha256.New, dateRegionServiceKey.Sum(nil))
	signingKey.Write([]byte("aws4_request"))

	signatureV4 := hmac.New(sha256.New, signingKey.Sum(nil))
	signatureV4.Write([]byte(cr.StringToSignV4()))

	return hex.EncodeToString(signatureV4.Sum(nil))
}

// https://docs.aws.amazon.com/IAM/latest/UserGuide/create-signed-request.html#add-signature-to-request
func (cr *CanonRequest) AuthHeader() string {
	date := cr.date
	if amzDate, ok := cr.amzHeaders["x-amz-date"]; ok {
		date = amzDate
	}
	return "AWS4-HMAC-SHA256 " +
		"Credential=" + cr.Creds.AccessKeyId + "/" + date[:8] + "/" + cr.Region + "/" + cr.Service + "/aws4_request" +
		",SignedHeaders=" + strings.Join(sortedKeys(cr.amzHeaders), ";") +
		",Signature=" + cr.SignatureV4()
}

func CreateCanonRequest(req *http.Request, payload []byte, crTemplate CanonRequest) *CanonRequest {
	now := time.Now()
	formatted := strings.ReplaceAll(
		strings.ReplaceAll(now.UTC().Format(time.RFC3339), "-", ""),
		":", "")
	amzDate := formatted[:len(formatted)-3] + "00Z"
	if date := req.Header.Get("date"); date == "" {
		req.Header.Set("date", now.Local().Format(time.RFC1123))
	}
	crPayload := payload
	if crPayload == nil {
		crPayload = []byte("")
	}
	sha := sha256.New()
	sha.Write(crPayload)
	req.Header.Set("X-Amz-Content-Sha256", hex.EncodeToString(sha.Sum(nil)))
	req.Header.Set("X-Amz-Date", amzDate)
	if crTemplate.Creds.SecurityToken != "" {
		req.Header.Set("x-amz-security-token", crTemplate.Creds.SecurityToken)
	}
	return updateCanonRequest(req, &crTemplate)
}

func updateCanonRequest(req *http.Request, cr *CanonRequest) *CanonRequest {
	m := map[string][]string(req.Header)
	headers := make(map[string]string, len(m))
	for k, vs := range m {
		headers[strings.ToLower(k)] = strings.TrimSpace(strings.Join(vs, ","))
	}
	cr.httpMethod = req.Method
	cr.date = headers["date"]
	cr.amzHeaders = headers
	if req.URL.Path != "" {
		cr.canonUri = strings.TrimSpace(req.URL.Path)
	}
	return cr
}

func canonString(in map[string]string, sep string, inter string, encoding bool) string {
	var c string
	keys := make([]string, 0, len(in))
	for k := range in {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if c != "" {
			c = c + inter
		}
		if encoding {
			c = c + fmt.Sprintf("%s%s%s", url.QueryEscape(k), sep, url.QueryEscape(in[k]))
		} else {
			c = c + fmt.Sprintf("%s%s%s", k, sep, in[k])
		}
	}
	return c
}

func sortedKeys(in map[string]string) []string {
	keys := make([]string, 0, len(in))
	for k := range in {
		keys = append(keys, strings.ToLower(k))
	}
	sort.Strings(keys)
	return keys
}
