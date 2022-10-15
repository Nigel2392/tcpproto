package tcpproto

import (
	"bytes"
	"encoding/base64"
	"errors"
	"strconv"
	"strings"
)

type Response struct {
	Headers   map[string]string
	SetValues map[string]string
	DelValues []string
	Vault     map[string]string
	Content   []byte
	File      *FileData
	Error     []error
}

func (resp *Response) AddError(err string) {
	resp.Error = append(resp.Error, errors.New(err))
}

func InitResponse() *Response {
	return &Response{
		Headers:   make(map[string]string),
		SetValues: make(map[string]string),
		DelValues: make([]string, 0),
		Vault:     make(map[string]string),
		Content:   make([]byte, 0),
		File:      &FileData{},
		Error:     make([]error, 0),
	}
}

func (resp *Response) Lock(key string, value string) *Response {
	resp.Vault[key] = value
	return resp
}

func (resp *Response) GetVault(key string) (string, bool) {
	vault, ok := resp.Vault[key]
	return vault, ok
}

func (resp *Response) Remember(k string, v string) *Response {
	resp.SetValues[k] = v
	return resp
}

func (resp *Response) Forget(key string) *Response {
	resp.DelValues = append(resp.DelValues, key)
	return resp
}

func (resp *Response) ForgetVault(key string) {
	resp.DelValues = append(resp.DelValues, "VAULT-"+key)
}

func (resp *Response) AddFile(filename string, file []byte, boundary string) {
	resp.File.Name = filename
	resp.File.Size = len(file)
	resp.File.Content = file
	resp.File.Present = true
	resp.File.Boundary = boundary
}

func (resp *Response) DecodeHeaders(headers map[string]string) (map[string]string, []string, error) {
	// Get cookie values
	forget := make([]string, 0)
	for k, v := range headers {
		if strings.HasPrefix(k, "REMEMBER-") {
			// Split the key and value
			split := strings.Split(v, "%EQUALS%")
			b64val, err := base64.StdEncoding.DecodeString(split[1])
			if err != nil {
				err := errors.New("invalid base64 value")
				return nil, nil, err
			}
			split[1] = string(b64val)
			if len(split) != 2 {
				return nil, nil, errors.New("length of split is not 2")
			}
			// Set the cookie
			resp.SetValues[split[0]] = split[1]
			delete(headers, k)
		} else if strings.HasPrefix(k, "VAULT-") {
			resp.SetValues[k] = v
			delete(headers, k)
		} else if strings.HasPrefix(k, "FORGET-") {
			// Delete the cookie
			delete(resp.SetValues, v)
			delete(headers, k)
			forget = append(forget, v)
		}
	}
	// Set headers
	for key, value := range headers {
		resp.Headers[key] = value
	}
	return resp.SetValues, forget, nil
}

func (resp *Response) ContentLength() int {
	if resp.File.Present {
		return len(resp.File.Content) + len(resp.File.StartBoundary()) + len(resp.File.EndBoundary()) + len(resp.Content)
	}
	return len(resp.Content)
	//content := resp.Content
	//if resp.File.Present {
	//	content = append(resp.File.EndBoundary(), content...)
	//	content = append(resp.File.Content, content...)
	//	content = append(resp.File.StartBoundary(), content...)
	//}
	//return len(content)
}

func (resp *Response) Generate() []byte {
	// Set up file if present
	content := resp.Content
	if resp.File.Present {
		resp.Headers["FILE_NAME"] = resp.File.Name
		resp.Headers["FILE_SIZE"] = strconv.Itoa(resp.File.Size)
		resp.Headers["FILE_BOUNDARY"] = resp.File.Boundary
		content = append(resp.File.EndBoundary(), content...)
		content = append(resp.File.Content, content...)
		content = append(resp.File.StartBoundary(), content...)
	}

	// Add content length
	resp.Headers["CONTENT_LENGTH"] = strconv.Itoa(resp.ContentLength())
	// Generate the header
	header := resp.GenHeader()
	// Add header to content
	content = append([]byte(header), content...)
	return content
}

