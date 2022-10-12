package tcpproto

import (
	"bytes"
	"errors"
	"net"
	"strconv"
)

// Format of message looks like this

// CONTENT_LENGTH: CONTENT_LENGTH int
// CLIENT_ID: CLIENT_ID int
// MESSAGE_TYPE: MESSAGE_TYPE string
// COMMAND: COMMAND string
//
// CONTENT CONTENT CONTENT
// CONTENT CONTENT CONTENT
// CONTENT CONTENT CONTENT

// Format of message with files looks like this

// CONTENT_LENGTH: CONTENT_LENGTH int
// CLIENT_ID: CLIENT_ID int
// MESSAGE_TYPE: MESSAGE_TYPE string
// COMMAND: COMMAND string
// HAS_FILE: HAS_FILE bool
// FILE_NAME: FILE_NAME string
// FILE_SIZE: FILE_SIZE int
// FILE_BOUNDARY: FILE_BOUNDARY string
//
// FILE_BOUNDARY
// FILE CONTENT
// FILE CONTENT
// FILE CONTENT
// FILE_BOUNDARY
//
// CONTENT CONTENT CONTENT
// CONTENT CONTENT CONTENT
// CONTENT CONTENT CONTENT

func (rq *Request) ParseFile() error {
	// Check if data includes a file
	has_file, ok := rq.Headers["HAS_FILE"]
	if ok {
		has_file, err := strconv.ParseBool(has_file)
		if err == nil {
			if has_file {
				// Read the file
				file_name, ok := rq.Headers["FILE_NAME"]
				if !ok {
					err := errors.New("file name not found")
					LOGGER.Error(err.Error())
					return err
				}
				file_size, ok := rq.Headers["FILE_SIZE"]
				if !ok {
					err := errors.New("file size not found")
					LOGGER.Error(err.Error())
					return err
				}
				file_boundary, ok := rq.Headers["FILE_BOUNDARY"]
				if !ok {
					err := errors.New("file boundary not found")
					LOGGER.Error(err.Error())
					return err
				}
				file_size_int, err := strconv.Atoi(file_size)
				if err != nil {
					err := errors.New("invalid file size")
					LOGGER.Error(err.Error())
					return err
				}
				// Set up file
				rq.File.Name = file_name
				rq.File.Size = file_size_int
				rq.File.Boundary = file_boundary
				// Parse and remove file from request.Content
				_, err = rq.ParseFileData()
				if err != nil {
					// rq.Errors = append(rq.Errors, err)
					LOGGER.Error(err.Error())
					return err
				}
			}
		}
	}
	return nil
}

func (rq *Request) ParseFileData() ([]byte, error) {
	// Set up the file boundary for parsing
	start_boundary := []byte("--" + rq.File.Boundary + "--")
	end_boundary := []byte("----" + rq.File.Boundary + "----")
	//
	// regex_str := string(start_boundary) + "(.*?)" + string(end_boundary)
	// regex := regexp.MustCompile(regex_str)
	// Find the file data
	// file_data := regex.FindSubmatch(rq.Content)
	// fmt.Println(file_data)
	// if len(file_data) == 2 {
	// Set the file data
	// rq.File.Content = file_data[1]
	// Remove the file data from the content
	// rq.Content = bytes.Replace(rq.Content, file_data[0], []byte(""), 1)
	// return file_data[1], nil
	// }
	// err := errors.New("file data not found")
	// return nil, err
	// Verify that the starting boundary is in the message
	start_index := bytes.Index(rq.Content, start_boundary)
	if start_index == -1 {
		// err := errors.New("file starting boundary not found")
		// rq.Errors = append(rq.Errors, err)
		return nil, nil
	}
	// Verify that the ending boundary is in the message
	end_index := bytes.Index(rq.Content, end_boundary)
	if end_index == -1 {
		// err := errors.New("file ending boundary not found")
		// rq.Errors = append(rq.Errors, err)
		return nil, nil
	}
	// Verify that start index is before end index
	if end_index < start_index {
		err := errors.New("file ending boundary before starting boundary")
		// rq.Errors = append(rq.Errors, err)
		return nil, err
	}
	// Extract the file data from the request.Content
	file := rq.Content[start_index+len(start_boundary) : end_index]
	if len(file) != rq.File.Size {
		err := errors.New("file size does not match")
		// rq.Errors = append(rq.Errors, err)
		return nil, err
	}
	// Remove the file and boundaries from the content
	rq.Content = bytes.Replace(rq.Content, end_boundary, []byte(""), 1)
	rq.Content = bytes.Replace(rq.Content, start_boundary, []byte(""), 1)
	rq.Content = bytes.Replace(rq.Content, file, []byte(""), -1)
	rq.Content = bytes.TrimLeft(rq.Content, "\n")
	// Set the file data
	rq.File.Present = true
	rq.File.Content = file
	return file, nil
}

