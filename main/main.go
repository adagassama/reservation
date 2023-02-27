package main

import (
	"fmt"
	"net/http"
	"reservation/web"
)

func main() {

	mux := web.NewHandler()

	err := http.ListenAndServe(":2020", mux)
	if err != nil {
		_ = fmt.Errorf("impossible de lancer le serveur : %w", err)
		return
	}
}
