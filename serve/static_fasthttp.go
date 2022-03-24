package serve

import (
	"bytes"
	"crypto/subtle"
	"github.com/spi-ca/eighty"
	"github.com/spi-ca/misc/strutil"
	"github.com/valyala/fasthttp"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

const cacheControlPrefix = "public, max-age="

type staticFileHandler struct {
	contentPool  sync.Pool
	pathPrefix   []byte
	baseVersion  int64
	cacheFiller  func(header *fasthttp.ResponseHeader, versionPrefix, discretePrefix []byte)
	fileProvider FileProvider
}

// NewStaticFileHandler returns a static handler for fasthttp with advanced caching and url prefix stripping.
func NewStaticFileHandler(debug bool, urlPrefix, staticUrl string, startupTime time.Time, fileProvider FileProvider) fasthttp.RequestHandler {
	h := staticFileHandler{
		pathPrefix:   []byte(path.Join(urlPrefix, staticUrl)),
		baseVersion:  startupTime.Unix(),
		fileProvider: fileProvider,
	}

	h.contentPool.New = func() any { return [sniffLen]byte{} }

	if debug {
		h.cacheFiller = func(header *fasthttp.ResponseHeader, versionPrefix, discretePrefix []byte) {}
	} else {
		h.cacheFiller = func(header *fasthttp.ResponseHeader, versionPrefix, discretePrefix []byte) {
			if receivedVersion, err := strconv.ParseInt(strutil.B2S(versionPrefix), 10, 64); err != nil || receivedVersion > 0 && receivedVersion < h.baseVersion {
				return
			}

			var cacheDuration int64 = 2592000
			if pushedDuration, err := strconv.ParseInt(strutil.B2S(discretePrefix), 10, 64); err == nil && pushedDuration > 0 && pushedDuration > cacheDuration {
				cacheDuration = pushedDuration
			}
			//add header
			header.Set(eighty.CacheControlHeader, cacheControlPrefix+strconv.FormatInt(cacheDuration, 10))
			header.Set(eighty.ExpiresHeader, strconv.FormatInt(cacheDuration, 10))
		}
	}
	return h.Handle
}

func (h *staticFileHandler) resolvFile(pathData []byte) (f File, stat os.FileInfo) {
	var err error
	// normalizePath
	requestPath := strutil.B2S(bytes.TrimPrefix(pathData, h.pathPrefix))
	strippedPath := path.Clean(requestPath)
	if f, err = h.fileProvider(strippedPath); err != nil || f == nil {
		panic(eighty.HandledErrorNotFound)
	}

	// resolvFileInfo
	if stat, err = f.Stat(); err != nil {
		_, code := toHTTPError(err)
		err, _ = eighty.HandledErrorCodeOf(code)
		panic(err)
	} else if stat.IsDir() {
		panic(eighty.HandledErrorNotFound)
	}
	return
}

func (h *staticFileHandler) resolvContentType(name string, header *fasthttp.ResponseHeader, file File) {
	if ctypes := header.ContentType(); len(ctypes) > 0 {
		return
	}

	if ctype := mime.TypeByExtension(filepath.Ext(name)); len(ctype) > 0 {
		header.SetContentType(ctype)
		return
	}

	buf := h.contentPool.Get().([sniffLen]byte)
	defer h.contentPool.Put(buf)

	// read a chunk to decide between utf-8 text and binary
	if n, _ := file.Read(buf[:]); n > 0 {
		// rewind to output whole file
		if _, seekErr := file.Seek(0, io.SeekStart); seekErr != nil {
			log.Printf("file %s seeker can't seek", name)
			panic(http.StatusInternalServerError)
		}
		header.SetContentType(http.DetectContentType(buf[:n]))
	} else {
		header.SetContentType("application/octet-stream")
	}
}

// Handle is a handler method for http request.
func (h *staticFileHandler) Handle(ctx *fasthttp.RequestCtx) {

	var (
		// step 1 stripping url
		// step 2 openfile
		f, stat   = h.resolvFile(ctx.Path())
		queryArgs = ctx.QueryArgs()
		modTime   = stat.ModTime()
		name      = f.Name()
		respHdr   = &ctx.Response.Header
		etag      []byte
	)
	// step 3 cache control
	h.cacheFiller(respHdr, queryArgs.Peek("v"), queryArgs.Peek("d"))
	respHdr.Set(eighty.VaryHeader, eighty.UserAgentHeader)
	// step 4 serve file
	if !isZeroTime(modTime) {
		respHdr.SetLastModified(modTime)
	}

	etag = f.Hash()
	if len(etag) > 0 {
		respHdr.SetBytesV(eighty.EtagHeader, etag)
	}

	// serveContent will check modification time
	if checkPreconditions(&ctx.Request, &ctx.Response, modTime, etag) {
		// not modified
		return
	}

	// step 5 resolve content type
	h.resolvContentType(name, respHdr, f)

	if size := stat.Size(); size <= 0 {
		ctx.Response.SkipBody = true
		ctx.SetStatusCode(http.StatusNoContent)
		return
	} else if subtle.ConstantTimeCompare(ctx.Method(), eighty.MethodHEAD) == 1 {
		ctx.Response.SkipBody = true
		ctx.SetStatusCode(http.StatusOK)
	} else {
		ctx.SetBodyStream(f, int(size))
		ctx.SetStatusCode(http.StatusOK)
	}
}
