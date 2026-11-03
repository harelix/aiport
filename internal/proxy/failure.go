package proxy

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"io"
	"net"
	"regexp"
	"strings"
	"syscall"

	"github.com/harelix/aiport/internal/refusal"
)

// errorBodyLimit bounds how much of a provider's error body is looked at to
// find its error type. Only error responses are looked at, and only the type
// is kept.
const errorBodyLimit = 16 << 10

// identifier is what a provider error type or code must look like to be
// kept: a short identifier, never free text that could echo a prompt.
var identifier = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_.-]{0,63}$`)

// ErrorType returns a provider's short error type from an error body: the
// OpenAI-style error.code, else error.type (also Anthropic's), else a
// top-level code or type. Message text is never returned; "" when there is no
// identifier-shaped value.
func ErrorType(body []byte) string {
	var shape struct {
		Error json.RawMessage `json:"error"`
		Type  any             `json:"type"`
		Code  any             `json:"code"`
	}
	if json.Unmarshal(body, &shape) != nil {
		return ""
	}
	var inner struct {
		Type any `json:"type"`
		Code any `json:"code"`
	}
	_ = json.Unmarshal(shape.Error, &inner)
	for _, v := range []any{inner.Code, inner.Type, shape.Code} {
		if s, ok := v.(string); ok && identifier.MatchString(s) {
			return s
		}
	}
	if s, ok := shape.Type.(string); ok && s != "error" && identifier.MatchString(s) {
		return s
	}
	return ""
}

// ConnectionError classifies why a connection to a provider failed into one
// of the fixed refusal connection classes. The error text is never kept.
func ConnectionError(err error) string {
	var dns *net.DNSError
	var cert *tls.CertificateVerificationError
	var record tls.RecordHeaderError
	var authority x509.UnknownAuthorityError
	var host x509.HostnameError
	var invalid x509.CertificateInvalidError
	var timeout net.Error
	switch {
	case errors.As(err, &dns):
		if dns.IsTimeout {
			return refusal.ConnectTimeout
		}
		return refusal.DNSFailure
	case errors.Is(err, syscall.ECONNREFUSED):
		return refusal.ConnectionRefused
	case errors.As(err, &cert), errors.As(err, &record), errors.As(err, &authority), errors.As(err, &host), errors.As(err, &invalid), strings.Contains(err.Error(), "tls: "):
		return refusal.TLSError
	case errors.Is(err, syscall.ECONNRESET), errors.Is(err, io.EOF), errors.Is(err, io.ErrUnexpectedEOF):
		return refusal.ConnectionReset
	case errors.As(err, &timeout) && timeout.Timeout():
		return refusal.ConnectTimeout
	}
	return refusal.ConnectionFailed
}

// errorBody keeps the start of a provider's error body to classify it.
type errorBody struct{ data []byte }

func (b *errorBody) add(p []byte) {
	if b == nil {
		return
	}
	if room := errorBodyLimit - len(b.data); room > 0 {
		if len(p) > room {
			p = p[:room]
		}
		b.data = append(b.data, p...)
	}
}
