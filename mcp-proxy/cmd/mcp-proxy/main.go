package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/ahaufox/mcp-gateway/mcp-proxy/internal/config"
	"github.com/ahaufox/mcp-gateway/mcp-proxy/internal/server"
)

var BuildVersion = "dev"

func main() {
	conf := flag.String("config", "", "path to a single config file (or a http(s) url). Mutually exclusive with --config-dir")
	confDir := flag.String("config-dir", "", "path to a directory of split config files (base.json + categories/*.json + overrides/*.json). Mutually exclusive with --config")
	insecure := flag.Bool("insecure", false, "allow insecure HTTPS connections by skipping TLS certificate verification")
	expandEnv := flag.Bool("expand-env", true, "expand environment variables in config file")
	httpHeaders := flag.String("http-headers", "", "optional HTTP headers for config URL, format: 'Key1:Value1;Key2:Value2'")
	httpTimeout := flag.Int("http-timeout", 20, "HTTP timeout in seconds when fetching config from URL")

	version := flag.Bool("version", false, "print version and exit")
	help := flag.Bool("help", false, "print help and exit")
	flag.Parse()
	if *help {
		flag.Usage()
		return
	}
	if *version {
		fmt.Println(BuildVersion)
		return
	}
	if *conf != "" && *confDir != "" {
		log.Fatalf("--config and --config-dir are mutually exclusive")
	}
	if *conf == "" && *confDir == "" {
		// Default: prefer split-config dir if it exists, otherwise fall back
		// to the legacy single-file config.json path.
		if _, err := os.Stat("configs"); err == nil {
			*confDir = "configs"
		} else {
			*conf = "config.json"
		}
	}

	var (
		cfg          *config.Config
		err          error
		configSource func() ([]byte, error)
	)
	if *confDir != "" {
		cfg, err = config.LoadDir(*confDir, *insecure, *expandEnv, *httpHeaders, *httpTimeout)
		if err != nil {
			log.Fatalf("Failed to load config dir %q: %v", *confDir, err)
		}
		// In dir mode, no single file on disk represents the effective
		// configuration, so the /api/config endpoint should serve the
		// in-memory merged view rather than a stale per-file snippet.
		configSource = func() ([]byte, error) { return json.Marshal(cfg) }
		// Keep server.StartHTTPServer's reload path happy — it watches the
		// legacy single-file path, so feed it base.json when in dir mode.
		*conf = filepath.Join(*confDir, "base.json")
	} else {
		cfg, err = config.Load(*conf, *insecure, *expandEnv, *httpHeaders, *httpTimeout)
		if err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}
	}
	err = server.StartHTTPServerWithConfigSource(cfg, *conf, configSource)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
