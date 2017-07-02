package gozilla

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"
)

var (
	bufferPool = sync.Pool{
		New: func() interface{} { return new(bytes.Buffer) },
	}
)

type loggingRequest struct {
	buf *bytes.Buffer
}

func newLoggingRequest(r *http.Request) *loggingRequest {
	l := &loggingRequest{
		buf: bufferPool.Get().(*bytes.Buffer),
	}
	l.buf.Reset()
	r.Body = ioutil.NopCloser(io.TeeReader(r.Body, l.buf))
	return l
}

func (l *loggingRequest) Body() []byte {
	b := l.buf.Bytes()
	bufferPool.Put(l.buf)
	return b
}

type loggingResponse struct {
	http.ResponseWriter
	status int
}

func newLoggingResponse(w http.ResponseWriter) *loggingResponse {
	return &loggingResponse{
		ResponseWriter: w,
	}
}

func (l *loggingResponse) WriteHeader(status int) {
	l.status = status
	l.ResponseWriter.WriteHeader(status)
}

func (l *loggingResponse) Status() int {
	return l.status
}

type logRecord struct {
	time      time.Time
	Time      string
	Remote    string
	Method    string
	Rawpath   string
	Status    int
	UserAgent string
	Refer     string
	Proto     string
	Body      string
	Used      float32
}

type logger struct {
	opt *LogOptions

	logtpl *template.Template

	next http.Handler

	lastFileName string
	logFile      *os.File

	writerLock    sync.Mutex
	logFileWriter *bufio.Writer

	ch chan *logRecord
}

func NewLoggerHandler(opt *LogOptions, h http.Handler) http.Handler {
	l := &logger{
		opt:    opt,
		logtpl: template.Must(template.New("log").Parse(opt.Format)),
		next:   h,
		ch:     make(chan *logRecord, 1000),
	}
	err := os.MkdirAll(opt.BaseDir, 0755)
	if err != nil {
		panic(err)
	}

	// test log format
	err = l.logtpl.Execute(ioutil.Discard, new(logRecord))
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			l.logRecord(<-l.ch)
		}
	}()

	go l.flushDaemon()

	return l
}

func compactBody(src string) string {
	return strings.Replace(string(src), "\n", " ", -1)
}

func (l *logger) flushDaemon() {
	tick := time.NewTicker(l.opt.FlushInterval)
	defer tick.Stop()
	for range tick.C {
		l.writerLock.Lock()
		if l.logFileWriter != nil {
			l.logFileWriter.Flush()
		}
		l.writerLock.Unlock()
	}
}

func (l *logger) logRecord(lr *logRecord) {
	fileName := fmt.Sprintf("%s.%s.log", l.opt.Prefix, time.Now().Format(l.opt.Suffix))
	fullName := filepath.Join(l.opt.BaseDir, fileName)
	if fullName > l.lastFileName {
		if l.logFile != nil {
			l.logFile.Close()
		}
		var err error
		l.logFile, err = os.OpenFile(fullName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			log.Printf("Error opening %q: %v", fullName, err)
			return
		}
		l.writerLock.Lock()
		l.logFileWriter = bufio.NewWriter(l.logFile)
		l.writerLock.Unlock()
		l.lastFileName = fullName
	}

	lr.Body = compactBody(lr.Body)
	if len(lr.Body) == 0 {
		lr.Body = "-"
	}

	lr.Time = lr.time.Format("02/Jan/2006:15:04:05 -0700")
	lr.UserAgent = strconv.Quote(lr.UserAgent)
	if lr.Status == 0 {
		lr.Status = http.StatusOK
	}

	if l.logFile != nil {
		l.writerLock.Lock()
		err := l.logtpl.Execute(l.logFileWriter, lr)
		if err != nil {
			log.Printf("execute errror:%s", err)
		}
		l.logFileWriter.WriteByte('\n')
		l.writerLock.Unlock()
	}
}

func (l *logger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rr := newLoggingRequest(r)
	ww := newLoggingResponse(w)
	begin := time.Now()
	l.next.ServeHTTP(ww, r)

	addr := r.RemoteAddr
	if colon := strings.LastIndex(addr, ":"); colon != -1 {
		addr = addr[:colon]
	}

	l.ch <- &logRecord{
		time:      time.Now(),
		Remote:    addr,
		Method:    r.Method,
		Rawpath:   r.URL.RequestURI(),
		UserAgent: r.UserAgent(),
		Refer:     r.Referer(),
		Status:    ww.Status(),
		Proto:     r.Proto,
		Body:      string(rr.Body()),
		Used:      float32(time.Now().Sub(begin).Seconds()),
	}
}
