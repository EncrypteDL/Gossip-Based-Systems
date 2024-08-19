package internal

import (
	"crypto/rand"
	"encoding/base64"
)

func EncodeBase64(hex []byte) string{
	return base64.StdEncoding.EncodeToString([]byte(hex))
}

func GetRandomBySlices(size int) []byte{
	data := make([]byte, size)
	_, err := rand.Read(data)
	if err != nil{
		panic(err)
	}
	return data
}