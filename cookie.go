package eighty

import (
	"net/http"
	"time"
)

var (
	oldTime = time.Unix(0, 0)
)

// SetCookieValue is a useful cookie generator for net.http.
func SetCookieValue(key string, expireDuration time.Duration, sessionSecure bool) func(http.ResponseWriter, string, string) {
	return func(w http.ResponseWriter, host, newCookieValue string) {
		if len(newCookieValue) > 0 {
			http.SetCookie(w,
				&http.Cookie{
					Name:     key,
					Value:    newCookieValue,
					Path:     "/",
					Domain:   host,
					Expires:  time.Now().Add(expireDuration),
					Secure:   sessionSecure,
					SameSite: http.SameSiteLaxMode,
					MaxAge:   int(expireDuration.Seconds()),
					HttpOnly: true,
				},
			)
		} else {
			http.SetCookie(w,
				&http.Cookie{
					Name:     key,
					Value:    "_",
					Path:     "/",
					Domain:   host,
					Expires:  oldTime,
					Secure:   sessionSecure,
					SameSite: http.SameSiteLaxMode,
					MaxAge:   -1,
					HttpOnly: true,
				},
			)
		}
	}
}

// GetCookieValue is the simple cookie getter.
func GetCookieValue(req *http.Request, name string) (cookieValue string) {
	if cookie, _ := req.Cookie(name); cookie != nil {
		cookieValue = cookie.Value
	}
	return
}
