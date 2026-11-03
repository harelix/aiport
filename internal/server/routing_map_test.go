package server

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"aiport/internal/config"
	"aiport/internal/telemetry"
)

// The Routing map over HTTP: the Owner sees, for every Application, each
// name it can send and the Connection and model serving it, built from the
// same Catalog as that Application's /v1/models.

var (
	mapManagement = strings.Repeat("a", 40)
	mapDocs       = strings.Repeat("b", 40)
	mapReview     = strings.Repeat("e", 40)
	mapKey        = "sk-routing-map-upstream-key-1234567890"
)

type mapFixture struct {
	h                     *Handler
	hosted, local, claude *catalogUpstream
}

// protocolOf is the protocol of each Connection in mapServer.
var mapProtocolOf = map[string]string{"hosted": "openai", "local": "openai", "claude": "anthropic"}

// mapServer: hosted and local speak OpenAI and list models, claude speaks
// Anthropic. Shared: Alias quick on hosted, Levels best (hosted gpt-4o),
// fast (local llama3:8b) and balanced (claude claude-sonnet). docs-bot
// follows the defaults and has its own Alias summarize; review-code is bound
// to local and has its own best on local.
func mapServer(t *testing.T) *mapFixture {
	t.Helper()
	f := &mapFixture{
		hosted: newCatalogUpstream(t, 200, `{"object":"list","data":[{"id":"gpt-4o"},{"id":"gpt-4o-mini"}]}`),
		local:  newCatalogUpstream(t, 200, `{"object":"list","data":[{"id":"llama3:8b"},{"id":"llama3:70b"},{"id":"qwen2.5"}]}`),
		claude: newCatalogUpstream(t, 200, `{"data":[{"id":"claude-haiku"},{"id":"claude-sonnet"}]}`),
	}
	yaml := "version: 1\nauth: {token_env: MAP_MGMT_TOKEN}\ntargets:\n" +
		"  - {id: hosted, provider: openai, protocol: openai, base_url: '" + f.hosted.srv.URL + "/v1', api_key_env: MAP_HOSTED_KEY}\n" +
		"  - {id: local, provider: ollama, protocol: openai, base_url: '" + f.local.srv.URL + "/v1'}\n" +
		"  - {id: claude, provider: anthropic, protocol: anthropic, base_url: '" + f.claude.srv.URL + "/v1'}\n" +
		"defaults: {openai_target: hosted, anthropic_target: claude}\n" +
		"aliases:\n" +
		"  - {target: hosted, alias: quick, model: gpt-4o-mini}\n" +
		"  - {target: hosted, alias: best, model: gpt-4o}\n" +
		"  - {target: local, alias: fast, model: 'llama3:8b'}\n" +
		"  - {target: claude, alias: balanced, model: claude-sonnet}\n" +
		"origins:\n" +
		"  - id: docs-bot\n    name: Docs Bot\n    token_env: MAP_DOCS_TOKEN\n    aliases:\n      - {alias: summarize, target: hosted, model: gpt-4o}\n" +
		"  - id: review-code\n    name: Review Code\n    token_env: MAP_REVIEW_TOKEN\n    target: local\n    aliases:\n      - {alias: best, target: local, model: 'llama3:70b'}\n"
	path := filepath.Join(t.TempDir(), "aiport.yaml")
	if err := os.WriteFile(path, []byte(yaml), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MAP_MGMT_TOKEN", mapManagement)
	t.Setenv("MAP_DOCS_TOKEN", mapDocs)
	t.Setenv("MAP_REVIEW_TOKEN", mapReview)
	t.Setenv("MAP_HOSTED_KEY", mapKey)
	c, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	s, err := c.Resolve(config.EnvKeys(path).Lookup(c))
	if err != nil {
		t.Fatal(err)
	}
	f.h = NewConfigured(c, s, telemetry.New(true, 16), path).(*Handler)
	now := time.UnixMilli(1_800_000_000_000)
	f.h.now = func() time.Time { return now }
	t.Cleanup(f.h.Close)
	return f
}

type mapRoute struct {
	Name       string `json:"name"`
	Kind       string `json:"kind"`
	Protocol   string `json:"protocol"`
	Connection string `json:"connection"`
	Model      string `json:"model"`
	Shared     bool   `json:"shared"`
}

type mapApplication struct {
	Application string     `json:"application"`
	Name        string     `json:"name"`
	Version     string     `json:"version"`
	Routes      []mapRoute `json:"routes"`
	Models      []struct {
		ID         string `json:"id"`
		Connection string `json:"connection"`
	} `json:"models"`
	Changes []map[string]any `json:"changes"`
}

func routingMap(t *testing.T, h *Handler, token, query string) (*httptest.ResponseRecorder, []mapApplication) {
	t.Helper()
	r := httptest.NewRequest("GET", "http://127.0.0.1:8787/api/v1/routing"+query, nil)
	r.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	var out struct {
		Applications []mapApplication `json:"applications"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &out)
	return w, out.Applications
}

func TestTheRoutingMapNeedsTheManagementToken(t *testing.T) {
	f := mapServer(t)
	for _, token := range []string{mapDocs, ""} {
		if w, _ := routingMap(t, f.h, token, ""); w.Code != 401 {
			t.Fatalf("GET /api/v1/routing with token %q: %d %s", token, w.Code, w.Body)
		}
	}
	if w, apps := routingMap(t, f.h, mapManagement, ""); w.Code != 200 || len(apps) != 2 {
		t.Fatalf("GET /api/v1/routing with the management token: %d %s", w.Code, w.Body)
	}
}

// Each Application's names on each protocol are exactly what its own
// /v1/models lists there, in the same order, and its version is the one its
// answers carry in X-Aiport-Catalog.
func TestTheRoutingMapMatchesEachApplicationsModels(t *testing.T) {
	f := mapServer(t)
	_, apps := routingMap(t, f.h, mapManagement, "")
	tokens := map[string]string{"docs-bot": mapDocs, "review-code": mapReview}
	if len(apps) != 2 {
		t.Fatalf("the map lists %d Applications, want 2", len(apps))
	}
	// Both halves are compared on real lists: the reached Connections'
	// models must be in the answer, not just agree on being empty.
	compared := 0
	defer func() {
		if compared == 0 && !t.Failed() {
			t.Errorf("no listed model was compared")
		}
	}()
	for _, app := range apps {
		compared += len(app.Models)
		token := tokens[app.Application]
		if token == "" {
			t.Fatalf("unexpected Application %q in the map", app.Application)
		}
		for _, protocol := range []string{"openai", "anthropic"} {
			var header []string
			if protocol == "anthropic" {
				header = []string{"anthropic-version", "2023-06-01"}
			}
			w, out := listModels(f.h, token, header...)
			if w.Code != 200 {
				t.Fatalf("/v1/models of %s: %d %s", app.Application, w.Code, w.Body)
			}
			if got := w.Header().Get(CatalogHeader); got != app.Version {
				t.Fatalf("%s: map version %q, X-Aiport-Catalog %q", app.Application, app.Version, got)
			}
			want := []string{}
			for _, e := range entries(t, out, "id") {
				want = append(want, e[0])
			}
			got := []string{}
			for _, r := range app.Routes {
				if r.Protocol == protocol && r.Kind != "default" {
					got = append(got, r.Name)
				}
			}
			for _, m := range app.Models {
				if mapProtocolOf[m.Connection] == protocol {
					got = append(got, m.ID)
				}
			}
			if strings.Join(got, ",") != strings.Join(want, ",") {
				t.Fatalf("%s on %s: map lists %v, /v1/models %v", app.Application, protocol, got, want)
			}
		}
	}
}

func TestTheRoutingMapNamesTheConnectionAndModelServingEachName(t *testing.T) {
	f := mapServer(t)
	_, apps := routingMap(t, f.h, mapManagement, "")
	byApp := map[string][]mapRoute{}
	for _, a := range apps {
		byApp[a.Application] = a.Routes
	}
	want := map[string][]mapRoute{
		// docs-bot: own task name, shared task name, the shared Levels per
		// protocol, then where any other name goes on each protocol.
		"docs-bot": {
			{Name: "summarize", Kind: "task_name", Protocol: "openai", Connection: "hosted", Model: "gpt-4o"},
			{Name: "quick", Kind: "task_name", Protocol: "openai", Connection: "hosted", Model: "gpt-4o-mini", Shared: true},
			{Name: "fast", Kind: "level", Protocol: "openai", Connection: "local", Model: "llama3:8b", Shared: true},
			{Name: "balanced", Kind: "level", Protocol: "anthropic", Connection: "claude", Model: "claude-sonnet", Shared: true},
			{Name: "best", Kind: "level", Protocol: "openai", Connection: "hosted", Model: "gpt-4o", Shared: true},
			{Kind: "default", Protocol: "openai", Connection: "hosted"},
			{Kind: "default", Protocol: "anthropic", Connection: "claude"},
		},
		// review-code is bound to local (OpenAI): no shared task name of
		// hosted, its own best, and no default on Anthropic.
		"review-code": {
			{Name: "fast", Kind: "level", Protocol: "openai", Connection: "local", Model: "llama3:8b", Shared: true},
			{Name: "balanced", Kind: "level", Protocol: "anthropic", Connection: "claude", Model: "claude-sonnet", Shared: true},
			{Name: "best", Kind: "level", Protocol: "openai", Connection: "local", Model: "llama3:70b"},
			{Kind: "default", Protocol: "openai", Connection: "local"},
		},
	}
	for app, rows := range want {
		got := byApp[app]
		if len(got) != len(rows) {
			t.Fatalf("%s routes %+v, want %+v", app, got, rows)
		}
		for i := range rows {
			if got[i] != rows[i] {
				t.Fatalf("%s route %d: %+v, want %+v", app, i, got[i], rows[i])
			}
		}
	}
}

func TestTheRoutingMapDisclosesNoKeysOrURLs(t *testing.T) {
	f := mapServer(t)
	r := httptest.NewRequest("GET", "http://127.0.0.1:8787/api/v1/routing", nil)
	r.Header.Set("Authorization", "Bearer "+mapManagement)
	w := httptest.NewRecorder()
	f.h.ServeHTTP(w, r)
	for _, secret := range []string{f.hosted.srv.URL, f.local.srv.URL, f.claude.srv.URL, "127.0.0.1", "http", mapKey, "MAP_", mapDocs, mapReview} {
		if strings.Contains(w.Body.String(), secret) {
			t.Fatalf("the Routing map discloses %q: %s", secret, w.Body)
		}
	}
}

// One Application's view (Integrate) is the same entry with its recent
// Catalog changes, newest first.
func TestOneApplicationsViewCarriesItsRecentCatalogChanges(t *testing.T) {
	f := mapServer(t)
	_, all := routingMap(t, f.h, mapManagement, "")
	if _, _, err := f.h.live.Change("add review", config.AddOriginAlias("docs-bot", config.OriginAlias{Alias: "review", Target: "hosted", Model: "gpt-4o"})); err != nil {
		t.Fatal(err)
	}
	w, one := routingMap(t, f.h, mapManagement, "?application=docs-bot")
	if w.Code != 200 || len(one) != 1 || one[0].Application != "docs-bot" || one[0].Name != "Docs Bot" {
		t.Fatalf("GET /api/v1/routing?application=docs-bot: %d %s", w.Code, w.Body)
	}
	if one[0].Version == all[0].Version {
		t.Fatalf("the version did not change with the Catalog")
	}
	if one[0].Routes[0].Name != "review" {
		t.Fatalf("the new task name is not listed first: %+v", one[0].Routes)
	}
	if len(one[0].Changes) != 1 || one[0].Changes[0]["after"] != one[0].Version {
		t.Fatalf("recent Catalog changes: %v", one[0].Changes)
	}
	if w, _ := routingMap(t, f.h, mapManagement, "?application=nobody"); w.Code != 404 {
		t.Fatalf("an unknown Application: %d %s", w.Code, w.Body)
	}
}

// Per-effort models a task name carries reach its row on the map unchanged.
func TestTheRoutingMapShowsPerEffortModels(t *testing.T) {
	f := mapServer(t)
	edit := config.AddOriginAlias("docs-bot", config.OriginAlias{Alias: "review", Target: "hosted", Model: "gpt-4o", Effort: config.EffortModels{Low: "gpt-4o-mini"}})
	if _, _, err := f.h.live.Change("add review", edit); err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest("GET", "http://127.0.0.1:8787/api/v1/routing?application=docs-bot", nil)
	r.Header.Set("Authorization", "Bearer "+mapManagement)
	w := httptest.NewRecorder()
	f.h.ServeHTTP(w, r)
	var out struct {
		Applications []struct {
			Routes []struct {
				Name   string            `json:"name"`
				Effort map[string]string `json:"effort"`
			} `json:"routes"`
		} `json:"applications"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil || len(out.Applications) != 1 {
		t.Fatalf("GET /api/v1/routing?application=docs-bot: %d %s", w.Code, w.Body)
	}
	got := out.Applications[0].Routes[0]
	if got.Name != "review" || got.Effort["low"] != "gpt-4o-mini" || len(got.Effort) != 1 {
		t.Fatalf("the review row: %+v", got)
	}
}
