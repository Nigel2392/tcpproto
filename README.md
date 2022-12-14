# tcpproto
Simple http-like layer ontop of TCP.
This package currently supports:
* Getting system information from requests (`request.SysInfo()`):
  * Hostname
  * CPU
  * Platform (OS)
  * Mac Address
  * Max Memory
  * Max Disk
* Encrypting certain headers client side, with an RSA public key
  * Only works if public/private key is provided, and CONF.Use_Crypto=true
* Cookies encrypted with the SECRET_KEY
* Non-encrypted cookies
* File inside of requests (Single file only)
* Max size of requests
* Handling based upon `COMMAND:` header
* Support for middleware before and after calling the main handler

## Installation:
```
go get github.com/Nigel2392/tcpproto
```

## Usage:
* The config struct is pretty basic, but you should not edit it manually. The config is predefined, and you can get it by calling `tcpproto.GetConfig()` or by using `SetConfig()`.
```go
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
// Predefined byte-sizes
const (
	DISABLED     = 0
	KILOBYTE     = 1024
	MEGABYTE     = 1024 * KILOBYTE
	GIGABYTE     = 1024 * MEGABYTE
	TEN_GIGABYTE = 10 * GIGABYTE
)
					       
conf := tcpproto.SetConfig(
	"SECRET_KEY", 		// Secret key for encryption
	"DEBUG",    		// Logger level
	2048,     			// Buffer size
	tcpproto.DISABLED, 	// Max content length
	true, 				// Include system info
	true,  				// Use crypto
	func(rq *Request, resp *Response) error {return nil} // Default authentication function.
)
```
Then we can get to start sending requests.
A typical response/request looks like this:
```go
CONTENT_LENGTH: CONTENT_LENGTH
COMMAND: COMMAND
CUSTOM_HEADER: CUSTOM_HEADER
CUSTOM_HEADER1: CUSTOM_HEADER1
CUSTOM_HEADER2: CUSTOM_HEADER2
CUSTOM_HEADER3: CUSTOM_HEADER3
HAS_FILE: HAS_FILE 
FILE_NAME: FILE_NAME 
FILE_SIZE: FILE_SIZE 
FILE_BOUNDARY: FILE_BOUNDARY 

FILE_BOUNDARY
FILE CONTENT
FILE_BOUNDARY

CONTENT CONTENT CONTENT
CONTENT CONTENT CONTENT
CONTENT CONTENT CONTENT
```
Where anything that has to do with files, can optionally be left out.  
To add a file, you can use the following:
```go
request.AddFile(filename string, content []byte, boundary string)
```
The following is needed:  
`CONTENT_LENGTH` and `COMMAND`  
Content length is used to make sure the whole request is parsed properly, and chunks we not forgotten.  
Command is used to add callbacks to the request/response cycle, where you can edit either one.  
## Server
A typical server looks like this:  
```go
func main() {
	ipaddr := "127.0.0.1"
	port := 22392
	// If CONF.Use_Crypto is disabled, you do not have to provide the private RSA key,
	// it can be left as an empty string.
	s := tcpproto.InitServer(ipaddr, port, "PRIVATE_KEY.pem")
	s.AddMiddlewareBeforeResp(AuthMiddleware) // Middleware to be called before the callback is called.
	s.AddMiddlewareAfterResp(tcpproto.LogMiddleware) // Middleware to be called after the callback is called.
	s.AddCallback("SET", SET) // The callback to be called, derived from "COMMAND" header.
	s.AddCallback("GET", GET)
	if err := s.Start(); err != nil {
		os.Exit(1)
	}
}
```
### Middleware
To add middleware, or callbacks, the function needs to take the following arguments:
```go
func MiddlewareOrCallback(rq *tcpproto.Request, resp *tcpproto.Response){
	// Do stuff here
}
```

In these middleware and callbacks you can ofcourse access all headers with request.Headers (Cookies are also stored here)
Or optionally, you can encrypt data with the SECRET_KEY provided to the CONFIG.
### Storing data client side
This data is then sent, like HTTP cookies, on every request.
To encrypt data, you can use the following:
```go
response.Lock(key string, data string)
```
To set some cookies in the response, you can use the following:
```go
response.Remember(key string, data string)
```
To remove the encrypted data from the client
```go
response.ForgetVault(key string)
```
To remove a cookie from the client
```go
response.Forget(key string)
```
When the client has sent this data, you could look at it like so:
```go
response.SetValues 	// Cookies
response.Vault		// Vault
response.Headers 	// Headers
response.Data		// Client side vault
```

### Client information
If you have enabled `Include_Sysinfo` in the CONFIG, you can access the following method in on the request:
```go
var system_information tcpproto.SysInfo
system_information = request.SysInfo()
```
You then have access to the following information:
```go
system_information.Hostname 	// The hostname of the client
system_information.Platform 	// The platform of the client
system_information.CPU 			// The CPU of the client
system_information.RAM 			// The RAM of the client
system_information.Disk 		// The disk of the client
system_information.MacAddr 		// The MAC address of the client
```

## Client:
A typical client looks like this:
```go
// Initialise request
request := InitRequest()

// Set some headers
// CONTENT_LENGTH is automatically added when generating the request.
request.Headers["COMMAND"] = "SET"
request.Headers["HEADER-TEST-1"] = "VALUE-TEST-1" 
request.Headers["HEADER-TEST-2"] = "VALUE-TEST-2" 
request.Headers["HEADER-TEST-3"] = "VALUE-TEST-3" 

// Set request content
request.Content = []byte("TEST_CONTENT\n")

// Initialize client
// If CONF.Use_Crypto is disabled, you do not have to provide the public RSA key,
// it can be left as an empty string.
client := InitClient("127.0.0.1", 12239, "PUBLIKKEY.pem")

// Connect to server
if err := client.Connect(); err != nil {
	err = errors.New("error connecting to server: " + err.Error())
	t.Error(err)
}

// Set some "cookies"
client.Cookies["TEST0"] = InitCookie("TEST0", "TEST0")

// Add something to the client side vault, this is encrypted with a public key, and decrypted by the server. 
// This only works if CONF.Use_Crypto is enabled.
client.Vault("Cookiename", "Cookievalue")

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
```go
// Response the server sends back
type Response struct {
	Headers   map[string]string
	SetValues map[string]string
	DelValues []string
	Vault     map[string]string
	Content   []byte
	File      *FileData
	Error     []error
}

// Request to send to the server
type Request struct {
	Headers            map[string]string
	Vault              map[string]string
	Content            []byte
	File               *FileData
	Data               map[string]string
	User               *User
	Conn               net.Conn
}
```