func (resp *Response) GenHeader() string {
	// Add headers
	// Generate the header
	header := ""
	for key, value := range resp.Headers {
		header += key + ":" + value + "\n"
	}

	// Write "cookie" values onto the header
	index := 0
	for key, value := range resp.SetValues {
		index += 1
		value = base64.StdEncoding.EncodeToString([]byte(value))
		header += "REMEMBER-" + strconv.Itoa(index) + ":" + key + "%EQUALS%" + value + "\n"
	}

	// Write to the vault
	index = 0
	for key, value := range resp.Vault {
		index += 1
		// Encrypt the vault key and value
		val, err := CONF.GenVault(key, value)
		if err != nil {
			continue
		}
		header += "VAULT-" + key + ":" + val + "\n"
	}

	// Write "forget" values onto the header
	for i, value := range resp.DelValues {
		header += "FORGET-" + strconv.Itoa(i) + ":" + value + "\n"
	}

	header += "\n"
	return header
}

func (resp *Response) Bytes() []byte {
	return resp.Generate()
}

func (resp *Response) ParseFile() error {
	// Check if data includes a file
	has_file, ok := resp.Headers["HAS_FILE"]
	if ok {
		has_file, err := strconv.ParseBool(has_file)
		if err == nil {
			if has_file {
				// Read the file
				file_name, ok := resp.Headers["FILE_NAME"]
				if !ok {
					err := errors.New("file name not found")
					CONF.LOGGER.Error(err.Error())
					return err
				}
				file_size, ok := resp.Headers["FILE_SIZE"]
				if !ok {
					err := errors.New("file size not found")
					CONF.LOGGER.Error(err.Error())
					return err
				}
				file_boundary, ok := resp.Headers["FILE_BOUNDARY"]
				if !ok {
					err := errors.New("file boundary not found")
					CONF.LOGGER.Error(err.Error())
					return err
				}
				file_size_int, err := strconv.Atoi(file_size)
				if err != nil {
					err := errors.New("invalid file size")
					CONF.LOGGER.Error(err.Error())
					return err
				}
				// Set up file
				resp.File.Name = file_name
				resp.File.Size = file_size_int
				resp.File.Boundary = file_boundary
				// Parse and remove file from request.Content
				_, err = resp.ParseFileData()
				if err != nil {
					// resp.Errors = append(resp.Errors, err)
					CONF.LOGGER.Error(err.Error())
					return err
				}
			}
		}
	}
	return nil
}

func (resp *Response) ParseFileData() ([]byte, error) {
	// Set up the file boundary for parsing
	start_boundary := []byte("--" + resp.File.Boundary + "--")
	end_boundary := []byte("----" + resp.File.Boundary + "----")
	// Verify that the starting boundary is in the message
	start_index := bytes.Index(resp.Content, start_boundary)
	if start_index == -1 {
		// err := errors.New("file starting boundary not found")
		// resp.Errors = append(resp.Errors, err)
		return nil, nil
	}
	// Verify that the ending boundary is in the message
	end_index := bytes.Index(resp.Content, end_boundary)
	if end_index == -1 {
		// err := errors.New("file ending boundary not found")
		// resp.Errors = append(resp.Errors, err)
		return nil, nil
	}
	// Verify that start index is before end index
	if end_index < start_index {
		err := errors.New("file ending boundary before starting boundary")
		// resp.Errors = append(resp.Errors, err)
		return nil, err
	}
	// Extract the file data from the request.Content
	file := resp.Content[start_index+len(start_boundary) : end_index]
	if len(file) != resp.File.Size {
		err := errors.New("file size does not match")
		// resp.Errors = append(resp.Errors, err)
		return nil, err
	}
	// Remove the file and boundaries from the content
	resp.Content = bytes.Replace(resp.Content, end_boundary, []byte(""), 1)
	resp.Content = bytes.Replace(resp.Content, start_boundary, []byte(""), 1)
	resp.Content = bytes.Replace(resp.Content, file, []byte(""), -1)
	resp.Content = bytes.TrimLeft(resp.Content, "\n")
	// Set the file data
	resp.File.Present = true
	resp.File.Content = file
	return file, nil
}
