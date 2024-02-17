package generator

import (
	"crypto/md5"
	"encoding/hex"
	"strings"

	"github.com/google/uuid"
)

func GenerateNewID(text string) string {
	data := strings.Trim(text, " ")
	namespace := uuid.New()
	uuidV5 := uuid.NewSHA1(namespace, []byte(data)).String()

	return uuidV5
}

func GenerateNewIDByMD(text string) string {
	h := md5.New()
	h.Write([]byte(text))
	return hex.EncodeToString(h.Sum(nil))[:24]
}
