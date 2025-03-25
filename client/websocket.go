package client

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// HeartbeatConfig 心跳配置
type HeartbeatConfig struct {
	MinInterval     time.Duration
	MaxInterval     time.Duration
	InitialInterval time.Duration
	AdjustFactor    float64
	RTTThreshold    time.Duration
}

// WSClient WebSocket客户端
type WSClient struct {
	conn           *websocket.Conn
	send           chan []byte
	receive        chan []byte
	done           chan struct{}
	closeOnce      sync.Once
	config         *Config
	heartbeatStats struct {
		currentInterval time.Duration
		lastPingTime    time.Time
		lastRTT         time.Duration
		mutex           sync.Mutex
	}
	heartbeatConfig HeartbeatConfig
}

// MessageHandler 消息处理函数类型
type MessageHandler func(message []byte)

// ConnectWebSocket 连接WebSocket服务器
func (c *Client) ConnectWebSocket() (*WSClient, error) {
	// 解析WebSocket URL
	u, err := url.Parse(c.config.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse base URL failed: %w", err)
	}

	// 将 http(s) 转换为 ws(s)
	switch u.Scheme {
	case "http":
		u.Scheme = "ws"
	case "https":
		u.Scheme = "wss"
	}
	u.Path = "/ws"

	// 添加认证token
	if c.config.AuthToken != "" {
		header := make(map[string][]string)
		header["Authorization"] = []string{"Bearer " + c.config.AuthToken}
		dialer := websocket.Dialer{
			HandshakeTimeout: c.config.Timeout,
		}

		// 建立连接
		conn, _, err := dialer.Dial(u.String(), header)
		if err != nil {
			return nil, fmt.Errorf("dial websocket failed: %w", err)
		}

		ws := &WSClient{
			conn:    conn,
			send:    make(chan []byte, 256),
			receive: make(chan []byte, 256),
			done:    make(chan struct{}),
			config:  c.config,
			heartbeatConfig: HeartbeatConfig{
				MinInterval:     15 * time.Second,
				MaxInterval:     60 * time.Second,
				InitialInterval: 30 * time.Second,
				AdjustFactor:    1.5,
				RTTThreshold:    time.Second,
			},
		}

		// 设置初始心跳间隔
		ws.heartbeatStats.currentInterval = ws.heartbeatConfig.InitialInterval

		// 启动读写goroutine
		go ws.readPump()
		go ws.writePump()

		return ws, nil
	}

	return nil, fmt.Errorf("authentication token required")
}

// readPump 从WebSocket连接读取消息
func (ws *WSClient) readPump() {
	defer func() {
		ws.Close()
	}()

	// 设置pong处理器来计算RTT和调整心跳间隔
	ws.conn.SetPongHandler(func(string) error {
		ws.heartbeatStats.mutex.Lock()
		defer ws.heartbeatStats.mutex.Unlock()

		// 计算RTT
		rtt := time.Since(ws.heartbeatStats.lastPingTime)
		ws.heartbeatStats.lastRTT = rtt

		// 根据RTT调整心跳间隔
		ws.adjustHeartbeatInterval(rtt)

		// 重置读取超时（当前心跳间隔的1.5倍）
		ws.conn.SetReadDeadline(time.Now().Add(time.Duration(float64(ws.heartbeatStats.currentInterval) * 1.5)))

		if ws.config.Debug {
			log.Printf("[DEBUG] Received pong, RTT: %v, new interval: %v", rtt, ws.heartbeatStats.currentInterval)
		}

		return nil
	})

	// 初始读取超时
	ws.conn.SetReadDeadline(time.Now().Add(time.Second * 45))

	for {
		_, message, err := ws.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("websocket read error: %v", err)
			}
			return
		}

		// 重置读取超时
		ws.conn.SetReadDeadline(time.Now().Add(time.Duration(float64(ws.heartbeatStats.currentInterval) * 1.5)))

		select {
		case ws.receive <- message:
		case <-ws.done:
			return
		}
	}
}

// writePump 向WebSocket连接写入消息
func (ws *WSClient) writePump() {
	ticker := time.NewTicker(ws.heartbeatStats.currentInterval)
	defer func() {
		ticker.Stop()
		ws.Close()
	}()

	for {
		select {
		case message := <-ws.send:
			ws.conn.SetWriteDeadline(time.Now().Add(time.Second * 10))
			if err := ws.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("websocket write error: %v", err)
				return
			}
		case <-ticker.C:
			// 发送ping并记录时间
			ws.heartbeatStats.mutex.Lock()
			ws.heartbeatStats.lastPingTime = time.Now()
			ws.heartbeatStats.mutex.Unlock()

			ws.conn.SetWriteDeadline(time.Now().Add(time.Second * 10))
			if err := ws.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("websocket ping error: %v", err)
				return
			}

			if ws.config.Debug {
				log.Printf("[DEBUG] Sent heartbeat ping with interval: %v", ws.heartbeatStats.currentInterval)
			}
		case <-ws.done:
			return
		}
	}
}

// adjustHeartbeatInterval 根据RTT调整心跳间隔
func (ws *WSClient) adjustHeartbeatInterval(rtt time.Duration) {
	// 如果RTT超过阈值，减小心跳间隔
	if rtt > ws.heartbeatConfig.RTTThreshold {
		newInterval := time.Duration(float64(ws.heartbeatStats.currentInterval) / ws.heartbeatConfig.AdjustFactor)
		if newInterval < ws.heartbeatConfig.MinInterval {
			newInterval = ws.heartbeatConfig.MinInterval
		}
		ws.heartbeatStats.currentInterval = newInterval
	} else {
		// 如果RTT正常，可以适当增加心跳间隔
		newInterval := time.Duration(float64(ws.heartbeatStats.currentInterval) * ws.heartbeatConfig.AdjustFactor)
		if newInterval > ws.heartbeatConfig.MaxInterval {
			newInterval = ws.heartbeatConfig.MaxInterval
		}
		ws.heartbeatStats.currentInterval = newInterval
	}
}

// Send 发送消息
func (ws *WSClient) Send(msgType string, content interface{}) error {
	msg := struct {
		Type    string      `json:"type"`
		Content interface{} `json:"content"`
	}{
		Type:    msgType,
		Content: content,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal message failed: %w", err)
	}

	select {
	case ws.send <- data:
		return nil
	case <-ws.done:
		return fmt.Errorf("websocket connection closed")
	}
}

// SendText 发送文本消息
func (ws *WSClient) SendText(text string) error {
	return ws.Send("text", TextMessage{Text: text})
}

// Receive 接收消息
func (ws *WSClient) Receive(ctx context.Context) ([]byte, error) {
	select {
	case message := <-ws.receive:
		return message, nil
	case <-ws.done:
		return nil, fmt.Errorf("websocket connection closed")
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Listen 监听消息
func (ws *WSClient) Listen(handler MessageHandler) {
	for {
		select {
		case message := <-ws.receive:
			handler(message)
		case <-ws.done:
			return
		}
	}
}

// Close 关闭连接
func (ws *WSClient) Close() {
	ws.closeOnce.Do(func() {
		close(ws.done)
		ws.conn.Close()
	})
}
