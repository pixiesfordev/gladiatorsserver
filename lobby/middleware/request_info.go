package middleware

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

func RequestInfoMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Infof("類型: (%s), 請求: (%s), 請求方: (%s)", r.Method, r.URL, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}
