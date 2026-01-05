package bedrock

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAWSSigner(t *testing.T) {
	tests := []struct {
		name         string
		region       string
		accessKey    string
		secretKey    string
		sessionToken string
	}{
		{
			name:      "basic credentials",
			region:    "us-east-1",
			accessKey: "AKIAIOSFODNN7EXAMPLE",
			secretKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		},
		{
			name:         "with session token",
			region:       "us-west-2",
			accessKey:    "ASIAIOSFODNN7EXAMPLE",
			secretKey:    "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
			sessionToken: "AQoDYXdzEJr...",
		},
		{
			name:      "different region",
			region:    "eu-west-1",
			accessKey: "AKIAIOSFODNN7EXAMPLE",
			secretKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signer := newAWSSigner(tt.region, tt.accessKey, tt.secretKey, tt.sessionToken)
			assert.NotNil(t, signer)
			assert.Equal(t, tt.region, signer.region)
			assert.Equal(t, tt.accessKey, signer.accessKey)
			assert.Equal(t, tt.secretKey, signer.secretKey)
			assert.Equal(t, tt.sessionToken, signer.sessionToken)
		})
	}
}

func TestAWSSigner_SignRequest(t *testing.T) {
	signer := newAWSSigner(
		"us-east-1",
		"AKIAIOSFODNN7EXAMPLE",
		"wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		"",
	)

	tests := []struct {
		name        string
		method      string
		urlStr      string
		body        string
		expectedErr bool
	}{
		{
			name:        "POST request with body",
			method:      "POST",
			urlStr:      "https://bedrock-runtime.us-east-1.amazonaws.com/model/anthropic.claude-v2/invoke",
			body:        `{"messages":[{"role":"user","content":[{"text":"Hello"}]}]}`,
			expectedErr: false,
		},
		{
			name:        "GET request",
			method:      "GET",
			urlStr:      "https://bedrock-runtime.us-east-1.amazonaws.com/foundation-models",
			body:        "",
			expectedErr: false,
		},
		{
			name:        "POST with empty body",
			method:      "POST",
			urlStr:      "https://bedrock-runtime.us-east-1.amazonaws.com/model/test/invoke",
			body:        "",
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, tt.urlStr, strings.NewReader(tt.body))
			require.NoError(t, err)

			req.Header.Set("Content-Type", "application/json")

			err = signer.signRequest(req, []byte(tt.body), time.Now())

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify required headers are present
				assert.NotEmpty(t, req.Header.Get("Authorization"))
				assert.NotEmpty(t, req.Header.Get("X-Amz-Date"))
				assert.NotEmpty(t, req.Header.Get("X-Amz-Content-Sha256"))

				// Verify Authorization header format
				authHeader := req.Header.Get("Authorization")
				assert.Contains(t, authHeader, "AWS4-HMAC-SHA256")
				assert.Contains(t, authHeader, "Credential=")
				assert.Contains(t, authHeader, "SignedHeaders=")
				assert.Contains(t, authHeader, "Signature=")
			}
		})
	}
}

func TestAWSSigner_SignRequestWithSessionToken(t *testing.T) {
	signer := newAWSSigner(
		"us-east-1",
		"ASIAIOSFODNN7EXAMPLE",
		"wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		"AQoDYXdzEJr...",
	)

	req, err := http.NewRequest("POST", "https://bedrock-runtime.us-east-1.amazonaws.com/model/test/invoke", strings.NewReader("test"))
	require.NoError(t, err)

	err = signer.signRequest(req, []byte("test"), time.Now())
	assert.NoError(t, err)

	// Verify session token header is present
	assert.Equal(t, "AQoDYXdzEJr...", req.Header.Get("X-Amz-Security-Token"))
}

func TestAWSSigner_CreateCanonicalRequest(t *testing.T) {
	signer := newAWSSigner("us-east-1", "key", "secret", "")

	tests := []struct {
		name           string
		method         string
		path           string
		query          string
		headers        map[string]string
		signedHeaders  string
		payloadHash    string
		expectedParts  []string
	}{
		{
			name:   "simple POST request",
			method: "POST",
			path:   "/model/anthropic.claude-v2/invoke",
			query:  "",
			headers: map[string]string{
				"content-type":         "application/json",
				"host":                 "bedrock-runtime.us-east-1.amazonaws.com",
				"x-amz-date":           "20231201T120000Z",
				"x-amz-content-sha256": "abc123",
			},
			signedHeaders: "content-type;host;x-amz-content-sha256;x-amz-date",
			payloadHash:   "abc123",
			expectedParts: []string{
				"POST",
				"/model/anthropic.claude-v2/invoke",
				"",
				"content-type:application/json",
				"host:bedrock-runtime.us-east-1.amazonaws.com",
			},
		},
		{
			name:   "GET request with query parameters",
			method: "GET",
			path:   "/foundation-models",
			query:  "byProvider=Anthropic&maxResults=10",
			headers: map[string]string{
				"host":       "bedrock.us-east-1.amazonaws.com",
				"x-amz-date": "20231201T120000Z",
			},
			signedHeaders: "host;x-amz-date",
			payloadHash:   "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			expectedParts: []string{
				"GET",
				"/foundation-models",
				"byProvider=Anthropic",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, "https://example.com"+tt.path, nil)
			require.NoError(t, err)

			if tt.query != "" {
				req.URL.RawQuery = tt.query
			}

			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			canonical := signer.createCanonicalRequest(req, tt.signedHeaders, tt.payloadHash)

			// Verify canonical request contains expected parts
			for _, part := range tt.expectedParts {
				assert.Contains(t, canonical, part)
			}

			// Verify starts with HTTP method
			assert.True(t, strings.HasPrefix(canonical, tt.method))
		})
	}
}

