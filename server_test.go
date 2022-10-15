package tcpproto

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/Nigel2392/typeutils"
)

var Request_Server *Request = InitRequest()
var SERVER_REQUEST chan *Request = make(chan *Request)

func TEST_REQUESTS(rq *Request, resp *Response) {
	resp.Headers["TEST"] = "TEST"
	resp.Headers["COMMAND"] = rq.Headers["COMMAND"]
	resp.Content = []byte("TEST")
	for i := 0; i < 3; i++ {
		resp.Remember("TEST"+strconv.Itoa(i), "TEST"+strconv.Itoa(i))
	}

	// Delete cookie "TEST1"
	resp.Forget("TEST1")
	SERVER_REQUEST <- rq
}

func Test_Requests(t *testing.T) {
	server := InitServer("127.0.0.1", 12239, "PRIVKEY.pem")
	server.AddMiddlewareBeforeResp(TEST_REQUESTS)
	request := InitRequest()
	request.Headers["COMMAND"] = "TEST_REQUESTS"
	request.Headers["MESSAGE_TYPE"] = "TEST_REQUESTS"
	// Add 10 headers
	for i := 0; i < 10; i++ {
		request.Headers["TEST"+strconv.Itoa(i)] = "TEST" + strconv.Itoa(i)
	}

	request.Content = []byte(typeutils.Repeat("TEST_CONTENT\n", 200))

	request.AddFile("TEST_FILE", []byte(strings.Repeat("TEST_FILE_CONTENT\n", 200)), "MYBOUNDS")

	go func() {
		err := server.Start()
		if err != nil {
			err = errors.New("Error starting server: " + err.Error())
			t.Error(err)
		}
	}()

	client := InitClient("127.0.0.1", 12239, "PUBKEY.pem")
	if err := client.Connect(); err != nil {
		err = errors.New("error connecting to server: " + err.Error())
		t.Error(err)
	}
	client.Cookies["TEST0"] = InitCookie("TEST0", "TEST0")
	client.Cookies["TEST1"] = InitCookie("TEST1", "TEST1")
	client.Cookies["TEST2"] = InitCookie("TEST2", "TEST2")
	go func(client *Client) {
		// var response *Response = InitResponse()
		var err error = nil
		_, err = client.Send(request)
		if err != nil {
			err = errors.New("error sending request: " + err.Error())
			t.Error(err)
		}
		// CONF.LOGGER.Test("(LONG) SetValues: " + fmt.Sprintf("%v", respone.SetValues))
		// CONF.LOGGER.Test("(LONG) Headers: " + fmt.Sprintf("%v", respone.Headers))
		// CONF.LOGGER.Test("(LONG) Closing client connection")
		if err := client.Close(); err != nil {
			err = errors.New("error closing client connection: " + err.Error())
			t.Error(err)
		}
		_, ok := client.Cookies["TEST1"]
		if ok {
			t.Error("Cookie TEST1 was not deleted")
		}
	}(client)

	Request_Server = <-SERVER_REQUEST

	rqtlen := request.ContentLength()
	rqslen := Request_Server.ContentLength()

	if rqtlen != rqslen {
		t.Error("Content length mismatch: request.ContentLength()  !=  Request_Server.ContentLength()")
		t.Error("request.ContentLength(): ", rqtlen)
		t.Error("Request_Server.ContentLength(): ", rqslen)
		t.Error("request.ContentLength() == Request_Server.ContentLength(): ", rqtlen == rqslen)
	}

	if string(request.Content) != string(Request_Server.Content) {
		t.Error("Content mismatch: string(request.Content) != string(Request_Server.Content)")
	}

	rqt, err := request.Generate()
	if err != nil {
		err = errors.New("error generating request from request test: " + err.Error())
		t.Error(err)
	}
	headers_test, content_test, err := parseHeader(rqt)
	if err != nil {
		err = errors.New("error parsing header: " + err.Error())
		t.Error(err)
	}
	for key, value := range request.Headers {
		val, ok := headers_test[key]
		if !ok {
			t.Error("Test header key mismatch: " + key)
		}
		if val != value {
			t.Error("Test header mismatch: " + val)
		}
	}

	rqs, err := Request_Server.Generate()
	if err != nil {
		err = errors.New("error generating request from server request: " + err.Error())
		t.Error(err)
	}
	headers_server, content_server, err := parseHeader(rqs)
	if err != nil {
		err = errors.New("error parsing header: " + err.Error())
		t.Error(err)
	}

	for key, value := range Request_Server.Headers {
		key, ok := headers_server[key]
		if !ok {
			t.Error("Server header key mismatch: " + key)
		} else {
			if key != value {
				t.Error("Server header mismatch: " + key)
			}
		}
	}

	if string(content_test) != string(content_server) {
		t.Error("Content mismatch: string(content_test) != string(content_server)")
	}

}

