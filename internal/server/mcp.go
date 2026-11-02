package server

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	"aiport/internal/decide"
)

// MCPPath serves the Model Context Protocol over Streamable HTTP for one
// Application, authenticated with that Application's own token. It is
// hand-written JSON-RPC 2.0 (initialize, ping, tools/list, tools/call and
// notifications) answered with application/json; there is no server-to-client
// stream, so GET is refused.
const MCPPath = "/mcp"

// mcpVersions are the protocol revisions aiport speaks, newest first.
var mcpVersions = []string{"2025-06-18", "2025-03-26", "2024-11-05"}

// mcpInstructions explain the loop to a coding agent in two sentences.
const mcpInstructions = "This Application's LLM calls go through aiport; call aiport_capabilities to see what it can use right now, and if aiport lacks something it needs, ask the Owner with aiport_request_change instead of working around it. The Owner decides in aiport's dashboard, and aiport_my_requests shows each request's answer and the Change that was made."

// mcpIssuesInstructions are added while aiport adapts.
const mcpIssuesInstructions = " At the start of a session, check aiport_my_issues for problems aiport keeps seeing with this Application's calls and their fix; after fixing one in the code, say so with aiport_issue_fixed."

type rpcMessage struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// JSON-RPC 2.0 error codes.
const (
	rpcParseError     = -32700
	rpcInvalidRequest = -32600
	rpcNoMethod       = -32601
	rpcInvalidParams  = -32602
)

// serveMCP handles MCPPath for an authenticated caller.
func (h *Handler) serveMCP(w http.ResponseWriter, r *http.Request, applicationID string, fail func(int, string, string)) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		fail(405, "method_not_allowed", "Use POST; aiport offers no server-to-client stream")
		return
	}
	if !h.isApplication(applicationID) {
		fail(403, "application_token_required", "The MCP endpoint serves one Application: send that Application's own token")
		return
	}
	if ct := r.Header.Get("Content-Type"); !strings.HasPrefix(strings.ToLower(strings.TrimSpace(ct)), "application/json") {
		fail(415, "unsupported_media_type", "Send JSON-RPC as application/json")
		return
	}
	if v := r.Header.Get("MCP-Protocol-Version"); v != "" && !supportedMCP(v) {
		fail(400, "unsupported_protocol_version", "Supported MCP protocol versions: "+strings.Join(mcpVersions, ", "))
		return
	}
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 2*decide.MaxBody))
	if err != nil {
		fail(413, "request_too_large", "An MCP message must be at most 128 KiB")
		return
	}
	var m rpcMessage
	if trimmed := bytes.TrimSpace(body); len(trimmed) > 0 && trimmed[0] == '[' {
		writeJSON(w, 400, rpcReply(nil, nil, &rpcError{rpcInvalidRequest, "Send one JSON-RPC message per request; batches are not supported"}))
		return
	}
	if json.Unmarshal(body, &m) != nil {
		writeJSON(w, 400, rpcReply(nil, nil, &rpcError{rpcParseError, "The body is not a JSON-RPC message"}))
		return
	}
	if m.JSONRPC != "2.0" || m.Method == "" {
		writeJSON(w, 400, rpcReply(m.ID, nil, &rpcError{rpcInvalidRequest, `Send {"jsonrpc":"2.0","method":...}`}))
		return
	}
	if len(m.ID) == 0 || string(m.ID) == "null" {
		// A notification (initialized, cancelled, ...): nothing to answer.
		w.WriteHeader(http.StatusAccepted)
		return
	}
	result, rerr := h.mcpCall(applicationID, m)
	writeJSON(w, 200, rpcReply(m.ID, result, rerr))
}

func (h *Handler) mcpCall(applicationID string, m rpcMessage) (any, *rpcError) {
	switch m.Method {
	case "initialize":
		var p struct {
			ProtocolVersion string `json:"protocolVersion"`
		}
		_ = json.Unmarshal(m.Params, &p)
		version := mcpVersions[0]
		if supportedMCP(p.ProtocolVersion) {
			version = p.ProtocolVersion
		}
		return map[string]any{
			"protocolVersion": version,
			"capabilities":    map[string]any{"tools": map[string]any{"listChanged": false}},
			"serverInfo":      map[string]any{"name": "aiport", "version": "1"},
			"instructions":    h.mcpInstructions(),
		}, nil
	case "ping":
		return map[string]any{}, nil
	case "tools/list":
		return map[string]any{"tools": h.mcpToolList()}, nil
	case "tools/call":
		var p struct {
			Name      string          `json:"name"`
			Arguments json.RawMessage `json:"arguments"`
		}
		if json.Unmarshal(m.Params, &p) != nil {
			return nil, &rpcError{rpcInvalidParams, "params must be {name, arguments}"}
		}
		args := bytes.TrimSpace(p.Arguments)
		if len(args) == 0 || string(args) == "null" {
			args = []byte("{}")
		}
		if args[0] != '{' {
			return nil, &rpcError{rpcInvalidParams, "arguments must be a JSON object"}
		}
		return h.mcpTool(applicationID, p.Name, args)
	}
	return nil, &rpcError{rpcNoMethod, "Unknown method " + m.Method}
}

