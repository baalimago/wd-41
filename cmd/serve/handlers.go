package serve

import (
	"net/http"

	"github.com/baalimago/go_away_boilerplate/pkg/ancli"
)

func slogHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Perform your middleware action (e.g., logging)
		ancli.PrintfOK("%s - %s", r.Method, r.URL.Path)

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}
