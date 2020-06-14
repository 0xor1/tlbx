package server

import (
	"context"
	"crypto/tls"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/log"
	"github.com/0xor1/tlbx/pkg/web/server/autocertcache"
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

func Run(configs ...func(c *Config)) {
	c := config(configs...)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	shutdownServers := func(servers ...*http.Server) func() {
		return func() {
			<-quit
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			for _, server := range servers {
				server.SetKeepAlivesEnabled(false)
				c.Log.Info("Server %s shutting down", server.Addr)
				c.Log.ErrorOn(server.Shutdown(ctx))
			}
		}
	}
	logDoneError := func(err error) {
		if err != nil && err.Error() != "http: Server closed" {
			c.Log.ErrorOn(err)
		}
	}

	if !c.UseHttps {
		c.Log.Info("Insecure app server running bound to %s", c.AppBindTo)
		appServer := appServer(c, c.Handler, nil)
		Go(shutdownServers(appServer), c.Log.ErrorOn)
		logDoneError(appServer.ListenAndServe())
	} else {
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
		Go(shutdownServers(appServer, certServer), c.Log.ErrorOn)
		logDoneError(appServer.ListenAndServeTLS("", ""))
	}
	c.Log.Info("Server stopped")
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
		CertCache:             autocertcache.Dir("acme_certs"),
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
