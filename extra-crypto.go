package tcpproto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
)

func BytesToBase64(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

func Base64ToBytes(s string) []byte {
	decoded, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return decoded
}

// NewEncryptionKey generates a random 256-bit key for Encrypt() and
// Decrypt(). It panics if the source of randomness fails.
func NewEncryptionKey() *[32]byte {
	key := [32]byte{}
	_, err := io.ReadFull(rand.Reader, key[:])
	if err != nil {
		panic(err)
	}
	return &key
}

// Encrypt encrypts data using 256-bit AES-GCM.  This both hides the content of
// the data and provides a check that it hasn't been altered. Output takes the
// form nonce|ciphertext|tag where '|' indicates concatenation.
func Encrypt(plaintext []byte, key *[32]byte) (ciphertext []byte, err error) {
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// Decrypt decrypts data using 256-bit AES-GCM.  This both hides the content of
// the data and provides a check that it hasn't been altered. Expects input
// form nonce|ciphertext|tag where '|' indicates concatenation.
func Decrypt(ciphertext []byte, key *[32]byte) (plaintext []byte, err error) {
	if key == nil {
		return nil, errors.New("key is nil")
	}
	if key == &[32]byte{} {
		return nil, errors.New("key is empty")
	}
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < gcm.NonceSize() {
		return nil, errors.New("malformed ciphertext")
	}

	return gcm.Open(nil,
		ciphertext[:gcm.NonceSize()],
		ciphertext[gcm.NonceSize():],
		nil,
	)
}

func EncryptString(s string, key *[32]byte) (string, error) {
	ciphertext, err := Encrypt([]byte(s), key)
	if err != nil {
		return "", err
	}
	return BytesToBase64(ciphertext), nil
}

func DecryptString(s string, key *[32]byte) (string, error) {
	ciphertext := Base64ToBytes(s)
	plaintext, err := Decrypt(ciphertext, key)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

func KeyToBase64(key *[32]byte) string {
	return BytesToBase64(key[:])
}

func Base64ToKey(s string) *[32]byte {
	key := [32]byte{}
	copy(key[:], s)
	return &key
}

func PadStr(s string, l int) string {
	if len(s) > l {
		return s[:l]
	}
	return s + strings.Repeat("$", l-len(s))
}

func UnpadStr(s string) string {
	for strings.HasSuffix(s, "$") {
		s = s[:len(s)-1]
	}
	return s
}

// EncryptWithPublicKey encrypts data with public key
func EncryptWithPublicKey(msg []byte, pub *rsa.PublicKey) []byte {
	hash := sha512.New()
	ciphertext, err := rsa.EncryptOAEP(hash, rand.Reader, pub, msg, nil)
	if err != nil {
		CONF.LOGGER.Error("Public key bit size is too small for this message: " + fmt.Sprintf("%d", pub.Size()))
		CONF.LOGGER.Error("Error encrypting data with public key (msglength: " + fmt.Sprintf("%v", len(msg)) + "): " + err.Error())
	}
	return ciphertext
}

func PubKeySTR_to_PubKey(pubkeystr string) *rsa.PublicKey {
	block, _ := pem.Decode([]byte(pubkeystr))
	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		CONF.LOGGER.Error("Error converting public key: " + err.Error())
	}
	return key.(*rsa.PublicKey)
}

func ExportPublic_PEM_Key(key *rsa.PublicKey, filename string) {
	block, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		CONF.LOGGER.Error("Error exporting public key: " + err.Error())
	}
	err = ioutil.WriteFile(filename, pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: block}), 0644)
	if err != nil {
		CONF.LOGGER.Error("Error exporting public key: " + err.Error())
	}
}

func ImportPublic_PEM_Key(filename string) *rsa.PublicKey {
	keyfile, err := ioutil.ReadFile(filename)
	if err != nil {
		CONF.LOGGER.Error("Error importing public key: " + err.Error())
	}
	block, _ := pem.Decode(keyfile)
	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		CONF.LOGGER.Error("Error importing public key: " + err.Error())
	}
	return key.(*rsa.PublicKey)
}

func GenerateKeyPair(bits int) (*rsa.PrivateKey, *rsa.PublicKey) {
	privkey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		CONF.LOGGER.Error("Error generating key pair: " + err.Error())
	}
	return privkey, &privkey.PublicKey
}

// DecryptWithPrivateKey decrypts data with private key
func DecryptWithPrivateKey(ciphertext []byte, priv *rsa.PrivateKey) []byte {
	hash := sha512.New()
	plaintext, err := rsa.DecryptOAEP(hash, rand.Reader, priv, ciphertext, nil)
	if err != nil {
		CONF.LOGGER.Error("Error decrypting data with private key: " + err.Error())
	}
	return plaintext
}

func ExportPrivate_PEM_Key(key *rsa.PrivateKey, filename string) {
	block, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		CONF.LOGGER.Error("Error marshalling private key: " + err.Error())
	}
	err = ioutil.WriteFile(filename, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: block}), 0644)
	if err != nil {
		CONF.LOGGER.Error("Error writing private key to file: " + err.Error())
	}
}

func PrivKeySTR_to_PrivKey(privkeystr string) *rsa.PrivateKey {
	block, _ := pem.Decode([]byte(privkeystr))
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		CONF.LOGGER.Error("Error converting private key: " + err.Error())
	}
	return key.(*rsa.PrivateKey)
}

func ImportPrivate_PEM_Key(filename string) *rsa.PrivateKey {
	keyfile, err := ioutil.ReadFile(filename)
	if err != nil {
		CONF.LOGGER.Error("Error importing private key: " + err.Error())
	}
	block, _ := pem.Decode(keyfile)
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		CONF.LOGGER.Error("Error parsing private key: " + err.Error())
	}
	return key.(*rsa.PrivateKey)
}