func TestAWSSigner_CreateStringToSign(t *testing.T) {
	signer := newAWSSigner("us-east-1", "key", "secret", "")

	timestamp := time.Date(2023, 12, 1, 12, 0, 0, 0, time.UTC)
	canonicalRequest := "POST\n/model/test/invoke\n\ncontent-type:application/json\nhost:bedrock-runtime.us-east-1.amazonaws.com\n\ncontent-type;host\nabc123"

	stringToSign := signer.createStringToSign(timestamp, canonicalRequest)

	// Verify string to sign format
	lines := strings.Split(stringToSign, "\n")
	assert.Len(t, lines, 4)
	assert.Equal(t, "AWS4-HMAC-SHA256", lines[0])
	assert.Equal(t, "20231201T120000Z", lines[1])
	assert.Contains(t, lines[2], "20231201/us-east-1/bedrock/aws4_request")
	assert.NotEmpty(t, lines[3]) // hash of canonical request
}

func TestAWSSigner_CalculateSignature(t *testing.T) {
	signer := newAWSSigner("us-east-1", "key", "secret", "")

	timestamp := time.Date(2023, 12, 1, 12, 0, 0, 0, time.UTC)
	stringToSign := "AWS4-HMAC-SHA256\n20231201T120000Z\n20231201/us-east-1/bedrock/aws4_request\nabc123"

	signature := signer.calculateSignature(timestamp, stringToSign)

	// Verify signature is a valid hex string
	assert.Len(t, signature, 64) // SHA256 produces 64 hex characters
	assert.Regexp(t, "^[0-9a-f]+$", signature)
}

func TestAWSSigner_GetSignedHeaders(t *testing.T) {
	signer := newAWSSigner("us-east-1", "key", "secret", "")

	req, err := http.NewRequest("POST", "https://example.com/test", nil)
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Amz-Date", "20231201T120000Z")
	req.Header.Set("Host", "bedrock-runtime.us-east-1.amazonaws.com")
	req.Header.Set("Authorization", "should-be-excluded")

	signedHeaders := signer.getSignedHeaders(req)

	// Verify signed headers are sorted and lowercase
	parts := strings.Split(signedHeaders, ";")
	assert.Greater(t, len(parts), 0)

	// Verify Authorization is not included
	assert.NotContains(t, signedHeaders, "authorization")

	// Verify headers are sorted alphabetically
	for i := 1; i < len(parts); i++ {
		assert.True(t, parts[i-1] < parts[i], "headers should be sorted")
	}
}

func TestHexEncode(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected string
	}{
		{
			name:     "empty input",
			input:    []byte{},
			expected: "",
		},
		{
			name:     "simple bytes",
			input:    []byte{0x01, 0x02, 0x03},
			expected: "010203",
		},
		{
			name:     "alphabetic bytes",
			input:    []byte{0xAB, 0xCD, 0xEF},
			expected: "abcdef",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hexEncode(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHashSHA256(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected string
	}{
		{
			name:     "empty string",
			input:    []byte(""),
			expected: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			name:     "hello world",
			input:    []byte("hello world"),
			expected: "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hashSHA256(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBuildHostHeader(t *testing.T) {
	tests := []struct {
		name     string
		urlStr   string
		expected string
	}{
		{
			name:     "standard HTTPS port",
			urlStr:   "https://bedrock-runtime.us-east-1.amazonaws.com/model/test/invoke",
			expected: "bedrock-runtime.us-east-1.amazonaws.com",
		},
		{
			name:     "custom port",
			urlStr:   "https://bedrock-runtime.us-east-1.amazonaws.com:8443/model/test/invoke",
			expected: "bedrock-runtime.us-east-1.amazonaws.com:8443",
		},
		{
			name:     "HTTP default port",
			urlStr:   "http://localhost/test",
			expected: "localhost",
		},
		{
			name:     "HTTP custom port",
			urlStr:   "http://localhost:8080/test",
			expected: "localhost:8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := url.Parse(tt.urlStr)
			require.NoError(t, err)

			result := buildHostHeader(u)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetCanonicalQueryString(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected string
	}{
		{
			name:     "no query parameters",
			query:    "",
			expected: "",
		},
		{
			name:     "single parameter",
			query:    "foo=bar",
			expected: "foo=bar",
		},
		{
			name:     "multiple parameters sorted",
			query:    "z=3&a=1&m=2",
			expected: "a=1&m=2&z=3",
		},
		{
			name:     "parameters with special characters",
			query:    "name=John Doe&email=john@example.com",
			expected: "email=john%40example.com&name=John+Doe", // Go's url.QueryEscape uses + for spaces
		},
		{
			name:     "parameter with no value",
			query:    "flag&other=value",
			expected: "flag=&other=value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := url.Parse("https://example.com?" + tt.query)
			require.NoError(t, err)

			result := getCanonicalQueryString(u)
			assert.Equal(t, tt.expected, result)
		})
	}
}
