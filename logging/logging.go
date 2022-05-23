package logging

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/google/martian/v3"
	"github.com/google/martian/v3/messageview"
)

// Logger is a modifier that logs requests and responses.
type Logger struct {
}

// NewLogger returns a logger that logs requests and responses, optionally
// logging the body. Log function defaults to martian.Infof.
func NewLogger() *Logger {
	return &Logger{}
}

// ModifyRequest simply put all the request header and body into the context for later use
func (l *Logger) ModifyRequest(req *http.Request) error {
	ctx := martian.NewContext(req)

	b := &bytes.Buffer{}

	mv := messageview.New()
	skipBody := true
	ct := req.Header.Get("Content-Type")
	if strings.HasPrefix(ct, "text") || strings.HasSuffix(ct, "json") {
		skipBody = false
	}
	mv.SkipBody(skipBody)
	if err := mv.SnapshotRequest(req); err != nil {
		return err
	}

	var opts []messageview.Option
	opts = append(opts, messageview.Decode())
	r, err := mv.Reader(opts...)
	if err != nil {
		return err
	}

	io.Copy(b, r)
	ctx.Set("req", b)
	return nil
}

func (l *Logger) logRequest(req *http.Request, b *bytes.Buffer) {
	ctx := martian.NewContext(req)
	if orig, ok := ctx.Get("req"); ok {
		o := orig.(*bytes.Buffer)
		io.Copy(b, o)
	}
	fmt.Fprint(b, "\n")
}

// ModifyResponse logs the response, optionally including the body.
func (l *Logger) ModifyResponse(res *http.Response) error {
	ctx := martian.NewContext(res.Request)
	if ctx.SkippingLogging() {
		return nil
	}

	b := &bytes.Buffer{}
	fmt.Fprintln(b, "")
	fmt.Fprintln(b, strings.Repeat("-", 80))
	fmt.Fprintf(b, "%s\n", res.Request.URL)
	fmt.Fprintln(b, strings.Repeat("-", 80))

	l.logRequest(res.Request, b)

	mv := messageview.New()
	skipBody := true
	ct := res.Header.Get("Content-Type")
	if strings.HasPrefix(ct, "text") || strings.HasSuffix(ct, "json") {
		skipBody = false
	}
	mv.SkipBody(skipBody)
	if err := mv.SnapshotResponse(res); err != nil {
		return err
	}

	var opts []messageview.Option
	opts = append(opts, messageview.Decode())

	r, err := mv.Reader(opts...)
	if err != nil {
		return err
	}

	io.Copy(b, r)

	fmt.Fprintln(b, "")
	fmt.Fprintln(b, strings.Repeat("-", 80))

	//l.log(b.String())
	req := res.Request
	log.Printf("%s %s %s => %s", req.RemoteAddr, req.Method, req.URL, res.Status)

	return nil
}
