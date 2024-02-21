package mintscan

import (
	"net/http"

	"github.com/tomasen/realip"

	"go.uber.org/zap"
)

// Middleware logs incoming requests and calls next handler.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		clientIP := realip.FromRequest(r)
		zap.S().Infof("%s %s [%s]", r.Method, r.URL, clientIP)

		next.ServeHTTP(rw, r)
	})
}
