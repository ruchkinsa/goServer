package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path"
	//"path/filepath"

	"./daemon"
)

var publicPath string

func processFlags() *daemon.Config {
	cfg := &daemon.Config{}

	flag.StringVar(&cfg.ListenHost, "listen", "localhost:10000", "HTTP listen")
	flag.StringVar(&cfg.Db.ConnectString, "db-connect", "root:admin@tcp(localhost:3306)/godb", "DB Connect String")
	flag.StringVar(&publicPath, "public-path", "web", "Path to public dir")

	flag.Parse()
	return cfg
}

func setupHttpAssets(cfg *daemon.Config) {
	log.Printf("PublicPath served from %q.", publicPath)
	workDir, _ := os.Getwd()
	cfg.API.PublicPath = publicPath
	cfg.API.PublicPathCSS = http.Dir(path.Join(workDir, publicPath, "css"))
	cfg.API.PublicPathJS = http.Dir(path.Join(workDir, publicPath, "js"))
	cfg.API.PublicPathTemplates = http.Dir(path.Join(workDir, publicPath, "templates"))
}

func main() {
	cfg := processFlags()

	setupHttpAssets(cfg)

	if err := daemon.Run(cfg); err != nil {
		log.Printf("Error in main(): %v", err)
	}
}
