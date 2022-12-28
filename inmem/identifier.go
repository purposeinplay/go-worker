package inmem

import (
	"crypto/rand"
	"fmt"
	"io"
)

func generateIdentifier() string {
	b := make([]byte, 12) // nolint

	_, err := io.ReadFull(rand.Reader, b)
	if err != nil {
		return ""
	}

	return fmt.Sprintf("%x", b)
}
