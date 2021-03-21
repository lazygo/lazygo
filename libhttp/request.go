package libhttp

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"encoding/xml"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"
)

//
type HttpRequest struct {
	url      string
	req      *http.Request
	params   url.Values
	files    map[string]string
	settings HttpSettings
	resp     *http.Response
	body     []byte
	dump     []byte
}

// http.Client settings
type HttpSettings struct {
	ShowDebug        bool
	UserAgent        string
	ConnectTimeout   time.Duration
	ReadWriteTimeout time.Duration
	// TLSClientConfig  *tls.Config
	Proxy         func(*http.Request) (*url.URL, error)
	CheckRedirect func(req *http.Request, via []*http.Request) error
	EnableCookie  bool
	Gzip          bool
	DumpBody      bool
	Retries       int // if set to -1 means will retry forever
}

//
var defaultSetting = HttpSettings{
	UserAgent:        "lazygo",
	ConnectTimeout:   60 * time.Second,
	ReadWriteTimeout: 60 * time.Second,
	Gzip:             true,
	DumpBody:         true,
	Retries:          0, // 失败不重试
}

var defaultCookieJar http.CookieJar

//
func NewRequest(rawUrl, method string) *HttpRequest {
	var resp http.Response
	u, err := url.Parse(rawUrl)
	checkError(err)
	req := http.Request{
		URL:        u,
		Method:     method,
		Header:     make(http.Header),
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
	}
	return &HttpRequest{
		url:      rawUrl,
		req:      &req,
		params:   url.Values{},
		files:    map[string]string{},
		settings: defaultSetting,
		resp:     &resp,
	}
}

// 设置请求配置
func (h *HttpRequest) Settings(settings HttpSettings) *HttpRequest {
	h.settings = settings
	return h
}

// 设置超时时间
func (h *HttpRequest) SetTimeout(connectTimeout, readWriteTimeout time.Duration) *HttpRequest {
	h.settings.ConnectTimeout = connectTimeout
	h.settings.ReadWriteTimeout = readWriteTimeout
	return h
}

// 设置请求失败重试次数 0:不重试(默认) -1:一直重试
func (h *HttpRequest) SetRetries(times int) *HttpRequest {
	h.settings.Retries = times
	return h
}

// SetTLSClientConfig sets tls connection configurations if visiting https url.
/*func (h *HttpRequest) SetTLSClientConfig(config *tls.Config) *HttpRequest {
	h.settings.TLSClientConfig = config
	return h
}*/

// 设置请求Host
func (h *HttpRequest) SetHost(host string) *HttpRequest {
	h.req.Host = host
	return h
}

// 设置User-Agent
func (h *HttpRequest) SetUserAgent(useragent string) *HttpRequest {
	h.settings.UserAgent = useragent
	return h
}

// SetEnableCookie sets enable/disable cookiejar
func (h *HttpRequest) SetEnableCookie(enable bool) *HttpRequest {
	h.settings.EnableCookie = enable
	return h
}

// 添加请求Cookie
func (h *HttpRequest) AddCookie(cookie *http.Cookie) *HttpRequest {
	h.req.Header.Add("Cookie", cookie.String())
	return h
}

// 设置Http代理
func (h *HttpRequest) SetProxy(proxy func(*http.Request) (*url.URL, error)) *HttpRequest {
	h.settings.Proxy = proxy
	return h
}

// SetBasicAuth sets the request's Authorization header to use HTTP Basic Authentication with the provided username and password.
func (h *HttpRequest) SetBasicAuth(username string, password string) *HttpRequest {
	h.req.SetBasicAuth(username, password)
	return h
}

// 设置请求Header
func (h *HttpRequest) SetHeader(key, value string) *HttpRequest {
	h.req.Header.Set(key, value)
	return h
}

// 增加请求参数
func (h *HttpRequest) AddParam(key, value string) *HttpRequest {
	h.params.Add(key, value)
	return h
}

// 设置请求参数
func (h *HttpRequest) SetParam(key, value string) *HttpRequest {
	h.params.Set(key, value)
	return h
}

// 设置键值对格式请求参数
func (h *HttpRequest) SetParams(params map[string]interface{}) *HttpRequest {
	h.params = url.Values{}
	for k, v := range params {
		switch v.(type) {
		case []string:
			for _, vv := range v.([]string) {
				h.params.Set(k, vv)
			}
		default:
			h.params.Add(k, toString(v))
		}
	}
	return h
}

// 设置请求Body
// it supports string and []byte.
func (h *HttpRequest) SetBody(data interface{}) *HttpRequest {
	switch t := data.(type) {
	case string:
		bf := bytes.NewBufferString(t)
		h.req.Body = ioutil.NopCloser(bf)
		h.req.ContentLength = int64(len(t))
	case []byte:
		bf := bytes.NewBuffer(t)
		h.req.Body = ioutil.NopCloser(bf)
		h.req.ContentLength = int64(len(t))
	}
	return h
}

// 设置XML格式请求Body
func (h *HttpRequest) SetXmlBody(obj interface{}) (*HttpRequest, error) {
	if h.req.Body == nil && obj != nil {
		xmlData, err := xml.Marshal(obj)
		if err != nil {
			return h, err
		}
		h.req.Body = ioutil.NopCloser(bytes.NewReader(xmlData))
		h.req.ContentLength = int64(len(xmlData))
		h.req.Header.Set("Content-Type", "application/xml")
	}
	return h, nil
}

