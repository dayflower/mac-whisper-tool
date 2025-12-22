package utils

import (
	"encoding/base64"
	"fmt"
)

// EncodeSessionID encodes a binary session ID to base64 without padding
func EncodeSessionID(id []byte) string {
	return base64.RawStdEncoding.EncodeToString(id)
}

// DecodeSessionID decodes a base64 session ID (without padding) to binary
func DecodeSessionID(encoded string) ([]byte, error) {
	decoded, err := base64.RawStdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("invalid session ID format: %w", err)
	}
	return decoded, nil
}
