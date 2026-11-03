package agents

import (
	"strings"
	"testing"

	"github.com/harelix/aiport/internal/config"
)

func TestAgentLaunch(t *testing.T) {
	c, err := config.Parse([]byte(`version: 1
auth: {token_env: LOCAL_TOKEN}
targets:
 - {id: openai, provider: openai, protocol: openai, base_url: 'https://api.openai.com/v1', api_key_env: OPENAI_API_KEY}
 - {id: anthropic, provider: anthropic, protocol: anthropic, base_url: 'https://api.anthropic.com/v1', api_key_env: ANTHROPIC_API_KEY}
origins:
 - {id: codex, name: Codex, token_env: CODEX_TOKEN, target: openai}
 - {id: claude, name: Claude Code, token_env: CLAUDE_TOKEN, target: anthropic}
`))
	if err != nil {
		t.Fatal(err)
	}
	environment := []string{"PATH=/usr/bin", "LOCAL_TOKEN=management-canary", "CODEX_TOKEN=" + strings.Repeat("c", 43), "CLAUDE_TOKEN=" + strings.Repeat("d", 43), "OPENAI_API_KEY=upstream-canary", "ANTHROPIC_API_KEY=upstream-canary", "ANTHROPIC_AUTH_TOKEN=old-secret"}
	for _, name := range []string{"codex", "claude"} {
		launch, err := Prepare(c, name, "synthetic", nil, environment)
		if err != nil {
			t.Fatal(err)
		}
		env := strings.Join(launch.Env, "\n")
		args := strings.Join(launch.Args, " ")
		if strings.Contains(env, "management-canary") || strings.Contains(env, "upstream-canary") {
			t.Fatal("upstream or management key exposed to agent")
		}
		if strings.Contains(args, strings.Repeat("c", 43)) || strings.Contains(args, strings.Repeat("d", 43)) {
			t.Fatal("token in process arguments")
		}
		if name == "codex" && (!strings.Contains(args, `wire_api="responses"`) || !strings.Contains(args, `request_max_retries=0`)) {
			t.Fatal("Codex native config missing")
		}
		if name == "claude" && (!strings.Contains(env, "ANTHROPIC_BASE_URL=http://127.0.0.1:8787") || strings.Contains(env, "old-secret")) {
			t.Fatal("Claude gateway env incorrect")
		}
	}
	c.Origins[0].Target = "anthropic"
	if _, err := Prepare(c, "codex", "synthetic", nil, environment); err == nil {
		t.Fatal("accepted protocol mismatch")
	}
}
