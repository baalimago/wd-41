package serve

import (
	"net/http"

	"github.com/baalimago/go_away_boilerplate/pkg/ancli"
)

func SlogHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ancli.Okf("%s - %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func CacheHandler(next http.Handler, cacheControl string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Cache-Control", cacheControl)
		next.ServeHTTP(w, r)
	})
}

func CrossOriginIsolationHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Cross-Origin-Opener-Policy", "same-origin")
		w.Header().Add("Cross-Origin-Embedder-Policy", "require-corp")
		next.ServeHTTP(w, r)
	})
}
