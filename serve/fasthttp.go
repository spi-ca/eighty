package serve

import (
	"bytes"
	"crypto/subtle"
	"github.com/spi-ca/eighty"
	"github.com/spi-ca/misc/strutil"
	"github.com/valyala/fasthttp"
	"net/http"
	"net/textproto"
	"time"
)

const sniffLen = 512

var (
	etagPrefix = []byte("W/")

	unixEpochTime = time.Unix(0, 0)
)

type condResult int

const (
	condNone condResult = iota
	condTrue
	condFalse
)

// checkPreconditions evaluates request preconditions and reports whether a precondition
// resulted in sending StatusNotModified or StatusPreconditionFailed.
func checkPreconditions(r *fasthttp.Request, w *fasthttp.Response, modtime time.Time, etag []byte) (done bool) {
	// This function carefully follows RFC 7232 section 6.
	ch := checkIfMatch(r, w)
	if ch == condNone {
		ch = checkIfUnmodifiedSince(r, modtime)
	}
	if ch == condFalse {
		w.SetStatusCode(http.StatusPreconditionFailed)
		return true
	}
	switch checkIfNoneMatch(etag, r, w) {
	case condFalse:
		if subtle.ConstantTimeCompare(r.Header.Method(), eighty.MethodGET) == 1 || subtle.ConstantTimeCompare(r.Header.Method(), eighty.MethodHEAD) == 1 {
			writeNotModified(w)
			return true
		} else {
			w.SetStatusCode(http.StatusPreconditionFailed)
			return true
		}
	case condNone:
		if checkIfModifiedSince(r, modtime) == condFalse {
			writeNotModified(w)
			return true
		}
	}
	return
}

func checkIfNoneMatch(providedEtag []byte, r *fasthttp.Request, w *fasthttp.Response) condResult {
	inm := r.Header.Peek("If-None-Match")
	if len(inm) == 0 {
		return condNone
	}
	buf := inm
	for {
		buf = textproto.TrimBytes(buf)
		if len(buf) == 0 {
			break
		}
		if buf[0] == ',' {
			buf = buf[1:]
		}
		if buf[0] == '*' {
			return condFalse
		}
		etag, remain := scanETag(buf)
		if len(etag) == 0 {
			break
		}
		if etagWeakMatch(etag, providedEtag) {
			return condFalse
		}
		buf = remain
	}
	return condTrue
}

func checkIfModifiedSince(r *fasthttp.Request, modtime time.Time) condResult {
	if subtle.ConstantTimeCompare(r.Header.Method(), eighty.MethodGET) != 1 && subtle.ConstantTimeCompare(r.Header.Method(), eighty.MethodHEAD) != 1 {
		return condNone
	}
	ims := r.Header.Peek("If-Modified-Since")
	if len(ims) == 0 || isZeroTime(modtime) {
		return condNone
	}
	t, err := http.ParseTime(strutil.B2S(ims))
	if err != nil {
		return condNone
	}
	// The Date-Modified header truncates sub-second precision, so
	// use mtime < t+1s instead of mtime <= t to check for unmodified.
	if modtime.Before(t.Add(1 * time.Second)) {
		return condFalse
	}
	return condTrue
}

func checkIfUnmodifiedSince(r *fasthttp.Request, modtime time.Time) condResult {
	ius := r.Header.Peek("If-Unmodified-Since")
	if len(ius) == 0 || isZeroTime(modtime) {
		return condNone
	}
	if t, err := http.ParseTime(strutil.B2S(ius)); err == nil {
		// The Date-Modified header truncates sub-second precision, so
		// use mtime < t+1s instead of mtime <= t to check for unmodified.
		if modtime.Before(t.Add(1 * time.Second)) {
			return condTrue
		}
		return condFalse
	}
	return condNone
}

func checkIfMatch(r *fasthttp.Request, w *fasthttp.Response) condResult {
	im := r.Header.Peek("If-Match")
	if len(im) == 0 {
		return condNone
	}
	for {
		im = textproto.TrimBytes(im)
		if len(im) == 0 {
			break
		}
		if im[0] == ',' {
			im = im[1:]
			continue
		}
		if im[0] == '*' {
			return condTrue
		}
		etag, remain := scanETag(im)
		if len(etag) == 0 {
			break
		}
		if etagStrongMatch(etag, w.Header.Peek("Etag")) {
			return condTrue
		}
		im = remain
	}

	return condFalse
}

func scanETag(s []byte) (etag []byte, remain []byte) {
	s = textproto.TrimBytes(s)
	start := 0
	if bytes.HasPrefix(s, etagPrefix) {
		start = 2
	}
	if len(s[start:]) < 2 || s[start] != '"' {
		return
	}
	// ETag is either W/"text" or "text".
	// See RFC 7232 2.3.
	for i := start + 1; i < len(s); i++ {
		c := s[i]
		switch {
		// Character values allowed in ETags.
		case c == 0x21 || c >= 0x23 && c <= 0x7E || c >= 0x80:
		case c == '"':
			return s[:i+1], s[i+1:]
		default:
			return
		}
	}
	return
}

func etagStrongMatch(a, b []byte) bool {
	return subtle.ConstantTimeCompare(a, b) == 1 && len(a) > 0 && a[0] == '"'
}

// etagWeakMatch reports whether a and b match using weak ETag comparison.
// Assumes a and b are valid ETags.
func etagWeakMatch(a, b []byte) bool {
	return subtle.ConstantTimeCompare(
		bytes.TrimPrefix(a, etagPrefix),
		bytes.TrimPrefix(b, etagPrefix),
	) == 1
}

func writeNotModified(w *fasthttp.Response) {
	// RFC 7232 section 4.1:
	// a sender SHOULD NOT generate representation metadata other than the
	// above listed fields unless said metadata exists for the purpose of
	// guiding cache updates (e.g., Last-Modified might be useful if the
	// response does not have an ETag field).
	w.Header.Del("Content-Type")
	w.Header.Del("Content-Length")
	if len(w.Header.Peek("Etag")) > 0 {
		w.Header.Del("Last-Modified")
	}
	w.SetStatusCode(http.StatusNotModified)
	w.SkipBody = true
}

func isZeroTime(t time.Time) bool {
	return t.IsZero() || t.Equal(unixEpochTime)
}