// 设置Json格式请求Body
func (h *HttpRequest) SetJsonBody(obj interface{}) (*HttpRequest, error) {
	if h.req.Body == nil && obj != nil {
		jsonData, err := json.Marshal(obj)
		if err != nil {
			return h, err
		}
		h.req.Body = ioutil.NopCloser(bytes.NewReader(jsonData))
		h.req.ContentLength = int64(len(jsonData))
		h.req.Header.Set("Content-Type", "application/json")
	}
	return h, nil
}

// PostFile add a post file to the request
func (h *HttpRequest) PostFile(formname, filename string) *HttpRequest {
	h.files[formname] = filename
	return h
}

// ===================================

//
func (h *HttpRequest) buildURL(paramBody string) {
	// build GET url with query string
	if h.req.Method == "GET" && len(paramBody) > 0 {
		if strings.Contains(h.url, "?") {
			h.url += "&" + paramBody
		} else {
			h.url = h.url + "?" + paramBody
		}
		return
	}

	// build POST/PUT/PATCH url and body
	if (h.req.Method == "POST" || h.req.Method == "PUT" || h.req.Method == "PATCH" || h.req.Method == "DELETE") && h.req.Body == nil {
		// with files
		if len(h.files) > 0 {
			pr, pw := io.Pipe()
			bodyWriter := multipart.NewWriter(pw)
			recoverGo(func() {
				for formname, filename := range h.files {
					fileWriter, err := bodyWriter.CreateFormFile(formname, filename)
					checkError(err)
					fh, err := os.Open(filename)
					checkError(err)
					//iocopy
					_, err = io.Copy(fileWriter, fh)
					_ = fh.Close()
					checkError(err)
				}
				for k, v := range h.params {
					for _, vv := range v {
						_ = bodyWriter.WriteField(k, vv)
					}
				}
				_ = bodyWriter.Close()
				_ = pw.Close()
			})
			h.SetHeader("Content-Type", bodyWriter.FormDataContentType())
			h.req.Body = ioutil.NopCloser(pr)
			return
		}

		// with params
		if len(paramBody) > 0 {
			h.SetHeader("Content-Type", "application/x-www-form-urlencoded")
			h.SetBody(paramBody)
		}
	}
}

// 发起请求
func (h *HttpRequest) doRequest() (*http.Response, error) {
	var paramBody string = h.params.Encode()
	h.buildURL(paramBody)

	reqUrl, err := url.Parse(h.url)
	if err != nil {
		return nil, err
	}
	h.req.URL = reqUrl

	// create default transport
	trans := &http.Transport{
		// TLSClientConfig:     h.settings.TLSClientConfig,
		Proxy:               h.settings.Proxy,
		Dial:                timeoutDialer(h.settings.ConnectTimeout, h.settings.ReadWriteTimeout),
		MaxIdleConnsPerHost: -1,
	}

	var jar http.CookieJar
	if h.settings.EnableCookie {
		if defaultCookieJar == nil {
			defaultCookieJar, _ = cookiejar.New(nil)
		}
		jar = defaultCookieJar
	}

	client := &http.Client{
		Transport: trans,
		Jar:       jar,
	}

	// 设置User-Agent
	if h.settings.UserAgent != "" && h.req.Header.Get("User-Agent") == "" {
		h.req.Header.Set("User-Agent", h.settings.UserAgent)
	}

	if h.settings.CheckRedirect != nil {
		client.CheckRedirect = h.settings.CheckRedirect
	}

	if h.settings.ShowDebug {
		dump, err := httputil.DumpRequest(h.req, h.settings.DumpBody)
		if err != nil {
			log.Println(err.Error())
		}
		h.dump = dump
	}
	// retries default value is 0, it will run once.
	// retries equal to -1, it will run forever until success
	// retries is setted, it will retries fixed times.
	var resp *http.Response
	for i := 0; h.settings.Retries == -1 || i <= h.settings.Retries; i++ {
		resp, err = client.Do(h.req)
		if err == nil {
			break
		}
	}
	return resp, err
}

// 获取http.response
func (h *HttpRequest) GetResponse() (*http.Response, error) {
	if h.resp.StatusCode != 0 {
		return h.resp, nil
	}
	resp, err := h.doRequest()
	if err != nil {
		return nil, err
	}
	h.resp = resp
	return resp, nil
}

func (h *HttpRequest) ToBytes() ([]byte, error) {
	if h.body != nil {
		return h.body, nil
	}
	resp, err := h.GetResponse()
	if err != nil {
		return nil, err
	}
	if resp.Body == nil {
		return nil, nil
	}
	defer resp.Body.Close()
	if h.settings.Gzip && resp.Header.Get("Content-Encoding") == "gzip" {
		reader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, err
		}
		h.body, err = ioutil.ReadAll(reader)
		return h.body, err
	}
	h.body, err = ioutil.ReadAll(resp.Body)
	return h.body, err
}

// 将相应转换成字符串
func (h *HttpRequest) ToString() (string, error) {
	data, err := h.ToBytes()
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// 将响应作为json解析
func (h *HttpRequest) ToJSON(v interface{}) error {
	data, err := h.ToBytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// ToXML returns the map that marshals from the body bytes as xml in response .
// it calls Response inner.
func (h *HttpRequest) ToXML(v interface{}) error {
	data, err := h.ToBytes()
	if err != nil {
		return err
	}
	return xml.Unmarshal(data, v)
}

// ToFile saves the body data in response to one file.
// it calls Response inner.
func (h *HttpRequest) ToFile(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	resp, err := h.GetResponse()
	if err != nil {
		return err
	}
	if resp.Body == nil {
		return nil
	}
	defer resp.Body.Close()
	_, err = io.Copy(f, resp.Body)
	return err
}