var Request_Server_LONG *Request
var SERVER_REQUEST_LONG chan *Request = make(chan *Request)

func TEST_REQUESTS_LONG(rq *Request, resp *Response) {
	resp.Headers["COMMAND"] = rq.Headers["COMMAND"]
	resp.Content = []byte("TEST")
	for i := 0; i < 3; i++ {
		resp.Remember("TEST"+strconv.Itoa(i), "TEST"+strconv.Itoa(i))
	}

	// Delete cookie "TEST1"
	CONF.LOGGER.Test("(LONG) Deleting cookie TEST1")
	resp.Forget("TEST1")
	if Request_Server_LONG == nil {
		CONF.LOGGER.Test("(LONG) Locking \"TEST\"")
		resp.Lock("TEST_LOCK", "TEST_LOCK")

		CONF.LOGGER.Test("(LONG) Pushing to channel")
		SERVER_REQUEST_LONG <- rq
	} else {
		CONF.LOGGER.Test("(LONG) Request already received")
		CONF.LOGGER.Info("RESP SETVALUES:" + fmt.Sprintf("%v", resp.SetValues))
		CONF.LOGGER.Info("RESP VAULT:" + fmt.Sprintf("%v", resp.Vault))
		CONF.LOGGER.Info("REQUEST HEADERS:" + fmt.Sprintf("%v", rq.Headers))
		CONF.LOGGER.Info("REQUEST DATA:" + fmt.Sprintf("%v", rq.Data))
	}
}

