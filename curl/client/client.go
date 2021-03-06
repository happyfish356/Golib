// Copyright 2016  cbping. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// client 包
// @author cbping
package client

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const (
	SlowReqRecord = "SlowReqRecord"
	ReqRecord = "ReqRecord"
	ErrorReqRecord = "ErrorReqRecord"
)

func init() {
	SetDefaultClient("", http.DefaultClient)
}

//----------------------------------------------------------------------------------------------------------------------

//
//  客户端
//  处理http请求
type Client struct {
	// 采用默认&http.Client{}
	*http.Client

	//
	UserAgent   string

	// 超过SlowReqLong时间长度的请求，将记录为慢请求
	// 默认为2秒
	SlowReqLong time.Duration

	// 函数参数
	// 记录信息；如日志记录
	Record      func(tag, msg string)

	// 和http.Client.Timeout相关
	Timeout     time.Duration

	//版本号
	Version     string
	//
	debug       bool
}

func (c *Client) SetDebug(debug bool) {
	c.debug = debug
}

func (c *Client) SetRecord(record func(tag, msg string)) {
	c.Record = record
}

func (c *Client) SetTimeOut(timeout time.Duration) {
	c.Timeout = timeout
}

func (c *Client) SetSlowReqLong(long time.Duration) {
	c.SlowReqLong = long
}

// 设置代理
// example:
//
//	func(req *http.Request) (*url.URL, error) {
// 		u, _ := url.ParseRequestURI("http://127.0.0.1:8118")
// 		return u, nil
// 	}
func (c *Client) SetProxy(proxy func(*http.Request) (*url.URL, error)) {
	//TODO::默认http.Client或者默认http.Transport时，是否值得改变代理（影响其他请求）？
	if nil != c.Client && nil != c.Client.Transport {
		c.Client.Transport.(*http.Transport).Proxy = proxy
	}

	return
}

// 处理请求
func (c *Client) DoRequest(req Request) (resp *Response, err error) {
	if nil == c.Client {
		c.Client = http.DefaultClient
	}

	defer func() {
		if nil != err {
			if (nil != c.Record) {
				c.Record(ErrorReqRecord, fmt.Sprintf("query:: %s error:: %v) ", req.String(), err))
			}
			err = clientError(err)
		}
	}()

	if nil == req {
		return nil, errors.New("Request is nil")
	}

	httpReq, err := req.HttpRequest()
	if nil != err {
		return nil, err
	}

	//必要头部信息设置
	httpReq.Header.Set("User-Agent", `Bping-Curl-` + c.UserAgent + "/" + c.Version)

	// 超时时间设置
	// XXX:并发下同步未做处理。
	// 会混乱被覆盖，未能按照预期执行
	// 建议client使用统一超时时间即可，
	// 不必细化到每一个request中去
	timeout := req.GetTimeOut()
	if timeout < 0 {
		timeout = c.Timeout
	}
	c.Client.Timeout = timeout

	t0 := time.Now()
	httpResp, err := c.Client.Do(httpReq)
	t1 := time.Now()
	if nil != err {
		return nil, err
	}
	resp = &Response{Response: httpResp}
	if nil != c.Record {
		resStr, _ := resp.Bytes()
		reqInfo := fmt.Sprintf("http query:: %s status:%d \n response:%s \n ts:(%v) \n", req.String(), httpResp.StatusCode, string(resStr), t1.Sub(t0))
		if t1.Sub(t0) >= c.SlowReqLong {
			c.Record(SlowReqRecord, reqInfo)
		}
		c.Record(ReqRecord, reqInfo)
	}
	return
}

func NewClient(title string, client *http.Client) *Client {
	return &Client{
		Client:      client,
		Version:     Version,
		UserAgent:   title,
		debug:       false,
		SlowReqLong: 2 * time.Second,
	}
}

func clientError(err error) error {
	if nil == err {
		return nil
	}
	return errors.New("Bping-Curl-Client-Failure:" + err.Error())
}

//----------------------------------------------------------------------------------------------------------------------

var DefaultClient *Client

// 设置DefaultClient
func SetDefaultClient(title string, client *http.Client) {
	DefaultClient = NewClient(title, client)
}

// 设置代理
// example:
//
//	func(req *http.Request) (*url.URL, error) {
// 		u, _ := url.ParseRequestURI("http://127.0.0.1:8118")
// 		return u, nil
// 	}
// 内部调用DefaultClient
func SetProxy(proxy func(*http.Request) (*url.URL, error)) {
	DefaultClient.SetProxy(proxy)
}

// 设置记录
// 内部调用DefaultClient
func SetRecord(record func(tag, msg string)) {
	DefaultClient.SetRecord(record)
}

// 设置慢请求时间限制
// 内部调用DefaultClient
func SetSlowReqLong(long time.Duration) {
	DefaultClient.SetSlowReqLong(long)
}

// 设置超时时间
// 内部调用DefaultClient
func SetTimeOut(timeout time.Duration) {
	DefaultClient.SetTimeOut(timeout)
}

// 处理请求，内部调用DefaultClient
func DoRequest(req Request) (*Response, error) {
	return DefaultClient.DoRequest(req)
}
