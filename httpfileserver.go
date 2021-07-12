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

func loadRouteHandlers(cfg *Config) (map[string]http.Handler, map[string]string) {
	handlers := make(map[string]http.Handler)
	paths := make(map[string]string)

	if len(cfg.Routes.Values) == 0 {
		log.Print("SETTING DOT ROUTE")
		_ = cfg.Routes.Set(".")
	}

	for _, route := range cfg.Routes.Values {
		handlers[route.Route] = filehandler.NewFileHandler(
			route.Route,
			route.Path,
			cfg.AllowUploadsFlag,
		)
		paths[route.Route] = route.Path
	}

	return handlers, paths
}

func addMuxRoutes(mux *http.ServeMux, paths map[string]string, handlers map[string]http.Handler) {
	for route, path := range paths {
		mux.Handle(route, handlers[route])
		log.Printf("serving local path %q on %q", path, route)
	}
}

func redirectRootRoute(cfg Config, mux *http.ServeMux, handlers map[string]http.Handler) {
	_, rootRouteTaken := handlers[cfg.RootRoute]
	if !rootRouteTaken {
		route := cfg.Routes.Values[0].Route
		mux.Handle(cfg.RootRoute, http.RedirectHandler(route, http.StatusTemporaryRedirect))
		log.Printf("redirecting to %q from %q", route, cfg.RootRoute)
	}
}

func getMux(cfg Config) *http.ServeMux {
	handlers, paths := loadRouteHandlers(&cfg)
	mux := http.NewServeMux()
	addMuxRoutes(mux, paths, handlers)
	redirectRootRoute(cfg, mux, handlers)
	return mux
}

func Serve(ctx context.Context, cfg Config) error {
	mux := getMux(cfg)
	exeName := getExeName()

	if cfg.SslCertificate != "" && cfg.SslKey != "" {
		log.Printf("%s (HTTPS) listening on %q", exeName, cfg.Addr)
		return http.ListenAndServeTLS(cfg.Addr, cfg.SslCertificate, cfg.SslKey, mux)
	}
	log.Printf("%s listening on %q", exeName, cfg.Addr)
	return http.ListenAndServe(cfg.Addr, mux)
}
