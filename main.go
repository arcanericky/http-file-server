package httpfileserver

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
)

// const (
// 	addrEnvVarName           = "ADDR"
// 	allowUploadsEnvVarName   = "UPLOADS"
// 	defaultAddr              = ":8080"
// 	portEnvVarName           = "PORT"
// 	quietEnvVarName          = "QUIET"
// 	rootRoute                = "/"
// 	sslCertificateEnvVarName = "SSL_CERTIFICATE"
// 	sslKeyEnvVarName         = "SSL_KEY"
// )

// var (
// 	addrFlag         = os.Getenv(addrEnvVarName)
// 	allowUploadsFlag = os.Getenv(allowUploadsEnvVarName) == "true"
// 	portFlag64, _    = strconv.ParseInt(os.Getenv(portEnvVarName), 10, 64)
// 	portFlag         = int(portFlag64)
// 	quietFlag        = os.Getenv(quietEnvVarName) == "true"
// 	routesFlag       Routes
// 	sslCertificate   = os.Getenv(sslCertificateEnvVarName)
// 	sslKey           = os.Getenv(sslKeyEnvVarName)
// )

type Config struct {
	AddrFlag         string
	AllowUploadsFlag bool
	DefaultAddr      string
	PortFlag64       int64
	PortFlag         int
	RootRoute        string
	SslCertificate   string
	SslKey           string
	QuietFlag        bool
	RoutesFlag       Routes
}

func Server(addr string, routes Routes, cfg Config) error {
	mux := http.DefaultServeMux
	handlers := make(map[string]http.Handler)
	paths := make(map[string]string)

	if len(routes.Values) == 0 {
		_ = routes.Set(".")
	}

	for _, route := range routes.Values {
		handlers[route.Route] = &fileHandler{
			route:       route.Route,
			path:        route.Path,
			allowUpload: cfg.AllowUploadsFlag,
		}
		paths[route.Route] = route.Path
	}

	for route, path := range paths {
		mux.Handle(route, handlers[route])
		log.Printf("serving local path %q on %q", path, route)
	}

	_, rootRouteTaken := handlers[cfg.RootRoute]
	if !rootRouteTaken {
		route := routes.Values[0].Route
		mux.Handle(cfg.RootRoute, http.RedirectHandler(route, http.StatusTemporaryRedirect))
		log.Printf("redirecting to %q from %q", route, cfg.RootRoute)
	}

	binaryPath, _ := os.Executable()
	if binaryPath == "" {
		binaryPath = "server"
	}
	if cfg.SslCertificate != "" && cfg.SslKey != "" {
		log.Printf("%s (HTTPS) listening on %q", filepath.Base(binaryPath), addr)
		return http.ListenAndServeTLS(addr, cfg.SslCertificate, cfg.SslKey, mux)
	}
	log.Printf("%s listening on %q", filepath.Base(binaryPath), addr)
	return http.ListenAndServe(addr, mux)
}

func Addr(cfg Config) (string, error) {
	portSet := cfg.PortFlag != 0
	addrSet := cfg.AddrFlag != ""
	switch {
	case portSet && addrSet:
		a, err := net.ResolveTCPAddr("tcp", cfg.AddrFlag)
		if err != nil {
			return "", err
		}
		a.Port = cfg.PortFlag
		return a.String(), nil
	case !portSet && addrSet:
		a, err := net.ResolveTCPAddr("tcp", cfg.AddrFlag)
		if err != nil {
			return "", err
		}
		return a.String(), nil
	case portSet && !addrSet:
		return fmt.Sprintf(":%d", cfg.PortFlag), nil
	case !portSet && !addrSet:
		fallthrough
	default:
		return cfg.DefaultAddr, nil
	}
}
