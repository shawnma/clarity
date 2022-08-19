package logging

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/martian/v3"
	"github.com/google/martian/v3/messageview"
	"shawnma.com/clarity/config"
)

type HttpLog struct {
	User       string
	RemoteAddr string
	Method     string
	Url        string

	RequestContentType string
	RequestLength      int
	RequestBody        string

	ResponseCode        int
	ResponseContentType string
	ResponseLength      int
	ResponseBody        string
	Title               string
}

func (l *HttpLog) String() string {
	return fmt.Sprintf("[%s | %s][%s | %d | %s][%d | %s | %d | %s | %s] %s",
		l.RemoteAddr, l.Method,
		l.RequestContentType, l.RequestLength, l.RequestBody,
		l.ResponseCode, l.ResponseContentType, l.ResponseLength, l.ResponseBody, l.Title,
		l.Url)
}

var titleExp = regexp.MustCompile(`(?i)<title>([^<>]*)</title>`)

type AccessLogger interface {
	Log(httpLog *HttpLog)
}

// Logger is a modifier that logs requests and responses.
type Logger struct {
	log AccessLogger
}

// NewLogger returns a logger that logs requests and responses, optionally
// logging the body. Log function defaults to martian.Infof.
func NewLogger(c *config.Config) *Logger {
	l, e := NewAccessLogger(c)
	if e != nil {
		log.Fatalf("Unable to create access logger: %s", e)
	}
	return &Logger{l}
}

// ModifyRequest simply put all the request header and body into the context for later use
func (l *Logger) ModifyRequest(req *http.Request) error {
	var httpLog HttpLog

	ct := sanitizeContentType(req.Header.Get("Content-Type"))
	httpLog.RequestContentType = ct
	length := req.Header.Get("Content-Length")
	if length != "" {
		if l, e := strconv.Atoi(length); e == nil {
			httpLog.RequestLength = l
		}
	}
	httpLog.Url = req.URL.String()
	if len(httpLog.Url) > 1000 {
		httpLog.Url = httpLog.Url[:1000]
	}
	httpLog.Method = req.Method
	httpLog.RemoteAddr = req.RemoteAddr

	skipBody := true
	if strings.HasPrefix(ct, "text") || strings.HasSuffix(ct, "json") || strings.HasSuffix(ct, "x-www-form-urlencoded") {
		skipBody = false
	}
	mv := messageview.New()
	mv.SkipBody(skipBody)
	if err := mv.SnapshotRequest(req); err != nil {
		return err
	}

	if !skipBody {
		b, err := l.readBody(mv)
		if err != nil {
			return err
		}
		httpLog.RequestBody = b
	}

	ctx := martian.NewContext(req)
	ctx.Set("log", &httpLog)
	return nil
}

func (*Logger) readBody(mv *messageview.MessageView) (string, error) {
	opts := []messageview.Option{messageview.Decode()}
	r, err := mv.BodyReader(opts...)
	if err != nil {
		return "", err
	}
	b, e := io.ReadAll(r)
	if e != nil {
		return "", e
	}
	return string(b), nil
}

// ModifyResponse logs the response, optionally including the body.
func (l *Logger) ModifyResponse(res *http.Response) error {
	ctx := martian.NewContext(res.Request)
	if ctx.SkippingLogging() {
		return nil
	}

	//l.logRequest(res.Request, b)
	httpLog, ok := ctx.Get("log")
	if !ok {
		return fmt.Errorf("unable to find log object in request for %s", res.Request.URL)
	}
	h := httpLog.(*HttpLog)

	ct := sanitizeContentType(res.Header.Get("Content-Type"))
	h.ResponseCode = res.StatusCode
	h.ResponseContentType = ct
	length := res.Header.Get("Content-Length")
	if length != "" {
		if l, e := strconv.Atoi(length); e == nil {
			h.ResponseLength = l
		} else {
			log.Printf("unable to parse length %s: %s\n", length, e)
		}
	}

	skipBody := true
	if strings.HasPrefix(ct, "text") || strings.HasSuffix(ct, "json") {
		skipBody = false
	}
	mv := messageview.New()
	mv.SkipBody(skipBody)
	if err := mv.SnapshotResponse(res); err != nil {
		return err
	}
	if !skipBody {
		b, err := l.readBody(mv)
		if err != nil {
			return err
		}
		r := []rune(b)
		if len(r) > 1000 {
			r = r[:1000]
		}
		b = string(r)
		match := titleExp.FindStringSubmatch(b)
		if len(match) > 1 {
			h.Title = match[1]
		}
	}
	l.log.Log(h)
	return nil
}

func sanitizeContentType(ct string) string {
	if strings.Contains(ct, ";") {
		return strings.Split(ct, ";")[0]
	}
	return ct
}
