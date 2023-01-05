package middleware

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/starblaze-ec/tracing/logger"
)

type GinLogger struct {
	// Filter 用户自定义过滤
	Filter func(c *gin.Context) bool
	// FilterKeyword 关键字过滤(key)
	FilterKeyword func(layout *logger.LogLayout) bool
	// AuthProcess 鉴权处理
	AuthProcess func(c *gin.Context, layout *logger.LogLayout)
	// 日志处理
	Print func(logger.LogLayout)
	// Source 服务唯一标识
	Source string
}

func (l GinLogger) SetLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		trace := c.Request.Header.Get("x-trace-id")
		if len(trace) > 0 { //如果请求有携带事务ID,则替换全局事务ID
			logger.SetGoid(trace)
		}
		var body []byte
		if l.Filter != nil && !l.Filter(c) {
			body, _ = c.GetRawData()
			// 将原body塞回去
			c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		}

		c.Next()
		cost := time.Since(start)
		layout := logger.LogLayout{
			Time:       time.Now(),
			Path:       path,
			Query:      query,
			IP:         c.ClientIP(),
			UserAgent:  c.Request.UserAgent(),
			Error:      strings.TrimRight(c.Errors.ByType(gin.ErrorTypePrivate).String(), "\n"),
			Cost:       cost,
			Method:     c.Request.Method,
			StatusCode: c.Writer.Status(),
			Traceid:    trace,
		}
		if l.Filter != nil && !l.Filter(c) {
			layout.Body = string(body)
		}
		// 处理鉴权需要的信息
		// l.AuthProcess(c, &layout)
		// l.AuthProcess(c, &layout)
		if l.FilterKeyword != nil {
			// 自行判断key/value 脱敏等
			l.FilterKeyword(&layout)
		}
		// 自行处理日志
		l.Print(layout)
		logger.Remove()
	}
}

const (
	green   = "\033[97;42m"
	white   = "\033[90;47m"
	yellow  = "\033[90;43m"
	red     = "\033[97;41m"
	blue    = "\033[97;44m"
	magenta = "\033[97;45m"
	cyan    = "\033[97;46m"
	reset   = "\033[0m"
)

func getStatusColor(code int) string {
	switch {
	case code >= http.StatusOK && code < http.StatusMultipleChoices:
		return green
	case code >= http.StatusMultipleChoices && code < http.StatusBadRequest:
		return white
	case code >= http.StatusBadRequest && code < http.StatusInternalServerError:
		return yellow
	default:
		return red
	}
}
func methodColor(method string) string {
	switch method {
	case http.MethodGet:
		return blue
	case http.MethodPost:
		return cyan
	case http.MethodPut:
		return yellow
	case http.MethodDelete:
		return red
	case http.MethodPatch:
		return green
	case http.MethodHead:
		return magenta
	case http.MethodOptions:
		return white
	default:
		return reset
	}
}

func DefaultGinLogger() gin.HandlerFunc {
	return GinLogger{

		Print: func(param logger.LogLayout) {
			// 标准输出,k8s做收集
			s := fmt.Sprintf("%s %3d %s %13v %15s %s %-7s %s %#v %s",
				getStatusColor(param.StatusCode), param.StatusCode, reset,
				param.Cost,
				param.IP,
				methodColor(param.Method), param.Method, reset,
				param.Path,
				param.Error,
			)
			logger.Log().Info(s)
		},
		Source: "Platform",
	}.SetLoggerMiddleware()
}
