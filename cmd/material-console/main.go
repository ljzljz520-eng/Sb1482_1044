package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"example.com/materialconsole/internal/httpapi"
	"example.com/materialconsole/internal/service"
	"example.com/materialconsole/internal/store"
)

func main() {
	databasePath := flag.String("db", "material-console.db", "bbolt database path")
	address := flag.String("addr", "127.0.0.1:8099", "HTTP listen address")
	flag.Parse()
	storage, err := store.Open(*databasePath)
	if err != nil {
		log.Fatal(err)
	}
	defer storage.Close()
	console := service.New(storage)
	server := &http.Server{Addr: *address, Handler: httpapi.New(console).Handler()}
	fmt.Printf("material-console listening on %s\n", *address)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
