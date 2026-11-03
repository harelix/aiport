package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// The MCP endpoint: a coding agent reaches aiport's Change request loop over
// the Model Context Protocol (Streamable HTTP, JSON responses) at /mcp, with
// its Application's own token.

func mcp(h *Handler, token, body string) (*httptest.ResponseRecorder, map[string]any) {
	r := httptest.NewRequest("POST", "http://127.0.0.1:8787"+MCPPath, strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Accept", "application/json, text/event-stream")
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	out := map[string]any{}
	_ = json.Unmarshal(w.Body.Bytes(), &out)
	return w, out
}

func rpc(id int, method string, params any) string {
	m := map[string]any{"jsonrpc": "2.0", "id": id, "method": method}
	if params != nil {
		m["params"] = params
	}
	b, _ := json.Marshal(m)
	return string(b)
}

func TestMCPInitializeExplainsTheLoop(t *testing.T) {
	h, _ := requestsServer(t)
	w, out := mcp(h, reqHelper, rpc(1, "initialize", map[string]any{
		"protocolVersion": "2025-06-18",
		"capabilities":    map[string]any{},
		"clientInfo":      map[string]any{"name": "test", "version": "1"},
	}))
	if w.Code != 200 || !strings.HasPrefix(w.Header().Get("Content-Type"), "application/json") {
		t.Fatalf("initialize: %d %q %s", w.Code, w.Header().Get("Content-Type"), w.Body)
	}
	if out["jsonrpc"] != "2.0" || out["id"] != float64(1) {
		t.Fatalf("not a JSON-RPC answer to id 1: %s", w.Body)
	}
	res, _ := out["result"].(map[string]any)
	if res["protocolVersion"] != "2025-06-18" {
		t.Fatalf("protocol version not agreed: %v", res)
	}
	if caps, _ := res["capabilities"].(map[string]any); caps["tools"] == nil {
		t.Fatalf("tools capability missing: %v", res)
	}
	if info, _ := res["serverInfo"].(map[string]any); info["name"] != "aiport" {
		t.Fatalf("serverInfo: %v", res)
	}
	instr, _ := res["instructions"].(string)
	// The same three things as the AGENTS.md line: a Level or task name,
	// never a provider's model id; effort with the header; the session-start
	// checks and confirming guidance.
	mustContain(t, "MCP instructions", instr, "aiport_request_change", "aiport_my_requests")
	mustContain(t, "MCP instructions", instr, levelWords...)
	mustContain(t, "MCP instructions", instr, sessionStartTools...)
}

func TestConnectAnnouncesTheMCPEndpoint(t *testing.T) {
	h, _ := requestsServer(t)
	_, doc := call(h, "GET", "/_connect", "", "", false)
	for _, f := range doc["features"].([]any) {
		f := f.(map[string]any)
		if f["name"] == "mcp" {
			d := f["description"].(string)
			if f["status"] != "available" || !strings.Contains(d, MCPPath) || !strings.Contains(d, "aiport_request_change") || !strings.Contains(d, "AIPORT_APP_TOKEN") {
				t.Fatalf("mcp feature: %v", f)
			}
			return
		}
	}
	t.Fatalf("/_connect does not list the MCP endpoint: %v", doc["features"])
}

func TestMCPRefusesWithoutAnApplicationToken(t *testing.T) {
	h, _ := requestsServer(t)
	init := rpc(1, "initialize", map[string]any{"protocolVersion": "2025-06-18"})
	for _, token := range []string{"", strings.Repeat("x", 40)} {
		if w, _ := mcp(h, token, init); w.Code != 401 || w.Header().Get("WWW-Authenticate") == "" {
			t.Fatalf("token %q: want 401, got %d %s", token, w.Code, w.Body)
		}
	}
	if w, out := mcp(h, reqManagement, init); w.Code != 403 || errorCode(out) != "application_token_required" {
		t.Fatalf("management token: want 403 application_token_required, got %d %s", w.Code, w.Body)
	}

	r := httptest.NewRequest("GET", "http://127.0.0.1:8787"+MCPPath, nil)
	r.Header.Set("Authorization", "Bearer "+reqHelper)
	r.Header.Set("Accept", "text/event-stream")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 405 || w.Header().Get("Allow") != "POST" {
		t.Fatalf("GET: want 405 Allow POST, got %d %q", w.Code, w.Header().Get("Allow"))
	}

	for name, set := range map[string]func(*http.Request){
		"other origin":  func(r *http.Request) { r.Header.Set("Origin", "https://evil.example") },
		"cross site":    func(r *http.Request) { r.Header.Set("Sec-Fetch-Site", "cross-site") },
		"other host":    func(r *http.Request) { r.Host = "evil.example:8787" },
		"not json body": func(r *http.Request) { r.Header.Set("Content-Type", "text/plain") },
	} {
		r := httptest.NewRequest("POST", "http://127.0.0.1:8787"+MCPPath, strings.NewReader(init))
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Authorization", "Bearer "+reqHelper)
		set(r)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code < 400 || strings.Contains(w.Body.String(), "result") {
			t.Fatalf("%s: a browser-like request was answered: %d %s", name, w.Code, w.Body)
		}
	}
}

func TestMCPNotificationsAndPing(t *testing.T) {
	h, _ := requestsServer(t)
	w, _ := mcp(h, reqHelper, `{"jsonrpc":"2.0","method":"notifications/initialized"}`)
	if w.Code != 202 || w.Body.Len() != 0 {
		t.Fatalf("notification: want 202 and no body, got %d %s", w.Code, w.Body)
	}
	if w, out := mcp(h, reqHelper, rpc(7, "ping", nil)); w.Code != 200 || out["result"] == nil || out["id"] != float64(7) {
		t.Fatalf("ping: %d %s", w.Code, w.Body)
	}
	if _, out := mcp(h, reqHelper, rpc(8, "resources/list", nil)); out["error"].(map[string]any)["code"] != float64(-32601) {
		t.Fatalf("unknown method: %v", out)
	}
	if _, out := mcp(h, reqHelper, `{"jsonrpc":"2.0","id":9,`); out["error"].(map[string]any)["code"] != float64(-32700) {
		t.Fatalf("parse error: %v", out)
	}
}

func toolCall(t *testing.T, h *Handler, token, name string, args any) map[string]any {
	t.Helper()
	params := map[string]any{"name": name}
	if args != nil {
		params["arguments"] = args
	}
	w, out := mcp(h, token, rpc(3, "tools/call", params))
	res, ok := out["result"].(map[string]any)
	if w.Code != 200 || !ok {
		t.Fatalf("tools/call %s: %d %s", name, w.Code, w.Body)
	}
	return res
}

// sameJSON fails unless a tool result carries exactly the HTTP body, both as
// structured content and as its text.
func sameJSON(t *testing.T, res map[string]any, httpBody []byte) {
	t.Helper()
	var want any
	if err := json.Unmarshal(httpBody, &want); err != nil {
		t.Fatalf("http body: %v %s", err, httpBody)
	}
	strip := func(v any) string {
		// request_id differs per request; everything else must match.
		if m, ok := v.(map[string]any); ok {
			if e, ok := m["error"].(map[string]any); ok {
				delete(e, "request_id")
			}
		}
		b, _ := json.Marshal(v)
		return string(b)
	}
	content, _ := res["content"].([]any)
	// A second block may only be the guidance reminder.
	if len(content) == 2 {
		if b, _ := content[1].(map[string]any); b["type"] != "text" || !strings.HasPrefix(b["text"].(string), "aiport's guidance for coding agents") {
			t.Fatalf("unexpected second content block: %v", res)
		}
	} else if len(content) != 1 {
		t.Fatalf("want one content block: %v", res)
	}
	block := content[0].(map[string]any)
	var text any
	if block["type"] != "text" || json.Unmarshal([]byte(block["text"].(string)), &text) != nil {
		t.Fatalf("content is not JSON text: %v", block)
	}
	if strip(text) != strip(want) || strip(res["structuredContent"]) != strip(want) {
		t.Fatalf("MCP result differs from HTTP\n mcp: %s\nhttp: %s", strip(res["structuredContent"]), strip(want))
	}
}

func TestMCPListsItsTools(t *testing.T) {
	h, _ := requestsServer(t)
	_, out := mcp(h, reqHelper, rpc(2, "tools/list", nil))
	result, _ := out["result"].(map[string]any)
	tools, _ := result["tools"].([]any)
	names := map[string]bool{}
	for _, tool := range tools {
		tool := tool.(map[string]any)
		schema, _ := tool["inputSchema"].(map[string]any)
		if tool["description"] == "" || schema["type"] != "object" {
			t.Fatalf("tool lacks a description or an object input schema: %v", tool)
		}
		names[tool["name"].(string)] = true
	}
	if len(names) != 8 || !names["aiport_guidance"] || !names["aiport_guidance_done"] || !names["aiport_catalog"] || !names["aiport_capabilities"] || !names["aiport_request_change"] || !names["aiport_my_requests"] || !names["aiport_my_issues"] || !names["aiport_issue_fixed"] {
		t.Fatalf("tools: %v", names)
	}
}

func TestMCPCapabilitiesMatchHTTP(t *testing.T) {
	h, _ := requestsServer(t)
	for _, token := range []string{reqHelper, reqNightly} {
		res := toolCall(t, h, token, "aiport_capabilities", map[string]any{})
		if res["isError"] == true {
			t.Fatalf("capabilities: %v", res)
		}
		w, _ := call(h, "GET", CapabilitiesPath, "", token, false)
		sameJSON(t, res, w.Body.Bytes())
	}
}

func TestMCPChangeRequestsMatchHTTP(t *testing.T) {
	h, _ := requestsServer(t)
	var args map[string]any
	_ = json.Unmarshal([]byte(requestBody(map[string]any{"filed_by": "Claude Code over MCP"})), &args)

	filed := toolCall(t, h, reqHelper, "aiport_request_change", args)
	sc, _ := filed["structuredContent"].(map[string]any)
	if filed["isError"] == true || sc["status"] != "pending" || sc["id"] == "" {
		t.Fatalf("file over MCP: %v", filed)
	}
	id := sc["id"].(string)
	w, _ := call(h, "GET", RequestsPath+"/"+id, "", reqHelper, false)
	sameJSON(t, filed, w.Body.Bytes())

	one := toolCall(t, h, reqHelper, "aiport_my_requests", map[string]any{"id": id})
	sameJSON(t, one, w.Body.Bytes())
	w, _ = call(h, "GET", RequestsPath, "", reqHelper, false)
	sameJSON(t, toolCall(t, h, reqHelper, "aiport_my_requests", nil), w.Body.Bytes())

	// Another Application sees nothing of it, through either door.
	other := toolCall(t, h, reqWriter, "aiport_my_requests", map[string]any{"id": id})
	w, _ = call(h, "GET", RequestsPath+"/"+id, "", reqWriter, false)
	if other["isError"] != true || w.Code != 404 {
		t.Fatalf("another Application read the request: %v", other)
	}
	sameJSON(t, other, w.Body.Bytes())
	w, _ = call(h, "GET", RequestsPath, "", reqWriter, false)
	sameJSON(t, toolCall(t, h, reqWriter, "aiport_my_requests", nil), w.Body.Bytes())

	// Refusals carry the same codes, messages and problems.
	refused := toolCall(t, h, reqNightly, "aiport_request_change", args)
	w, _ = call(h, "POST", RequestsPath, requestBody(nil), reqNightly, false)
	if refused["isError"] != true || errorCode(refused["structuredContent"].(map[string]any)) != "change_requests_not_allowed" {
		t.Fatalf("an Application without permission filed over MCP: %v", refused)
	}
	sameJSON(t, refused, w.Body.Bytes())

	incomplete := toolCall(t, h, reqHelper, "aiport_request_change", map[string]any{"kind": "application_alias", "color": "blue"})
	w, _ = call(h, "POST", RequestsPath, `{"kind":"application_alias","color":"blue"}`, reqHelper, false)
	if incomplete["isError"] != true || w.Code != 400 {
		t.Fatalf("incomplete request: %v", incomplete)
	}
	sameJSON(t, incomplete, w.Body.Bytes())

	// The Owner's decision reads back through the tool.
	if w, _ := call(h, "POST", "/api/v1/decisions/"+id, `{"action":"dismiss","note":"not now"}`, reqManagement, true); w.Code != 200 {
		t.Fatalf("decide: %d %s", w.Code, w.Body)
	}
	decided := toolCall(t, h, reqHelper, "aiport_my_requests", map[string]any{"id": id})
	if decided["structuredContent"].(map[string]any)["note"] != "not now" {
		t.Fatalf("decision not visible over MCP: %v", decided)
	}

	if _, out := mcp(h, reqHelper, rpc(4, "tools/call", map[string]any{"name": "aiport_delete_everything"})); out["error"] == nil {
		t.Fatalf("an unknown tool was run: %v", out)
	}
}
