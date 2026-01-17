package server

import (
	"bufio"
	"net"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

type testEmbedStruct struct {
	Name string `json:"name" bind:"query,form"`
}

type testBindStruct struct {
	UID     int    `json:"uid" bind:"ctx"`
	Name    string `json:"name" bind:"query,form"`
	Age     int    `json:"age" bind:"query,form"`
	Email   string `json:"email" bind:"header"`
	Session string `json:"session" bind:"cookie"`
	ID      string `json:"id" bind:"param"`
	IDs     []int  `json:"ids" bind:"form"`
}

type testNestedStruct struct {
	Basic    testBindStruct  `json:"basic" bind:"form"`
	Embed    testEmbedStruct `json:"embed" bind:"form"`
	Pointer  *testBindStruct `json:"pointer" bind:"form"`
	NoTag    testBindStruct
	NoExport testBindStruct `json:"-"`
}

func TestContextBind(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		setup       func(*context)
		input       string
		want        testBindStruct
		wantErr     bool
	}{
		{
			name:        "bind json",
			contentType: MIMEApplicationJSON,
			setup:       func(c *context) {},
			input:       `{"uid":12, "name":"test","age":18}`,
			want: testBindStruct{
				UID:  0,
				Name: "test",
				Age:  18,
			},
		},
		{
			name:        "bind form",
			contentType: MIMEApplicationForm,
			setup: func(c *context) {
				c.request.Form = map[string][]string{
					"name": {"test"},
					"age":  {"18"},
				}
			},
			want: testBindStruct{
				Name: "test",
				Age:  18,
			},
		},
		{
			name:        "bind form array",
			contentType: MIMEApplicationForm,
			setup: func(c *context) {
				c.request.Form = map[string][]string{
					"name": {"test"},
					"ids":  {"18,19,20"},
				}
			},
			want: testBindStruct{
				Name: "test",
				IDs:  []int{18, 19, 20},
			},
		},
		{
			name:        "bind form json",
			contentType: MIMEApplicationForm,
			setup: func(c *context) {
				c.request.Form = map[string][]string{
					"name": {"test"},
					"ids":  {`[18,19,20]`},
				}
			},
			want: testBindStruct{
				Name: "test",
				IDs:  []int{18, 19, 20},
			},
		},
		{
			name:        "bind query",
			contentType: MIMEApplicationForm,
			setup: func(c *context) {
				c.request.URL.RawQuery = "name=test&age=18"
			},
			want: testBindStruct{
				Name: "test",
				Age:  18,
			},
		},
		{
			name:        "bind header",
			contentType: MIMEApplicationForm,
			setup: func(c *context) {
				c.request.Header.Set("email", "test@example.com")
			},
			want: testBindStruct{
				Email: "test@example.com",
			},
		},
		{
			name:        "bind cookie",
			contentType: MIMEApplicationForm,
			setup: func(c *context) {
				c.request.AddCookie(&http.Cookie{Name: "session", Value: "123456"})
			},
			want: testBindStruct{
				Session: "123456",
			},
		},
		{
			name:        "bind param",
			contentType: MIMEApplicationForm,
			setup: func(c *context) {
				c.pnames = []string{"id"}
				c.pvalues = []string{"abc123"}
			},
			want: testBindStruct{
				ID: "abc123",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.input))
			r.Header.Set(HeaderContentType, tt.contentType)

			c := &context{
				request:        r,
				responseWriter: NewResponseWriter(w),
				server:         New(),
			}

			tt.setup(c)

			var got testBindStruct
			err := c.Bind(&got)

			if (err != nil) != tt.wantErr {
				t.Errorf("context.Bind() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("context.Bind() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContextBindNested(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		setup       func(*context)
		input       string
		want        testNestedStruct
		wantErr     bool
	}{
		{
			name:        "bind nested json",
			contentType: MIMEApplicationJSON,
			setup:       func(c *context) {},
			input:       `{"name":"li","basic":{"name":"test","age":18},"embed":{"name":"embed"},"pointer":{"name":"ptr","age":20}}`,
			want: testNestedStruct{
				Basic: testBindStruct{
					Name: "test",
					Age:  18,
				},
				Embed: testEmbedStruct{
					Name: "embed",
				},
				NoTag: testBindStruct{
					Name: "li",
				},
				Pointer: &testBindStruct{
					Name: "ptr",
					Age:  20,
				},
			},
		},
		{
			name:        "bind nested form",
			contentType: MIMEApplicationForm,
			setup: func(c *context) {
				c.request.Form = map[string][]string{
					"basic":   {`{"name":"test","age":18}`},
					"embed":   {`{"name":"embed"}`},
					"pointer": {`{"name":"ptr","age":20}`},
				}
			},
			want: testNestedStruct{
				Basic: testBindStruct{
					Name: "test",
					Age:  18,
				},
				Embed: testEmbedStruct{
					Name: "embed",
				},
				Pointer: &testBindStruct{
					Name: "ptr",
					Age:  20,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.input))
			r.Header.Set(HeaderContentType, tt.contentType)

			c := &context{
				request:        r,
				responseWriter: NewResponseWriter(w),
				server:         New(),
			}

			tt.setup(c)

			var got testNestedStruct
			err := c.Bind(&got)

			if (err != nil) != tt.wantErr {
				t.Errorf("context.Bind() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got.Basic, tt.want.Basic) {
				t.Errorf("context.Bind() Basic = %v, want %v", got.Basic, tt.want.Basic)
			}
			if got.Embed != tt.want.Embed {
				t.Errorf("context.Bind() Embed = %v, want %v", got.Embed, tt.want.Embed)
			}
			if got.Pointer == nil || !reflect.DeepEqual(*got.Pointer, *tt.want.Pointer) {
				t.Errorf("context.Bind() Pointer = %v, want %v", got.Pointer, tt.want.Pointer)
			}
		})
	}
}

type hrw struct {
}

func (h *hrw) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (h *hrw) WriteHeader(statusCode int) {
}

func (h *hrw) Header() http.Header {
	return http.Header{}
}

func (h *hrw) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, nil
}

func (h *hrw) Flush() {
}

func TestContextResponseWriter(t *testing.T) {

	resp := NewResponseWriter(&hrw{})
	resp.Write([]byte("test"))
	resp.WriteHeader(http.StatusOK)
	resp.Flush()
	resp.Hijack()
	resp.Write([]byte("test"))
	resp.WriteHeader(http.StatusOK)
	resp.Flush()
	resp.Hijack()
}
