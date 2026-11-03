// Package agents configures coding clients for a single gateway-routed launch.
package agents

import (
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/harelix/aiport/internal/config"
)

type Launch struct {
	Name      string
	Args, Env []string
}

func Prepare(c config.Config, name, model string, extra, environment []string) (Launch, error) {
	if name != "codex" && name != "claude" {
		return Launch{}, errors.New("choose codex or claude")
	}
	if model == "" || strings.ContainsAny(model, "\r\n\x00") || len(model) > 256 {
		return Launch{}, errors.New("provide --model with an upstream model identifier")
	}
	var origin config.Origin
	for _, o := range c.Origins {
		if o.ID == name {
			origin = o
		}
	}
	if origin.ID == "" {
		return Launch{}, errors.New("configure an origin whose id matches the coding agent first")
	}
	values := map[string]string{}
	for _, entry := range environment {
		parts := strings.SplitN(entry, "=", 2)
		if len(parts) == 2 {
			values[parts[0]] = parts[1]
		}
	}
	token := values[origin.TokenEnv]
	if token == "" && c.Auth.Disabled {
		token = "local-auth-disabled-placeholder-token"
	}
	if len(token) < 32 || len(token) > 512 {
		return Launch{}, errors.New("agent local token is missing or invalid; source the onboarding credential file")
	}
	protocol := "openai"
	if name == "claude" {
		protocol = "anthropic"
	}
	targetID := origin.Target
	if targetID == "" {
		if protocol == "openai" {
			targetID = c.Defaults.OpenAITarget
		} else {
			targetID = c.Defaults.AnthropicTarget
		}
	}
	var target config.Target
	for _, t := range c.Targets {
		if t.ID == targetID {
			target = t
		}
	}
	if target.ID == "" || target.Disabled || target.Protocol != protocol {
		return Launch{}, errors.New("agent target is missing, disabled, or uses an incompatible native protocol")
	}
	// The agent needs its local credential, not upstream keys or the management token.
	delete(values, c.Auth.TokenEnv)
	for _, o := range c.Origins {
		delete(values, o.TokenEnv)
	}
	for _, t := range c.Targets {
		if t.APIKeyEnv != "" {
			delete(values, t.APIKeyEnv)
		}
	}
	delete(values, "ANTHROPIC_AUTH_TOKEN")
	base := "http://" + c.Listen
	launch := Launch{Name: name}
	if name == "codex" {
		values[origin.TokenEnv] = token
		settings := []string{`model_provider="aiport"`, `model_providers.aiport.name="aiport"`, `model_providers.aiport.base_url=` + strconv.Quote(base+"/v1"), `model_providers.aiport.env_key=` + strconv.Quote(origin.TokenEnv), `model_providers.aiport.wire_api="responses"`, `model_providers.aiport.requires_openai_auth=false`, `model_providers.aiport.supports_websockets=false`, `model_providers.aiport.request_max_retries=0`, `model_providers.aiport.stream_max_retries=0`}
		for _, setting := range settings {
			launch.Args = append(launch.Args, "-c", setting)
		}
	} else {
		delete(values, "ANTHROPIC_API_KEY")
		delete(values, "CLAUDE_CODE_OAUTH_TOKEN")
		values["ANTHROPIC_BASE_URL"] = base
		values["ANTHROPIC_AUTH_TOKEN"] = token
		values["CLAUDE_CODE_USE_BEDROCK"] = "0"
		values["CLAUDE_CODE_USE_VERTEX"] = "0"
		values["CLAUDE_CODE_USE_FOUNDRY"] = "0"
	}
	launch.Args = append(launch.Args, "--model", model)
	launch.Args = append(launch.Args, extra...)
	for key, value := range values {
		launch.Env = append(launch.Env, fmt.Sprintf("%s=%s", key, value))
	}
	return launch, nil
}
func Installed() []string {
	var names []string
	for _, n := range []string{"codex", "claude"} {
		if executableExists(n) {
			names = append(names, n)
		}
	}
	return names
}
func executableExists(name string) bool { _, err := exec.LookPath(name); return err == nil }
