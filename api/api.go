package api

import (
	"iamtuxmeister/bible-app/pkg/server"
	"net/http"
)

func Handler(w http.ResponseWriter, r *http.Request) {
    server.ApiRequestHandler(w, r)
}
