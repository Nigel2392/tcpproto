import base64
import datetime
import socket
from cryptography.hazmat.backends import default_backend
from cryptography.hazmat.primitives import (
    hashes,
    serialization 
)
from cryptography.hazmat.primitives.asymmetric import padding
import os
import traceback

class Logger:
    """Basic logging with colors and levels"""
    reset  = "\033[0m"
    error    = "\033[31m"
    debug  = "\033[32m"
    info   = "\033[34m"
    test = "\033[35m"

    def __init__(self, loglevel=0):
        """
        loglevel: 0 = test, 1 = debug, 2 = info, 3 = warning, 4 = error
        """
        if isinstance(loglevel, int):
            self.loglevel = loglevel
        else:
            if isinstance(loglevel, str):
                self.getlevelfromstr(loglevel)

    def getlevelfromstr(self, loglevel: str):
        """Converts a string to a loglevel"""
        if isinstance(loglevel, str):
            match loglevel.lower():
                case "error":
                    self.loglevel = 4
                case "warning":
                    self.loglevel = 3
                case "info":
                    self.loglevel = 2
                case "debug":
                    self.loglevel = 1
                case "test":
                    self.loglevel = 0
                case _:
                    self.loglevel = 0
        else:
            raise TypeError("loglevel must be a string to call getlevelfromstr")

    def craft(self, msg, loglevel: str, prefix: str=""):
        nowtime = datetime.datetime.now().strftime("%Y-%m-%d %H:%M:%S")
        if not hasattr(self, loglevel.lower()):
            raise AttributeError("Invalid loglevel")
        return f"{getattr(self, loglevel.lower())}{prefix}[{nowtime} {loglevel.upper()}] {msg}{self.reset}"

    def log(self, msg, color, level=0):
        if level >= self.loglevel:
            print(self.craft(msg, color))

    def Except(self, func, exceptions=[Exception], *args, **kwargs):
        """
        Used for logging exceptions and returning a value
        Either returns the value of the function or the value of the exception

        Raises exception if the exception is not in the exceptions list.
        """
        excepted: Exception = None
        try:
            retval = func(*args, **kwargs)
        except Exception as e:
            flag = False
            for exception in exceptions:
                if isinstance(e, exception):
                    flag = True
                    excepted = e
                    break
            if not flag:
                raise e
        else:
            return retval
        finally:
            msg = self.craft(
                f"""Function raised an exception:\n"""
                f"""  Function: {func.__name__}\n"""
                f"""  Exception: {excepted.__class__.__name__}\n"""
                f"""  Args: {excepted.args}\n"""
                f"""  Traceback:    V-V-V-V-V-V-V\n\n""" +
                "\n".join(traceback.format_tb(excepted.__traceback__)) + 
                "-" * 40, 
                "error", 
                prefix="-" * 40 + "\n"
            )
            print(msg)
            return excepted

    def Error(self, msg, level=4):
        """Used for logging error messages"""
        self.log(msg, "error", level)

    def Warning(self, msg, level=3):
        """Used for logging warning messages, does not nescicarily mean exception"""
        self.log(msg, "warning", level)

    def Info(self, msg, level=2):
        '''Used for logging generic info messages'''
        self.log(msg, "info", level)

    def Debug(self, msg, level=1):
        """Used for logging debug messages"""
        self.log(msg, "debug", level)

    def Test(self, msg, level=0):
        """Used for logging test messages"""
        self.log(msg, "test", level)


def ParseHeader(data: bytes):
    """
    Function for parsing headers
    """
    header_dict = {}
    # Split the last \r\n\r\n
    data_list = data.split(b'\r\n\r\n', 1)
    
    if len(data_list) != 2:
        raise Exception("Invalid header")
    header, content = data_list
    for line in header.split(b"\r\n"):
        
        line_list = line.split(b":")
        key = line_list[0].decode()
        value_list = []
        for item in line_list[1:]:
            value_list.append(item.decode())

        value = line_list[1:]
        value = ":".join(value_list)
        header_dict[key] = value
    return header_dict, content

class File:
    """
    File to add to the data
    """
    def __init__(self, filename: str=None, data: bytes=None, border: str=None):
        self.filename   = filename
        self.border     = border
        self.data       = data
        if self.filename and self.border and self.data:
            self.has_file = True

    def __dict__(self) -> dict:
        """
        File as dictionary
        """
        return {
            "filename": self.filename,
            "border": self.border,
            "data": self.data,
            "has_file": self.has_file
        }
    
    def Generate(self) -> bytes:
        """
        File content -> prepend to the request content
        """
        return self.starting_border + self.data + self.ending_border

    def Size(self) -> int:
        """
        File size
        """
        return len(self.data)

    @property
    def starting_border(self) -> bytes:
        """
        File starting border
        """
        return b"--" + self.border.encode() + b"--"
    
    @property
    def ending_border(self) -> bytes:
        """
        File ending border
        """
        return b"----" + self.border.encode() + b"----"

