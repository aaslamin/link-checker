package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/aaslamin/link-checker/linkworker"
	"github.com/aaslamin/link-checker/storage"
	"github.com/julienschmidt/httprouter"
)

var (
	addr = flag.String("listen", ":8080", "address to run the link-checker svc - e.g. \":8181\"")
)

func main() {
	flag.Parse()

	router := httprouter.New()
	linkworker.NewHandler(storage.NewMemoryStorage()).SetupRoutes(router)

	log.Printf("svc listening on port %s\n", *addr)
	log.Fatal(http.ListenAndServe(*addr, router))
}
