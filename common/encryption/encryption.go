package encryption

import (
	"encoding/base64"
)

const key string = "1rG" // XOR encryption key

// Encrypt provided text.
func Encrypt(text []byte) string {
	textBase64 := base64.StdEncoding.EncodeToString(text)
	textEncrypted := encryptDecryptXOR(textBase64)
	return textEncrypted
}

// Decrypt provided text.
// Return error if base64 part fail.
func Decrypt(secret string) ([]byte, error) {
	textBase64 := encryptDecryptXOR(secret)
	textDecrypted, err := base64.StdEncoding.DecodeString(textBase64)
	if err != nil {
		return nil, err
	}
	return textDecrypted, nil
}

// encryptDecryptXOR runs a XOR encryption on the input string, encrypting it if it hasn't already been,
// and decrypting it if it has, using the key provided.
func encryptDecryptXOR(input string) (output string) {
	for i := 0; i < len(input); i++ {
		output += string(input[i] ^ key[i%len(key)])
	}
	return output
}