func parseHeader(data []byte) (map[string]string, []byte, error) {
	header := make(map[string]string)
	// Split the data into header and message
	// The header is split into key value pairs
	// The header is the first section before the first double newline
	// The message is the rest of the data

	// Split the data into header and message. Only split on the first double newline
	content := bytes.SplitN(data, []byte("\n\n"), 2)
	if len(content) == 2 {
		header_data := content[0]
		message_data := content[1]
		// Split the header into key value pairs
		header_lines := bytes.Split(header_data, []byte("\n"))
		for _, line := range header_lines {
			// Remove spaces
			line = bytes.Replace(line, []byte(" "), []byte(""), -1)
			// Split the line into key value
			line_data := bytes.Split(bytes.TrimSpace(line), []byte(":"))
			if len(line_data) != 2 {
				err := errors.New("invalid header")
				LOGGER.Error(err.Error())
				return nil, nil, err
			}
			header[string(line_data[0])] = string(line_data[1])
		}
		// Return the header and message
		return header, message_data, nil
	} else {
		// Message is not formatted correctly
		err := errors.New("invalid header")
		LOGGER.Error(err.Error())
		return nil, nil, err
	}
}

func ParseConnection(conn net.Conn) (*Request, *Response, error) {
	// Read the header when one is sent.
	data_part_one, err := getHeader(conn)
	if err != nil {
		return nil, nil, err
	}
	// Parse the header
	header, recv_data, err := parseHeader(data_part_one)
	if err != nil {
		LOGGER.Error(err.Error())
		return nil, nil, err
	}

	// Get the content length
	content_length, err := strconv.Atoi(header["CONTENT_LENGTH"])
	if err != nil {
		err = errors.New("invalid content length")
		LOGGER.Error(err.Error())
		return nil, nil, err
	}

	// Read the rest of the content.
	recv_data, err = getContent(conn, recv_data, content_length)
	if err != nil {
		LOGGER.Error(err.Error())
		return nil, nil, err
	}

	// Verify content length
	if len(recv_data) != content_length {
		err = errors.New("content length mismatch")
		LOGGER.Error(err.Error())
		return nil, nil, err
	}
	// Initialize request
	rq := InitRequest()
	rq.Headers = header
	rq.Content = recv_data
	rq.Conn = conn

	// Create the response
	resp := InitResponse()

	// Parse the file if one exists:
	err = rq.ParseFile()
	if err != nil {
		return nil, nil, err
	}
	return rq, resp, nil
}

func getHeader(conn net.Conn) ([]byte, error) {
	data_part_one := make([]byte, BUFF_SIZE)
	_, err := conn.Read(data_part_one)
	if err != nil {
		err = errors.New("error reading header")
		return nil, err
	}

	for !bytes.Contains(data_part_one, []byte("\n\n")) {
		data_part_two := make([]byte, BUFF_SIZE)
		_, err := conn.Read(data_part_two)
		if err != nil {
			err = errors.New("invalid header")
			return nil, err
		}
		data_part_one = append(data_part_one, data_part_two...)
	}
	return data_part_one, nil
}

func getContent(conn net.Conn, recv_data []byte, content_length int) ([]byte, error) {
	if len(recv_data) >= content_length {
		recv_data = recv_data[:content_length]
	} else {
		// Read the rest of the data
		for len(recv_data) < content_length {
			data := make([]byte, BUFF_SIZE)
			_, err := conn.Read(data)
			if err != nil {
				return nil, err
			}
			recv_data = append(recv_data, data...)
		}
		recv_data = recv_data[:content_length]
	}

	return recv_data, nil
}
