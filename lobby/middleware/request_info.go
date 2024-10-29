package middleware

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

func RequestInfoMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Infof("Request Method: (%s), URL: (%s), Remote Address: (%s)", r.Method, r.URL, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}
