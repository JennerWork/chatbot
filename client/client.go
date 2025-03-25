package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Config 客户端配置
type Config struct {
	BaseURL   string        // API基础URL
	Timeout   time.Duration // HTTP请求超时时间
	AuthToken string        // JWT认证token
	UserAgent string        // User-Agent
	Debug     bool          // 是否开启调试模式
}

// Client 聊天机器人客户端
type Client struct {
	config     *Config
	httpClient *http.Client
}

// NewClient 创建新的客户端实例
func NewClient(config *Config) *Client {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.BaseURL == "" {
		config.BaseURL = "http://localhost:8080"
	}
	if config.UserAgent == "" {
		config.UserAgent = "ChatBot-Client/1.0"
	}

	return &Client{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// do 执行HTTP请求
func (c *Client) do(method, path string, body interface{}, result interface{}) error {
	// 构建完整URL
	url := fmt.Sprintf("%s%s", c.config.BaseURL, path)

	// 准备请求体
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body failed: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	// 创建请求
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return fmt.Errorf("create request failed: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.config.UserAgent)
	if c.config.AuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.config.AuthToken)
	}

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("do request failed: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body failed: %w", err)
	}

	// 调试模式下打印请求和响应信息
	if c.config.Debug {
		fmt.Printf("[DEBUG] Request: %s %s\n", method, url)
		if body != nil {
			fmt.Printf("[DEBUG] Request Body: %+v\n", body)
		}
		fmt.Printf("[DEBUG] Response Status: %d\n", resp.StatusCode)
		fmt.Printf("[DEBUG] Response Body: %s\n", string(respBody))
	}

	// 检查响应状态码
	if resp.StatusCode >= 400 {
		var errResp ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err != nil {
			return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
		}
		return &errResp
	}

	// 解析响应结果
	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("unmarshal response body failed: %w", err)
		}
	}

	return nil
}

// ErrorResponse API错误响应
type ErrorResponse struct {
	Code     int    `json:"code"`
	Message  string `json:"message"`
	ErrorMsg string `json:"error,omitempty"`
}

func (e *ErrorResponse) Error() string {
	if e.ErrorMsg != "" {
		return fmt.Sprintf("%s: %s", e.Message, e.ErrorMsg)
	}
	return e.Message
}

// LoginRequest 登录请求参数
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// TokenResponse token响应
type TokenResponse struct {
	Token string `json:"token"`
}

// Login 用户登录
func (c *Client) Login(email, password string) error {
	req := LoginRequest{
		Email:    email,
		Password: password,
	}

	var resp TokenResponse
	if err := c.do(http.MethodPost, "/api/auth/login", req, &resp); err != nil {
		return err
	}

	c.config.AuthToken = resp.Token
	return nil
}

// RefreshToken 刷新token
func (c *Client) RefreshToken() error {
	var resp TokenResponse
	if err := c.do(http.MethodPost, "/api/auth/refresh", nil, &resp); err != nil {
		return err
	}

	c.config.AuthToken = resp.Token
	return nil
}

// RegisterRequest 注册请求参数
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

// Register 注册新用户
func (c *Client) Register(email, password, name string) error {
	req := RegisterRequest{
		Email:    email,
		Password: password,
		Name:     name,
	}

	return c.do(http.MethodPost, "/api/customers/register", req, nil)
}

// UpdatePasswordRequest 更新密码请求参数
type UpdatePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

// UpdatePassword 更新密码
func (c *Client) UpdatePassword(oldPassword, newPassword string) error {
	req := UpdatePasswordRequest{
		OldPassword: oldPassword,
		NewPassword: newPassword,
	}

	return c.do(http.MethodPut, "/api/customers/password", req, nil)
}

// UpdateProfileRequest 更新资料请求参数
type UpdateProfileRequest struct {
	Name string `json:"name"`
}

// UpdateProfile 更新用户资料
func (c *Client) UpdateProfile(name string) error {
	req := UpdateProfileRequest{
		Name: name,
	}

	return c.do(http.MethodPut, "/api/customers/profile", req, nil)
}
