package core

import (
	"bytes"
	"fmt"
	"io"
	"syscall"
)

type Client struct {
	io.ReadWriteCloser
	Fd     int
	cqueue RedisCmds
	isTxn  bool
}

func (c *Client) Read(b []byte) (n int, err error) {
	return syscall.Read(c.Fd, b)
}

func (c *Client) Write(b []byte) (n int, err error) {
	return syscall.Write(c.Fd, b)
}

func (c *Client) Close() error {
	return syscall.Close(c.Fd)
}

func (c *Client) TxnBegin() {
	c.isTxn = true
}

func (c *Client) TxnExec() []byte {
	buf := bytes.NewBuffer(nil)

	// encode it as array
	fmt.Fprintf(buf, "*%d\r\n", len(c.cqueue))

	// go through the queue
	for _, cmd := range c.cqueue {
		buf.Write(evaluateCommand(cmd, c))
	}

	/// clear the queue
	c.cqueue = make(RedisCmds, 0)
	c.isTxn = false

	return buf.Bytes()
}

func (c *Client) TxnDiscard() {
	c.cqueue = make(RedisCmds, 0)
	c.isTxn = false
}

func (c *Client) TxnQueue(cmd *RedisCmd) {
	c.cqueue = append(c.cqueue, cmd)
}

func NewClient(fd int) *Client {
	return &Client{
		Fd:     fd,
		cqueue: RedisCmds{},
		isTxn:  false,
	}
}
