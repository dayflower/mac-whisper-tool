package utils

import (
	"encoding/base64"
	"fmt"
)

// EncodeSessionID encodes a binary session ID to URL-safe base64 without padding
// Uses base64.RawURLEncoding which replaces + with - and / with _ for URI safety
func EncodeSessionID(id []byte) string {
	return base64.RawURLEncoding.EncodeToString(id)
}

// DecodeSessionID decodes a URL-safe base64 session ID (without padding) to binary
func DecodeSessionID(encoded string) ([]byte, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("invalid session ID format: %w", err)
	}
	return decoded, nil
}
