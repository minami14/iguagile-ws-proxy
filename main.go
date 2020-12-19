package main

import (
	"log"
	"os"

	"github.com/minami14/iguagile-ws-proxy/proxy"
)

func main() {
	addr := os.Getenv("PROXY_HOST")
	if addr == "" {
		addr = ":80"
	}

	p := proxy.New()
	log.Fatal(p.Start(addr))
}
