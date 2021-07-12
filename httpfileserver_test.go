package httpfileserver

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"testing"

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
		name  string
		args  args
		want  []string
		want1 map[string]string
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
			want1: map[string]string{
				curBase: curPath,
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
			want1: map[string]string{
				"/route": "/path",
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
			want1: map[string]string{
				"/route1": "/path1",
				"/route2": "/path2",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := loadRouteHandlers(&tt.args.cfg)

			for _, v := range tt.want {
				if _, ok := got[v]; !ok {
					t.Errorf("loadRouteHandlers() did not load map entry for %v", v)
				}
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("loadRouteHandlers() got1 = %v, want %v", got1, tt.want1)
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

type testHandler struct{}

func (th *testHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {}

func Test_addMuxRoutes(t *testing.T) {
	type args struct {
		mux      *http.ServeMux
		paths    map[string]string
		handlers map[string]http.Handler
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
				handlers: map[string]http.Handler{
					"/route": &testHandler{},
				},
			},
			want: "/route",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addMuxRoutes(tt.args.mux, tt.args.paths, tt.args.handlers)
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
		handlers map[string]http.Handler
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
				// handlers: map[string]http.Handler{},
				handlers: map[string]http.Handler{
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
