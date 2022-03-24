package eighty

import (
	"github.com/spi-ca/misc/networking"
	"github.com/spi-ca/misc/strutil"
	"github.com/valyala/fasthttp"
	"log"
	"net"
	"strings"
	"time"
)

type (
	cookieWriterFasthttpImpl struct {
		key            string
		expireDuration time.Duration
		sessionSecure  bool
	}

	CookieWriterFasthttp func(*fasthttp.Response, []byte, string)
)

// NewCookieWriter is a useful cookie generator for fasthttp.
func NewCookieWriter(key string, expireDuration time.Duration, secured bool) CookieWriterFasthttp {
	return (&cookieWriterFasthttpImpl{
		key:            key,
		expireDuration: expireDuration,
		sessionSecure:  secured,
	}).Write
}
func (cw *cookieWriterFasthttpImpl) validateCookiePathByte(b byte) bool {
	return 0x20 <= b && b < 0x7f && b != ';'
}

func (cw *cookieWriterFasthttpImpl) validateCookieValueByte(b byte) bool {
	return 0x20 <= b && b < 0x7f && b != '"' && b != ';' && b != '\\'
}

func (cw *cookieWriterFasthttpImpl) validateCookieDomain(v []byte) (valid bool) {
	// isCookieDomainName
	if len(v) == 0 {
		return false
	}
	if len(v) > 255 {
		return false
	}

	if v[0] == '.' {
		// A cookie a domain attribute may start with a leading dot.
		v = v[1:]
	}
	var last byte = '.'
	partlen := 0
	for i := 0; i < len(v); i++ {
		c := v[i]
		switch {
		default:
			return false
		case 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z':
			// No '_' allowed here (in contrast to package net).
			valid = true
			partlen++
		case '0' <= c && c <= '9':
			// fine
			partlen++
		case c == '-':
			// Byte before dash cannot be dot.
			if last == '.' {
				return false
			}
			partlen++
		case c == '.':
			// Byte before dot cannot be dot, dash.
			if last == '.' || last == '-' {
				return false
			}
			if partlen > 63 || partlen == 0 {
				return false
			}
			partlen = 0
		}
		last = c
	}

	if last == '-' || partlen > 63 {
		return false
	} else if valid {
		// isCookieDomainName
		return
	}
	// isCookieValidIp
	addr := networking.ParseIPv4(v)
	return addr != nil &&
		!addr.Equal(net.IPv4bcast) &&
		!addr.IsUnspecified() &&
		!addr.IsMulticast() &&
		!addr.IsLinkLocalUnicast()
}

func (cw *cookieWriterFasthttpImpl) sanitizeOrWarn(fieldName string, valid func(byte) bool, v string) string {
	ok := true
	for i := 0; i < len(v); i++ {
		if valid(v[i]) {
			continue
		}
		log.Printf("invalid byte %q in %s; dropping invalid bytes", v[i], fieldName)
		ok = false
		break
	}
	if ok {
		return v
	}
	var build strings.Builder
	for i := 0; i < len(v); i++ {
		if b := v[i]; valid(b) {
			build.WriteByte(b)
		}
	}
	return build.String()
}

func (cw *cookieWriterFasthttpImpl) sanitizeCookiePath(v string) string {
	return cw.sanitizeOrWarn("Cookie.Path", cw.validateCookiePathByte, v)
}

func (cw *cookieWriterFasthttpImpl) sanitizeCookieValue(v string) string {
	v = cw.sanitizeOrWarn("Cookie.Value", cw.validateCookieValueByte, v)
	if len(v) == 0 {
		return v
	}
	if strings.IndexByte(v, ' ') >= 0 || strings.IndexByte(v, ',') >= 0 {
		return `"` + v + `"`
	}
	return v
}

func (cw *cookieWriterFasthttpImpl) Write(w *fasthttp.Response, host []byte, newCookieValue string) {
	cookie := fasthttp.AcquireCookie()
	defer fasthttp.ReleaseCookie(cookie)

	cookie.SetKey(cw.key)
	cookie.SetPath(cw.sanitizeCookiePath("/"))

	if len(host) > 0 {
		if cw.validateCookieDomain(host) {
			cookie.SetDomainBytes(host)
		} else {
			log.Printf("invalid Cookie.Domain %s; dropping domain attribute", strutil.B2S(host))
		}
	}
	cookie.SetSecure(cw.sessionSecure)
	cookie.SetSameSite(fasthttp.CookieSameSiteLaxMode)
	cookie.SetHTTPOnly(true)

	if len(newCookieValue) > 0 {
		cookie.SetValue(cw.sanitizeCookieValue(newCookieValue))
		cookie.SetExpire(time.Now().Add(cw.expireDuration))
		cookie.SetMaxAge(int(cw.expireDuration.Seconds()))
	} else {
		cookie.SetValue("-")
		cookie.SetExpire(oldTime)
		cookie.SetMaxAge(-1)
	}
	w.Header.SetCookie(cookie)
}
