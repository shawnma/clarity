package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/google/martian/v3"
	mapi "github.com/google/martian/v3/api"
	"github.com/google/martian/v3/cors"
	"github.com/google/martian/v3/fifo"
	"github.com/google/martian/v3/header"
	mlog "github.com/google/martian/v3/log"
	"github.com/google/martian/v3/martianhttp"
	"github.com/google/martian/v3/mitm"
	"github.com/google/martian/v3/servemux"
	"shawnma.com/clarity/config"
	"shawnma.com/clarity/filter"
	"shawnma.com/clarity/logging"
)

var (
	addr          = flag.String("addr", ":8080", "host:port of the proxy")
	apiAddr       = flag.String("api-addr", ":8181", "host:port of the configuration API")
	tlsAddr       = flag.String("tls-addr", ":4443", "host:port of the transparent proxy over TLS")
	apiHost       = flag.String("api", "clarity.proxy", "hostname for the API")
	cert          = flag.String("cert", "", "filepath to the CA certificate used to sign MITM certificates")
	key           = flag.String("key", "", "filepath to the private key of the CA used to sign MITM certificates")
	organization  = flag.String("organization", "Clarity Proxy", "organization name for MITM certificates")
	validity      = flag.Duration("validity", time.Hour, "window of time that MITM certificates are valid")
	allowCORS     = flag.Bool("cors", false, "allow CORS requests to configure the proxy")
	skipTLSVerify = flag.Bool("skip-tls-verify", false, "skip TLS server verification; insecure")
	level         = flag.Int("v", 0, "log level")
)

func main() {
	flag.Parse()
	mlog.SetLevel(*level)
	config := config.NewConfig()

	p := martian.NewProxy()
	defer p.Close()

	l, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatal(err)
	}

	lAPI, err := net.Listen("tcp", *apiAddr)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("martian: starting proxy on %s and api on %s", l.Addr().String(), lAPI.Addr().String())

	mux := http.NewServeMux()
	startMitm(p, mux)

	stack := newStack(config)
	filter := filter.NewFilter(config)
	stack.AddRequestModifier(filter)
	configure("/config", filter.HttpHandler(), mux)

	// static content serving
	fs := http.StripPrefix("/filter", http.FileServer(http.Dir("./public/")))
	configure("/filter/", fs, mux)

	// Redirect API traffic to API server.
	if *apiAddr != "" {
		addrParts := strings.Split(lAPI.Addr().String(), ":")
		apip := addrParts[len(addrParts)-1]
		port, err := strconv.Atoi(apip)
		if err != nil {
			log.Fatal(err)
		}
		host := strings.Join(addrParts[:len(addrParts)-1], ":")

		// Forward traffic that pattern matches in http.DefaultServeMux
		log.Printf("Setting up API fwd at %s:%d", host, port)
		apif := servemux.NewFilter(mux)
		apif.SetRequestModifier(mapi.NewForwarder(host, port))
		stack.AddRequestModifier(apif)
	}

	p.SetRequestModifier(stack)
	p.SetResponseModifier(stack)

	go p.Serve(l)
	go http.Serve(lAPI, mux)

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt)

	<-sigc

	log.Println("martian: shutting down")
	os.Exit(0)
}

func startMitm(p *martian.Proxy, mux *http.ServeMux) {
	tr := &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: *skipTLSVerify,
		},
	}
	p.SetRoundTripper(tr)

	var x509c *x509.Certificate
	var priv interface{}

	if *cert == "" || *key == "" {
		log.Printf("Generating a new ROOT CA for testing purpose. To use persist CA, generate one using openssl and supply -cert and -key")
		var err error
		x509c, priv, err = mitm.NewAuthority("martian.proxy", "Martian Authority", 30*24*time.Hour)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		tlsc, err := tls.LoadX509KeyPair(*cert, *key)
		if err != nil {
			log.Fatal(err)
		}
		priv = tlsc.PrivateKey

		x509c, err = x509.ParseCertificate(tlsc.Certificate[0])
		if err != nil {
			log.Fatal(err)
		}
	}

	mc, err := mitm.NewConfig(x509c, priv)
	if err != nil {
		log.Fatal(err)
	}

	mc.SetValidity(*validity)
	mc.SetOrganization(*organization)
	mc.SkipTLSVerify(*skipTLSVerify)

	p.SetMITM(mc)

	ah := martianhttp.NewAuthorityHandler(x509c)
	configure("/authority.cer", ah, mux)

	tl, err := net.Listen("tcp", *tlsAddr)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Starting transparent TLS listener on %s\n", tl.Addr())

	go p.Serve(tls.NewListener(tl, mc.TLS()))
}

// configure installs a configuration handler at path.
func configure(pattern string, handler http.Handler, mux *http.ServeMux) {
	if *allowCORS {
		handler = cors.NewHandler(handler)
	}

	// register handler for martian.proxy to be forwarded to
	// local API server
	mux.Handle(path.Join(*apiHost, pattern), handler)

	// register handler for local API server
	mux.Handle(pattern, handler)
}

func newStack(c *config.Config) (grp *fifo.Group) {
	grp = fifo.NewGroup()
	logger := logging.NewLogger(c)
	grp.AddRequestModifier(logger) // required to save a copy of the request
	grp.AddResponseModifier(logger)
	grp.AddRequestModifier(header.NewBadFramingModifier())
	return grp
}
