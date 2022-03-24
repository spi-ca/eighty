package middleware

import (
	"bytes"
	"context"
	"github.com/spi-ca/eighty"
	"github.com/spi-ca/eighty/routing"
	"github.com/spi-ca/logging"
	"github.com/spi-ca/misc/q"
	"github.com/spi-ca/misc/strutil"
	"github.com/valyala/fasthttp"
	"io"
	"io/ioutil"
	"net"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	dateFormat = "02/Jan/2006:15:04:05 -0700"
)

type accessLogMiddleware struct {
	writer io.WriteCloser

	apiUrlPrefix string

	logInChan  chan<- string
	logOutChan <-chan string

	logger    logging.Logger
	logWaiter sync.Mutex
	waitGroup sync.WaitGroup
	ctx       context.Context
	closer    func()

	errorViewTemplateRenderer eighty.PageRenderer
}

func (m *accessLogMiddleware) Handle(h routing.Router) routing.Router {
	return func(ctx *fasthttp.RequestCtx) {
		m.waitGroup.Add(1)
		defer m.waitGroup.Done()
		// access log 기록
		defer m.recordAccess()(ctx)
		// 내부 panic 해소
		defer m.handlePanic(ctx)
		m.writeBasicHeader(&ctx.Response)
		h(ctx)
	}
}

// source returns a space-trimmed slice of the n'th line.
func (m *accessLogMiddleware) source(buf *strings.Builder, lines [][]byte, n int) {
	n-- // in stack trace, lines are 1-indexed but our array is 0-indexed
	if n < 0 || n >= len(lines) {
		buf.WriteString("???")
	} else {
		buf.Write(lines[n])
	}
}

// function returns, if possible, the name of the function containing the PC.
func (m *accessLogMiddleware) function(buf *strings.Builder, pc uintptr) {
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		buf.WriteString("???")
	}
	name := fn.Name()
	// The name include the path name to the package, which is unnecessary
	// since the file name is already included.  Plus, it has center dots.
	// That is, we see
	//	runtime/debug.*T·ptrmethod
	// and want
	//	*T.ptrmethod
	// Also the package path might contains dot (e.g. code.google.com/...),
	// so first eliminate the path prefix
	if lastslash := strings.LastIndexByte(name, '/'); lastslash >= 0 {
		name = name[lastslash+1:]
	}
	if period := strings.IndexByte(name, '.'); period >= 0 {
		name = name[period+1:]
	}
	buf.WriteString(strings.Replace(name, "·", ".", -1))
}

// 라우팅 로직에서 panic이 발생 했을경우 해당 스택을 보여준다.
// stack returns a nicely formated stack frame, skipping skip frames
func (m *accessLogMiddleware) getStack(buf *strings.Builder, skip int) {
	// As we loop, we open files and read them. These variables record the currently
	// loaded file.
	var lines [][]byte
	var lastFile string
	for i := skip; ; i++ {
		// Skip the expected number of frames
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		} else if i > skip {
			buf.WriteByte('\n')
		}
		if paths := strings.SplitN(file, "src/", 2); len(paths) == 1 {
			// Print this much at least.  If we can't find the source, it won't show.
			//_, _ = fmt.Fprintf(buf, "%s:%d (0x%x)\n", file, line, pc)
			buf.WriteString(file)
		} else if vendors := strings.SplitN(paths[1], "vendor/", 2); len(vendors) == 1 {
			// Print this much at least.  If we can't find the source, it won't show.
			//_, _ = fmt.Fprintf(buf, "%s:%d (0x%x)\n", paths[1], line, pc)
			buf.WriteString(paths[1])
		} else {
			// Print this much at least.  If we can't find the source, it won't show.
			//_, _ = fmt.Fprintf(buf, "%s:%d (0x%x)\n", vendors[1], line, pc)
			buf.WriteString(vendors[1])
		}
		buf.WriteByte(':')
		buf.WriteString(strconv.FormatInt(int64(line), 10))
		buf.WriteString(" (0x")
		buf.WriteString(strconv.FormatInt(int64(pc), 16))
		buf.WriteString(")\n")

		// -----
		//_, _ = fmt.Fprintf(buf, "\t%s: %s\n", function(pc), source(lines, line))
		buf.WriteByte('\t')
		m.function(buf, pc)
		buf.WriteString(": ")
		if file == lastFile {
			buf.WriteString("???")
		} else if data, err := ioutil.ReadFile(file); err != nil {
			buf.WriteString("???")
		} else {
			lines = bytes.Split(data, []byte{'\n'})
			lastFile = file
			m.source(buf, lines, line)
		}
	}
}

func (m *accessLogMiddleware) handlePanic(ctx *fasthttp.RequestCtx) {
	panicObj := recover()
	if panicObj == nil {
		return
	}
	errorType, err := eighty.WrapHandledError(panicObj)
	if err != nil {
		var buf strings.Builder
		buf.WriteString("PANIC! ")
		buf.WriteString(err.Error())
		buf.WriteString("\n--------\nREQUEST\n")
		_, _ = ctx.Request.WriteTo(&buf)
		buf.WriteString("\n--------\nSTACK\n")
		m.getStack(&buf, 3)
		buf.WriteString("\n--------")
		m.logger.Error(buf.String())
	}
	isAPI := bytes.HasPrefix(ctx.RequestURI(), []byte(m.apiUrlPrefix))
	if isAPI {
		errorType.RenderAPI(ctx, err)
	} else {
		errorType.RenderPage(ctx, m.errorViewTemplateRenderer, err)
	}
}

