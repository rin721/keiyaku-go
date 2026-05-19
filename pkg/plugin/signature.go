package plugin

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	HeaderPluginKey = "X-Keiyaku-Plugin-Key"
	HeaderTimestamp = "X-Keiyaku-Timestamp"
	HeaderNonce     = "X-Keiyaku-Nonce"
	HeaderSignature = "X-Keiyaku-Signature"

	DefaultSignatureSkew = 5 * time.Minute
)

type SignatureParts struct {
	PluginKey string
	Timestamp string
	Nonce     string
	Signature string
}

type NonceChecker interface {
	UseNonce(pluginKey string, nonce string, expiresAt time.Time, now time.Time) error
}

func SignRequest(req *http.Request, pluginKey string, secret string, body []byte, now time.Time, nonce string) error {
	if req == nil {
		return runtimeError("sign request", "request is nil", nil)
	}
	if !ValidPluginKey(pluginKey) {
		return runtimeError("sign request", "invalid plugin key", ErrInvalidSignature)
	}
	if strings.TrimSpace(secret) == "" {
		return runtimeError("sign request", "secret is required", ErrInvalidSignature)
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	if nonce == "" {
		generated, err := NewNonce()
		if err != nil {
			return runtimeError("sign request", "generate nonce", err)
		}
		nonce = generated
	}
	timestamp := now.UTC().Format(time.RFC3339Nano)
	signature := Sign(req.Method, req.URL.EscapedPath(), timestamp, nonce, BodySHA256(body), secret)
	req.Header.Set(HeaderPluginKey, pluginKey)
	req.Header.Set(HeaderTimestamp, timestamp)
	req.Header.Set(HeaderNonce, nonce)
	req.Header.Set(HeaderSignature, signature)
	return nil
}

func VerifySignedRequest(req *http.Request, secret string, maxBodyBytes int64, now time.Time, skew time.Duration) ([]byte, SignatureParts, error) {
	if req == nil {
		return nil, SignatureParts{}, runtimeError("verify signed request", "request is nil", nil)
	}
	body, err := ReadLimitedBody(req.Body, maxBodyBytes)
	if err != nil {
		return nil, SignatureParts{}, err
	}
	req.Body = io.NopCloser(bytes.NewReader(body))
	parts := SignatureFromHeader(req.Header)
	if err := Verify(req.Method, req.URL.EscapedPath(), body, parts, secret, now, skew); err != nil {
		return nil, parts, err
	}
	return body, parts, nil
}

func ReadLimitedBody(reader io.Reader, limit int64) ([]byte, error) {
	if reader == nil {
		return nil, nil
	}
	if limit <= 0 {
		content, err := io.ReadAll(reader)
		if err != nil {
			return nil, runtimeError("read request body", "read body", err)
		}
		return content, nil
	}
	content, err := io.ReadAll(io.LimitReader(reader, limit+1))
	if err != nil {
		return nil, runtimeError("read request body", "read body", err)
	}
	if int64(len(content)) > limit {
		return nil, &Error{Kind: ErrorKindValidation, Op: "read request body", Msg: "request body exceeds limit", Err: ErrBodyTooLarge}
	}
	return content, nil
}

func SignatureFromHeader(header http.Header) SignatureParts {
	return SignatureParts{
		PluginKey: strings.TrimSpace(header.Get(HeaderPluginKey)),
		Timestamp: strings.TrimSpace(header.Get(HeaderTimestamp)),
		Nonce:     strings.TrimSpace(header.Get(HeaderNonce)),
		Signature: strings.TrimSpace(header.Get(HeaderSignature)),
	}
}

func Verify(method string, path string, body []byte, parts SignatureParts, secret string, now time.Time, skew time.Duration) error {
	if !ValidPluginKey(parts.PluginKey) {
		return signatureError("plugin key is invalid")
	}
	if strings.TrimSpace(secret) == "" {
		return signatureError("secret is required")
	}
	if strings.TrimSpace(parts.Timestamp) == "" || strings.TrimSpace(parts.Nonce) == "" || strings.TrimSpace(parts.Signature) == "" {
		return signatureError("signature headers are required")
	}
	if skew <= 0 {
		skew = DefaultSignatureSkew
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	timestamp, err := ParseSignatureTimestamp(parts.Timestamp)
	if err != nil {
		return err
	}
	if timestamp.Before(now.Add(-skew)) || timestamp.After(now.Add(skew)) {
		return signatureError("timestamp is outside allowed skew")
	}
	expected := Sign(method, path, parts.Timestamp, parts.Nonce, BodySHA256(body), secret)
	if subtle.ConstantTimeCompare([]byte(expected), []byte(parts.Signature)) != 1 {
		return signatureError("signature mismatch")
	}
	return nil
}

func Sign(method string, path string, timestamp string, nonce string, bodyHash string, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = io.WriteString(mac, CanonicalString(method, path, timestamp, nonce, bodyHash))
	return hex.EncodeToString(mac.Sum(nil))
}

func CanonicalString(method string, path string, timestamp string, nonce string, bodyHash string) string {
	return strings.ToUpper(method) + "\n" + path + "\n" + timestamp + "\n" + nonce + "\n" + bodyHash
}

func BodySHA256(body []byte) string {
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}

func ParseSignatureTimestamp(raw string) (time.Time, error) {
	timestamp, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(raw))
	if err != nil {
		return time.Time{}, signatureError("timestamp is invalid")
	}
	return timestamp.UTC(), nil
}

func NewNonce() (string, error) {
	var buf [16]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf[:]), nil
}

func signatureError(msg string) error {
	return &Error{Kind: ErrorKindValidation, Op: "verify signature", Msg: msg, Err: ErrInvalidSignature}
}

func runtimeError(op string, msg string, err error) error {
	return &Error{Kind: ErrorKindRuntime, Op: op, Msg: msg, Err: err}
}

func SignatureErrorf(format string, args ...interface{}) error {
	return signatureError(fmt.Sprintf(format, args...))
}