class BaseRqResp:
    headers:    dict    = {}
    content:    bytes   = b""
    file:       File    = None
    cookies:    dict    = {}
    command:    str     = ""
    vault:      dict     = {}

    def __dict__(self) -> dict:
        """
        data as dictionary
        """
        return {
            "command": self.command,
            "headers": self.headers,
            "content": self.content,
            "file": dict(self.file),
            "cookies": self.cookies,
            "vault": self.vault
        }

    def __init__(self, **kwargs):
        """
        Initialize the data
        """
        for key, value in kwargs.items():
            if hasattr(self, key.lower()):
                setattr(self, key.lower(), value)

    @property
    def ContentLength(self) -> int:
        """
        data content length
        """
        if self.file:
            if self.file.has_file:
                return len(self.content) + len(self.file.Generate())
        return len(self.content)

    def Generate(self) -> bytes:
        """
        Generate the data
        """
        if self.file:
            if self.file.has_file:
                return self.GenerateHeaders() + self.file.Generate() + self.content
        return self.GenerateHeaders() + self.content

    def GenerateHeaders(self) -> bytes:
        """
        Generate the data headers
        """
        headers = ""
        headers += f"CONTENT_LENGTH:{self.ContentLength}\r\n"
        headers += f"COMMAND:{self.command}\r\n"

        if self.file:
            if self.file.has_file:
                headers += "FILE_NAME:" + self.file.filename + "\r\n"
                headers += "FILE_SIZE:" + str(self.file.Size()) + "\r\n"
                headers += "FILE_BOUNDARY:" + self.file.border + "\r\n"
                headers += "HAS_FILE:" + "true" + "\r\n"

        for key, value in self.headers.items():
            headers += f"{key}: {value}\r\n"

        for key, value in self.cookies.items():
            headers += f"REMEMBER-{key}:{value}\r\n"

        for key, value in self.vault.items():
            headers += f"VAULT-{key}:{value}\r\n"

        headers += "\r\n"
        return headers.encode()

class Response(BaseRqResp):
    """
    Response class
    """
    pass

class Request(BaseRqResp):
    """
    Request class
    """
    pass

class Client():
    cookies: dict = {}
    vault: dict = {}
    client_vault: dict = {}

    def __init__(self, host: str, port: int, rsa_file: str="PUBKEY.pem", buffer_size: int=2048):
        """
        Initialize the client

        Private key is not required. 
        If it is not provided, you will not be able to encrypt (client.Lock()) data client-side.
        """
        self.host = host
        self.port = port
        self.sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self.sock.connect((self.host, self.port))
        self.buffer_size = buffer_size
        try:
            os.chdir(os.path.dirname(os.path.abspath(__file__)))
            with open(".\\"+ rsa_file, "rb") as f:
                fdata = f.read()
                self.rsa_key = serialization.load_pem_public_key(fdata, backend=default_backend())
        except:
            pass

    def __dict__(self) -> dict:
        """
        Client as dictionary
        """
        return {
            "host": self.host,
            "port": self.port,
            "cookies": self.cookies,
            "vault": self.vault
        }

    def Send(self, request: Request) -> Response:
        """
        Send a request to the server. 
        The server will return a response.
        """
        request.vault = self.vault
        request.cookies = self.cookies
        if hasattr(self, "rsa_key"):
            for key, value in self.client_vault.items():
                value = self.rsa_key.encrypt(value.encode(), padding.OAEP(mgf=padding.MGF1(algorithm=hashes.SHA512()), algorithm=hashes.SHA512(), label=None))
                value = base64.b64encode(value)
                request.headers["CLIENT_VAULT-"+key] = value.decode()
        self.client_vault = {}
        self.sock.send(request.Generate())
        resp = self.Receive()
        return resp

    def Close(self):
        """
        Close the connection
        """
        self.sock.close()

    def Receive(self) -> Response:
        """
        Receive a response from the server
        """
        headers, content = self.rcv_header()
        content_length = int(headers["CONTENT_LENGTH"])
        content += self.rcv_content(content_length, content)
        # Initialize the request (Will be used as response)
        resp = Response()
        resp.headers = headers
        resp.content = content
        resp.cookies = self.cookies
        resp.vault = self.vault
        return resp


    def rcv_header(self, data: bytes=b""):
        """
        :param data: The data that has been received so far
        :return: [dict, bytes] Headers, content 
        """
        chunk = self.sock.recv(self.buffer_size)
        data += chunk
        if b"\r\n\r\n" in data:
            headers, content = ParseHeader(data)
            keys = list(headers.keys())
            for key in keys:
                if key.startswith("REMEMBER-"):
                    self.cookies[key[9:]] = headers[key]
                    del headers[key]
                elif key.startswith("VAULT-"):
                    self.vault[key[6:]] = headers[key]
                    del headers[key]
                elif key.startswith("CLIENT_VAULT-"):
                    self.client_vault[key[13:]] = headers[key]
                    del headers[key]
                elif key.startswith("FORGET-"):
                    del self.cookies[headers[key]]
                    del self.vault[headers[key]]
                    del headers[key]
            return headers, content
        else:
            return self.rcv_header(data)

    def rcv_content(self, content_length: int, data: bytes=b"") -> bytes:
        while len(data) < content_length:
            data += self.sock.recv(self.buffer_size)
        return data[content_length:]

    def Lock(self, key, value):
        self.client_vault[key] = value


if __name__ == "__main__":
    file = File(filename="test.txt", data=b"Hello World!", border="FILE_BOUDNAKSFDJBADFS")
    request = Request(file=file, content=b"sdfsdf Worfsdfsdfsdfld!", command="SET")
    request.headers["Content-Type"] = "text/plain"
    client = Client("127.0.0.1", 22392, "PUBKEY.pem", 4096)
    resp = client.Send(request)
    logger = Logger("debug")
    logger.Test("Response 1:")
    print(resp.headers)
    print(resp.cookies)
    print(resp.vault)
    print(resp.content)
    request = Request(content=b"Hello world!", command="GET")
    response = client.Send(request)
    logger.Test("Response 2:")
    print(response.headers)
    print(response.cookies)
    print(response.vault)
    print(response.content)
    def WillExcept():
        dict = {}
        dict["test"] = "test"
        key = dict["test2"]

    logger.Except(WillExcept, [Exception])



