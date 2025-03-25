# Chat Client 使用说明

这是一个命令行聊天客户端，用于连接聊天服务器，支持用户注册、登录、发送消息和查看历史消息等功能。

## 编译

在项目根目录下执行：

```bash
go build -o chatclient cmd/client/main.go
```

## 使用方法

### 命令行参数

- `-server`: 服务器地址，默认为 "http://localhost:8080"
- `-email`: 用户邮箱
- `-password`: 用户密码
- `-name`: 用户名称（仅注册时需要）
- `-register`: 是否为注册新用户

### 注册新用户

```bash
./chatclient -register -email="user@example.com" -password="yourpassword" -name="Your Name"
```

### 登录并开始聊天

```bash
./chatclient -email="user@example.com" -password="yourpassword"
```

### 聊天命令

连接成功后，可以使用以下命令：

- 发送消息：直接输入消息内容并按回车发送
- `/history`: 查看最近的10条消息历史
- `/quit`: 退出程序

### 示例会话

```
$ ./chatclient -email="user@example.com" -password="yourpassword"
Login successful!

Recent messages:
[2024-01-20 10:30:15] Alice: Hello everyone!
[2024-01-20 10:31:20] Bob: Hi Alice!

Connected to chat server. Type your message and press Enter to send.
Type '/quit' to exit, '/history' to view message history.
> Hello, I'm new here!
Received: Welcome! How can I help you today?
> /history
Recent messages:
[2024-01-20 10:30:15] Alice: Hello everyone!
[2024-01-20 10:31:20] Bob: Hi Alice!
[2024-01-20 10:32:00] You: Hello, I'm new here!
[2024-01-20 10:32:05] Bot: Welcome! How can I help you today?
> /quit
Goodbye!
```

## 注意事项

1. 确保服务器地址正确且可访问
2. 密码在传输过程中会进行加密处理
3. 如果连接断开，程序会自动退出，需要重新登录
4. 使用 Ctrl+C 可以安全地退出程序

## 错误处理

常见错误及解决方案：

1. 连接失败
   - 检查服务器地址是否正确
   - 确认服务器是否在运行
   - 检查网络连接

2. 登录失败
   - 确认邮箱和密码是否正确
   - 检查账号是否已注册

3. 消息发送失败
   - 检查网络连接
   - 确认 WebSocket 连接是否正常

## 调试模式

客户端默认开启调试模式，会显示详细的请求和响应信息。如需关闭，可以修改源码中的 `Debug: true` 为 `Debug: false`。 