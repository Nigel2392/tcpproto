package tcpproto

import (
	"crypto/rsa"
	"encoding/base64"
	"errors"
	"net"
	"strconv"
)

type Client struct {
	IP          string
	Port        int
	Conn        net.Conn
	Cookies     map[string]*Cookie
	ClientVault map[string]string
	PUBKEY      *rsa.PublicKey
}

func (c *Client) Addr() string {
	return c.IP + ":" + strconv.Itoa(c.Port)
}

func InitClient(ip string, port int, pubkey_file string) *Client {
	client := &Client{
		IP:          ip,
		Port:        port,
		Cookies:     map[string]*Cookie{},
		ClientVault: map[string]string{},
	}
	if CONF.Use_Crypto {
		// Load the public key
		client.PUBKEY = ImportPublic_PEM_Key(pubkey_file)
	}
	return client
}

func (c *Client) Vault(key string, value string) error {
	if CONF.Use_Crypto {
		c.ClientVault[key] = value
		return nil
	}
	return errors.New("crypto is disabled")
}

func (c *Client) Connect() error {
	var err error
	c.Conn, err = net.Dial("tcp", c.Addr())
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) Close() error {
	return c.Conn.Close()
}

func (c *Client) Send(rq *Request) (*Response, error) {
	for key, val := range c.Cookies {
		rq.AddCookie(key, val.Value)
	}

	if CONF.Use_Crypto {
		if c.PUBKEY != nil {
			for key, val := range c.ClientVault {
				val := EncryptWithPublicKey([]byte(val), c.PUBKEY)
				strval := base64.StdEncoding.EncodeToString(val)
				rq.AddHeader("CLIENT_VAULT-"+key, strval)
			}
		} else {
			err := errors.New("no public key provided")
			CONF.LOGGER.Error(err.Error())
			return nil, err
		}
	}

	if CONF.Include_Sysinfo {
		sysinfo := GetSysInfo()
		rq.Headers["SYSINFO"] = sysinfo.ToJSON()
	}

	// Send the request
	err := c.send(rq)
	if err != nil {
		return nil, err
	}
	header, recv_data, err := c.recv_data()
	if err != nil {
		return nil, err
	}
	// Parse the response
	resp, err := c.ParseResponse(rq, header, recv_data)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) ParseResponse(rq *Request, header map[string]string, recv_data []byte) (*Response, error) {
	// Initialize response
	resp := InitResponse()
	// Decode the response
	remember, forget, err := resp.DecodeHeaders(header)
	if err != nil {
		return nil, err
	}
	c.UpdateCookies(remember, forget)
	// Set the content
	resp.Content = recv_data
	// Parse possible included files
	err = rq.ParseFile()
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) UpdateCookies(remember map[string]string, forget []string) {
	// Set client cookies
	for key, val := range remember {
		c.Cookies[key] = InitCookie(key, val)
	}
	// Remove client cookies
	for _, key := range forget {
		delete(c.Cookies, key)
	}

}

func (c *Client) recv_data() (map[string]string, []byte, error) {
	// Receive response
	buf, err := getHeader(c.Conn)
	if err != nil {
		return nil, nil, err
	}
	// Parse response
	header, recv_data, err := parseHeader(buf)
	if err != nil {
		return nil, nil, err
	}
	// Get the content length
	content_length, err := strconv.Atoi(header["CONTENT_LENGTH"])
	if err != nil {
		err = errors.New("invalid content length")
		CONF.LOGGER.Error(err.Error())
		return nil, nil, err
	}
	// Get the rest of content
	recv_data, err = getContent(c.Conn, recv_data, content_length)
	if err != nil {
		CONF.LOGGER.Error(err.Error())
		return nil, nil, err
	}
	return header, recv_data, nil
}

func (c *Client) send(rq *Request) error {
	// Send the request
	content, err := rq.Generate()
	if err != nil {
		return err
	}
	// Write to server
	_, err = c.Conn.Write(content)
	if err != nil {
		return err
	}
	return nil
}
