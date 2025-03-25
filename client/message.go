package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Message 消息结构
type Message struct {
	ID        uint            `json:"id"`
	Content   string          `json:"content"`
	Type      string          `json:"type"`
	Sender    string          `json:"sender"`
	Seq       uint            `json:"seq"`
	SessionID uint            `json:"session_id"`
	Extra     json.RawMessage `json:"extra,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
}

// MessageQueryParams 消息查询参数
type MessageQueryParams struct {
	SessionID *uint     `json:"session_id,omitempty"`
	StartTime time.Time `json:"start_time,omitempty"`
	EndTime   time.Time `json:"end_time,omitempty"`
	Page      int       `json:"page,omitempty"`
	PageSize  int       `json:"page_size,omitempty"`
}

// MessageQueryResult 消息查询结果
type MessageQueryResult struct {
	Total    int64     `json:"total"`
	Messages []Message `json:"messages"`
}

// GetMessageHistory 获取消息历史
func (c *Client) GetMessageHistory(params MessageQueryParams) (*MessageQueryResult, error) {
	// 构建查询参数
	query := make(map[string]string)
	if params.SessionID != nil {
		query["session_id"] = fmt.Sprintf("%d", *params.SessionID)
	}
	if !params.StartTime.IsZero() {
		query["start_time"] = params.StartTime.Format("2006-01-02 15:04:05")
	}
	if !params.EndTime.IsZero() {
		query["end_time"] = params.EndTime.Format("2006-01-02 15:04:05")
	}
	if params.Page > 0 {
		query["page"] = fmt.Sprintf("%d", params.Page)
	}
	if params.PageSize > 0 {
		query["page_size"] = fmt.Sprintf("%d", params.PageSize)
	}

	// 构建查询字符串
	path := "/api/message/list"
	if len(query) > 0 {
		path += "?"
		for k, v := range query {
			path += fmt.Sprintf("%s=%s&", k, v)
		}
		path = path[:len(path)-1] // 移除最后的 &
	}

	var result MessageQueryResult
	if err := c.do(http.MethodGet, path, nil, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// TextMessage 文本消息
type TextMessage struct {
	Text string `json:"text"`
}

// SendMessage 发送消息
func (c *Client) SendMessage(msgType string, content interface{}) error {
	msg := struct {
		Type    string      `json:"type"`
		Content interface{} `json:"content"`
	}{
		Type:    msgType,
		Content: content,
	}

	return c.do("POST", "/api/message/send", msg, nil)
}

// SendText 发送文本消息
func (c *Client) SendText(text string) error {
	return c.SendMessage("text", TextMessage{Text: text})
}
