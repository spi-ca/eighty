package eighty

import (
	"github.com/spi-ca/misc"
	"github.com/valyala/fasthttp"
	"log"
)

type (
	// HandledError is a http status handler type
	HandledError int
	// PageRenderer is a function interface type for the http status page renderer.
	PageRenderer = func(r *fasthttp.RequestCtx, name string, context map[string]any) error
)

var (
	// Is the interface compatible with the actual dto
	_ error = (*HandledError)(nil)
)

// Collection of predefined HandledError.
const (
	// HandledErrorBadRequest : 400, BadRequest http status
	HandledErrorBadRequest HandledError = 400
	// HandledErrorUnauthorized : 401, Unauthorized http status
	HandledErrorUnauthorized HandledError = 401
	// HandledErrorForbidden : 403, Forbidden http status
	HandledErrorForbidden HandledError = 403
	// HandledErrorNotFound : 404, NotFound http status
	HandledErrorNotFound HandledError = 404
	// HandledErrorMethodNotAllowed : 405, MethodNotAllowed http status
	HandledErrorMethodNotAllowed HandledError = 405
	// HandledErrorNotAcceptable : 406, NotAcceptable http status
	HandledErrorNotAcceptable HandledError = 406
	// HandledErrorRequestTimeout : 408, RequestTimeout http status
	HandledErrorRequestTimeout HandledError = 408
	// HandledErrorGone : 410, Gone http status
	HandledErrorGone HandledError = 410
	// HandledErrorTooManyRequests : 429, TooManyRequests http status
	HandledErrorTooManyRequests HandledError = 429
	// HandledErrorInternalServerError : 500, InternalServerError http status
	HandledErrorInternalServerError HandledError = 500
	// HandledErrorNotImplemented : 501, NotImplemented http status
	HandledErrorNotImplemented HandledError = 501
	// HandledErrorBadGateway : 502, BadGateway http status
	HandledErrorBadGateway HandledError = 502
	// HandledErrorServiceUnavailable : 503, ServiceUnavailable http status
	HandledErrorServiceUnavailable HandledError = 503
	// HandledErrorGatewayTimeout : 504, GatewayTimeout http status
	HandledErrorGatewayTimeout HandledError = 504
)

// HandledErrorCodeOf is the conversion function with the http status code to HandledError.
func HandledErrorCodeOf(value int) (HandledError, bool) {
	switch value {
	case int(HandledErrorBadRequest):
		return HandledErrorBadRequest, true
	case int(HandledErrorUnauthorized):
		return HandledErrorUnauthorized, true
	case int(HandledErrorForbidden):
		return HandledErrorForbidden, true
	case int(HandledErrorNotFound):
		return HandledErrorNotFound, true
	case int(HandledErrorMethodNotAllowed):
		return HandledErrorMethodNotAllowed, true
	case int(HandledErrorNotAcceptable):
		return HandledErrorNotAcceptable, true
	case int(HandledErrorGone):
		return HandledErrorGone, true
	case int(HandledErrorNotImplemented):
		return HandledErrorNotImplemented, true
	case int(HandledErrorBadGateway):
		return HandledErrorBadGateway, true
	case int(HandledErrorServiceUnavailable):
		return HandledErrorServiceUnavailable, true
	case int(HandledErrorGatewayTimeout):
		return HandledErrorGatewayTimeout, true
	case int(HandledErrorInternalServerError):
		return HandledErrorInternalServerError, true
	default:
		return HandledErrorInternalServerError, false
	}
}

// HandledErrorOf is the conversion function with the generic error object to HandledError.
func HandledErrorOf(value any) (HandledError, bool) {
	switch value {
	case HandledErrorBadRequest:
		return HandledErrorBadRequest, true
	case HandledErrorUnauthorized:
		return HandledErrorUnauthorized, true
	case HandledErrorForbidden:
		return HandledErrorForbidden, true
	case HandledErrorNotFound:
		return HandledErrorNotFound, true
	case HandledErrorMethodNotAllowed:
		return HandledErrorMethodNotAllowed, true
	case HandledErrorNotAcceptable:
		return HandledErrorNotAcceptable, true
	case HandledErrorRequestTimeout:
		return HandledErrorRequestTimeout, true
	case HandledErrorGone:
		return HandledErrorGone, true
	case HandledErrorTooManyRequests:
		return HandledErrorTooManyRequests, true
	case HandledErrorNotImplemented:
		return HandledErrorNotImplemented, true
	case HandledErrorBadGateway:
		return HandledErrorBadGateway, true
	case HandledErrorServiceUnavailable:
		return HandledErrorServiceUnavailable, true
	case HandledErrorGatewayTimeout:
		return HandledErrorGatewayTimeout, true
	case HandledErrorInternalServerError:
		return HandledErrorInternalServerError, true
	default:
		return HandledErrorInternalServerError, false
	}
}

// StatusCode returns the http status code.
func (handler HandledError) StatusCode() int {
	return int(handler)
}

