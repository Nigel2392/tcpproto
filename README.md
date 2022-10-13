# tcpproto
Simple http-like layer ontop of TCP.

This layer is capable of transfering files, authentication, client-side storage and more!

A typical response/request looks like this:
```
CONTENT_LENGTH: CONTENT_LENGTH int
COMMAND: COMMAND string
CUSTOM_HEADER: CUSTOM_HEADER string
CUSTOM_HEADER: CUSTOM_HEADER string
CUSTOM_HEADER: CUSTOM_HEADER string
CUSTOM_HEADER: CUSTOM_HEADER string
HAS_FILE: HAS_FILE bool
FILE_NAME: FILE_NAME string
FILE_SIZE: FILE_SIZE int
FILE_BOUNDARY: FILE_BOUNDARY string

FILE_BOUNDARY
FILE CONTENT
FILE CONTENT
FILE CONTENT
FILE_BOUNDARY

CONTENT CONTENT CONTENT
CONTENT CONTENT CONTENT
CONTENT CONTENT CONTENT
```
Where anything that has to do with files, can optionally be left out.  
To add a file, you can use the following:
```
request.AddFile(filename string, content []byte, boundary string)
```
The following is needed:  
`CONTENT_LENGTH` and `COMMAND`  
Content length is used to make sure the whole request is parsed properly, and chunks we not forgotten.  
Command is used to add callbacks to the request/response cycle, where you can edit either one.  
## Server
A typical server looks like this:  
```
func main() {
	ipaddr := "127.0.0.1"
	port := 22392
	s := tcpproto.InitServer(ipaddr, port)
	s.AddMiddlewareBeforeResp(AuthMiddleware) // Middleware to be called before the callback is called.
	s.AddMiddlewareAfterResp(tcpproto.LogMiddleware) // Middleware to be called after the callback is called.
	s.AddCallback("SET", SET) // The callback to be called, derived from "COMMAND" header.
	s.AddCallback("GET", GET)
	if err := s.Start(); err != nil {
		os.Exit(1)
	}
}
```

To add middleware, or callbacks, the function needs to take the following arguments:
```
func MiddlewareOrCallback(rq *tcpproto.Request, resp *tcpproto.Response)
```

## Client:
A typical client looks like this:
```
// Initialise request
request := InitRequest()

// Set some headers
// CONTENT_LENGTH is automatically added when generating the request.
request.Headers["COMMAND"] = "SET"
for i := 0; i < 10; i++ {
	request.Headers["HEADER-TEST"+strconv.Itoa(i)] = "VALUE-TEST" + strconv.Itoa(i)
}

// Set request content
request.Content = []byte(strings.Repeat("TEST_CONTENT\n", 200))

// Initialize client
client := InitClient("127.0.0.1", 12239)
// Connect to server
if err := client.Connect(); err != nil {
	err = errors.New("error connecting to server: " + err.Error())
	t.Error(err)
}
// Set some "cookies"
client.Cookies["TEST0"] = InitCookie("TEST0", "TEST0")
client.Cookies["TEST1"] = InitCookie("TEST1", "TEST1")
client.Cookies["TEST2"] = InitCookie("TEST2", "TEST2")
// Get the response when sending the data to the server
response, err = client.Send(request)
if err != nil {
	err = errors.New("error sending request: " + err.Error())
	t.Error(err)
}
// Close the client when done
if err := client.Close(); err != nil {
	err = errors.New("error closing client connection: " + err.Error())
	t.Error(err)
}
```
As you can see, the client receives the response back when sending data to a server. 
This data fits into the following struct:
```
// Response the server sends back
type Response struct {
	Headers   map[string]string
	SetValues map[string]string
	DelValues []string
	Content   []byte
	File      *FileData
}
// Request to send to the server
type Request struct {
	Headers map[string]string
	Content []byte
	File    *FileData
	Data    map[string]string
	User    *User
	Conn    net.Conn
}
```