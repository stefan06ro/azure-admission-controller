package app

import (
	"net/http"
)

type ResourceHandler interface {
	Resource() string
}

type HttpRequestHandler interface {
	Handle(pattern string, handler http.Handler)
}
