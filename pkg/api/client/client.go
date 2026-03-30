package client

import "net/http"

func Setup(mux *http.ServeMux, servedir string) {
	mux.Handle("/", http.FileServer(http.Dir(servedir)))
}
