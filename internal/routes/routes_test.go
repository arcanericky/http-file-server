package routes

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestRoutes_Help(t *testing.T) {
	const helpText = "a route definition ROUTE%sPATH (ROUTE defaults to basename of PATH if omitted)"
	type fields struct {
		Separator string
		Values    []struct {
			Route string
			Path  string
		}
		Texts []string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name:   "default separater",
			fields: fields{},
			want:   fmt.Sprintf(helpText, "="),
		},
		{
			name: "custom separater",
			fields: fields{
				Separator: "*",
			},
			want: fmt.Sprintf(helpText, "*"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fv := &Routes{
				Separator: tt.fields.Separator,
				Values:    tt.fields.Values,
				Texts:     tt.fields.Texts,
			}
			if got := fv.Help(); got != tt.want {
				t.Errorf("Routes.Help() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoutes_Set(t *testing.T) {
	currentPath, _ := os.Getwd()

	type fields struct {
		Separator string
		Values    []struct {
			Route string
			Path  string
		}
		Texts []string
	}
	type args struct {
		v string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   struct {
			Route string
			Path  string
		}
		wantErr bool
	}{
		{
			name:   "default separator, same base path and full route",
			fields: fields{},
			args: args{
				v: ".",
			},
			want: struct {
				Route string
				Path  string
			}{

				Route: "/" + filepath.Base(currentPath) + "/",
				Path:  currentPath,
			},
			wantErr: false,
		},
		{
			name:   "default separator, different base path and full route",
			fields: fields{},
			args: args{
				v: "testroute=" + currentPath,
			},
			want: struct {
				Route string
				Path  string
			}{

				Route: "/testroute/",
				Path:  currentPath,
			},
			wantErr: false,
		},
		{
			name: "custom separator, same base path and full route",
			fields: fields{
				Separator: "*",
			},
			args: args{
				v: ".",
			},
			want: struct {
				Route string
				Path  string
			}{

				Route: "/" + filepath.Base(currentPath) + "/",
				Path:  currentPath,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fv := &Routes{
				Separator: tt.fields.Separator,
				Values:    tt.fields.Values,
				Texts:     tt.fields.Texts,
			}
			if err := fv.Set(tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("Routes.Set() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got := fv.Values[0]; got != tt.want {
				t.Errorf("fv.Values[0] = %v, want %v", got, tt.want)
			}
			// fmt.Println(fv.Values)
			fmt.Println(fv.Texts)
			fmt.Println(fv.String())
			fmt.Println("---")
		})
	}
}

func TestRoutes_String(t *testing.T) {
	type fields struct {
		Separator string
		Values    []struct {
			Route string
			Path  string
		}
		Texts []string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "success",
			fields: fields{
				Texts: []string{"first", "second", "third"},
			},
			want: "first, second, third",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fv := &Routes{
				Separator: tt.fields.Separator,
				Values:    tt.fields.Values,
				Texts:     tt.fields.Texts,
			}
			if got := fv.String(); got != tt.want {
				t.Errorf("Routes.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
