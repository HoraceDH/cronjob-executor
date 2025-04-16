package httpclients

// Created in 2025-03-19 08:41.
// @author Horace

import (
	"bytes"
	"encoding/json"
	"fmt"
	logger "github.com/cihub/seelog"
	"github.com/horacedh/cronjob-executor/utils"
	"github.com/horacedh/cronjob-executor/webresult"
	"io"
	"net/http"
	"net/http/cookiejar"
	neturl "net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

// 单例模式
var (
	httpClient     HttpClient
	httpClientOnce sync.Once
)

type Options struct {
	// 超时时间，毫秒
	Timeout time.Duration
	// 签名Key
	SignKey string
}

// HttpResult 响应结果
type HttpResult struct {
	Status        int
	Body          []byte
	Header        http.Header
	Cookies       map[string]*http.Cookie
	ContentLength int64
	Err           error
	// 请求耗时，毫秒
	Elapsed   int64
	MsgObject webresult.MsgObject
}

// IsStatusSuccess 是否成功，没有错误，状态码在200-300之间，是否业务成功需要单独判断响应体
func (result *HttpResult) IsStatusSuccess() bool {
	if result.Status == 200 || result.Status == 201 || result.Status == 302 {
		if result.Err == nil {
			return true
		}
	}
	return false
}

// IsSuccess 是否成功，状态成功，响应体成功
func (result *HttpResult) IsSuccess() bool {
	if result.Status == 200 || result.Status == 201 || result.Status == 302 {
		if result.Err == nil {
			msgObject, err := result.toMsgObject()
			if err != nil {
				return false
			}
			return msgObject.Code == webresult.SUCCESS.Code
		}
	}
	return false
}

// ToString 将结果转换为字符串
func (result *HttpResult) ToString() (string, error) {
	if result.IsStatusSuccess() {
		return string(result.Body), nil
	}
	return "", result.Err
}

// ToUniDecodeString 将结果转换为字符串
func (result *HttpResult) ToUniDecodeString() (string, error) {
	if result.IsStatusSuccess() {
		return utils.UniDecode(string(result.Body)), nil
	}
	return "", result.Err
}

// ToMap 将结果转换为map
func (result *HttpResult) ToMap() (map[string]interface{}, error) {
	if !result.IsStatusSuccess() {
		return nil, result.Err
	}
	var jsonMap map[string]interface{}
	err := json.Unmarshal(result.Body, &jsonMap)
	if err != nil {
		return nil, err
	}
	return jsonMap, nil
}

// ToMapArray 将结果转换为map数组
func (result *HttpResult) ToMapArray() ([]map[string]interface{}, error) {
	if !result.IsStatusSuccess() {
		return nil, result.Err
	}
	var jsonMap []map[string]interface{}
	err := json.Unmarshal(result.Body, &jsonMap)
	if err != nil {
		return nil, err
	}
	return jsonMap, nil
}

// toMsgObject 将结果转换为MsgObject
func (result *HttpResult) toMsgObject() (*webresult.MsgObject, error) {
	if !result.IsStatusSuccess() {
		return nil, result.Err
	}
	var msgObject = &webresult.MsgObject{}
	err := json.Unmarshal(result.Body, msgObject)
	if err != nil {
		return nil, err
	}
	result.MsgObject = *msgObject
	return msgObject, nil
}

// HttpClient 接口
type HttpClient interface {
	// Request 发送HTTP请求，返回结果
	Request(url string, method string, headers map[string]interface{}, params map[string]interface{}, body io.Reader) HttpResult
	// GetRequest 发送Get请求，返回结果
	GetRequest(url string, headers map[string]interface{}, params map[string]interface{}, body io.Reader) HttpResult
	// PostRequest 发送Post请求，返回结果
	PostRequest(url string, headers map[string]interface{}, params map[string]interface{}, body io.Reader) HttpResult
}

// HttpClientImpl 实现类
type httpClientImpl struct {
	Options *Options
	client  *http.Client
}

// GetRequest 发送Get请求，返回结果
func (httpClient *httpClientImpl) GetRequest(url string, headers map[string]interface{}, params map[string]interface{}, body io.Reader) HttpResult {
	return httpClient.Request(url, "GET", headers, params, body)
}

// PostRequest 发送Post请求，返回结果
func (httpClient *httpClientImpl) PostRequest(url string, headers map[string]interface{}, params map[string]interface{}, body io.Reader) HttpResult {
	return httpClient.Request(url, "POST", headers, params, body)
}

// encodeParams 编码参数
func encodeParams(params map[string]interface{}) string {
	if params != nil {
		var form = neturl.Values{}
		for key, value := range params {
			switch value.(type) {
			case string:
				form.Set(key, value.(string))
				break
			case int:
				form.Set(key, strconv.Itoa(value.(int)))
			}
		}
		return form.Encode()
	}
	return ""
}

// Request 发送HTTP请求，返回字节数组
func (httpClient *httpClientImpl) Request(url string, method string, headers map[string]interface{}, params map[string]interface{}, body io.Reader) HttpResult {
	now := time.Now()
	// 设置参数
	paramString := encodeParams(params)
	var bodyString = ""
	if body != nil {
		bodyBytes, _ := io.ReadAll(body)
		bodyString = string(bodyBytes)
		body = bytes.NewReader(bodyBytes)
	}

	var request *http.Request
	var err error
	if strings.EqualFold("POST", method) && body == nil {
		request, err = http.NewRequest(method, url, strings.NewReader(paramString))
	} else {
		request, err = http.NewRequest(method, url, body)
	}
	request.URL.RawQuery = paramString
	if err != nil {
		logger.Errorf("create request error, method: %s, url: %s, headers: %v, params: %v", method, url, headers, params)
		return HttpResult{
			Status:        -1,
			Body:          nil,
			Header:        nil,
			Cookies:       nil,
			ContentLength: 0,
			Err:           err,
			Elapsed:       time.Since(now).Milliseconds(),
		}
	}

	if headers == nil {
		headers = make(map[string]interface{})
		// 如果没有设置Content-Type，则默认设置为application/json
		if headers["Content-Type"] == nil {
			headers["Content-Type"] = "application/json"
		}
	}

	// 参数签名
	var token = headers["token"]
	if token == nil {
		token = "not-need-token"
	}
	var times = fmt.Sprintf("%d", time.Now().UnixMilli())
	var sign = utils.Sign(httpClient.Options.SignKey, token.(string), times, bodyString, params)
	headers["sign"] = sign
	headers["times"] = times
	headers["token"] = token

	// 设置请求头
	for key, value := range headers {
		switch value.(type) {
		case string:
			request.Header.Set(key, value.(string))
			break
		case int:
			request.Header.Set(key, strconv.Itoa(value.(int)))
		}
	}

	// 发送HTTP请求
	response, err := httpClient.client.Do(request)
	code := -1
	if response != nil {
		code = response.StatusCode
	}

	elapsed := time.Since(now).Milliseconds()
	if elapsed >= 200 {
		logger.Warnf("send http request, take too long, code: %d, elapsed: %d, method: %s, url: %s, headers: %v, params: %v", code, elapsed, method, url, headers, params)
	}
	if err != nil {
		logger.Errorf("failed to send http request, code: %d, elapsed: %d, method: %s, url: %s, headers: %v, params: %v, Err: %v", code, elapsed, method, url, headers, params, err)
		logger.Flush()
		return HttpResult{
			Status:        code,
			Body:          nil,
			Header:        nil,
			Cookies:       nil,
			ContentLength: 0,
			Err:           err,
			Elapsed:       elapsed,
		}
	} else {
		defer response.Body.Close()
		cookies := response.Cookies()
		cookieMaps := make(map[string]*http.Cookie)
		for _, cookie := range cookies {
			cookieMaps[cookie.Name] = cookie
		}
		if response.StatusCode == 200 || response.StatusCode == 201 || response.StatusCode == 302 {
			bytes, err := io.ReadAll(response.Body)
			if err != nil {
				logger.Errorf("failed to read body data, code: %d, elapsed: %d, method: %s, url: %s, headers: %v, params: %v, Err: %v", code, elapsed, method, url, headers, params, err)
				logger.Flush()
			}

			return HttpResult{
				Status:        response.StatusCode,
				Body:          bytes,
				Header:        response.Header,
				Cookies:       cookieMaps,
				ContentLength: response.ContentLength,
				Err:           err,
				Elapsed:       elapsed,
			}
		} else {
			logger.Errorf("failed to read body data, code: %d, elapsed: %d, url: %s, headers: %v, params: %v, Err: %v", code, elapsed, url, headers, params, err)
			logger.Flush()
			return HttpResult{
				Status:        response.StatusCode,
				Body:          nil,
				Header:        response.Header,
				Cookies:       cookieMaps,
				ContentLength: response.ContentLength,
				Err:           nil,
				Elapsed:       elapsed,
			}
		}
	}
}

// Init 初始化HttpClient实例对象
func Init(options Options) HttpClient {
	httpClientOnce.Do(func() {
		httpClient = &httpClientImpl{
			Options: &options,
		}
		// 初始化
		jar, _ := cookiejar.New(&cookiejar.Options{}) // 已正确创建cookie jar
		httpClient.(*httpClientImpl).client = &http.Client{
			Jar:     jar,
			Timeout: options.Timeout, // 默认超时时间
			// 默认的Transport中包含连接池相关的能力
			// Transport: &http.Transport{},
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
	})
	return httpClient
}

// GetHttpClient 获取实例对象，需要先调用Init方法初始化
func GetHttpClient() HttpClient {
	return httpClient
}
