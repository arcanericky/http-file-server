package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"github.com/sgreben/httpfileserver"
)

const (
	addrEnvVarName           = "ADDR"
	allowUploadsEnvVarName   = "UPLOADS"
	portEnvVarName           = "PORT"
	quietEnvVarName          = "QUIET"
	sslCertificateEnvVarName = "SSL_CERTIFICATE"
	sslKeyEnvVarName         = "SSL_KEY"
)

var Version = ":unknown:"

func configureRuntime(cfg httpfileserver.Config) httpfileserver.Config {
	log.SetFlags(log.LUTC | log.Ldate | log.Ltime)
	log.SetOutput(os.Stderr)
	if cfg.AddrFlag == "" {
		cfg.AddrFlag = cfg.DefaultAddr
	}
	cfg.PortFlag64, _ = strconv.ParseInt(os.Getenv(portEnvVarName), 10, 64)
	flag.StringVar(&cfg.AddrFlag, "addr", cfg.AddrFlag, fmt.Sprintf("address to listen on (environment variable %q)", addrEnvVarName))
	flag.StringVar(&cfg.AddrFlag, "a", cfg.AddrFlag, "(alias for -addr)")
	flag.IntVar(&cfg.PortFlag, "port", cfg.PortFlag, fmt.Sprintf("port to listen on (overrides -addr port) (environment variable %q)", portEnvVarName))
	flag.IntVar(&cfg.PortFlag, "p", cfg.PortFlag, "(alias for -port)")
	flag.BoolVar(&cfg.QuietFlag, "quiet", cfg.QuietFlag, fmt.Sprintf("disable all log output (environment variable %q)", quietEnvVarName))
	flag.BoolVar(&cfg.QuietFlag, "q", cfg.QuietFlag, "(alias for -quiet)")
	flag.BoolVar(&cfg.AllowUploadsFlag, "uploads", cfg.AllowUploadsFlag, fmt.Sprintf("allow uploads (environment variable %q)", allowUploadsEnvVarName))
	flag.BoolVar(&cfg.AllowUploadsFlag, "u", cfg.AllowUploadsFlag, "(alias for -uploads)")
	flag.Var(&cfg.Routes, "route", cfg.Routes.Help())
	flag.Var(&cfg.Routes, "r", "(alias for -route)")
	flag.StringVar(&cfg.SslCertificate, "ssl-cert", cfg.SslCertificate, fmt.Sprintf("path to SSL server certificate (environment variable %q)", sslCertificateEnvVarName))
	flag.StringVar(&cfg.SslKey, "ssl-key", cfg.SslKey, fmt.Sprintf("path to SSL private key (environment variable %q)", sslKeyEnvVarName))
	flag.Parse()
	if cfg.QuietFlag {
		log.SetOutput(ioutil.Discard)
	}
	for i := 0; i < flag.NArg(); i++ {
		arg := flag.Arg(i)
		err := cfg.Routes.Set(arg)
		if err != nil {
			log.Fatalf("%q: %v", arg, err)
		}
	}

	return cfg
}

func newConfig() httpfileserver.Config {
	portFlag64, _ := strconv.ParseInt(os.Getenv(portEnvVarName), 10, 64)
	return httpfileserver.Config{
		AddrFlag:         os.Getenv(addrEnvVarName),
		AllowUploadsFlag: os.Getenv(allowUploadsEnvVarName) == "true",
		PortFlag:         int(portFlag64),
		QuietFlag:        os.Getenv(quietEnvVarName) == "true",
		SslCertificate:   os.Getenv(sslCertificateEnvVarName),
		SslKey:           os.Getenv(sslKeyEnvVarName),
		DefaultAddr:      ":8080",
		RootRoute:        "/",
	}
}

func main() {
	cfg := configureRuntime(newConfig())
	log.Printf("httpfileserver v%s", Version)

	addr, err := httpfileserver.Addr(cfg)
	if err != nil {
		log.Fatalf("address/port: %v", err)
	}
	err = httpfileserver.Server(addr, cfg)
	if err != nil {
		log.Fatalf("start server: %v", err)
	}
}
