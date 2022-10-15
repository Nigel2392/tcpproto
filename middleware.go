package tcpproto

func LogMiddleware(rq *Request, resp *Response) {
	if rq.File.Present {
		CONF.LOGGER.Info(rq.Headers["CLIENT_ID"] + " " + rq.Headers["COMMAND"] + " sent file: " + rq.File.Name)
	} else {
		CONF.LOGGER.Info(rq.Headers["CLIENT_ID"] + " " + rq.Headers["COMMAND"])
	}
}