// mcpTool runs one tool by sending the equivalent request through the HTTP
// handlers of /v1/aiport/capabilities, /v1/aiport/requests and
// /v1/aiport/issues, so both doors
// share one implementation and the tool result is exactly the HTTP body.
func (h *Handler) mcpTool(applicationID, name string, args []byte) (any, *rpcError) {
	var r *http.Request
	switch name {
	case "aiport_capabilities":
		r = internalRequest(http.MethodGet, CapabilitiesPath, nil)
	case "aiport_request_change":
		r = internalRequest(http.MethodPost, RequestsPath, args)
	case "aiport_my_requests":
		var p struct {
			ID string `json:"id"`
		}
		if json.Unmarshal(args, &p) != nil {
			return nil, &rpcError{rpcInvalidParams, "id must be a string"}
		}
		path := RequestsPath
		if p.ID != "" {
			path += "/" + p.ID
		}
		r = internalRequest(http.MethodGet, path, nil)
	case "aiport_my_issues", "aiport_issue_fixed":
		var p struct {
			ID string `json:"id"`
		}
		if json.Unmarshal(args, &p) != nil {
			return nil, &rpcError{rpcInvalidParams, "id must be a string"}
		}
		if name == "aiport_issue_fixed" && p.ID == "" {
			return nil, &rpcError{rpcInvalidParams, "id is required"}
		}
		path, method := IssuesPath, http.MethodGet
		if p.ID != "" {
			path += "/" + url.PathEscape(p.ID)
		}
		if name == "aiport_issue_fixed" {
			path, method = path+"/fixed", http.MethodPost
		}
		r = internalRequest(method, path, nil)
	default:
		return nil, &rpcError{rpcInvalidParams, "Unknown tool " + name}
	}
	tw := &toolWriter{header: http.Header{}, status: 200}
	tw.header.Set("X-Aiport-Request-Id", requestID())
	fail := func(status int, code, message string) {
		writeJSON(tw, status, map[string]any{"error": map[string]string{"code": code, "message": message, "request_id": tw.header.Get("X-Aiport-Request-Id")}})
	}
	switch {
	case name == "aiport_capabilities":
		h.serveCapabilities(tw, r, applicationID, fail)
	case strings.HasPrefix(r.URL.Path, IssuesPath):
		h.serveIssues(tw, r, applicationID, fail)
	default:
		h.serveRequests(tw, r, applicationID, fail)
	}
	body := bytes.TrimSpace(tw.body.Bytes())
	return map[string]any{
		"content":           []map[string]string{{"type": "text", "text": string(body)}},
		"structuredContent": json.RawMessage(body),
		"isError":           tw.status >= 400,
	}, nil
}

func internalRequest(method, path string, body []byte) *http.Request {
	r := &http.Request{Method: method, URL: &url.URL{Path: path}, Header: http.Header{}, Body: http.NoBody}
	if body != nil {
		r.Body = io.NopCloser(bytes.NewReader(body))
		r.ContentLength = int64(len(body))
		r.Header.Set("Content-Type", "application/json")
	}
	return r
}

// toolWriter collects an HTTP handler's answer for a tool result.
type toolWriter struct {
	header http.Header
	status int
	body   bytes.Buffer
}

func (w *toolWriter) Header() http.Header         { return w.header }
func (w *toolWriter) Write(b []byte) (int, error) { return w.body.Write(b) }
func (w *toolWriter) WriteHeader(status int)      { w.status = status }

