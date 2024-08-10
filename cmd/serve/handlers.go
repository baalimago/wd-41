package serve

import (
	"fmt"
	"net/http"

	"github.com/baalimago/go_away_boilerplate/pkg/ancli"
)

func byteCountSI(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB",
		float64(b)/float64(div), "kMGTPE"[exp])
}

func slogHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Perform your middleware action (e.g., logging)
		ancli.PrintfOK("%s - %s", r.Method, r.URL.Path)

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}
