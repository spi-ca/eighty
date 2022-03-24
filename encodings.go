package eighty

import (
	"github.com/spi-ca/misc/strutil"
	"github.com/valyala/fasthttp"
	"mime"
	"strings"
)

// Collection of predefined request header names.
const (
	ContentTypeHeader    = "Content-Type"
	ContentLengthHeader  = "Content-Length"
	EtagHeader           = "Etag"
	UserAgentHeader      = "User-Agent"
	LastModifiedHeader   = "Last-Modified"
	ExpiresHeader        = "Expires"
	CacheControlHeader   = "Cache-Control"
	IfModifiedSince      = "If-Modified-Since"
	IfNoneMatch          = "If-None-Match"
	Server               = "Server"
	VaryHeader           = "Vary"
	ForwardedForIPHeader = "X-Forwarded-For"
)

// Collection of predefined response header names.
const (
	RetryAfterHeader        = "Retry-After"
	LocationHeader          = "Location"
	FrameOptionHeader       = "X-Frame-Options"
	ContentTypeOptionHeader = "X-Content-Type-Options"
	XssProtectionHeader     = "X-XSS-Protection"
	XCsrfToken              = "X-CSRF-Token"
	XForwardedProto         = "X-Forwarded-Proto"
)

// Collection of predefined cache header values.
const (
	CacheControlNoCache = "private, no-cache, no-store, no-transform, max-age=0, must-revalidate"
	ExpiresNone         = "0"
)

// Collection of predefined mime types.
var (
	HtmlContentUTF8Type      = []string{"text/html; charset=utf-8"}
	HtmlContentType          = []string{"text/html"}
	TextContentType          = []string{"text/text"}
	TextContentUTF8Type      = []string{"text/text; charset=utf-8"}
	UrlencodeContentUTF8Type = []string{"application/x-www-form-urlencoded; charset=utf-8"}
	UrlencodeContentType     = []string{"application/x-www-form-urlencoded"}
	JsonContentUTF8Type      = []string{"application/json; charset=utf-8"}
	JsonContentType          = []string{"application/json"}
)

// Collection of predefined CSRF header values.
var (
	FrameOptionDeny       = []string{"DENY"}
	FrameOptionSameOrigin = []string{"SAMEORIGIN"}

	ContentTypeOptionNoSniffing = []string{"nosniff"}
	XssProtectionBlocking       = []string{"1; mode=block"}
)

// Collection of predefined http method names.
var (
	MethodHEAD = []byte("HEAD")
	MethodGET  = []byte("GET")
	MethodPOST = []byte("POST")
)

// HasContentTypeFasthttp is a simple checker that checks if an incoming request satisfies a given mime-type.
func HasContentTypeFasthttp(r *fasthttp.Request, mimetype string) bool {
	contentType := strutil.B2S(r.Header.ContentType())
	if len(contentType) == 0 {
		return mimetype == "application/octet-stream"
	}

	for _, v := range strings.Split(contentType, ",") {
		t, _, err := mime.ParseMediaType(v)
		if err != nil {
			break
		}
		if t == mimetype {
			return true
		}
	}
	return false
}
