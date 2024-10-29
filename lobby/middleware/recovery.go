package middleware

import (
	"net/http"

	"gladiatorsGoModule/logger"

	log "github.com/sirupsen/logrus"
)

func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Errorf("%s Recovered from panic: %v", logger.LOG_Middleware, rec)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
