package cdp

import (
	"encoding/json"

	"nhooyr.io/websocket"
)

// readLoop continuously reads messages from the WebSocket connection.
// On any read failure (Chrome crashed / closed the WS / network reset) the loop
// must wake up every pending sender by cancelling the client context, otherwise
// callers waiting on respChan in SendToSession block forever — a 2-minute test
// timeout vs. a fast ErrConnectionClosed.
func (c *Client) readLoop() {
	defer c.shutdown()
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

// shutdown marks the client closed and cancels the client context so every
// blocked Send/SendToSession unblocks via its `case <-c.ctx.Done()` arm.
// Idempotent — safe whether the close came from readLoop or Client.Close.
func (c *Client) shutdown() {
	if !c.closed.CompareAndSwap(false, true) {
		return
	}
	c.cancel()
}

// handleMessage processes an incoming CDP message — either a command response
// (has an "id" field) or an event notification (has "method" but no "id").
func (c *Client) handleMessage(data []byte) {
	var resp Response
	var err error = json.Unmarshal(data, &resp)
	if err == nil && resp.ID > 0 {
		c.handleResponse(&resp)
		return
	}

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

// handleEvent dispatches an event to registered handlers — one goroutine per
// handler so a slow handler can't block the read loop or other event handlers.
func (c *Client) handleEvent(event *Event) {
	c.log.Debug("← received event", "method", event.Method, "params", event.Params)

	c.handlersMu.RLock()
	var handlers []EventHandler = c.handlers[event.Method]
	c.handlersMu.RUnlock()

	for _, handler := range handlers {
		go handler(event.Params)
	}
}
