package httpmiddleware

import (
	"net/http"
)

// Quite ambiguous, maybe first step is to spin up an HTTP server and write down the middleware type

// This is the Middleware type that I am supposed to work with
type Middleware func(http.Handler) http.Handler
