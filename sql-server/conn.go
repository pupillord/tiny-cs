package main

import (
	"bufio"
	"context"
	"encoding/binary"
	"errors"
	"io"
	"log"
	"net"
	"sync/atomic"
	"time"
	"tinysql-cs/protocol"
	"tinysql-cs/util"
)

const defaultWriterSize = 16 * 1024

type clientConn struct {
	server       *Server           // a reference of server instance.
	connectionID uint32            // atomically allocated by a global variable, unique in process scope.
	user         string            // user of the client.
	dbname       string            // default database name.
	bufWriter    *bufio.Writer
	bufReadConn  *util.BufferedReadConn

	session		 *Session

	lastActive   time.Time
}

// newClientConn creates a *clientConn object.
func newClientConn(s *Server) *clientConn {
	return &clientConn{
		server:       s,
		connectionID: atomic.AddUint32(&baseConnID, 1),
		lastActive:   time.Now(),
	}
}

func (cc *clientConn) setConn(conn net.Conn) {
	cc.bufReadConn = util.NewBufferedReadConn(conn)
	cc.bufWriter = bufio.NewWriterSize(cc.bufReadConn, defaultWriterSize)
}

// handshake confirm connection with client,
// and get some information like user's name etc.
// the protocol rule can reference /tinysql-cs/protocol/protocol.go
func (cc *clientConn) handshake(ctx context.Context) error {
	msg, err := cc.readPacket(ctx)
	if err != nil {
		return err
	}

	if msg.MsgType != protocol.StartupMessageType {
		return errors.New("the first message must be startup message")
	}

	cc.user = string(msg.Content)

	completeMsg := protocol.Message{
		MsgType: protocol.CompleteMessageType,
	}

	cc.bufWriter.Write(completeMsg.Encode(nil))

	if err := cc.flush(); err != nil {
		return err
	}

	return cc.OpenSession()
}


// Run reads client query and writes query result to client in for loop,
func (cc *clientConn) Run(ctx context.Context) {
	defer closeConn(cc)

	for	{
		msg, err := cc.readPacket(ctx)
		if err != nil {
			log.Printf("failed to read client's packet: %s", err.Error())
			return
		}

		switch msg.MsgType {
		case protocol.QueryMessageType:
			if err := cc.handleQuery(msg, ctx); err != nil {
				log.Printf("failed to handle query: %s\n", err)
			}
		case protocol.ExitMessageType:
			return
		default:
			log.Printf("unknown package type: %v\n", msg.MsgType)
			return
		}
	}
}

func closeConn(cc *clientConn) error {
	return cc.bufReadConn.Close()
}

func (cc *clientConn) readPacket(ctx context.Context) (*protocol.Message, error) {
	msgLength := make([]byte, 4)
	if _, err := io.ReadFull(cc.bufReadConn, msgLength); err != nil {
		return nil, err
	}

	msgLen := binary.BigEndian.Uint32(msgLength)

	msgType := make([]byte, 1)

	if _, err := io.ReadFull(cc.bufReadConn, msgType); err != nil {
		return nil, err
	}

	msgContent := make([]byte, msgLen - 5)

	if _, err := io.ReadFull(cc.bufReadConn, msgContent); err != nil {
		return nil, err
	}

	return &protocol.Message{
		MsgType: msgType[0],
		MsgLength: msgLen,
		Content: msgContent,
	}, nil
}

func (cc *clientConn) handleQuery(msg *protocol.Message, ctx context.Context) error {
	log.Printf("sql: %s\n", util.String(msg.Content))
	replyMsg := protocol.Message{
		MsgType: protocol.DataMessageType,
		Content: util.Slice("the query has been completed"),
	}
	cc.bufWriter.Write(replyMsg.Encode(nil))
	return cc.flush()
}

func (cc *clientConn) flush() error {
	return cc.bufWriter.Flush()
}

func (cc *clientConn) OpenSession() error {
	cc.session = OpenSession()
	return nil
}