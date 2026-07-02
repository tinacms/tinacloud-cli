package utils

import (
	"bufio"
	"fmt"
	"net"
)

// Generate a listener and port [9999-10011]
func CallbackListener() (net.Listener, int, error) {
	for port := 9999; port <= 10011; port++ {
		ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			continue
		}
		return ln, port, nil
	}

	return nil, 0, fmt.Errorf("failed to bind local callback server on port range :9999-:10011")
}

func CallbackResponse(connection net.Conn) error {
	writer := bufio.NewWriter(connection)
	_, err := writer.WriteString("HTTP/1.1 200 OK\r\n" +
		"Content-Type: text/html\r\n" +
		"Connection: close\r\n" +
		"\r\n" +
		"<html><body><p>You may now close this tab.<p><script>window.close()</script></body></html>")
	if err != nil {
		return fmt.Errorf("failed to write back to connection: %w", err)
	}
	err = writer.Flush()
	if err != nil {
		return fmt.Errorf("failed to flush writer to connection: %w", err)
	}

	return nil
}
