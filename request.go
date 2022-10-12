package tcpproto

import (
	"net"
	"strconv"
)

type User struct {
	ID              int
	Username        string
	Email           string
	Password        string
	IsAdmin         bool
	IsAuthenticated bool
}

type Request struct {
	Headers map[string]string
	Content []byte
	File    *FileData
	Data    map[string]string
	User    *User
	Conn    net.Conn
}

func InitRequest() *Request {
	return &Request{
		Headers: make(map[string]string),
		Content: []byte{},
		File: &FileData{
			Name:     "",
			Size:     0,
			Boundary: "",
			Present:  false,
			Content:  []byte{},
		},
		Data: make(map[string]string),
		User: &User{},
	}
}

func (rq *Request) AddFile(filename string, file []byte, boundary string) {
	rq.File.Name = filename
	rq.File.Size = len(file)
	rq.File.Content = file
	rq.File.Present = true
	rq.File.Boundary = boundary
}

func (rq *Request) ContentLength() int {
	content := rq.Content
	if rq.File.Present {
		content = append(rq.File.EndBoundary(), content...)
		content = append(rq.File.Content, content...)
		content = append(rq.File.StartBoundary(), content...)
	}
	return len(content)
}

func (rq *Request) Generate() ([]byte, error) {
	// nowtime := time.Now()
	// LOGGER.Debug("Starting headers")

	// Generate the header
	header := ""
	// Set up file if present
	content := rq.Content
	if rq.File.Present {
		rq.Headers["FILE_NAME"] = rq.File.Name
		rq.Headers["FILE_SIZE"] = strconv.Itoa(rq.File.Size)
		rq.Headers["FILE_BOUNDARY"] = rq.File.Boundary
		rq.Headers["HAS_FILE"] = "true"
		// Generate the content
		content = append(rq.File.EndBoundary(), content...)
		content = append(rq.File.Content, content...)
		content = append(rq.File.StartBoundary(), content...)
	}
	rq.Headers["CONTENT_LENGTH"] = strconv.Itoa(len(content))
	// Generate headers
	for key, value := range rq.Headers {
		header += key + ":" + value + "\n"
	}
	header += "\n"
	// Generate the final request
	return append([]byte(header), content...), nil
}
