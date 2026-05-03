package cdp

import (
	"context"
	"encoding/json"
	"fmt"

	"nhooyr.io/websocket"
)

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
	var req *Request = &Request{ID: id, Method: method, Params: params, SessionID: sessionID}
	var respChan chan *Response = c.registerPending(id)
	defer c.unregisterPending(id)

	c.log.Debug("→ sending command", "id", id, "method", method, "params", params, "sessionId", sessionID)
	var err error = c.marshalAndWrite(ctx, req)
	if err != nil {
		return nil, err
	}
	return c.awaitResponse(ctx, id, respChan)
}

// registerPending creates a buffered response channel and stores it under id
// so the read loop can route the matching response back to this caller.
func (c *Client) registerPending(id int64) chan *Response {
	var respChan chan *Response = make(chan *Response, 1)
	c.pendingMu.Lock()
	c.pending[id] = respChan
	c.pendingMu.Unlock()
	return respChan
}

// unregisterPending removes the pending entry. Safe to call after a response
// arrives (no-op) or when bailing out early.
func (c *Client) unregisterPending(id int64) {
	c.pendingMu.Lock()
	delete(c.pending, id)
	c.pendingMu.Unlock()
}

// marshalAndWrite serialises req and writes it to the WebSocket. Returns the
// first failing step's wrapped error.
func (c *Client) marshalAndWrite(ctx context.Context, req *Request) error {
	var data []byte
	var err error
	data, err = json.Marshal(req)
	if err != nil {
		return fmt.Errorf("cdp: failed to marshal request: %w", err)
	}
	err = c.conn.Write(ctx, websocket.MessageText, data)
	if err != nil {
		c.log.Error("failed to send command", "err", err)
		return fmt.Errorf("cdp: failed to send command: %w", err)
	}
	return nil
}

// awaitResponse blocks until the matching response arrives or ctx (caller or
// connection-level) is cancelled.
func (c *Client) awaitResponse(ctx context.Context, id int64, respChan chan *Response) (map[string]interface{}, error) {
	var resp *Response
	select {
	case resp = <-respChan:
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
