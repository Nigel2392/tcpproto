import socket


class File:
    def __init__(self, filename: str=None, data: bytes=None, border: str=None):
        self.filename   = filename
        self.border     = border
        self.data       = data
        if self.filename and self.border and self.data:
            self.has_file = True
    
    def Generate(self):
        return self.starting_border + self.data + self.ending_border

    def Size(self):
        return len(self.Generate())
                
    @property
    def starting_border(self):
        return b"--" + self.border.encode() + b"--"
    
    @property
    def ending_border(self):
        return b"----" + self.border.encode() + b"----"

class Request:
    headers:    dict    = {}
    content:    bytes   = b""
    file:       File    = None
    cookies:    dict    = {}
    COMMAND:    str     = "SET"

    def __init__(self, **kwargs):
        for key, value in kwargs.items():
            if hasattr(self, key):
                setattr(self, key, value)

    @property
    def ContentLength(self):
        return len(self.content) + self.file.Size()

    def Generate(self):
        return self.GenerateHeaders() + self.file.Generate() + self.content

    def GenerateHeaders(self):
        headers = ""
        headers += f"CONTENT_LENGTH:{self.ContentLength}\r\n"
        headers += f"COMMAND:{self.COMMAND}\r\n"

        if self.file.has_file:
            headers += "FILE_NAME:" + self.file.filename + "\r\n"
            headers += "FILE_SIZE:" + str(self.file.Size()) + "\r\n"
            headers += "FILE_BOUNDARY:" + self.file.border + "\r\n"
            headers += "HAS_FILE:" + "true" + "\r\n"

        for key, value in self.headers.items():
            headers += f"{key}: {value}\r\n"

        for key, value in self.cookies.items():
            headers += f"REMEMBER-{key}:{value}\r\n"
        headers += "\r\n"
        return headers.encode()

class Client():
    def __init__(self, host: str, port: int, buffer_size: int=2048):
        self.host = host
        self.port = port
        self.sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self.sock.connect((self.host, self.port))
        self.buffer_size = buffer_size

    def Send(self, request: Request):
        self.sock.send(request.Generate())
        return self.Receive()

    def Close(self):
        self.sock.close()

    def Receive(self):
        data = b""
        chunk = self.sock.recv(self.buffer_size)
        data += chunk
        while b"\r\n\r\n" not in chunk:
            chunk = self.sock.recv(self.buffer_size)
        data += chunk
        return data

if __name__ == "__main__":
    file = File(filename="test.txt", data=b"Hello World!", border="test")
    request = Request(headers={"Content-Type": "multipart/form-data; boundary=" + file.border}, file=file, content=b"Hello World!")
    print(request.Generate())