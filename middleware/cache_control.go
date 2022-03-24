package middleware

import (
	"github.com/spi-ca/eighty"
	"net/http"
	"strconv"
	"time"
)

// CacheControlFunc returns a func(next http.Handler) http.Handler that handles cache control header.
func CacheControlFunc(debug bool, startupTime time.Time) func(next http.Handler) http.Handler {
	baseVersion := startupTime.Unix()
	return func(next http.Handler) http.Handler {
		if debug {
			return next
		} else {
			fn := func(w http.ResponseWriter, r *http.Request) {
				defer next.ServeHTTP(w, r)
				if receivedVersion, err := strconv.ParseInt(r.URL.Query().Get("v"), 10, 64); err == nil && receivedVersion >= baseVersion {
					var cacheDuration int64 = 2592000
					if pushedDuration, err := strconv.ParseInt(r.URL.Query().Get("d"), 10, 64); err == nil && pushedDuration > cacheDuration {
						cacheDuration = pushedDuration
					}
					//add header
					cacheDurationStr := strconv.FormatInt(cacheDuration, 10)
					w.Header().Add(eighty.CacheControlHeader, "public, max-age="+cacheDurationStr)
					w.Header().Add(eighty.ExpiresHeader, cacheDurationStr)
				}
				w.Header().Add(eighty.VaryHeader, "User-Agent")
			}
			return http.HandlerFunc(fn)
		}
	}
}
