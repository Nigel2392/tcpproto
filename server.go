package tcpproto

import (
	"crypto/rsa"
	"errors"
	"net"
	"strconv"
	"sync"
)

type Middleware struct {
	BeforeResponse func(rq *Request, resp *Response)
	AfterResponse  func(rq *Request, resp *Response)
	Before         bool
}

type Server struct {
	ln         net.Listener
	IP         string
	Config     *Config
	Port       int
	Callbacks  map[string]func(rq *Request, resp *Response)
	Middleware []*Middleware
	PRIVKEY    *rsa.PrivateKey
}

func InitServer(ip string, port int, privkey_file string) *Server {
	srv := &Server{
		IP:         ip,
		Port:       port,
		Callbacks:  make(map[string]func(rq *Request, resp *Response)),
		Middleware: []*Middleware{},
		PRIVKEY:    nil,
	}
	if CONF.Use_Crypto && privkey_file != "" {
		srv.PRIVKEY = ImportPrivate_PEM_Key(privkey_file)
	}
	return srv
}

func (s *Server) Addr() string {
	str_port := strconv.Itoa(s.Port)
	return s.IP + ":" + str_port
}

func (s *Server) Start() error {
	var err error
	s.ln, err = net.Listen("tcp", s.Addr())
	if err != nil {
		return err
	}
	err = s.Serve()
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) Serve() error {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			return err
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	wg := &sync.WaitGroup{}
	for {
		wg.Add(1)
		// Parse the request
		go func(wg *sync.WaitGroup) {
			defer wg.Done()
			rq, resp, err := s.ParseConnection(conn)
			if err != nil {
				// LOGGER.Error(err)
				return
			}

			// Execute authentication
			err = CONF.Default_Auth(rq, resp)
			if err != nil {
				// LOGGER.Error(err)
				return
			}

			// Handle middleware before response
			s.MiddlewareBeforeResponse(rq, resp)

			// Handle the request
			s.ExecCallback(rq, resp)

			// Handle middleware after response
			s.MiddlewareAfterResponse(rq, resp)

			// LOGGER.Debug("Sending response")
			err = s.Send(conn, resp)
			if err != nil {
				return
			}
		}(wg)
		wg.Wait()
	}
}

// Middleware to be used before the response is created
func (s *Server) AddMiddlewareBeforeResp(middleware func(rq *Request, resp *Response)) {
	Middleware := &Middleware{
		BeforeResponse: middleware,
	}
	s.Middleware = append(s.Middleware, Middleware)
}

// Middleware to be used after the response is created
func (s *Server) AddMiddlewareAfterResp(middleware func(rq *Request, resp *Response)) {
	Middleware := &Middleware{
		AfterResponse: middleware,
	}
	s.Middleware = append(s.Middleware, Middleware)
}

// Add a COMMAND callback to the server
func (s *Server) AddCallback(key string, callback func(rq *Request, resp *Response)) {
	s.Callbacks[key] = callback
}

// Execute middleware before the response is created
func (s *Server) MiddlewareBeforeResponse(rq *Request, resp *Response) {
	for _, middleware := range s.Middleware {
		if middleware.BeforeResponse != nil {
			middleware.BeforeResponse(rq, resp)
		}
	}
}

// Execute middleware after the response is created
func (s *Server) MiddlewareAfterResponse(rq *Request, resp *Response) {
	for _, middleware := range s.Middleware {
		if middleware.AfterResponse != nil {
			middleware.AfterResponse(rq, resp)
		}
	}
}

// Execute the callback for the given request
func (s *Server) ExecCallback(rq *Request, resp *Response) error {
	callback, ok := s.Callbacks[rq.Headers["COMMAND"]]
	if ok {
		callback(rq, resp)
	} else {
		return errors.New("no callback for command: " + rq.Headers["COMMAND"])
	}
	return nil
}

func (s *Server) Send(conn net.Conn, resp *Response) error {
	if resp.Error != nil {
		if len(resp.Error) > 0 {
			resp = InitResponse()
			resp.Headers["ERROR"] = "Internal Server Error"
			err_resp := ""
			for _, err := range resp.Error {
				err_resp += err.Error() + "\n"
			}
			resp.Content = []byte(err_resp)
		}
	}
	_, err := conn.Write(resp.Bytes())
	if err != nil {
		err = errors.New("error sending response: " + err.Error())
		CONF.LOGGER.Error(err.Error())
		return err
	}
	return nil
}
