# Chat Client User Guide

This is a command-line chat client for connecting to a chat server, supporting user registration, login, message sending, and viewing message history.

## Compilation

Execute the following in the project root directory:

```bash
go build -o chatclient cmd/client/main.go
```

## Usage

### Command Line Arguments

- `-server`: Server address, default is "http://localhost:8080"
- `-email`: User email
- `-password`: User password
- `-name`: User name (required only for registration)
- `-register`: Register a new user

### Register a New User

```bash
./chatclient -register -email="user@example.com" -password="yourpassword" -name="Your Name"
```

### Login and Start Chatting

```bash
./chatclient -email="user@example.com" -password="yourpassword"
```

### Chat Commands

Once connected, you can use the following commands:

- Send a message: Type the message content and press Enter to send
- `/history`: View the last 10 messages
- `/quit`: Exit the program
- `feedback`: Enter feedback mode to rate the service and provide comments

### Example Session

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
> feedback
Please rate this service on a scale of 1-5, with 5 being very satisfied and 1 being very dissatisfied.
> 5
Thank you for your rating! Do you have any suggestions or comments about our service?
> No, everything is great!
Thank you very much for your feedback! We will continue to strive to provide better service.
> /quit
Goodbye!
```

## Notes

1. Ensure the server address is correct and accessible
2. Passwords are encrypted during transmission
3. If the connection is lost, the program will automatically exit, requiring re-login
4. Use Ctrl+C to safely exit the program

## Error Handling

Common errors and solutions:

1. Connection failure
   - Check if the server address is correct
   - Ensure the server is running
   - Check network connection

2. Login failure
   - Verify email and password are correct
   - Check if the account is registered

3. Message sending failure
   - Check network connection
   - Ensure WebSocket connection is stable

## Debug Mode

The client defaults to debug mode, displaying detailed request and response information. To disable, change `Debug: true` to `Debug: false` in the source code.