func Test_Requests_LONG(t *testing.T) {
	wg := &sync.WaitGroup{}
	server := InitServer("127.0.0.1", 22239, "PRIVKEY.pem")
	server.AddMiddlewareBeforeResp(TEST_REQUESTS_LONG)
	request := InitRequest()
	request.Headers["COMMAND"] = "not-needed-for-tests"
	// Add file
	request.AddFile("test.txt", []byte(strings.Repeat("TEST_FILE_CONTENT\n", 200)), "THIS_IS_MY_FILE_BOUNDARY")

	for i := 0; i < 100; i++ {
		request.Headers["TEST"+strconv.Itoa(i)] = "TEST" + strconv.Itoa(i)
	}

	request.Content = []byte(strings.Repeat("TEST_CONTENT\n", 1000000000/16)) // 1000000000/16*13 = 0.8125GB

	go func() {
		err := server.Start()
		if err != nil {
			err = errors.New("Error starting server (LONG): " + err.Error())
			t.Error(err)
		}
	}()

	// Initialize client
	client := InitClient("127.0.0.1", 22239, "PUBKEY.pem")
	if err := client.Connect(); err != nil {
		err = errors.New("error connecting to server (LONG): " + err.Error())
		t.Error(err)
	}

	// Set client cookies
	client.Cookies["TEST0"] = InitCookie("TEST0", "TEST0")
	client.Cookies["TEST1"] = InitCookie("TEST1", "TEST1")
	client.Cookies["TEST2"] = InitCookie("TEST2", "TEST2")

	// Add data to client-side vault
	err := client.Vault("TEST_CLIENT_VAULT", "TEST_CLIENT_VAULT")
	if CONF.Use_Crypto {
		if err != nil {
			err = errors.New("error adding value to client vault (LONG): " + err.Error())
			t.Error(err)
		}
	} else {
		if err == nil {
			err = errors.New("client vault returns no error (LONG): " + err.Error())
			t.Error(err)
		}
	}

	// Send request
	wg.Add(1)
	go func(client *Client, wg *sync.WaitGroup) {
		defer wg.Done()
		CONF.LOGGER.Test("(LONG) Sending request")
		var respone *Response = InitResponse()
		var err error = nil
		respone, err = client.Send(request)
		if err != nil {
			err = errors.New("error sending request (LONG): " + err.Error())
			t.Error(err)
		}
		CONF.LOGGER.Test("(LONG) Response received, Content length: " + strconv.Itoa(respone.ContentLength()))
		// CONF.LOGGER.Test("(LONG) SetValues: " + fmt.Sprintf("%v", respone.SetValues))
		// CONF.LOGGER.Test("(LONG) Headers: " + fmt.Sprintf("%v", respone.Headers))
		_, ok := client.Cookies["TEST1"]
		CONF.LOGGER.Test("(LONG) Testing cookie TEST1 found: " + strconv.FormatBool(ok) + "\n")
		if ok {
			t.Error("Cookie TEST1 was not deleted")
		}
		CONF.LOGGER.Test("(LONG) Cookies: " + fmt.Sprintf("%v", client.Cookies))
	}(client, wg)

	// Wait for server to receive request
	Request_Server_LONG = <-SERVER_REQUEST_LONG

	// Set up contentlengths
	rqtlen := request.ContentLength()
	rqslen := Request_Server_LONG.ContentLength()

	// Verify contentlengths
	CONF.LOGGER.Test("(LONG) Testing contentlengths: " + strconv.Itoa(rqtlen) + " == " + strconv.Itoa(rqslen))
	if rqtlen != rqslen {
		t.Error("Content length mismatch: request.ContentLength()  !=  Request_Server_LONG.ContentLength()")
		t.Error("request.ContentLength(): ", rqtlen)
		t.Error("Request_Server_LONG.ContentLength(): ", rqslen)
		t.Error("request.ContentLength() == Request_Server_LONG.ContentLength(): ", rqtlen == rqslen)
	}

	// Test content
	CONF.LOGGER.Test("(LONG) Testing content")
	if string(request.Content) != string(Request_Server_LONG.Content) {
		t.Error("Content mismatch (LONG): string(request.Content) != string(Request_Server_LONG.Content)")
		CONF.LOGGER.Error("Length content: " + strconv.Itoa(len(request.Content)))
		CONF.LOGGER.Error("Server content length: " + strconv.Itoa(len(Request_Server_LONG.Content)))
	}

	// Generate request client-side
	CONF.LOGGER.Test("(LONG) Generating request")
	rqt, err := request.Generate()
	if err != nil {
		err = errors.New("error generating request from request test (LONG): " + err.Error())
		t.Error(err)
	}

	// parse headers from client-side request
	CONF.LOGGER.Test("(LONG) Parsing header")
	headers_test, content_test, err := parseHeader(rqt)
	if err != nil {
		err = errors.New("error parsing header (LONG): " + err.Error())
		t.Error(err)
	}

	// Validate client headers
	CONF.LOGGER.Test("(LONG) Validating headers")
	for key, value := range request.Headers {
		val, ok := headers_test[key]
		if !ok {
			t.Error("Test header key mismatch (LONG): " + key)
		}
		if val != value {
			t.Error("Test header mismatch (LONG): " + val + " != " + value)
		}
	}

	// generate server-side request
	CONF.LOGGER.Test("(LONG) Generating request")
	rqs, err := Request_Server_LONG.Generate()
	if err != nil {
		err = errors.New("error generating request from server request (LONG): " + err.Error())
		t.Error(err)
	}

	// parse headers from server-side request
	CONF.LOGGER.Test("(LONG) Parsing header")
	headers_server, content_server, err := parseHeader(rqs)
	if err != nil {
		err = errors.New("error parsing header (LONG): " + err.Error())
		t.Error(err)
	}

	// Validate server headers
	CONF.LOGGER.Test("(LONG) Validating headers")
	for key, value := range Request_Server_LONG.Headers {
		key, ok := headers_server[key]
		if !ok {
			t.Error("Server header key mismatch (LONG): " + key)
		} else {
			if key != value {
				t.Error("Server header mismatch (LONG): " + key)
			}
		}
	}

	// Validate file content
	if request.File.Present {
		CONF.LOGGER.Test("(LONG) Validating file")
		if string(content_test) != string(content_server) {
			t.Error("Content mismatch (LONG): string(content_test) != string(content_server)")
			CONF.LOGGER.Error("Length content: " + strconv.Itoa(len(request.Content)))
			CONF.LOGGER.Error("Server content length: " + strconv.Itoa(len(Request_Server_LONG.Content)))
		}

		if string(request.File.Content) != string(Request_Server_LONG.File.Content) {
			t.Error("File content mismatch (LONG): string(request.File.Content) != string(Request_Server_LONG.File.Content)")
			CONF.LOGGER.Error("Length content: " + strconv.Itoa(len(request.File.Content)))
			CONF.LOGGER.Error("Server content length: " + strconv.Itoa(len(Request_Server_LONG.File.Content)))
		}
	}

	// Create new request
	request = InitRequest()
	wg.Add(1)
	go func(client *Client, wg *sync.WaitGroup) {
		defer wg.Done()
		request.Headers["COMMAND"] = "not-needed-for-tests"
		request.Content = []byte(strings.Repeat("TEST_CONTENT\n", 10000000/16)) // 10000000/16*13 = 0.008125GB
		resp, err := client.Send(request)
		if err != nil {
			err = errors.New("error sending request (LONG): " + err.Error())
			t.Error(err.Error())
		}
		if resp == nil {
			t.Error("response is nil (LONG)")
		}
		if err != nil {
			err = errors.New("error generating vault (LONG): " + err.Error())
			t.Error(err.Error())
		}
		flag := false
		for _, cookie := range client.Cookies {
			if cookie.Name == "VAULT-TEST_LOCK" {
				flag = true
				key, value, ok := CONF.GetVault(cookie.Value)
				if !ok {
					t.Error("vault not found (LONG)")
				}
				if key != "TEST_LOCK" || value != "TEST_LOCK" {
					t.Error("vault mismatch (LONG): " + cookie.Value)
				}
				CONF.LOGGER.Test("(LONG) Vault found: " + key + " == " + value)
			}
		}
		if !flag {
			t.Error("Vault not found (LONG)")
		}
	}(client, wg)
	wg.Wait()

	// Validate client-side system information
	if CONF.Include_Sysinfo {
		sysinfo := GetSysInfo()
		ret_sysinfo_json := Request_Server_LONG.Headers["SYSINFO"]
		var ret_info SysInfo
		json.Unmarshal([]byte(ret_sysinfo_json), &ret_info)
		if ret_info.Hostname != sysinfo.Hostname {
			t.Error("Hostname mismatch: " + ret_info.Hostname + " != " + sysinfo.Hostname)
		}
		if ret_info.Platform != sysinfo.Platform {
			t.Error("Platform mismatch: " + ret_info.Platform + " != " + sysinfo.Platform)
		}
		if ret_info.RAM != sysinfo.RAM {
			t.Error("RAM mismatch: " + fmt.Sprintf("%v", ret_info.RAM) + " != " + fmt.Sprintf("%v", sysinfo.RAM))
		}
		if ret_info.CPU != sysinfo.CPU {
			t.Error("CPU mismatch: " + ret_info.CPU + " != " + sysinfo.CPU)
		}
		if ret_info.Disk != sysinfo.Disk {
			t.Error("Disk mismatch: " + fmt.Sprintf("%v", ret_info.Disk) + " != " + fmt.Sprintf("%v", sysinfo.Disk))
		}
		if ret_info.MacAddr != sysinfo.MacAddr {
			t.Error("MacAddr mismatch: " + ret_info.MacAddr + " != " + sysinfo.MacAddr)
		}
	}

	// Validate client-side vault
	if CONF.Use_Crypto {
		if Request_Server_LONG.Data["TEST_CLIENT_VAULT"] != "TEST_CLIENT_VAULT" {
			t.Error("CLIENT vault mismatch (LONG): " + Request_Server_LONG.Data["TEST_CLIENT_VAULT"])
		}
	}

	// CONF.LOGGER.Test("(LONG) Closing client connection")
	if err := client.Close(); err != nil {
		err = errors.New("error closing client connection: " + err.Error())
		t.Error(err)
	}
}
