package client

import (
	"errors"
	"fmt"
	"net/http"
	"time"
)

//
//  请求接口
//    所有请求对象继承的接口，也是Client接受处理的请求接口
type Request interface {
	//返回*http.Request
	HttpRequest() (*http.Request, error)
	//返回请求相关内容格式化字符串
	String() string
	// 获取超时时间
	// =0代表此请求不启用超时设置
	// <0代表默认使用全局
	// >0代表自定义超时时间
	GetTimeOut() time.Duration
	//克隆
	Clone() interface{}
}

type BaseRequest struct {
}

func (b *BaseRequest) HttpRequest() (*http.Request, error) {
	return nil, errors.New("Implement Interface's Method::HttpRequest")
}

func (b *BaseRequest) String() string {
	return fmt.Sprintf("Request:%v", b)
}

func (b *BaseRequest) GetTimeOut() time.Duration {
	return -1
}

func (b *BaseRequest) Clone() interface{} {
	new_obj := (*b)
	return &new_obj
}
