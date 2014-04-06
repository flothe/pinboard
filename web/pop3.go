// Simple POP3 Client implementation
package web

import (
	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net"
	"net/mail"
	"strconv"
	"strings"
)

// The POP3 client.
type Client struct {
	connection net.Conn
	data       *bufio.Reader
}

func Dial(addr string) (*Client, error) {
	connection, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return NewClient(connection)
}

// DialTLS creates a TLS-secured connection to the POP3 server at the given
// address and returns the corresponding Client.
func DialTLS(addr string) (*Client, error) {
	connection, err := tls.Dial("tcp", addr, nil)
	if err != nil {
		return nil, err
	}
	return NewClient(connection)
}

// NewClient returns a new Client object using an existing connection. if log is not nil
// logging output is written to it
func NewClient(con net.Conn) (*Client, error) {

	var client *Client
	client = &Client{
		data:       bufio.NewReader(con),
		connection: con,
	}

	// send dud command, to read a line
	_, err := client.Cmd("")
	if err != nil {
		return nil, err
	}
	return client, nil
}

// Convenience function to synchronously run an arbitrary command and wait for
// output. The terminating CRLF must be included in the format string.
//
// Output sent after the first line must be retrieved via readLines.
func (c *Client) Cmd(format string, args ...interface{}) (string, error) {
	// do some logging
	log.Printf("SEND: %q\n", fmt.Sprintf(format, args...))

	fmt.Fprintf(c.connection, format, args...)
	line, _, err := c.data.ReadLine()
	log.Printf("RECEIVE (only first line is shown): %q\n", string(line))

	if err != nil {
		return "", err
	}
	l := string(line)
	if l[0:3] != "+OK" {
		err = errors.New(l[5:])
	}
	if len(l) >= 4 {
		return l[4:], err
	}
	return "", err
}

func (c *Client) ReadLines() (lines []string, err error) {
	lines = make([]string, 0)
	l, _, err := c.data.ReadLine()
	line := string(l)
	for err == nil && line != "." {
		if len(line) > 0 && line[0] == '.' {
			line = line[1:]
		}
		lines = append(lines, line)
		l, _, err = c.data.ReadLine()
		line = string(l)
	}
	return
}

// User sends the given username to the server. Generally, there is no reason
// not to use the Auth convenience method.
func (c *Client) User(username string) (err error) {
	_, err = c.Cmd("USER %s\r\n", username)
	return
}

// Pass sends the given password to the server. The password is sent
// unencrypted unless the connection is already secured by TLS (via DialTLS or
// some other mechanism). Generally, there is no reason not to use the Auth
// convenience method.
func (c *Client) Pass(password string) (err error) {
	_, err = c.Cmd("PASS %s\r\n", password)
	return
}

// Auth sends the given username and password to the server, calling the User
// and Pass methods as appropriate.
func (c *Client) Auth(username, password string) (err error) {
	err = c.User(username)
	if err != nil {
		return
	}
	err = c.Pass(password)
	return
}

// Stat retrieves a drop listing for the current maildrop, consisting of the
// number of messages and the total size (in octets) of the maildrop.
// Information provided besides the number of messages and the size of the
// maildrop is ignored. In the event of an error, all returned numeric values
// will be 0.
func (c *Client) Stat() (count, size int, err error) {
	l, err := c.Cmd("STAT\r\n")
	if err != nil {
		return 0, 0, err
	}
	parts := strings.Fields(l)
	count, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, errors.New("Invalid server response")
	}
	size, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, errors.New("Invalid server response")
	}
	return
}

// List returns the size of the given message, if it exists. If the message
// does not exist, or another error is encountered, the returned size will be
// 0.
func (c *Client) List(msg int) (size int, err error) {
	l, err := c.Cmd("LIST %d\r\n", msg)
	if err != nil {
		return 0, err
	}
	size, err = strconv.Atoi(strings.Fields(l)[1])
	if err != nil {
		return 0, errors.New("Invalid server response")
	}
	return size, nil
}

// ListAll returns a list of all messages and their sizes.
func (c *Client) ListAll() (msgs []int, sizes []int, err error) {
	_, err = c.Cmd("LIST\r\n")
	if err != nil {
		return
	}
	lines, err := c.ReadLines()
	if err != nil {
		return
	}
	msgs = make([]int, len(lines), len(lines))
	sizes = make([]int, len(lines), len(lines))
	for i, l := range lines {
		var m, s int
		fs := strings.Fields(l)
		m, err = strconv.Atoi(fs[0])
		if err != nil {
			return
		}
		s, err = strconv.Atoi(fs[1])
		if err != nil {
			return
		}
		msgs[i] = m
		sizes[i] = s
	}
	return
}

// Retr downloads and returns the given message. The lines are separated by LF,
// whatever the server sent.
func (c *Client) Retr(msg int) (text string, err error) {
	_, err = c.Cmd("RETR %d\r\n", msg)
	if err != nil {
		return "", err
	}
	lines, err := c.ReadLines()
	text = strings.Join(lines, "\n")
	return
}

// Retr downloads and returns the given message.
func (c *Client) RetrMsg(nr int) (*mail.Message, error) {
	text, err := c.Retr(nr)
	if err != nil {
		return nil, err
	}
	m, err := mail.ReadMessage(strings.NewReader(text)) // Read email using Go's net/mail
	if err != nil {
		return nil, err
	}
	return m, err
}

// Dele marks the given message as deleted.
func (c *Client) Dele(msg int) (err error) {
	_, err = c.Cmd("DELE %d\r\n", msg)
	return
}

// Noop does nothing, but will prolong the end of the connection if the server
// has a timeout set.
func (c *Client) Noop() (err error) {
	_, err = c.Cmd("NOOP\r\n")
	return
}

// Rset unmarks any messages marked for deletion previously in this session.
func (c *Client) Rset() (err error) {
	_, err = c.Cmd("RSET\r\n")
	return
}

// Quit sends the QUIT message to the POP3 server and closes the connection.
func (c *Client) Quit() error {
	_, err := c.Cmd("QUIT\r\n")
	if err != nil {
		return err
	}
	c.connection.Close()
	return nil
}
