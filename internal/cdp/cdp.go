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
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/kexas-project/kexas/internal/logger"
	"nhooyr.io/websocket"
)

// Common CDP errors.
var (
	ErrConnectionClosed error = errors.New("cdp: connection closed")
	ErrTimeout          error = errors.New("cdp: operation timed out")
	ErrInvalidMessage   error = errors.New("cdp: invalid message format")
)

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

	// Start message reader goroutine
	go client.readLoop()

	log.Info("connected to CDP", "url", wsURL)
	return client, nil
}

// Send sends a CDP command and waits for the response.
func (c *Client) Send(ctx context.Context, method string, params map[string]interface{}) (map[string]interface{}, error) {
	return c.SendToSession(ctx, "", method, params)
}

// SendToSession sends a CDP command to a specific session and waits for the response.
func (c *Client) SendToSession(ctx context.Context, sessionID string, method string, params map[string]interface{}) (map[string]interface{}, error) {
	if c.closed.Load() {
		return nil, ErrConnectionClosed
	}

	var id int64 = c.nextID.Add(1)

	// For flat protocol, include sessionId as a top-level field in the request
	var req *Request = &Request{
		ID:        id,
		Method:    method,
		Params:    params,
		SessionID: sessionID, // Flat protocol routing
	}

	var respChan chan *Response = make(chan *Response, 1)
	c.pendingMu.Lock()
	c.pending[id] = respChan
	c.pendingMu.Unlock()

	defer func() {
		c.pendingMu.Lock()
		delete(c.pending, id)
		c.pendingMu.Unlock()
	}()

	c.log.Debug("→ sending command", "id", id, "method", method, "params", params, "sessionId", sessionID)

	var data []byte
	var err error
	data, err = json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("cdp: failed to marshal request: %w", err)
	}

	err = c.conn.Write(ctx, websocket.MessageText, data)
	if err != nil {
		c.log.Error("failed to send command", "err", err)
		return nil, fmt.Errorf("cdp: failed to send command: %w", err)
	}

	// Wait for response
	select {
	case resp := <-respChan:
		if resp.Error != nil {
			c.log.Error("← command failed", "id", id, "error", resp.Error.Message)
			return nil, resp.Error
		}
		c.log.Debug("← received response", "id", id, "result", resp.Result)
		return resp.Result, nil
	case <-ctx.Done():
		return nil, ErrTimeout
	case <-c.ctx.Done():
		return nil, ErrConnectionClosed
	}
}

// On registers an event handler for the given event method.
// Multiple handlers can be registered for the same event.
func (c *Client) On(method string, handler EventHandler) {
	c.handlersMu.Lock()
	defer c.handlersMu.Unlock()
	c.handlers[method] = append(c.handlers[method], handler)
	c.log.Debug("registered event handler", "method", method)
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

// readLoop continuously reads messages from the WebSocket connection.
func (c *Client) readLoop() {
	for {
		var msgType websocket.MessageType
		var data []byte
		var err error
		msgType, data, err = c.conn.Read(c.ctx)
		if err != nil {
			if !c.closed.Load() {
				c.log.Error("read error", "err", err)
			}
			return
		}

		if msgType != websocket.MessageText {
			c.log.Warn("unexpected message type", "type", msgType)
			continue
		}

		c.handleMessage(data)
	}
}

// handleMessage processes an incoming CDP message.
func (c *Client) handleMessage(data []byte) {
	// Try to parse as response first (has "id" field)
	var resp Response
	var err error = json.Unmarshal(data, &resp)
	if err == nil && resp.ID > 0 {
		c.handleResponse(&resp)
		return
	}

	// Try to parse as event (has "method" field, no "id")
	var event Event
	err = json.Unmarshal(data, &event)
	if err == nil && event.Method != "" {
		c.handleEvent(&event)
		return
	}

	c.log.Warn("unrecognized message format", "data", string(data))
}

// handleResponse delivers a response to the waiting command sender.
func (c *Client) handleResponse(resp *Response) {
	c.pendingMu.RLock()
	var respChan chan *Response
	var ok bool
	respChan, ok = c.pending[resp.ID]
	c.pendingMu.RUnlock()

	if !ok {
		c.log.Warn("received response for unknown request", "id", resp.ID)
		return
	}

	select {
	case respChan <- resp:
	default:
		c.log.Warn("response channel full", "id", resp.ID)
	}
}

// handleEvent dispatches an event to registered handlers.
func (c *Client) handleEvent(event *Event) {
	c.log.Debug("← received event", "method", event.Method, "params", event.Params)

	c.handlersMu.RLock()
	var handlers []EventHandler = c.handlers[event.Method]
	c.handlersMu.RUnlock()

	for _, handler := range handlers {
		go handler(event.Params)
	}
}
