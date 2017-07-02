package gozilla

import (
	"log"
	"os"
	"time"
)

type Options struct {
	// ErrorLog specifies an optional logger for errors
	// Default to os.Stderr
	ErrorLog *log.Logger

	// EnableValidator controls enable validtor or not.
	// If true, the validator in `github.com/go-playground/validator`
	// will be used.
	// For validator syntax, see `https://godoc.org/gopkg.in/go-playground/validator.v9`.
	// Default true
	EnableValidator bool
}

func DefaultOptions() *Options {
	opt := new(Options)
	opt.ErrorLog = log.New(os.Stderr, "", log.LstdFlags)
	opt.EnableValidator = true
	return opt
}

type LogOptions struct {
	// 日志文件的目录，默认"log"
	BaseDir string

	// 日志文件的前缀, 默认"gozilla"
	Prefix string

	// 日志文件分隔的后缀，按照go的时间格式，默认 "20060102"
	Suffix string

	// 日志文件的内容字段
	// 字段有
	// {{.Remote}}    客户端IP
	// {{.Time}}      访问时间，格式:02/Jan/2006:15:04:05 -0700
	// {{.Method}}    HTTP方法
	// {{.Rawpath}}   带query string的访问路径
	// {{.Proto}}     HTTP协议版本
	// {{.Status}}    HTTP状态码
	// {{.UserAgent}} HTTP user agent
	// {{.Body}}      HTTP客户端发送的Body
	// {{.Used}}      HTTP请求使用的时间，按秒计算
	// 默认 {{.Remote}} [{{.Time}}] "{{.Method}} {{.Rawpath}} {{.Proto}}" {{.Status}} {{.UserAgent}} {{.Used}}
	Format string

	// 日志文件flush的间隔，默认3s
	FlushInterval time.Duration
}

func DefaultLogOptions() *LogOptions {
	opt := new(LogOptions)
	opt.BaseDir = "log"
	opt.Prefix = "gozilla"
	opt.Suffix = "20060102"
	opt.FlushInterval = time.Second * 3
	opt.Format = `{{.Remote}} [{{.Time}}] "{{.Method}} {{.Rawpath}} {{.Proto}}" {{.Status}} {{.UserAgent}} {{.Used}}`
	return opt
}
