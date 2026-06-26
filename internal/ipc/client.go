package ipc

import (
	"bufio"
	"net"
)

type Client struct {
	conn net.Conn
	sc   *bufio.Scanner
}

func NewClient(socketPath string) (*Client, error) {
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return nil, err
	}
	return &Client{conn: conn, sc: bufio.NewScanner(conn)}, nil
}

func (c *Client) Send(v any) error {
	data, err := Encode(v)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	_, err = c.conn.Write(data)
	return err
}

func (c *Client) Recv() (Envelope, error) {
	if !c.sc.Scan() {
		if err := c.sc.Err(); err != nil {
			return Envelope{}, err
		}
		return Envelope{}, net.ErrClosed
	}
	return Decode(c.sc.Bytes())
}

func (c *Client) Close() error {
	return c.conn.Close()
}
