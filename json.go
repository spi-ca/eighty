package eighty

import (
	"github.com/spi-ca/misc"
	"github.com/valyala/fasthttp"
)

const (
	jsonMimeType = "application/json; charset=utf-8"
)

// DumpJSONFasthttp is a simple JSON renderer for the fasthttp.
func DumpJSONFasthttp(ctx *fasthttp.RequestCtx, code int, serializable any) {
	stream := misc.JSONCodec.BorrowStream(nil)
	defer misc.JSONCodec.ReturnStream(stream)

	if stream.WriteVal(serializable); stream.Error != nil {
		panic(stream.Error)
	} else if _, err := ctx.Write(stream.Buffer()); err != nil {
		panic(err)
	}

	ctx.SetContentType(jsonMimeType)
	ctx.SetStatusCode(code)
}
