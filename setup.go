package tcpproto

import (
	"embed"
	"encoding/base64"
	"strings"
)

//go:embed PUBKEY.pem
//go:embed PRIVKEY.pem
var PEM embed.FS

type Config struct {
	SecretKey          string
	LOGGER             *Logger
	BUFF_SIZE          int
	Default_Auth       func(rq *Request, resp *Response) error
	Include_Sysinfo    bool
	Use_Crypto         bool
	MAX_CONTENT_LENGTH int
	MAX_HEADER_SIZE    int
}

func InitConfig(secret_key string, loglevel string, buff_size int, max_length int, use_crypto bool, include_sysinfo bool, authenticate func(rq *Request, resp *Response) error) *Config {
	return &Config{
		SecretKey:          secret_key,
		LOGGER:             NewLogger(loglevel),
		BUFF_SIZE:          buff_size,
		Include_Sysinfo:    include_sysinfo,
		Default_Auth:       authenticate,
		Use_Crypto:         use_crypto,
		MAX_CONTENT_LENGTH: max_length,
		MAX_HEADER_SIZE:    max_length,
	}
}

const (
	DISABLED     = 0
	KILOBYTE     = 1024
	MEGABYTE     = 1024 * KILOBYTE
	GIGABYTE     = 1024 * MEGABYTE
	TEN_GIGABYTE = 10 * GIGABYTE
)

var CONF = InitConfig("SECRET_KEY", "DEBUG", 2048, DISABLED, true, true, Authenticate)

func SetConfig(secret_key string, loglevel string, buff_size int, max_length int, use_crypto bool, include_sysinfo bool, authenticate func(rq *Request, resp *Response) error) *Config {
	CONF = InitConfig(secret_key, loglevel, buff_size, max_length, use_crypto, include_sysinfo, authenticate)
	return CONF
}

func GetConfig() *Config {
	return CONF
}

func Authenticate(rq *Request, resp *Response) error {
	return nil
}

func (c *Config) GenVault(key string, value string) (string, error) {
	// Encrypt the value
	enc_key := &[32]byte{}
	copy(enc_key[:], []byte(PadStr(c.SecretKey, 32)))

	encrypted, err := Encrypt([]byte(key+"%EQUALS%"+value), enc_key)
	if err != nil {
		c.LOGGER.Error(err.Error())
		return "", err
	}
	b64 := base64.StdEncoding.EncodeToString(encrypted)
	return b64, nil
}

func (c *Config) GetVault(value string) (string, string, bool) {
	// Decrypt the value
	decoded, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		c.LOGGER.Error(err.Error())
		return "", "", false
	}
	enc_key := &[32]byte{}
	copy(enc_key[:], []byte(PadStr(c.SecretKey, 32)))

	decrypted, err := Decrypt(decoded, enc_key)
	if err != nil {
		c.LOGGER.Error(err.Error())
		return "", "", false
	}
	// Split the key and value
	split := strings.Split(string(decrypted), "%EQUALS%")
	return split[0], split[1], true
}
