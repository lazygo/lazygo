package libhttp

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"encoding/xml"
	"io"
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

// HttpClient
type HttpClient struct {
	url      string
	req      *http.Request
	params   url.Values
	files    map[string]string
	settings HttpSettings
	resp     *http.Response
	body     []byte
	dump     []byte
	lastErr  error
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

var defaultSetting = HttpSettings{
	UserAgent:        "lazygo",
	ConnectTimeout:   60 * time.Second,
	ReadWriteTimeout: 60 * time.Second,
	Gzip:             true,
	DumpBody:         true,
	Retries:          0, // 失败不重试
}

var defaultCookieJar http.CookieJar

// NewClient http客户端
func NewClient(rawUrl, method string) *HttpClient {
	client := &HttpClient{}
	u, err := url.Parse(rawUrl)
	if err != nil {
		client.lastErr = err
		return client
	}
	req := http.Request{
		URL:        u,
		Method:     method,
		Header:     make(http.Header),
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
	}
	client.url = rawUrl
	client.req = &req
	client.params = url.Values{}
	client.files = map[string]string{}
	client.settings = defaultSetting
	client.resp = &http.Response{}
	return client
}

// Settings 设置请求配置
func (h *HttpClient) Settings(settings HttpSettings) *HttpClient {
	h.settings = settings
	return h
}

// SetTimeout 设置超时时间
func (h *HttpClient) SetTimeout(connectTimeout, readWriteTimeout time.Duration) *HttpClient {
	h.settings.ConnectTimeout = connectTimeout
	h.settings.ReadWriteTimeout = readWriteTimeout
	return h
}

// SetRetries 设置请求失败重试次数 0:不重试(默认) -1:一直重试
func (h *HttpClient) SetRetries(times int) *HttpClient {
	h.settings.Retries = times
	return h
}

// SetTLSClientConfig sets tls connection configurations if visiting https url.
/*func (h *HttpRequest) SetTLSClientConfig(config *tls.Config) *HttpRequest {
	h.settings.TLSClientConfig = config
	return h
}*/

// SetHost 设置请求Host
func (h *HttpClient) SetHost(host string) *HttpClient {
	h.req.Host = host
	return h
}

// SetUserAgent 设置User-Agent
func (h *HttpClient) SetUserAgent(useragent string) *HttpClient {
	h.settings.UserAgent = useragent
	return h
}

// SetEnableCookie sets enable/disable cookiejar
func (h *HttpClient) SetEnableCookie(enable bool) *HttpClient {
	h.settings.EnableCookie = enable
	return h
}

// AddCookie 添加请求Cookie
func (h *HttpClient) AddCookie(cookie *http.Cookie) *HttpClient {
	h.req.Header.Add("Cookie", cookie.String())
	return h
}

// SetProxy 设置Http代理
func (h *HttpClient) SetProxy(proxy func(*http.Request) (*url.URL, error)) *HttpClient {
	h.settings.Proxy = proxy
	return h
}

// SetBasicAuth sets the request's Authorization header to use HTTP Basic Authentication with the provided username and password.
func (h *HttpClient) SetBasicAuth(username string, password string) *HttpClient {
	h.req.SetBasicAuth(username, password)
	return h
}

// SetHeader 设置请求Header
func (h *HttpClient) SetHeader(key, value string) *HttpClient {
	h.req.Header.Set(key, value)
	return h
}

// AddParam 增加请求参数
func (h *HttpClient) AddParam(key, value string) *HttpClient {
	h.params.Add(key, value)
	return h
}

// SetParam 设置请求参数
func (h *HttpClient) SetParam(key, value string) *HttpClient {
	h.params.Set(key, value)
	return h
}

// SetParams 设置键值对格式请求参数
func (h *HttpClient) SetParams(params map[string]any) *HttpClient {
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

// SetBody 设置请求Body  it supports string and []byte.
func (h *HttpClient) SetBody(data any) *HttpClient {
	switch t := data.(type) {
	case string:
		bf := bytes.NewBufferString(t)
		h.req.Body = io.NopCloser(bf)
		h.req.ContentLength = int64(len(t))
	case []byte:
		bf := bytes.NewBuffer(t)
		h.req.Body = io.NopCloser(bf)
		h.req.ContentLength = int64(len(t))
	}
	return h
}

// SetXmlBody 设置XML格式请求Body
func (h *HttpClient) SetXmlBody(obj any) *HttpClient {
	if h.req.Body == nil && obj != nil {
		xmlData, err := xml.Marshal(obj)
		if err != nil {
			h.lastErr = err
			return h
		}
		h.req.Body = io.NopCloser(bytes.NewReader(xmlData))
		h.req.ContentLength = int64(len(xmlData))
		h.req.Header.Set("Content-Type", "application/xml")
	}
	return h
}

// SetJsonBody 设置Json格式请求Body
func (h *HttpClient) SetJsonBody(obj any) *HttpClient {
	if h.req.Body == nil && obj != nil {
		jsonData, err := json.Marshal(obj)
		if err != nil {
			h.lastErr = err
			return h
		}
		h.req.Body = io.NopCloser(bytes.NewReader(jsonData))
		h.req.ContentLength = int64(len(jsonData))
		h.req.Header.Set("Content-Type", "application/json")
	}
	return h
}

// PostFile add a post file to the request
func (h *HttpClient) PostFile(formname, filename string) *HttpClient {
	h.files[formname] = filename
	return h
}

// ===================================

func (h *HttpClient) buildURL(paramBody string) {
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
			h.req.Body = io.NopCloser(pr)
			return
		}

		// with params
		if len(paramBody) > 0 {
			h.SetHeader("Content-Type", "application/x-www-form-urlencoded")
			h.SetBody(paramBody)
		}
	}
}

// doRequest 发起请求
func (h *HttpClient) doRequest() (*http.Response, error) {
	var paramBody = h.params.Encode()
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
		DialContext:         timeoutDialer(h.settings.ConnectTimeout, h.settings.ReadWriteTimeout),
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

// GetResponse 获取http.response
func (h *HttpClient) GetResponse() (*http.Response, error) {
	if h.lastErr != nil {
		return nil, h.lastErr
	}
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

func (h *HttpClient) ToBytes() ([]byte, error) {
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
		h.body, err = io.ReadAll(reader)
		return h.body, err
	}
	h.body, err = io.ReadAll(resp.Body)
	return h.body, err
}

// ToString 将相应转换成字符串
func (h *HttpClient) ToString() (string, error) {
	data, err := h.ToBytes()
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ToJSON 将响应作为json解析
func (h *HttpClient) ToJSON(v any) error {
	data, err := h.ToBytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// ToXML returns the map that marshals from the body bytes as xml in response .
// it calls Response inner.
func (h *HttpClient) ToXML(v any) error {
	data, err := h.ToBytes()
	if err != nil {
		return err
	}
	return xml.Unmarshal(data, v)
}

// ToFile saves the body data in response to one file.
// it calls Response inner.
func (h *HttpClient) ToFile(filename string) error {
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
