package cryptoutil

import (
	"crypto/sha1"
	"fmt"
)

func SHA1Sum(data []byte) []byte {
	sum := sha1.Sum(data)
	return sum[:]
}

func SHA1String(data []byte) string {
	return fmt.Sprintf("%x", sha1.Sum(data))
}
