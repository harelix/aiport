package agentrun

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"strconv"
	"sync"
)

// aiport's own small client of the Agent Client Protocol (ACP) version 1:
// JSON-RPC 2.0, one message per line, over the agent's stdin and stdout.
// aiport calls the agent (initialize, session/new, session/prompt, the
// session/cancel notification); the agent streams session/update
// notifications and calls back session/request_permission. aiport offers
// no file system or terminal to the agent, so every other call from the
// agent is answered "method not found".

// ProtocolVersion is the ACP version aiport speaks.
const ProtocolVersion = 1

// maxLine bounds one message from the agent.
const maxLine = 8 << 20

type message struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *rpcError) Error() string { return e.Message }

// conn is one JSON-RPC connection to an agent.
type conn struct {
	out  io.Writer
	wmu  sync.Mutex
	mu   sync.Mutex
	next int
	wait map[string]chan message
	// request answers a call from the agent; notify takes a notification.
	// notify runs on the reading goroutine, in order; request in its own.
	request func(method string, params json.RawMessage) (any, *rpcError)
	notify  func(method string, params json.RawMessage)
	closed  chan struct{}
	err     error // why reading ended, set before closed is closed
}

func newConn(out io.Writer) *conn {
	return &conn{out: out, wait: map[string]chan message{}, closed: make(chan struct{})}
}

func (c *conn) write(m message) error {
	m.JSONRPC = "2.0"
	line, err := json.Marshal(m)
	if err != nil {
		return err
	}
	c.wmu.Lock()
	defer c.wmu.Unlock()
	_, err = c.out.Write(append(line, '\n'))
	return err
}

// read reads the agent's messages until in ends, then fails every call
// still waiting.
func (c *conn) read(in io.Reader) {
	lines := bufio.NewScanner(in)
	lines.Buffer(make([]byte, 0, 64<<10), maxLine)
	for lines.Scan() {
		var m message
		if json.Unmarshal(lines.Bytes(), &m) != nil {
			continue
		}
		switch {
		case m.Method != "" && len(m.ID) > 0:
			go func() {
				result, e := c.request(m.Method, m.Params)
				reply := message{ID: m.ID, Error: e}
				if e == nil {
					reply.Result, _ = json.Marshal(result)
				}
				_ = c.write(reply)
			}()
		case m.Method != "":
			c.notify(m.Method, m.Params)
		case len(m.ID) > 0:
			c.mu.Lock()
			ch := c.wait[string(m.ID)]
			delete(c.wait, string(m.ID))
			c.mu.Unlock()
			if ch != nil {
				ch <- m
			}
		}
	}
	c.err = lines.Err()
	if c.err == nil {
		c.err = io.EOF
	}
	close(c.closed)
}

// call calls method on the agent and decodes its result into out.
func (c *conn) call(ctx context.Context, method string, params, out any) error {
	c.mu.Lock()
	c.next++
	id := strconv.Itoa(c.next)
	ch := make(chan message, 1)
	c.wait[id] = ch
	c.mu.Unlock()
	raw, err := json.Marshal(params)
	if err != nil {
		return err
	}
	if err := c.write(message{ID: json.RawMessage(id), Method: method, Params: raw}); err != nil {
		return errAgentGone
	}
	select {
	case m := <-ch:
		if m.Error != nil {
			return m.Error
		}
		if out != nil && json.Unmarshal(m.Result, out) != nil {
			return errors.New("the agent's answer to " + method + " is not what ACP describes")
		}
		return nil
	case <-c.closed:
		return errAgentGone
	case <-ctx.Done():
		c.mu.Lock()
		delete(c.wait, id)
		c.mu.Unlock()
		return ctx.Err()
	}
}

// send sends a notification.
func (c *conn) send(method string, params any) error {
	raw, err := json.Marshal(params)
	if err != nil {
		return err
	}
	return c.write(message{Method: method, Params: raw})
}

var errAgentGone = errors.New("the agent stopped talking")

// The ACP v1 shapes aiport uses.

type initializeParams struct {
	ProtocolVersion    int                `json:"protocolVersion"`
	ClientCapabilities clientCapabilities `json:"clientCapabilities"`
	ClientInfo         implementation     `json:"clientInfo"`
}

type clientCapabilities struct {
	FS       fsCapabilities `json:"fs"`
	Terminal bool           `json:"terminal"`
}

type fsCapabilities struct {
	ReadTextFile  bool `json:"readTextFile"`
	WriteTextFile bool `json:"writeTextFile"`
}

type implementation struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

type initializeResult struct {
	ProtocolVersion int `json:"protocolVersion"`
}

type newSessionParams struct {
	Cwd        string `json:"cwd"`
	MCPServers []any  `json:"mcpServers"`
}

type newSessionResult struct {
	SessionID string `json:"sessionId"`
}

type contentBlock struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

type promptParams struct {
	SessionID string         `json:"sessionId"`
	Prompt    []contentBlock `json:"prompt"`
}

type promptResult struct {
	StopReason string `json:"stopReason"`
}

type cancelParams struct {
	SessionID string `json:"sessionId"`
}

type sessionNotification struct {
	SessionID string        `json:"sessionId"`
	Update    sessionUpdate `json:"update"`
}

// sessionUpdate holds the fields of the session/update kinds aiport shows.
type sessionUpdate struct {
	SessionUpdate string        `json:"sessionUpdate"`
	Content       *contentBlock `json:"content,omitempty"`
	ToolCallID    string        `json:"toolCallId,omitempty"`
	Title         string        `json:"title,omitempty"`
	Kind          string        `json:"kind,omitempty"`
	Status        string        `json:"status,omitempty"`
	Entries       []struct {
		Content string `json:"content"`
		Status  string `json:"status"`
	} `json:"entries,omitempty"`
}

type toolCall struct {
	ToolCallID string `json:"toolCallId"`
	Title      string `json:"title"`
	Kind       string `json:"kind"`
	Locations  []struct {
		Path string `json:"path"`
	} `json:"locations"`
	RawInput json.RawMessage `json:"rawInput,omitempty"`
}

type permissionOption struct {
	OptionID string `json:"optionId"`
	Name     string `json:"name"`
	Kind     string `json:"kind"` // allow_once, allow_always, reject_once, reject_always
}

type permissionParams struct {
	SessionID string             `json:"sessionId"`
	ToolCall  toolCall           `json:"toolCall"`
	Options   []permissionOption `json:"options"`
}

type permissionOutcome struct {
	Outcome  string `json:"outcome"` // selected or cancelled
	OptionID string `json:"optionId,omitempty"`
}

type permissionResult struct {
	Outcome permissionOutcome `json:"outcome"`
}
