package pulsesms

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"

	"golang.org/x/crypto/pbkdf2"
)

func hashPasswordSalt(password string, salt []byte) string {
	derivedKey := pbkdf2.Key([]byte(password), salt, 10000, 32, sha1.New)
	b64Hash := base64.StdEncoding.EncodeToString(derivedKey)
	return b64Hash
}

func genAesKey(accountId string, pwHash string, salt []byte) []byte {
	combinedKey := fmt.Sprintf("%s:%s\n", accountId, pwHash)
	key := pbkdf2.Key([]byte(combinedKey), salt, 10000, 32, sha1.New)
	return key

}

func decrypt(block cipher.Block, data string) string {
	if data == "" {
		return data
	}

	// https://stackoverflow.com/questions/46475204/golang-decrypt-aes-256-cbc-base64-from-nodejs
	parts := strings.Split(data, "-:-")

	ciphertext, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		panic(err)
	}

	iv, err := base64.StdEncoding.DecodeString(parts[0])
	if err != nil {
		panic(err)
	}

	if len(ciphertext)%block.BlockSize() != 0 {
		err := fmt.Errorf("ciphertext of len %v is not a multiple of the block size: %v", len(ciphertext), block.BlockSize())
		panic(err)
	}

	mode := cipher.NewCBCDecrypter(block, iv)

	mode.CryptBlocks(ciphertext, ciphertext)
	return string(ciphertext)

}

func decryptConversation(block cipher.Block, convo *Conversation) (err error) {

	// Removes miliiseconds from timestamp
	convo.Timestamp = convo.Timestamp / 1000 >> 0 // Remove ms
	convo.Timestamp = convo.Timestamp * 1000      // Add back zero timestamp

	convo.Title = decrypt(block, convo.Title)
	convo.PhoneNumbers = decrypt(block, convo.PhoneNumbers)
	return nil

}
func decryptMessage(block cipher.Block, m *Message) (err error) {

	// Removes miliiseconds from timestamp
	m.Timestamp = m.Timestamp / 1000 >> 0 // Remove ms
	m.Timestamp = m.Timestamp * 1000      // Add back zero timestamp

	m.MimeType = decrypt(block, m.MimeType)

	// TODO handle emojis

	m.Data = decrypt(block, m.Data)
	m.From = decrypt(block, m.From)

	if m.DeviceID == 0 {
		m.DeviceID = m.ID
	}

	return nil

}

// https://gist.github.com/yingray/57fdc3264b1927ef0f984b533d63abab
func Ase256(plaintext string, key string, iv string, blockSize int) string {
	bKey := []byte(key)
	bIV := []byte(iv)
	bPlaintext := PKCS5Padding([]byte(plaintext), blockSize, len(plaintext))
	block, _ := aes.NewCipher(bKey)

	ciphertext := make([]byte, len(bPlaintext))
	mode := cipher.NewCBCEncrypter(block, bIV)
	mode.CryptBlocks(ciphertext, bPlaintext)
	return hex.EncodeToString(ciphertext)
}

func PKCS5Padding(ciphertext []byte, blockSize int, after int) []byte {
	padding := (blockSize - len(ciphertext)%blockSize)
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}
