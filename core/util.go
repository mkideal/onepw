package core

import (
	"crypto/md5"
	"crypto/sha1"
	"fmt"
	"hash"
)

func md5sum(i interface{}) string {
	return hashsum(i, md5.New())
}

func sha1sum(i interface{}) string {
	return hashsum(i, sha1.New())
}

func hashsum(i interface{}, h hash.Hash) string {
	dst := []byte{}
	switch v := i.(type) {
	case string:
		h.Write([]byte(v))

	case []byte:
		h.Write(v)

	default:
		h.Write([]byte(fmt.Sprintf("%v", v)))
	}
	return fmt.Sprintf("%x", h.Sum(dst))
}

func copyNonEmptyString(dst *string, src string) {
	if src != "" {
		*dst = src
	}
}
