package logging

import "log"

type ConsoleLogger struct{}

func (c *ConsoleLogger) Log(l *HttpLog) {
	log.Printf("ACCESS: [%s | %s][%s | %d | %s][%d | %s | %d | %s] %s",
		l.RemoteAddr, l.Method,
		l.RequestContentType, l.RequestLength, l.RequestBody,
		l.ResponseCode, l.ResponseContentType, l.ResponseLength, l.Title,
		l.Url)
}