// mcpTools are the tools aiport offers an Application's coding agent.
var mcpTools = []map[string]any{
	{
		"name":        "aiport_capabilities",
		"title":       "What aiport can do for this Application",
		"description": "What aiport can do right now for this Application: its Aliases, the Connections it can reach, its default routing, and the kinds of Change it may ask for. Same answer as GET " + CapabilitiesPath + ".",
		"inputSchema": map[string]any{"type": "object", "properties": map[string]any{}, "additionalProperties": false},
		"annotations": map[string]any{"readOnlyHint": true, "openWorldHint": false},
	},
	{
		"name":        "aiport_request_change",
		"title":       "Ask the Owner for a Change",
		"description": "File a Change request with aiport's Owner when this Application needs something aiport lacks (for example an Alias). The Owner decides; read the answer with aiport_my_requests. Plain text only; it is never sent to a model. Same as POST " + RequestsPath + ".",
		"inputSchema": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"kind":       map[string]any{"type": "string", "description": "the kind of Change, such as application_alias or bind_application (see aiport_capabilities)"},
				"need":       map[string]any{"type": "string", "description": "one line saying what is needed"},
				"details":    map[string]any{"type": "string", "description": "the details the Owner needs to make the Change"},
				"why":        map[string]any{"type": "string", "description": "why the Application needs it"},
				"scope":      map[string]any{"type": "string", "enum": []string{string(decide.ScopeApplication), string(decide.ScopeEvery)}, "description": "application (only this Application) or every_application"},
				"scope_case": map[string]any{"type": "string", "description": "the argument for that Scope"},
				"filed_by":   map[string]any{"type": "string", "description": "optional: who is asking, such as the coding agent and the project"},
			},
			"required":             []string{"kind", "need", "details", "why", "scope", "scope_case"},
			"additionalProperties": false,
		},
		"annotations": map[string]any{"readOnlyHint": false, "destructiveHint": false, "idempotentHint": false, "openWorldHint": false},
	},
	{
		"name":        "aiport_my_requests",
		"title":       "This Application's Change requests",
		"description": "This Application's Change requests, newest first, with the Owner's decision, note, granted Scope and Change id; give id for just one. Same as GET " + RequestsPath + " and GET " + RequestsPath + "/{id}.",
		"inputSchema": map[string]any{
			"type":                 "object",
			"properties":           map[string]any{"id": map[string]any{"type": "string", "description": "optional: one request's id"}},
			"additionalProperties": false,
		},
		"annotations": map[string]any{"readOnlyHint": true, "openWorldHint": false},
	},
}

// mcpIssueTools are offered only while aiport adapts (Issues exist only
// then).
var mcpIssueTools = []map[string]any{
	{
		"name":        "aiport_my_issues",
		"title":       "This Application's Issues",
		"description": "Problems aiport keeps seeing with this Application's calls (open, being resolved, and recently resolved), newest first: what is wrong, the evidence, application_fix (what to change in this Application, starting with the root cause when the Owner had it explained), explanation (that root cause and who wrote it) and whether an aiport fix is proposed, applied or reverted; give id for just one. Check it at the start of a session. Same as GET " + IssuesPath + " and GET " + IssuesPath + "/{id}.",
		"inputSchema": map[string]any{
			"type":                 "object",
			"properties":           map[string]any{"id": map[string]any{"type": "string", "description": "optional: one Issue's id"}},
			"additionalProperties": false,
		},
		"annotations": map[string]any{"readOnlyHint": true, "openWorldHint": false},
	},
	{
		"name":        "aiport_issue_fixed",
		"title":       "Say an Issue is fixed in this Application",
		"description": "Tell aiport you fixed an open Issue in this Application's code (for example it now sends a model this gateway serves). The Issue becomes resolving; aiport marks it resolved by the Application fix only when real calls succeed again. Same as POST " + IssuesPath + "/{id}/fixed.",
		"inputSchema": map[string]any{
			"type":                 "object",
			"properties":           map[string]any{"id": map[string]any{"type": "string", "description": "the Issue's id"}},
			"required":             []string{"id"},
			"additionalProperties": false,
		},
		"annotations": map[string]any{"readOnlyHint": false, "destructiveHint": false, "idempotentHint": true, "openWorldHint": false},
	},
}

// mcpInstructions is what initialize tells the agent now.
func (h *Handler) mcpInstructions() string {
	if h.decisions == nil || !h.decisions.On() {
		return mcpInstructions
	}
	return mcpInstructions + mcpIssuesInstructions
}

// mcpToolList is what tools/list offers now.
func (h *Handler) mcpToolList() []map[string]any {
	if h.decisions == nil || !h.decisions.On() {
		return mcpTools
	}
	return append(append([]map[string]any{}, mcpTools...), mcpIssueTools...)
}

func rpcReply(id json.RawMessage, result any, err *rpcError) map[string]any {
	out := map[string]any{"jsonrpc": "2.0", "id": json.RawMessage("null")}
	if len(id) > 0 {
		out["id"] = id
	}
	if err != nil {
		out["error"] = err
	} else {
		out["result"] = result
	}
	return out
}

func supportedMCP(v string) bool {
	for _, s := range mcpVersions {
		if s == v {
			return true
		}
	}
	return false
}
