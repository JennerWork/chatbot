package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/JennerWork/chatbot/client"
)

var (
	serverURL = flag.String("server", "http://localhost:8080", "Chat server URL")
	email     = flag.String("email", "", "User email")
	password  = flag.String("password", "", "User password")
	name      = flag.String("name", "", "User name (only required for registration)")
	register  = flag.Bool("register", false, "Register new user")
)

func main() {
	flag.Parse()

	// 创建客户端实例
	c := client.NewClient(&client.Config{
		BaseURL: *serverURL,
		Debug:   true,
	})

	// 根据命令行参数选择注册或登录
	if *register {
		if *email == "" || *password == "" || *name == "" {
			log.Fatal("Email, password and name are required for registration")
		}
		if err := c.Register(*email, *password, *name); err != nil {
			log.Fatalf("Registration failed: %v", err)
		}
		fmt.Println("Registration successful!")
	}

	// 登录
	if *email == "" || *password == "" {
		log.Fatal("Email and password are required")
	}
	if err := c.Login(*email, *password); err != nil {
		log.Fatalf("Login failed: %v", err)
	}
	fmt.Println("Login successful!")

	// 获取最近的消息历史
	result, err := c.GetMessageHistory(client.MessageQueryParams{
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		log.Printf("Failed to get message history: %v", err)
	} else {
		fmt.Println("\nRecent messages:")
		for _, msg := range result.Messages {
			fmt.Printf("[%s] %s: %s\n", msg.CreatedAt.String(), msg.Sender, msg.Content)
		}
		fmt.Println()
	}

	// 连接WebSocket
	ws, err := c.ConnectWebSocket()
	if err != nil {
		log.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer ws.Close()

	// 处理系统信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// 启动消息接收goroutine
	go func() {
		ws.Listen(func(message []byte) {
			var msg struct {
				Type    string          `json:"type"`
				Content json.RawMessage `json:"content"`
			}
			if err := json.Unmarshal(message, &msg); err != nil {
				log.Printf("Failed to parse message: %v", err)
				return
			}

			switch msg.Type {
			case "text":
				var textMsg client.TextMessage
				if err := json.Unmarshal(msg.Content, &textMsg); err != nil {
					log.Printf("Failed to parse text message: %v", err)
					return
				}
				fmt.Printf("\rReceived: %s\n> ", textMsg.Text)
			default:
				fmt.Printf("\rReceived unknown message type: %s\n> ", msg.Type)
			}
		})
	}()

	// 命令行交互
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Connected to chat server. Type your message and press Enter to send.")
	fmt.Println("Type '/quit' to exit, '/history' to view message history.")
	fmt.Print("> ")

	for {
		select {
		case <-sigCh:
			fmt.Println("\nReceived signal, shutting down...")
			return
		default:
			input, err := reader.ReadString('\n')
			if err != nil {
				log.Printf("Error reading input: %v", err)
				continue
			}

			input = strings.TrimSpace(input)
			if input == "" {
				fmt.Print("> ")
				continue
			}

			// 处理命令
			switch input {
			case "/quit":
				fmt.Println("Goodbye!")
				return
			case "/history":
				result, err := c.GetMessageHistory(client.MessageQueryParams{
					Page:     1,
					PageSize: 10,
				})
				if err != nil {
					log.Printf("Failed to get message history: %v", err)
				} else {
					fmt.Println("\nRecent messages:")
					for _, msg := range result.Messages {
						fmt.Printf("[%s] %s: %s\n", msg.CreatedAt.String(), msg.Sender, msg.Content)
					}
				}
			default:
				// 发送消息
				if err := ws.SendText(input); err != nil {
					log.Printf("Failed to send message: %v", err)
				}
			}
			fmt.Print("> ")
		}
	}
}
