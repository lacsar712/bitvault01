package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"example.com/bitvault/internal/config"
	"example.com/bitvault/internal/httpapi"
)

func main() {
	var (
		addr   = flag.String("addr", "127.0.0.1:8080", "listen address")
		dbPath = flag.String("db", "bitvault.sqlite", "sqlite path")
		data   = flag.String("data", "data", "attachments directory")
		token  = flag.String("token", "dev-token", "bootstrap admin token")
	)
	flag.Parse()
	cfg := config.Config{
		Addr:           *addr,
		DBPath:         *dbPath,
		DataDir:        *data,
		AdminToken:     *token,
		RequestTimeout: config.DefaultTimeout,
	}
	cfg = cfg.Normalized()
	if err := cfg.Validate(); err != nil {
		log.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Clean(cfg.DataDir), 0o755); err != nil {
		log.Fatal(err)
	}
	srv, err := httpapi.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Bitvault A/B listening on %s", cfg.Addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
