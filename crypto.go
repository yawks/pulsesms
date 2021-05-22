package pulsesms

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/pbkdf2"
)

func hashPasswordSalt(password string, salt []byte) string {
	derivedKey := pbkdf2.Key([]byte(password), salt, 10000, 256, sha1.New)
	b64Hash := base64.StdEncoding.EncodeToString(derivedKey)
	return b64Hash
}

func genAesKey(accountId string, pwHash string, salt []byte) []byte  {
    combinedKey := fmt.Sprintf("%s:%s\n", accountId, pwHash)
    key := pbkdf2.Key([]byte(combinedKey), salt, 10000, 256, sha1.New)
    return key
}




