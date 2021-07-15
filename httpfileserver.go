package httpfileserver

import (
	"context"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/sgreben/httpfileserver/internal/filehandler"
	"github.com/sgreben/httpfileserver/internal/routes"
)

type Config struct {
	Addr             string
	AllowUploadsFlag bool
	RootRoute        string
	SslCertificate   string
	SslKey           string
	Routes           routes.Routes
}

type routeEntry interface {
	http.Handler
	GetRoute() string
	GetPath() string
}

var listenAndServe = http.ListenAndServe
var listenAndServeTLS = http.ListenAndServeTLS

func NewConfig() Config {
	return Config{
		Addr:             ":8080",
		AllowUploadsFlag: false,
		RootRoute:        "/",
		SslCertificate:   "",
		SslKey:           "",
	}
}

func getExeName() string {
	binaryPath, _ := os.Executable()
	if binaryPath == "" {
		binaryPath = "http-file-server"
	}

	exeName := filepath.Base(binaryPath)

	return exeName
}

func loadRouteHandlers(cfg *Config) map[string]routeEntry {
	handlers := make(map[string]routeEntry)

	if len(cfg.Routes.Values) == 0 {
		_ = cfg.Routes.Set(".")
	}

	for _, route := range cfg.Routes.Values {
		handlers[route.Route] = filehandler.NewFileHandler(
			route.Route,
			route.Path,
			cfg.AllowUploadsFlag,
		)
	}

	return handlers
}

func addMuxRoutes(mux *http.ServeMux, handlers map[string]routeEntry) {
	for _, p := range handlers {
		mux.Handle(p.GetRoute(), p)
		log.Printf("serving local path %q on %q", p.GetPath(), p.GetRoute())
	}
}

func redirectRootRoute(cfg Config, mux *http.ServeMux, handlers map[string]routeEntry) {
	if len(cfg.Routes.Values) == 0 {
		log.Print("no routes registered")
		return
	}

	_, rootRouteTaken := handlers[cfg.RootRoute]
	if !rootRouteTaken {
		route := cfg.Routes.Values[0].Route
		mux.Handle(cfg.RootRoute, http.RedirectHandler(route, http.StatusTemporaryRedirect))
		log.Printf("redirecting to %q from %q", route, cfg.RootRoute)
	}
}

func getMux(cfg Config) *http.ServeMux {
	handlers := loadRouteHandlers(&cfg)
	mux := http.NewServeMux()
	addMuxRoutes(mux, handlers)
	redirectRootRoute(cfg, mux, handlers)
	return mux
}

func Serve(ctx context.Context, cfg Config) error {
	mux := getMux(cfg)
	exeName := getExeName()

	if cfg.SslCertificate != "" && cfg.SslKey != "" {
		log.Printf("%s (HTTPS) listening on %q", exeName, cfg.Addr)
		return listenAndServeTLS(cfg.Addr, cfg.SslCertificate, cfg.SslKey, mux)
	}
	log.Printf("%s listening on %q", exeName, cfg.Addr)
	return listenAndServe(cfg.Addr, mux)
}
