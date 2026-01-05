package bedrock

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

const (
	// AWS Signature V4 constants
	algorithm       = "AWS4-HMAC-SHA256"
	service         = "bedrock"
	awsRequestType  = "aws4_request"
	timeFormat      = "20060102T150405Z"
	shortTimeFormat = "20060102"
)

// awsSigner handles AWS Signature Version 4 signing
type awsSigner struct {
	region       string
	accessKey    string
	secretKey    string
	sessionToken string
}

// newAWSSigner creates a new AWS Signature V4 signer
func newAWSSigner(region, accessKey, secretKey, sessionToken string) *awsSigner {
	return &awsSigner{
		region:       region,
		accessKey:    accessKey,
		secretKey:    secretKey,
		sessionToken: sessionToken,
	}
}

// signRequest signs an HTTP request using AWS Signature Version 4
func (s *awsSigner) signRequest(req *http.Request, body []byte, timestamp time.Time) error {
	// Set required headers
	req.Header.Set("X-Amz-Date", timestamp.Format(timeFormat))

	// Add session token if present
	if s.sessionToken != "" {
		req.Header.Set("X-Amz-Security-Token", s.sessionToken)
	}

	// Calculate payload hash
	payloadHash := hashSHA256(body)
	req.Header.Set("X-Amz-Content-Sha256", payloadHash)

	// Set host header if not already set
	if req.Header.Get("Host") == "" {
		req.Header.Set("Host", buildHostHeader(req.URL))
	}

	// Create canonical request
	signedHeaders := s.getSignedHeaders(req)
	canonicalRequest := s.createCanonicalRequest(req, signedHeaders, payloadHash)

	// Create string to sign
	stringToSign := s.createStringToSign(timestamp, canonicalRequest)

	// Calculate signature
	signature := s.calculateSignature(timestamp, stringToSign)

	// Create authorization header
	authHeader := s.createAuthorizationHeader(timestamp, signedHeaders, signature)
	req.Header.Set("Authorization", authHeader)

	return nil
}

// createCanonicalRequest creates the canonical request string
func (s *awsSigner) createCanonicalRequest(req *http.Request, signedHeaders, payloadHash string) string {
	// Canonical URI
	canonicalURI := req.URL.EscapedPath()
	if canonicalURI == "" {
		canonicalURI = "/"
	}

	// Canonical query string
	canonicalQueryString := getCanonicalQueryString(req.URL)

	// Canonical headers
	canonicalHeaders := s.getCanonicalHeaders(req, signedHeaders)

	// Build canonical request
	return fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s",
		req.Method,
		canonicalURI,
		canonicalQueryString,
		canonicalHeaders,
		signedHeaders,
		payloadHash,
	)
}

// getCanonicalHeaders returns the canonical headers string
func (s *awsSigner) getCanonicalHeaders(req *http.Request, signedHeaders string) string {
	headers := strings.Split(signedHeaders, ";")
	var canonicalHeaders []string

	for _, header := range headers {
		value := strings.TrimSpace(req.Header.Get(header))
		canonicalHeaders = append(canonicalHeaders, fmt.Sprintf("%s:%s", header, value))
	}

	return strings.Join(canonicalHeaders, "\n") + "\n"
}

// getSignedHeaders returns the list of signed headers
func (s *awsSigner) getSignedHeaders(req *http.Request) string {
	var headers []string

	for name := range req.Header {
		lowerName := strings.ToLower(name)
		// Skip authorization header
		if lowerName == "authorization" {
			continue
		}
		headers = append(headers, lowerName)
	}

	// Sort headers alphabetically
	sort.Strings(headers)

	return strings.Join(headers, ";")
}

// createStringToSign creates the string to sign
func (s *awsSigner) createStringToSign(timestamp time.Time, canonicalRequest string) string {
	credentialScope := s.getCredentialScope(timestamp)
	hashedCanonicalRequest := hashSHA256([]byte(canonicalRequest))

	return fmt.Sprintf("%s\n%s\n%s\n%s",
		algorithm,
		timestamp.Format(timeFormat),
		credentialScope,
		hashedCanonicalRequest,
	)
}

// calculateSignature calculates the request signature
func (s *awsSigner) calculateSignature(timestamp time.Time, stringToSign string) string {
	// Derive signing key
	date := timestamp.Format(shortTimeFormat)
	kDate := hmacSHA256([]byte("AWS4"+s.secretKey), []byte(date))
	kRegion := hmacSHA256(kDate, []byte(s.region))
	kService := hmacSHA256(kRegion, []byte(service))
	kSigning := hmacSHA256(kService, []byte(awsRequestType))

	// Calculate signature
	signature := hmacSHA256(kSigning, []byte(stringToSign))
	return hexEncode(signature)
}

// createAuthorizationHeader creates the Authorization header value
func (s *awsSigner) createAuthorizationHeader(timestamp time.Time, signedHeaders, signature string) string {
	credentialScope := s.getCredentialScope(timestamp)
	credential := fmt.Sprintf("%s/%s", s.accessKey, credentialScope)

	return fmt.Sprintf("%s Credential=%s, SignedHeaders=%s, Signature=%s",
		algorithm,
		credential,
		signedHeaders,
		signature,
	)
}

// getCredentialScope returns the credential scope string
func (s *awsSigner) getCredentialScope(timestamp time.Time) string {
	return fmt.Sprintf("%s/%s/%s/%s",
		timestamp.Format(shortTimeFormat),
		s.region,
		service,
		awsRequestType,
	)
}

// hmacSHA256 computes HMAC-SHA256
func hmacSHA256(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}

// hashSHA256 computes SHA256 hash and returns hex-encoded string
func hashSHA256(data []byte) string {
	hash := sha256.Sum256(data)
	return hexEncode(hash[:])
}

// hexEncode converts bytes to lowercase hex string
func hexEncode(data []byte) string {
	return fmt.Sprintf("%x", data)
}

// buildHostHeader builds the host header value from URL
func buildHostHeader(u *url.URL) string {
	host := u.Hostname()
	port := u.Port()

	// Omit port if it's the default for the scheme
	if port != "" {
		if (u.Scheme == "https" && port == "443") || (u.Scheme == "http" && port == "80") {
			return host
		}
		return fmt.Sprintf("%s:%s", host, port)
	}

	return host
}

// getCanonicalQueryString returns the canonical query string
func getCanonicalQueryString(u *url.URL) string {
	if u.RawQuery == "" {
		return ""
	}

	// Parse query parameters
	values := u.Query()

	// Sort keys
	var keys []string
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// Build canonical query string
	var parts []string
	for _, key := range keys {
		for _, value := range values[key] {
			// URL encode key and value
			encodedKey := url.QueryEscape(key)
			encodedValue := url.QueryEscape(value)
			parts = append(parts, fmt.Sprintf("%s=%s", encodedKey, encodedValue))
		}
	}

	return strings.Join(parts, "&")
}