// StatusMessage returns the http status message.
func (handler HandledError) StatusMessage() (msg string) {
	switch handler {
	case HandledErrorBadRequest:
		msg = "Bad Request"
	case HandledErrorUnauthorized:
		msg = "Unauthorized"
	case HandledErrorForbidden:
		msg = "Forbidden"
	case HandledErrorNotFound:
		msg = "Not Found"
	case HandledErrorMethodNotAllowed:
		msg = "Method Not Allowed"
	case HandledErrorNotAcceptable:
		msg = "Not Acceptable"
	case HandledErrorRequestTimeout:
		msg = "Request Timeout"
	case HandledErrorGone:
		msg = "Gone"
	case HandledErrorTooManyRequests:
		msg = "Too Many Requests"
	case HandledErrorNotImplemented:
		msg = "Not Implemented"
	case HandledErrorBadGateway:
		msg = "Bad Gateway"
	case HandledErrorServiceUnavailable:
		msg = "Service Unavailable"
	case HandledErrorGatewayTimeout:
		msg = "Gateway Timeout"
	case HandledErrorInternalServerError:
		fallthrough
	default:
		msg = "Internal Server Error"
	}
	return
}

// StatusDescription returns the http status description.
func (handler HandledError) StatusDescription() (msg string) {
	switch handler {
	case HandledErrorBadRequest:
		msg = "The request could not be understood by the server due to malformed syntax."
	case HandledErrorUnauthorized:
		msg = "The request requires user authentication."
	case HandledErrorForbidden:
		msg = "The server understood the request, but is refusing to fulfill it."
	case HandledErrorNotFound:
		msg = "The server has not found anything matching the Request-URI."
	case HandledErrorMethodNotAllowed:
		msg = "The method specified in the Request-Line is not allowed for the resource identified by the Request-URI."
	case HandledErrorNotAcceptable:
		msg = "The resource identified by the request is only capable of generating response entities which have content characteristics not acceptable according to the accept headers sent in the request."
	case HandledErrorRequestTimeout:
		msg = "The client did not produce a request within the time that the server was prepared to wait."
	case HandledErrorGone:
		msg = "The requested resource is no longer available at the server and no forwarding address is known."
	case HandledErrorTooManyRequests:
		msg = "The client has sent too many requests in a given amount of time."
	case HandledErrorNotImplemented:
		msg = "The server does not support the functionality required to fulfill the request."
	case HandledErrorBadGateway:
		msg = "The server, while acting as a gateway or proxy, received an invalid response from the upstream server it accessed in attempting to fulfill the request."
	case HandledErrorServiceUnavailable:
		msg = "The server is currently unable to handle the request due to a temporary overloading or maintenance of the server."
	case HandledErrorGatewayTimeout:
		msg = "The server, while acting as a gateway or proxy, did not receive a timely response from the upstream server specified by the URI."
	case HandledErrorInternalServerError:
		fallthrough
	default:
		msg = "The server encountered an unexpected condition which prevented it from fulfilling the request."
	}
	return
}

// RenderPage is a html page renderer function, that follows the http status code with context.
func (handler HandledError) RenderPage(ctx *fasthttp.RequestCtx, templateRenderer PageRenderer, err error) {
	defer func() {
		if len(ctx.Response.Header.ContentType()) == 0 {
			ctx.SetContentType(HtmlContentUTF8Type[0])
		}
		ctx.SetStatusCode(handler.StatusCode())
	}()

	tmplCtx := map[string]any{
		"title":       handler.StatusMessage(),
		"description": handler.StatusDescription(),
		"nofollow":    true,
	}
	if err != nil {
		tmplCtx["message"] = err.Error()
	}

	if err := templateRenderer(ctx, "error", tmplCtx); err != nil {
		log.Print("cannot render error page: ", err)
	}
	return
}

// RenderAPI is a json renderer function, that follows the http status code with context.
func (handler HandledError) RenderAPI(ctx *fasthttp.RequestCtx, _ error) {
	defer func() {
		if len(ctx.Response.Header.ContentType()) == 0 {
			ctx.SetContentType(JsonContentType[0])
		}
		ctx.SetStatusCode(handler.StatusCode())
	}()
	stream := misc.JSONCodec.BorrowStream(ctx)
	defer misc.JSONCodec.ReturnStream(stream)
	stream.WriteObjectStart()
	stream.WriteObjectField("code")
	stream.WriteInt(handler.StatusCode())
	stream.WriteMore()
	stream.WriteObjectField("message")
	stream.WriteString(handler.StatusMessage())
	stream.WriteObjectEnd()
	_ = stream.Flush()
	return
}

// Error implements the built-in interface type error.
func (handler HandledError) Error() string {
	return handler.StatusMessage()
}

// WrapHandledError is the panic handler function with a thrown panic object.
func WrapHandledError(panicObj any) (handler HandledError, err error) {
	var panicObjIsErr bool
	if err, panicObjIsErr = panicObj.(error); panicObjIsErr {
		var errIsDefined bool
		if handler, errIsDefined = HandledErrorOf(err); errIsDefined {
			err = nil
		}
	} else {
		log.Printf("panic object(%v) isn't error interface", panicObj)
		handler = HandledErrorInternalServerError
	}
	return
}
