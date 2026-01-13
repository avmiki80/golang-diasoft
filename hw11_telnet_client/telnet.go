package main

import (
	"fmt"
	"io"
	"net"
	"time"
)

type TelnetClient interface {
	Connect() error
	io.Closer
	Send() error
	Receive() error
}

type telnetClient struct {
	address string
	timeout time.Duration
	in      io.ReadCloser
	out     io.Writer
	conn    net.Conn
}

func (t *telnetClient) Connect() error {
	dialer := net.Dialer{Timeout: t.timeout}
	conn, err := dialer.Dial("tcp", t.address)
	t.conn = conn
	return err
}

func (t *telnetClient) Close() error {
	if t.conn == nil {
		return fmt.Errorf("telnet connection already closed")
	}
	return t.conn.Close()
}

func (t *telnetClient) Send() error {
	if t.conn == nil {
		return fmt.Errorf("connection not established")
	}
	_, err := io.Copy(t.conn, t.in)
	return err
}

func (t *telnetClient) Receive() error {
	if t.conn == nil {
		return fmt.Errorf("connection not established")
	}
	_, err := io.Copy(t.out, t.conn)
	return err
}

func NewTelnetClient(address string, timeout time.Duration, in io.ReadCloser, out io.Writer) TelnetClient {
	return &telnetClient{
		address: address,
		timeout: timeout,
		in:      in,
		out:     out,
	}
}