func (m *accessLogMiddleware) writeBasicHeader(w *fasthttp.Response) {
	w.Header.Set(eighty.FrameOptionHeader, eighty.FrameOptionSameOrigin[0])
	w.Header.Set(eighty.ContentTypeOptionHeader, eighty.ContentTypeOptionNoSniffing[0])
	w.Header.Set(eighty.XssProtectionHeader, eighty.XssProtectionBlocking[0])
}

func (m *accessLogMiddleware) recordAccess() routing.Router {
	now := time.Now()
	return func(ctx *fasthttp.RequestCtx) {
		var (
			dur     = time.Since(now)
			builder strings.Builder
		)
		_, _ = builder.Write(m.remoteAddr(ctx.RemoteAddr(), &ctx.Request))
		_, _ = builder.WriteString(` - - [`)
		_, _ = builder.WriteString(now.Format(dateFormat))
		_, _ = builder.WriteString(`] "`)
		_, _ = builder.Write(ctx.Method())
		_ = builder.WriteByte(' ')
		_, _ = builder.Write(ctx.RequestURI())
		_ = builder.WriteByte(' ')
		if ctx.Request.Header.IsHTTP11() {
			_, _ = builder.WriteString("HTTP/1.1")
		} else {
			_, _ = builder.WriteString("HTTP/1.0")
		}
		_, _ = builder.WriteString(`" `)
		_, _ = builder.Write(strutil.FormatIntToBytes(ctx.Response.StatusCode()))
		_ = builder.WriteByte(' ')
		_, _ = builder.Write(strutil.FormatIntToBytes(ctx.Response.Header.ContentLength()))
		_, _ = builder.WriteString(` "`)
		_, _ = builder.Write(ctx.Request.Header.Referer())
		_, _ = builder.WriteString(`" "`)
		_, _ = builder.Write(ctx.Request.Header.UserAgent())
		_, _ = builder.WriteString(`" `)
		_, _ = builder.Write(strutil.FormatIntToBytes(int(dur.Nanoseconds() / time.Millisecond.Nanoseconds())))
		_ = builder.WriteByte(' ')
		_, _ = builder.Write(ctx.Request.Host())
		_ = builder.WriteByte('\n')

		select {
		case <-m.ctx.Done():
			m.logger.Error("cannot accesslog record: ", builder.String())
		default:
			m.logInChan <- builder.String()
		}
	}
}

func (m *accessLogMiddleware) Close() {
	defer m.writer.Close()
	m.closer()
	close(m.logInChan)
	m.logWaiter.Lock()
	defer m.logWaiter.Unlock()
}

func (m *accessLogMiddleware) lineByLineWriter() {
	var (
		ticker  = time.NewTicker(200 * time.Millisecond)
		maxsz   = 1024 * 1024
		sz      = 0
		rcvsz   = 0
		buf     = make([]byte, maxsz)
		flusher = func() {
			if sz > 0 {
				if _, err := m.writer.Write(buf[:sz]); err != nil {
					m.logger.Error("cannot write accesslog chunk : ", err)
				}
				//reset
				sz = 0
			}
		}
	)

	m.logWaiter.Lock()
	defer func() {
		defer m.logWaiter.Unlock()
		ticker.Stop()
		flusher()
	}()
	for {
		select {
		case logItem, ok := <-m.logOutChan:
			//or do the next job
			if !ok {
				return
			}
			rcvsz = len(logItem)
			if maxsz < sz+rcvsz {
				flusher()
			}
			if rcvsz > 0 {
				// append
				copy(buf[sz:], logItem)
				sz += rcvsz
			}
		case <-ticker.C:
			// if deadline exceeded write
			flusher()
		}
	}
}

// strip port from addresses with hostname, ipv4 or ipv6
func (m *accessLogMiddleware) stripPort(address string) string {
	if h, _, err := net.SplitHostPort(address); err == nil {
		return h
	}

	return address
}

// The remote address of the client. When the 'X-Forwarded-For'
// header is set, then it is used instead.
func (m *accessLogMiddleware) remoteAddr(remoteAddr net.Addr, r *fasthttp.Request) (ret []byte) {
	if ret = r.Header.Peek(eighty.ForwardedForIPHeader); ret == nil {
		ret = []byte(remoteAddr.String())
	}
	return
}

func (m *accessLogMiddleware) remoteHost(remoteAddr net.Addr, r *fasthttp.Request) string {
	a := m.remoteAddr(remoteAddr, r)
	h := m.stripPort(strutil.B2S(a))
	if h != "" {
		return h
	}

	return "-"
}

// AccessLogMiddleware returns a routing.Middleware that handles error handling and access logging.
func AccessLogMiddleware(
	apiUrlPrefix string,
	logWriter io.WriteCloser,
	templateRenderer eighty.PageRenderer,
	logger logging.Logger) (handler routing.Middleware, closer func(), err error) {
	ctx, canceler := context.WithCancel(context.Background())
	inchan, outchan := q.NewStringQueue()
	impl := &accessLogMiddleware{
		writer:                    logWriter,
		apiUrlPrefix:              apiUrlPrefix,
		logInChan:                 inchan,
		logOutChan:                outchan,
		logger:                    logger,
		ctx:                       ctx,
		errorViewTemplateRenderer: templateRenderer,
	}
	impl.closer = func() {
		if ctx.Err() == nil {
			canceler()
		}
		impl.waitGroup.Wait()
	}

	go impl.lineByLineWriter()

	return impl.Handle, impl.Close, nil
}
