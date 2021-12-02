package uuid

import (
	"crypto/rand"
	"encoding/hex"
	"io"
)

type UUID string

func (u UUID) String() string {
	return string(u)
}

func New() UUID {
	token := make([]byte, 18)
	if _, err := io.ReadFull(rand.Reader, token); err != nil {
		panic(err.Error())
	}

	result := make([]byte, 36)
	hex.Encode(result, token)

	result[8] = 0x2D
	result[13] = 0x2D
	result[14] = 0x35
	result[18] = 0x2D
	result[19] = 0x61
	result[23] = 0x2D

	return UUID(result)
}
