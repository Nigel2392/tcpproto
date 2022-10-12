# tcpproto
Simple http-like layer ontop of TCP.

This layer is capable of transfering files, authentication, client-side storage and more!

A typical response/request looks like this:
```
CONTENT_LENGTH: CONTENT_LENGTH int
CLIENT_ID: CLIENT_ID int
MESSAGE_TYPE: MESSAGE_TYPE string
COMMAND: COMMAND string
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
The following is needed:
`CONTENT_LENGTH` and `COMMAND`
Content length is used to make sure the whole request is parsed properly, and chunks we not forgotten.
Command is used to add callbacks to the request/response cycle, where you can edit either one.
A typical server looks like this:
```
func main() {
	ipaddr := "127.0.0.1"
	port := 22392
	s := tcpproto.InitServer(ipaddr, port)
	s.AddMiddlewareAfterResp(tcpproto.LogMiddleware)
	s.AddMiddlewareBeforeResp(AuthMiddleware)
	s.AddCallback("SET", SET)
	s.AddCallback("GET", GET)
	if err := s.Start(); err != nil {
		os.Exit(1)
	}
}
```
