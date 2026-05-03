// Package cdp implements a Chrome DevTools Protocol client over WebSocket.
//
// This package provides low-level communication with Chromium browsers
// using the CDP protocol. It handles WebSocket connection management,
// message serialization/deserialization, command-response matching,
// and event subscription.
//
// This package is internal to Kexas — end users should use the
// higher-level kexas.Browser and kexas.Page APIs instead.
package cdp

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/jankylewis/kexas/internal/logger"
	"nhooyr.io/websocket"
)

// Common CDP errors.
var (
	ErrConnectionClosed error = errors.New("cdp: connection closed")
	ErrTimeout          error = errors.New("cdp: operation timed out")
	ErrInvalidMessage   error = errors.New("cdp: invalid message format")
)

// NodeID represents a DOM node identifier in CDP.
type NodeID int64

// SessionID represents a CDP session identifier.
type SessionID string

// Client represents a CDP WebSocket client connection.
type Client struct {
	wsURL      string
	conn       *websocket.Conn
	log        *logger.Logger
	nextID     atomic.Int64
	pending    map[int64]chan *Response
	pendingMu  sync.RWMutex
	handlers   map[string][]EventHandler
	handlersMu sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
	closed     atomic.Bool
}

// Request represents a CDP command request.
type Request struct {
	ID        int64                  `json:"id"`
	Method    string                 `json:"method"`
	Params    map[string]interface{} `json:"params,omitempty"`
	SessionID string                 `json:"sessionId,omitempty"` // For flat protocol
}

// Response represents a CDP command response.
type Response struct {
	ID     int64                  `json:"id"`
	Result map[string]interface{} `json:"result,omitempty"`
	Error  *ResponseError         `json:"error,omitempty"`
}

// ResponseError represents an error in a CDP response.
type ResponseError struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
}

// Error implements the error interface for ResponseError.
func (e *ResponseError) Error() string {
	return fmt.Sprintf("cdp error %d: %s", e.Code, e.Message)
}

// Event represents a CDP event notification.
type Event struct {
	Method string                 `json:"method"`
	Params map[string]interface{} `json:"params,omitempty"`
}

// EventHandler is a function that handles CDP events.
type EventHandler func(params map[string]interface{})

// Connect establishes a WebSocket connection to the CDP endpoint.
func Connect(ctx context.Context, wsURL string) (*Client, error) {
	var log *logger.Logger = logger.New("cdp")

	log.Debug("connecting to CDP", "url", wsURL)

	var conn *websocket.Conn
	var err error
	conn, _, err = websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		log.Error("failed to connect", "err", err)
		return nil, fmt.Errorf("cdp: failed to connect: %w", err)
	}

	// Set read limit to 100MB to handle large messages like screenshots
	conn.SetReadLimit(100 * 1024 * 1024)

	var clientCtx context.Context
	var cancel context.CancelFunc
	clientCtx, cancel = context.WithCancel(ctx)

	var client *Client = &Client{
		wsURL:    wsURL,
		conn:     conn,
		log:      log,
		pending:  make(map[int64]chan *Response),
		handlers: make(map[string][]EventHandler),
		ctx:      clientCtx,
		cancel:   cancel,
	}

	go client.readLoop()

	log.Info("connected to CDP", "url", wsURL)
	return client, nil
}

// Close closes the CDP connection and cleans up resources.
func (c *Client) Close() error {
	if !c.closed.CompareAndSwap(false, true) {
		return nil // Already closed
	}

	c.log.Debug("closing CDP connection")
	c.cancel()

	var err error = c.conn.Close(websocket.StatusNormalClosure, "")
	if err != nil {
		c.log.Error("error closing connection", "err", err)
		return fmt.Errorf("cdp: failed to close connection: %w", err)
	}

	c.log.Info("CDP connection closed")
	return nil
}
