package server

import (
	"crypto/tls"
	"net/http"
	"time"

	"github.com/0xor1/wtf/pkg/log"
	"golang.org/x/crypto/acme/autocert"
)

type Config struct {
	Log                   log.Log
	UseHttps              bool
	AppBindTo             string
	CertBindTo            string
	HostWhitelist         []string
	CertReadTimeout       time.Duration
	CertReadHeaderTimeout time.Duration
	CertWriteTimeout      time.Duration
	CertCache             autocert.Cache
	Handler               http.HandlerFunc
}

func Run(configs ...func(c *Config)) error {
	c := config(configs...)

	if !c.UseHttps {
		c.Log.Info("Insecure app server running bound to %s", c.AppBindTo)
		appServer := appServer(c, c.Handler, nil)
		return appServer.ListenAndServe()
	}

	certManager := certManager(c)
	certServer := certServer(c, certManager)
	c.Log.Info("cert server running bound to %s", c.CertBindTo)
	go certServer.ListenAndServe()

	appServer := &http.Server{
		Addr:      c.AppBindTo,
		Handler:   c.Handler,
		TLSConfig: &tls.Config{GetCertificate: certManager.GetCertificate},
	}
	c.Log.Info("Secure app server running bound to %s", c.AppBindTo)
	return appServer.ListenAndServeTLS("", "")
}

func config(configs ...func(c *Config)) *Config {
	c := &Config{
		Log:                   log.New(),
		UseHttps:              false,
		AppBindTo:             ":8080",
		CertBindTo:            ":http",
		HostWhitelist:         nil,
		CertReadTimeout:       50 * time.Millisecond,
		CertReadHeaderTimeout: 50 * time.Millisecond,
		CertWriteTimeout:      50 * time.Millisecond,
		CertCache:             autocert.DirCache("acme_certs"),
		Handler: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotImplemented)
			w.Write([]byte(http.StatusText(http.StatusNotImplemented)))
		},
	}
	for _, config := range configs {
		config(c)
	}
	return c
}

func certManager(c *Config) *autocert.Manager {
	return &autocert.Manager{
		Cache:      c.CertCache,
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(c.HostWhitelist...),
	}
}

func certServer(c *Config, certManager *autocert.Manager) *http.Server {
	return &http.Server{
		Addr:              c.CertBindTo,
		Handler:           certManager.HTTPHandler(nil),
		ReadTimeout:       c.CertReadTimeout,
		ReadHeaderTimeout: c.CertReadHeaderTimeout,
		WriteTimeout:      c.CertWriteTimeout,
	}
}

func appServer(c *Config, handler http.HandlerFunc, tlsConfig *tls.Config) *http.Server {
	return &http.Server{
		Addr:      c.AppBindTo,
		Handler:   handler,
		TLSConfig: tlsConfig,
	}
}
