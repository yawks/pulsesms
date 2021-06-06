package pulsesms

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
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

func randomHex(length int) (string, error) {
	n := length / 2
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil

}

func encrypt(block cipher.Block, data string) (string, error) {
	if data == "" {
		return data, nil
	}
	rhex, err := randomHex(128)
	if err != nil {
		return "", err
	}

	hexString, _ := hex.DecodeString(rhex)
	iv := hexString[:block.BlockSize()]

	if len(iv) != block.BlockSize() {
		err := fmt.Errorf("iv length %v does not equal block size %v", len(iv), block.BlockSize())
		panic(err)
	}

	plain := []byte(data)

	plain, _ = pkcs7Pad(plain, block.BlockSize())

	if len(plain)%block.BlockSize() != 0 {
		err := fmt.Errorf("ciphertext of len %v is not a multiple of the block size: %v", len(plain), block.BlockSize())
		panic(err)
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(plain, plain)
	return base64.StdEncoding.EncodeToString(iv) + "-:-" + base64.StdEncoding.EncodeToString(plain), nil

}

func decrypt(block cipher.Block, data string) (string, error) {
	if data == "" {
		return data, nil
	}

	// https://stackoverflow.com/questions/46475204/golang-decrypt-aes-256-cbc-base64-from-nodejs
	parts := strings.Split(data, "-:-")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid encrypted data format")
	}

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
	content := string(ciphertext)
	content = strings.TrimSpace(content)
	return content, nil

}

// pkcs7Pad right-pads the given byte slice with 1 to n bytes, where
// n is the block size. The size of the result is x times n where x
// is at least 1.
// https://gist.github.com/huyinghuan/7bf174017bf54efb91ece04a48589b22
func pkcs7Pad(b []byte, blocksize int) ([]byte, error) {
	if blocksize <= 0 {
		return nil, fmt.Errorf("invalid block size")
	}
	if b == nil || len(b) == 0 {
		return nil, fmt.Errorf("invalid pkcs7 data format")

	}
	n := blocksize - (len(b) % blocksize)
	pb := make([]byte, len(b)+n)
	copy(pb, b)
	copy(pb[len(b):], bytes.Repeat([]byte{byte(n)}, n))
	return pb, nil
}
func decryptConversation(block cipher.Block, convo *conversation) (err error) {

	// Removes miliiseconds from timestamp
	convo.Timestamp = convo.Timestamp / 1000 >> 0 // Remove ms
	convo.Timestamp = convo.Timestamp * 1000      // Add back zero timestamp

	convo.Title, err = decrypt(block, convo.Title)
	convo.PhoneNumbers, err = decrypt(block, convo.PhoneNumbers)

	return nil

}
func decryptMessage(block cipher.Block, m *Message) (err error) {

	// Removes miliiseconds from timestamp
	m.Timestamp = m.Timestamp / 1000 >> 0 // Remove ms
	m.Timestamp = m.Timestamp * 1000      // Add back zero timestamp

	// TODO handle emojis

	m.Data, err = decrypt(block, m.Data)
	if err != nil {
		return err
	}
	m.From, err = decrypt(block, m.From)
	if err != nil {
		return err
	}

	// if m.DeviceID == 0 {
	// 	m.DeviceID = m.ID
	// }

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
