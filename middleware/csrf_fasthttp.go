package middleware

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"github.com/spi-ca/eighty"
	"github.com/spi-ca/eighty/routing"
	"github.com/valyala/fasthttp"
	"gitlab.com/NebulousLabs/fastrand"
	"io"
	"net/http"
	"time"
)

const (
	// the name of CSRF cookie
	CsrfCookieName = "csrf_token"

	// the name of CSRF header
	csrfContextKey = "csrf"

	csrfTokenLength = 32
)

// reasons for CSRF check failures
var (
	csrfSafeMethods = [][]byte{
		[]byte(http.MethodGet),
		[]byte(http.MethodHead),
		[]byte(http.MethodOptions),
		[]byte(http.MethodTrace),
	}
	mockCSRFRouterMiddleware = func(next routing.Router) routing.Router { return next }
)

type (
	csrfToken struct {
		payload string
	}
	csrfMiddleware struct {
		writer eighty.CookieWriterFasthttp
	}
)

// CSRFToken returns a CSRF token in the current request context.
// If the token was not found in the request, zero-value returned.
func CSRFToken(ctx *fasthttp.RequestCtx) (token string) {
	if ctx, ok := ctx.UserValue(csrfContextKey).(*csrfToken); ok && ctx != nil {
		token = ctx.payload
	}
	return
}

// Masks/unmasks the given data *in place*
// with the given key
// Slices must be of the same length, or csrfOneTimePad will panic
func (m *csrfMiddleware) csrfOneTimePad(data, key []byte) {
	n := len(data)
	if n != len(key) {
		panic("Lengths of slices are not equal")
	}

	for i := 0; i < n; i++ {
		data[i] ^= key[i]
	}
}

func (m *csrfMiddleware) isMethodSafe(s []byte) (safe bool) {
	// checks if the given slice contains the given string
	for _, v := range csrfSafeMethods {
		if safe = subtle.ConstantTimeCompare(v, s) == 1; safe {
			break
		}
	}
	return
}

// A token is generated by returning csrfTokenLength bytes
// from crypto/rand
func (m *csrfMiddleware) generateToken() []byte {
	bytes := make([]byte, csrfTokenLength)

	if _, err := io.ReadFull(rand.Reader, bytes); err != nil {
		panic(err)
	}

	return bytes
}

func (m *csrfMiddleware) tokenSerializer(data []byte, mask bool) (encoded string) {
	if !mask || len(data) != csrfTokenLength {
		return
	}

	// csrfTokenLength*2 == len(enckey + token)
	result := make([]byte, 2*csrfTokenLength)
	// the first half of the result is the OTP
	// the second half is the masked token itself
	key := result[:csrfTokenLength]
	token := result[csrfTokenLength:]
	copy(token, data)

	// generate the random token
	if _, err := io.ReadFull(fastrand.Reader, key); err != nil {
		panic(err)
	}
	m.csrfOneTimePad(token, key)

	return base64.StdEncoding.EncodeToString(result)
}

func (m *csrfMiddleware) tokenDeserializer(data []byte, unmask bool) (decoded []byte) {
	payloadSize := base64.StdEncoding.DecodedLen(len(data))
	if payloadSize != csrfTokenLength*2 {
		return
	}

	decoded = make([]byte, payloadSize)
	n, err := base64.StdEncoding.Decode(decoded, data)
	if err != nil || n < payloadSize {
		return nil
	}

	decoded = decoded[:n]
	if unmask {
		key := decoded[:csrfTokenLength]
		decoded = decoded[csrfTokenLength:]
		m.csrfOneTimePad(decoded, key)
	}
	return
}

func (m *csrfMiddleware) verifyToken(realToken, sentToken []byte) bool {
	realN := len(realToken)
	sentN := len(sentToken)
	if realN == csrfTokenLength && sentN == csrfTokenLength {
		return subtle.ConstantTimeCompare(realToken, sentToken) == 1
	}
	return false
}

func (m *csrfMiddleware) Handle(h routing.Router) routing.Router {
	return func(ctx *fasthttp.RequestCtx) {
		var (
			realToken     []byte
			internalToken csrfToken
			tokenCreated  bool
		)

		if cookieValue := ctx.Request.Header.Cookie(CsrfCookieName); len(cookieValue) > 0 {
			realToken = m.tokenDeserializer(cookieValue, false)
		}
		tokenCreated = len(realToken) != csrfTokenLength
		if tokenCreated {
			realToken = m.generateToken()
		}
		internalToken = csrfToken{
			payload: m.tokenSerializer(realToken, true),
		}
		ctx.SetUserValue(csrfContextKey, &internalToken)

		if m.isMethodSafe(ctx.Method()) {
			h(ctx)
		} else if sentToken := m.tokenDeserializer(ctx.Request.Header.Peek(eighty.XCsrfToken), true); !m.verifyToken(realToken, sentToken) {
			panic(eighty.HandledErrorBadRequest)
		} else {
			h(ctx)
		}
		ctx.Response.Header.Set(eighty.VaryHeader, "Cookie")
		if tokenCreated {
			m.writer(&ctx.Response, ctx.Host(), m.tokenSerializer(realToken, false))
		}
	}
}

// CSRFFunc returns a routing.Middleware that handles CSRF validation logic.
func CSRFFunc(isDebug bool, expire time.Duration, secure bool) (w routing.Middleware) {
	if isDebug {
		return mockCSRFRouterMiddleware
	}
	return (&csrfMiddleware{
		writer: eighty.NewCookieWriter(CsrfCookieName, expire, secure),
	}).Handle
}
