package tcpproto

import (
	"encoding/base64"
	"strings"
)

type Config struct {
	SecretKey       string
	LOGGER          *Logger
	BUFF_SIZE       int
	Default_Auth    func(rq *Request, resp *Response) error
	Include_MACaddr bool
	Include_Sysinfo bool
}

func InitConfig(secret_key string, loglevel string, buff_size int, include_MACaddr bool, include_sysinfo bool, authenticate func(rq *Request, resp *Response) error) *Config {
	return &Config{
		SecretKey:       secret_key,
		LOGGER:          NewLogger(loglevel),
		BUFF_SIZE:       buff_size,
		Include_MACaddr: include_MACaddr,
		Include_Sysinfo: include_sysinfo,
		Default_Auth:    authenticate,
	}
}

var CONF = InitConfig("SECRET_KEY", "DEBUG", 2048, true, true, Authenticate)

func SetConfig(secret_key string, loglevel string, buff_size int, include_MACaddr bool, include_sysinfo bool, authenticate func(rq *Request, resp *Response) error) *Config {
	CONF = InitConfig(secret_key, loglevel, buff_size, include_MACaddr, include_sysinfo, authenticate)
	return CONF
}

func Authenticate(rq *Request, resp *Response) error {
	return nil
}

func (c *Config) GenVault(key string, value string) (string, error) {
	// Encrypt the value
	encrypted, err := Encrypt([]byte(PadStr(c.SecretKey, 32)), []byte(key+"%EQUALS%"+value))
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
	decrypted, err := Decrypt([]byte(PadStr(c.SecretKey, 32)), decoded)
	if err != nil {
		c.LOGGER.Error(err.Error())
		return "", "", false
	}
	// Split the key and value
	split := strings.Split(string(decrypted), "%EQUALS%")
	return split[0], split[1], true
}
