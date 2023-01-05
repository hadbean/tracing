package http

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/starblaze-ec/tracing/logger"
)

func init() {
	fmt.Println("设置环境变量 system.mode.debug = 1 开启日志debug模式")
}
func Post(url, contentType string, body io.Reader) (resp *http.Response, err error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return DoRequest(req, nil)
}

func PostJson(url string, data interface{}) (resp *http.Response, err error) {
	switch v := data.(type) {
	case []byte:
		{
			return Post(url, "application/json", bytes.NewBuffer(v))
		}
	case string:
		{
			return Post(url, "application/json", strings.NewReader(v))
		}
	default:
		{
			if bs, err := json.Marshal(&data); err != nil {
				return nil, err
			} else {
				return Post(url, "application/json", bytes.NewBuffer(bs))

			}
		}
	}
}

func Get(url string) (resp *http.Response, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	return DoRequest(req, nil)
}

func DoRequest(req *http.Request, cli *http.Client) (resp *http.Response, err error) {
	handleRequest(req)
	start := time.Now()
	defer func() {
		var layout = logger.LogLayout{}
		layout.Cost = time.Since(start)
		layout.Path = req.URL.String()
		layout.Query = req.URL.RawQuery
		layout.Method = req.Method
		var body []byte
		if req.Body != nil {
			body, _ = ioutil.ReadAll(req.Body)
			req.Body = ioutil.NopCloser(bytes.NewBuffer(body))
			layout.Body = string(body)
		}
		layout.Traceid = req.Header.Get("x-trace-id")
		layout.Time = start
		if err != nil {
			layout.Error = err.Error()
			layout.StatusCode = 400
		} else {
			if resp != nil {
				layout.StatusCode = resp.StatusCode
				var cnt []byte
				cnt, _ = ioutil.ReadAll(resp.Body)
				resp.Body = ioutil.NopCloser(bytes.NewBuffer(cnt))
				layout.Resp = string(cnt)
			}
		}

		if os.Getenv("system.mode.debug") == "1" { //debug模式下,打印所有请求,包括请求头
			if len(layout.Body) > 1000 {
				layout.Body = layout.Body[:1000]
			}
			s := fmt.Sprintf("[debug]%3d %13v %s [body]%s [response]%s [path]%s %s\n",
				layout.StatusCode,
				layout.Cost,
				layout.Method,
				layout.Body,
				layout.Resp,
				layout.Path,
				layout.Error,
			)
			logger.Log().Info(s)
		} else {
			s := fmt.Sprintf("%3d %13v %s %s %s\n",
				layout.StatusCode,
				layout.Cost,
				layout.Method,
				layout.Path,
				layout.Error,
			)
			logger.Log().Info(s)
		}
	}()
	if cli == nil {
		cli := http.Client{Timeout: time.Second * 7}
		if req.URL.Scheme == "https" { //跳过https验证
			cli.Transport = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
		}
		resp, err = cli.Do(req)
	} else {
		resp, err = cli.Do(req)
	}
	return
}
func handleRequest(req *http.Request) {
	traceId := logger.Goid()
	req.Header.Set("x-trace-id", traceId)
}

func PostForm(url string, data url.Values) (resp *http.Response, err error) {
	return Post(url, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
}

func PostFile(url string, file []byte) (resp *http.Response, err error) {
	return Post(url, "multipart/form-data", bytes.NewBuffer(file))
}
