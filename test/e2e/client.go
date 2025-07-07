package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// TestClient HTTP测试客户端
type TestClient struct {
	client    *http.Client
	authToken string
	baseURL   string
}

// NewTestClient 创建新的测试客户端
func NewTestClient() *TestClient {
	return &TestClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SetAuthToken 设置认证Token
func (c *TestClient) SetAuthToken(token string) {
	c.authToken = token
}

// SetBaseURL 设置基础URL
func (c *TestClient) SetBaseURL(url string) {
	c.baseURL = url
}

// Get 发送GET请求
func (c *TestClient) Get(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	c.setHeaders(req)
	return c.client.Do(req)
}

// Post 发送POST请求
func (c *TestClient) Post(url string, data interface{}) (*http.Response, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	c.setHeaders(req)
	req.Header.Set("Content-Type", "application/json")
	return c.client.Do(req)
}

// Put 发送PUT请求
func (c *TestClient) Put(url string, data interface{}) (*http.Response, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	c.setHeaders(req)
	req.Header.Set("Content-Type", "application/json")
	return c.client.Do(req)
}

// Delete 发送DELETE请求
func (c *TestClient) Delete(url string) (*http.Response, error) {
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return nil, err
	}

	c.setHeaders(req)
	return c.client.Do(req)
}

// setHeaders 设置通用请求头
func (c *TestClient) setHeaders(req *http.Request) {
	req.Header.Set("User-Agent", "Heimdall-E2E-Test/1.0")
	
	if c.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.authToken)
	}
}

// ParseJSONResponse 解析JSON响应
func (c *TestClient) ParseJSONResponse(resp *http.Response, target interface{}) error {
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(body, target)
}

// GetResponseBody 获取响应体内容
func (c *TestClient) GetResponseBody(resp *http.Response) (string, error) {
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// TestResponse 通用测试响应结构
type TestResponse struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp string      `json:"timestamp"`
}

// AssertSuccess 断言响应成功
func (c *TestClient) AssertSuccess(resp *http.Response) (*TestResponse, error) {
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := c.GetResponseBody(resp)
		return nil, fmt.Errorf("HTTP request failed with status %d: %s", resp.StatusCode, body)
	}

	var testResp TestResponse
	if err := c.ParseJSONResponse(resp, &testResp); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %v", err)
	}

	if testResp.Code != 200 {
		return nil, fmt.Errorf("API request failed with code %d: %s", testResp.Code, testResp.Message)
	}

	return &testResp, nil
}

// WaitForService 等待服务启动
func (c *TestClient) WaitForService(url string, maxAttempts int) error {
	for i := 0; i < maxAttempts; i++ {
		resp, err := c.Get(url + "/health")
		if err == nil && resp.StatusCode == http.StatusOK {
			return nil
		}
		
		time.Sleep(2 * time.Second)
	}
	
	return fmt.Errorf("service not available after %d attempts", maxAttempts)
}