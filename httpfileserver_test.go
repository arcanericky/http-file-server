package httpfileserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/sgreben/httpfileserver/internal/filehandler"
	"github.com/sgreben/httpfileserver/internal/routes"
)

func Test_getExeName(t *testing.T) {
	// tough to unit test because OS dependent so
	// basically just making sure it gets populated
	// with something
	tests := []struct {
		name string
	}{
		{
			name: "executable name",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getExeName(); got == "" {
				t.Errorf("getExeName() = returned empty string")
			}
		})
	}
}

func Test_loadRouteHandlers(t *testing.T) {
	curPath, _ := os.Getwd()
	curBase := "/" + filepath.Base(curPath) + "/"

	type args struct {
		cfg Config
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "empty route value",
			args: args{
				cfg: Config{
					Routes: routes.Routes{},
				},
			},
			want: []string{
				curBase,
			},
		},
		{
			name: "single route value",
			args: args{
				cfg: Config{
					Routes: routes.Routes{
						Values: []struct {
							Route string
							Path  string
						}{
							{
								Route: "/route",
								Path:  "/path",
							},
						},
					},
				},
			},
			want: []string{
				"/route",
			},
		},
		{
			name: "multiple route value",
			args: args{
				cfg: Config{
					Routes: routes.Routes{
						Values: []struct {
							Route string
							Path  string
						}{
							{
								Route: "/route1",
								Path:  "/path1",
							},
							{
								Route: "/route2",
								Path:  "/path2",
							},
						},
					},
				},
			},
			want: []string{
				"/route1",
				"/route2",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := loadRouteHandlers(&tt.args.cfg)

			for _, v := range tt.want {
				if _, ok := got[v]; !ok {
					t.Errorf("loadRouteHandlers() did not load map entry for %v", v)
				}
			}
		})
	}
}

func TestNewConfig(t *testing.T) {
	tests := []struct {
		name string
		want Config
	}{
		{
			name: "success",
			want: Config{
				Addr:             ":8080",
				AllowUploadsFlag: false,
				RootRoute:        "/",
				SslCertificate:   "",
				SslKey:           "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewConfig(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

type testHandler filehandler.FileHandler

func (th *testHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {}
func (th *testHandler) GetRoute() string {
	return "/route"
}
func (th *testHandler) GetPath() string {
	return "/route"
}

func Test_addMuxRoutes(t *testing.T) {
	type args struct {
		mux      *http.ServeMux
		paths    map[string]string
		handlers map[string]routeEntry
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "success",
			args: args{
				mux: http.NewServeMux(),
				paths: map[string]string{
					"/route": "/path",
				},
				handlers: map[string]routeEntry{
					"/route": &testHandler{},
				},
			},
			want: "/route",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addMuxRoutes(tt.args.mux /* tt.args.paths, */, tt.args.handlers)
			r := &http.Request{
				URL: &url.URL{
					Path: tt.want,
				},
			}
			h, _ := tt.args.mux.Handler(r)
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			if w.Code == http.StatusNotFound {
				t.Error("handler not registered")
			}
		})
	}
}

func Test_redirectRootRoute(t *testing.T) {
	type args struct {
		cfg      Config
		mux      *http.ServeMux
		handlers map[string]routeEntry
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "success",
			args: args{
				cfg: Config{
					RootRoute: "/",
					Routes: routes.Routes{
						Values: []struct {
							Route string
							Path  string
						}{
							{
								Route: "/route",
								Path:  "/path",
							},
						},
					},
				},
				mux: http.NewServeMux(),
				handlers: map[string]routeEntry{
					"/route": nil,
				},
			},
		},
		{
			name: "failure",
			args: args{
				cfg: Config{
					RootRoute: "/",
					Routes:    routes.Routes{},
				},
				mux: http.NewServeMux(),
				handlers: map[string]routeEntry{
					"/route": nil,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			redirectRootRoute(tt.args.cfg, tt.args.mux, tt.args.handlers)
		})
	}
}

func Test_getMux(t *testing.T) {
	type args struct {
		cfg Config
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "success",
			args: args{
				cfg: Config{
					RootRoute: "/",
					Routes: routes.Routes{
						Values: []struct {
							Route string
							Path  string
						}{
							{
								Route: "/route",
								Path:  "/path",
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getMux(tt.args.cfg); got == nil {
				t.Errorf("getMux() returned nil mux")
			}
		})
	}
}

func TestServe(t *testing.T) {
	httpServe := func(addr string, handler http.Handler) error {
		return nil
	}
	httpsServe := func(addr, cert, key string, handler http.Handler) error {
		return nil
	}
	tlsConfig := NewConfig()
	tlsConfig.SslCertificate = "certfile"
	tlsConfig.SslKey = "keyfile"

	type args struct {
		ctx context.Context
		cfg Config
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "http success",
			args: args{
				ctx: context.Background(),
				cfg: NewConfig(),
			},
			wantErr: false,
		},
		{
			name: "https success",
			args: args{
				ctx: context.Background(),
				cfg: tlsConfig,
			},
			wantErr: false,
		},
	}

	listenAndServe = httpServe
	listenAndServeTLS = httpsServe

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Serve(tt.args.ctx, tt.args.cfg); (err != nil) != tt.wantErr {
				t.Errorf("Serve() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
