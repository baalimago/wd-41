package serve

import (
	"net/http"

	"github.com/baalimago/go_away_boilerplate/pkg/ancli"
)

func slogHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ancli.PrintfOK("%s - %s", r.Method, r.URL.Path)

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

func cacheHandler(next http.Handler, cacheControl string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Cache-Control", cacheControl)
		// Call the next handler
		next.ServeHTTP(w, r)
	})
}
