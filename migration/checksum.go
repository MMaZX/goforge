package migration

import (
	"crypto/sha256"
	"encoding/hex"
)

// Checksum returns the hex-encoded SHA-256 digest of content. It is used to
// detect modifications to a migration after it has been applied.
func Checksum(content []byte) string {
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}
