package main

func LogMiddleware(rq *Request, resp *Response) {
	if rq.File.Present {
		LOGGER.Info(rq.Headers["CLIENT_ID"] + " " + rq.Headers["COMMAND"] + " sent file: " + rq.File.Name)
	} else {
		LOGGER.Info(rq.Headers["CLIENT_ID"] + " " + rq.Headers["COMMAND"])
	}
}

//func AuthMiddleware(rq *Request, resp *Response) {
//	_, ok := rq.Headers["AUTH_SESSION"]
//	if ok {
//		// Check if session is valid
//		// If valid, set user to request
//		// If not valid, set  IsAuthenticated to false
//
//	} else {
//		// Set user IsAuthenticated to false
//	}
//
//}
//
