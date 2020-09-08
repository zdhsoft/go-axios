// Copyright 2019 tree xie
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package axios

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	HT "github.com/vicanso/http-trace"
)

type (
	// OnError on error function
	OnError func(err error, config *Config) (newErr error)
	// OnDone on done event
	OnDone func(config *Config, resp *Response, err error)
	// BeforeNewRequest before new request
	BeforeNewRequest func(config *Config) (err error)
	// Config http request config
	Config struct {
		Request  *http.Request
		Response *Response
		// Route the request route
		Route string
		// URL the request url
		URL string
		// Method http request method, default is `get`
		Method string
		// BaseURL http request base url
		BaseURL string
		// TransformRequest transform requset body
		TransformRequest []TransformRequest
		// TransformResponse transofrm response body
		TransformResponse []TransformResponse
		// Headers  custom headers for request
		Headers http.Header
		// Params params for request route
		Params map[string]string
		// Query query for requset
		Query url.Values

		// Body the request body
		Body interface{}

		// Concurrency current amount handling request of instance
		Concurrency uint32

		// Timeout request timeout
		Timeout time.Duration

		// Context context
		Context context.Context

		// Client http client
		Client *http.Client
		// Adapter custom handling of requset
		Adapter Adapter
		// RequestInterceptors request interceptor list
		RequestInterceptors []RequestInterceptor
		// ResponseInterceptors response interceptor list
		ResponseInterceptors []ResponseInterceptor

		// OnError on error function
		OnError OnError
		// OnDone on done event
		OnDone OnDone
		// BeforeNewRequest before new request
		BeforeNewRequest BeforeNewRequest

		HTTPTrace   *HT.HTTPTrace
		enableTrace bool
		data        map[string]interface{}
	}
	// InstanceConfig config of instance
	InstanceConfig struct {
		// BaseURL http request base url
		BaseURL string
		// TransformRequest transform requset body
		TransformRequest []TransformRequest
		// TransformResponse transofrm response body
		TransformResponse []TransformResponse
		// Headers  custom headers for request
		Headers http.Header
		// Timeout request timeout
		Timeout time.Duration

		// Client http client
		Client *http.Client
		// Adapter custom adapter
		Adapter Adapter

		// RequestInterceptors request interceptor list
		RequestInterceptors []RequestInterceptor
		// ResponseInterceptors response interceptor list
		ResponseInterceptors []ResponseInterceptor

		// EnableTrace enable http trace
		EnableTrace bool
		// OnError on error function
		OnError OnError
		// OnDone on done event
		OnDone OnDone
		// BeforeNewRequest before new request
		BeforeNewRequest BeforeNewRequest
	}
)

// Get get value from config
func (conf *Config) Get(key string) interface{} {
	if conf.data == nil {
		return nil
	}
	return conf.data[key]
}

// Set set value to config
func (conf *Config) Set(key string, value interface{}) {
	if conf.data == nil {
		conf.data = make(map[string]interface{})
	}
	conf.data[key] = value
}

// GetString get string value
func (conf *Config) GetString(key string) string {
	v := conf.Get(key)
	if v == nil {
		return ""
	}
	str, ok := v.(string)
	if !ok {
		return ""
	}
	return str
}

// GetBool get bool value
func (conf *Config) GetBool(key string) bool {
	v := conf.Get(key)
	if v == nil {
		return false
	}
	b, ok := v.(bool)
	if !ok {
		return false
	}
	return b
}

// GetInt get int value
func (conf *Config) GetInt(key string) int {
	v := conf.Get(key)
	if v == nil {
		return 0
	}
	i, ok := v.(int)
	if !ok {
		return 0
	}
	return i
}

// AddQuery add query
func (conf *Config) AddQuery(key, value string) *Config {
	if conf.Query == nil {
		conf.Query = make(url.Values)
	}
	conf.Query.Add(key, value)
	return conf
}

// AddParam add param
func (conf *Config) AddParam(key, value string) *Config {
	if conf.Params == nil {
		conf.Params = make(map[string]string)
	}
	conf.Params[key] = value
	return conf
}
func urlJoin(basicURL, url string) string {
	if basicURL == "" ||
		strings.HasPrefix(url, "http://") ||
		strings.HasPrefix(url, "https://") {
		return url
	}
	if strings.HasSuffix(basicURL, "/") && strings.HasPrefix(url, "/") {
		return basicURL + url[1:]
	}
	return basicURL + url
}

// getURL generate the url of request config
func (conf *Config) getURL() string {
	url := urlJoin(conf.BaseURL, conf.URL)
	if conf.Params != nil {
		for key, value := range conf.Params {
			url = strings.ReplaceAll(url, ":"+key, value)
		}
	}

	if conf.Query != nil {
		if strings.Contains(url, "?") {
			url += ("&" + conf.Query.Encode())
		} else {
			url += ("?" + conf.Query.Encode())
		}
	}
	return url
}

// getRequestBody get requet body
func (conf *Config) getRequestBody() (r io.Reader, err error) {
	if conf.Body == nil || !isNeedToTransformRequestBody(conf.Method) {
		return
	}
	data := conf.Body
	for _, fn := range conf.TransformRequest {
		buf, e := fn(data, conf.Headers)
		if e != nil {
			err = e
			return
		}
		data = buf
	}
	r = bytes.NewReader(data.([]byte))
	return
}

// CURL convert config to curl
func (conf *Config) CURL() string {
	builder := new(strings.Builder)
	builder.WriteString(fmt.Sprintf("curl -X%s ", conf.Method))

	r, _ := conf.getRequestBody()
	if r != nil {
		buf, _ := ioutil.ReadAll(r)
		builder.WriteString(fmt.Sprintf(`-d '%s' `, string(buf)))
	}

	for key, values := range conf.Headers {
		for _, value := range values {
			builder.WriteString(fmt.Sprintf(`-H '%s:%s' `, key, value))
		}
	}

	builder.WriteString(fmt.Sprintf(`'%s'`, conf.getURL()))

	return builder.String()
}
