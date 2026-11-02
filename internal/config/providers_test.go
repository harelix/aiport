package config

import (
	"os"
	"strings"
	"testing"
)

const head = "version: 1\nauth: {token_env: T}\ntargets:\n"

func TestRegionalPresetIsGeneric(t *testing.T) {
	c, err := Parse([]byte(head + "  - {id: bedrock, provider: bedrock, region: eu-central-1}\n"))
	if err != nil {
		t.Fatal(err)
	}
	b := c.Targets[0]
	if b.URL() != "https://bedrock-runtime.eu-central-1.amazonaws.com/openai/v1" || b.Protocol != "openai" || b.APIKeyEnv != "AWS_BEARER_TOKEN_BEDROCK" || b.Location != Remote {
		t.Fatalf("unexpected resolution: %+v", b)
	}
	if b.BaseURL != "" {
		t.Fatal("derived URL must not become configuration")
	}
}

func TestRegionTemplateOverride(t *testing.T) {
	c, err := Parse([]byte(head + "  - {id: mantle, provider: bedrock, region: us-west-2, base_url: 'https://bedrock-mantle.{region}.api.aws/v1'}\n"))
	if err != nil || c.Targets[0].URL() != "https://bedrock-mantle.us-west-2.api.aws/v1" {
		t.Fatal(err, c.Targets)
	}
}

func TestRegionErrors(t *testing.T) {
	for name, target := range map[string]string{
		"missing region":  "{id: b, provider: bedrock}",
		"bad region":      "{id: b, provider: bedrock, region: 'US East'}",
		"unused region":   "{id: m, provider: meta, region: us-east-1}",
		"custom no url":   "{id: x, provider: acme, protocol: openai}",
		"custom no proto": "{id: x, provider: acme, base_url: 'https://acme.example/v1'}",
	} {
		if _, err := Parse([]byte(head + "  - " + target + "\n")); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
}

func TestLocations(t *testing.T) {
	c, err := Parse([]byte(head + "  - {id: o, provider: ollama}\n  - {id: c, provider: acme, protocol: openai, base_url: 'http://localhost:9000/v1'}\n  - {id: a, provider: anthropic}\n"))
	if err != nil {
		t.Fatal(err)
	}
	got := []string{}
	for _, t := range c.Targets {
		got = append(got, string(t.Location))
	}
	if strings.Join(got, ",") != "local,local,remote" {
		t.Fatal(got)
	}
}

func TestBedrockEndpointsAndProtocols(t *testing.T) {
	cases := map[string]string{
		"{region: us-east-1}":                                        "https://bedrock-runtime.us-east-1.amazonaws.com/openai/v1",
		"{region: us-east-1, protocol: anthropic}":                   "https://bedrock-runtime.us-east-1.amazonaws.com/anthropic/v1",
		"{region: eu-west-1, endpoint: mantle}":                      "https://bedrock-mantle.eu-west-1.api.aws/v1",
		"{region: eu-west-1, endpoint: mantle, protocol: anthropic}": "https://bedrock-mantle.eu-west-1.api.aws/anthropic/v1",
	}
	for extra, want := range cases {
		body := strings.Replace(extra, "{", "{id: b, provider: bedrock, ", 1)
		c, err := Parse([]byte(head + "  - " + body + "\n"))
		if err != nil || c.Targets[0].URL() != want {
			t.Errorf("%s: got %v %v", extra, c.Targets, err)
		}
	}
	for _, bad := range []string{
		"{id: b, provider: bedrock, region: us-east-1, endpoint: nope}",
		"{id: b, provider: bedrock, region: us-east-1, endpoint: mantle, base_url: 'https://x.example/v1'}",
		"{id: m, provider: meta, endpoint: mantle}",
	} {
		if _, err := Parse([]byte(head + "  - " + bad + "\n")); err == nil {
			t.Errorf("accepted %s", bad)
		}
	}
}

func TestSaveTargetPreservesFile(t *testing.T) {
	path := t.TempDir() + "/aiport.yaml"
	original := head + "  # my bedrock\n  - id: b\n    provider: bedrock\n    region: us-east-1\n  - id: o\n    provider: ollama\n"
	if err := os.WriteFile(path, []byte(original), 0600); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	live := NewLive(path, loaded)
	save := func(id string, s TargetSettings) (Config, error) {
		c, _, err := live.Change("edit "+id, EditTarget(id, s))
		return c, err
	}
	c, err := save("b", TargetSettings{Region: "eu-central-1", Endpoint: "mantle", Protocol: "anthropic"})
	if err != nil || c.Targets[0].URL() != "https://bedrock-mantle.eu-central-1.api.aws/anthropic/v1" {
		t.Fatal(err, c.Targets)
	}
	data, _ := os.ReadFile(path)
	if !strings.Contains(string(data), "# my bedrock") || !strings.Contains(string(data), "provider: ollama") {
		t.Fatalf("file not preserved:\n%s", data)
	}
	// Back to defaults: keys removed, not set to empty.
	if _, err = save("b", TargetSettings{Region: "us-east-1"}); err != nil {
		t.Fatal(err)
	}
	data, _ = os.ReadFile(path)
	if strings.Contains(string(data), "endpoint") || strings.Contains(string(data), "protocol") {
		t.Fatalf("defaults not restored:\n%s", data)
	}
	if _, err = save("b", TargetSettings{}); err == nil {
		t.Fatal("saved bedrock without a region")
	}
	if after, _ := os.ReadFile(path); string(after) != string(data) {
		t.Fatal("invalid update modified the file")
	}
	if st, _ := os.Stat(path); st.Mode().Perm() != 0600 {
		t.Fatal("permissions changed")
	}
}
