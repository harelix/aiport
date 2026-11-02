// Package onboarding creates an explicit, private first-run configuration.
package onboarding

import (
	"bufio"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"aiport/internal/agents"
	"aiport/internal/config"
	"aiport/internal/terminal"
	"go.yaml.in/yaml/v3"
)

func Init(args []string, in io.Reader, out io.Writer) error {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	path := fs.String("config", "aiport.yaml", "")
	provider := fs.String("provider", "", "")
	base := fs.String("base-url", "", "")
	region := fs.String("region", "", "")
	agent := fs.String("agent", "", "")
	listen := fs.String("listen", "127.0.0.1:8787", "")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fmt.Fprintln(out, "aiport init [--config PATH] [--provider "+config.ProviderIDs()+"] [--agent codex|claude|none] [--region REGION] [--base-url URL]")
			return nil
		}
		return errors.New("invalid init arguments")
	}
	if fs.NArg() != 0 {
		return errors.New("unexpected init arguments")
	}
	scanner := bufio.NewScanner(in)
	scanner.Buffer(make([]byte, 4096), 4096)
	prompt := func(label, def string) (string, error) {
		fmt.Fprintf(out, "  %s [%s]: ", label, def)
		if !scanner.Scan() {
			return "", errors.New("onboarding input ended; rerun with explicit flags for noninteractive setup")
		}
		v := strings.TrimSpace(scanner.Text())
		if v == "" {
			v = def
		}
		return v, nil
	}
	terminal.Frost(out, "\n  aiport · Set up your local gateway\n\n  1  Choose providers\n  2  Connect coding agents\n  3  Start the gateway\n\n")
	if *provider == "" {
		fmt.Fprintln(out, "  "+strings.ReplaceAll(config.ProviderIDs(), "|", " · "))
		v, err := prompt("Providers (comma-separated)", "ollama")
		if err != nil {
			return err
		}
		*provider = v
	}
	selected := []config.ProviderSpec{}
	seen := map[string]bool{}
	for _, id := range strings.Split(*provider, ",") {
		id = strings.TrimSpace(id)
		found := false
		for _, preset := range config.Providers() {
			if preset.ID == id {
				if seen[id] {
					return errors.New("duplicate provider selection")
				}
				seen[id] = true
				selected = append(selected, preset)
				found = true
				break
			}
		}
		if !found {
			return errors.New("unknown provider; choose one listed by aiport init --help")
		}
	}
	if *base != "" && len(selected) != 1 {
		return errors.New("--base-url requires a single provider")
	}
	regional := false
	for _, p := range selected {
		regional = regional || p.Regional
	}
	if regional && *region == "" {
		def := os.Getenv("AWS_REGION")
		if def == "" {
			def = os.Getenv("AWS_DEFAULT_REGION")
		}
		if def == "" {
			def = "us-east-1"
		}
		v, err := prompt("Region", def)
		if err != nil {
			return err
		}
		*region = v
	}
	if *region != "" && !config.ValidRegion(*region) {
		return errors.New("--region must be a region name such as us-east-1")
	}
	if installed := agents.Installed(); len(installed) > 0 {
		fmt.Fprintln(out, "  Detected coding clients:", strings.Join(installed, ", "))
	}
	if *agent == "" {
		v, err := prompt("Coding agents (codex, claude, none; comma-separated)", "none")
		if err != nil {
			return err
		}
		*agent = v
	}
	c, err := config.Parse([]byte("version: 1\nauth: {token_env: AIPORT_LOCAL_TOKEN}\ntelemetry: {enabled: true}\n"))
	if err != nil {
		return err
	}
	c.Listen = *listen
	for _, preset := range selected {
		// Presets are written generically: provider (+ region); aiport derives
		// protocol, endpoint and key variable when it loads the file.
		t := config.Target{ID: preset.ID, Provider: preset.ID, BaseURL: *base}
		if preset.Regional {
			t.Region = *region
		}
		c.Targets = append(c.Targets, t)
		if preset.Protocol == "openai" && c.Defaults.OpenAITarget == "" {
			c.Defaults.OpenAITarget = preset.ID
		}
		if preset.Protocol == "anthropic" && c.Defaults.AnthropicTarget == "" {
			c.Defaults.AnthropicTarget = preset.ID
		}
	}
	agentNames := []string{}
	seenAgents := map[string]bool{}
	for _, name := range strings.Split(*agent, ",") {
		name = strings.TrimSpace(name)
		if name == "none" && *agent == "none" {
			break
		}
		if name != "codex" && name != "claude" || seenAgents[name] {
			return errors.New("agents must be distinct codex, claude, or none")
		}
		seenAgents[name] = true
		target := c.Defaults.OpenAITarget
		if name == "claude" {
			target = c.Defaults.AnthropicTarget
		}
		if target == "" {
			return errors.New("agent requires a compatible native protocol; no protocol translation is provided")
		}
		agentNames = append(agentNames, name)
		c.Origins = append(c.Origins, config.Origin{ID: name, Name: name, TokenEnv: "AIPORT_" + strings.ToUpper(name) + "_TOKEN", Target: target})
	}
	local, err := token()
	if err != nil {
		return err
	}
	environment := "# Generated local aiport credentials. Keep private; never commit.\nexport AIPORT_LOCAL_TOKEN='" + local + "'\n"
	for _, origin := range c.Origins {
		v, err := token()
		if err != nil {
			return err
		}
		environment += "export " + origin.TokenEnv + "='" + v + "'\n"
	}
	if err = c.Validate(); err != nil {
		return err
	}
	encoded, err := yaml.Marshal(c)
	if err != nil {
		return errors.New("cannot encode configuration")
	}
	// Both files are exclusive: an existing setup is never overwritten.
	envPath := *path + ".env"
	for _, p := range []string{*path, envPath} {
		if _, err := os.Lstat(p); err == nil || !os.IsNotExist(err) {
			return errors.New("setup file already exists or is inaccessible; select a new --config path")
		}
	}
	if err = exclusiveWrite(envPath, []byte(environment)); err != nil {
		return err
	}
	if err = exclusiveWrite(*path, encoded); err != nil {
		_ = os.Remove(envPath)
		return err
	}
	fmt.Fprintf(out, "\n  Created %s and %s with private permissions.\n", *path, envPath)
	for _, t := range c.Targets {
		spec, _ := config.LookupProvider(t.Provider)
		fmt.Fprintf(out, "\n  %s\n    Endpoint: %s\n", spec.Name, t.URL())
		if spec.KeyEnv != "" {
			fmt.Fprintf(out, "    Add the API key in the dashboard (Settings), or set %s in your shell.\n", spec.KeyEnv)
		}
	}
	fmt.Fprintln(out, "\n  Upstream keys are never written to YAML or printed.")
	fmt.Fprintf(out, "\n  Next:\n    source %s\n    aiport check --config %s\n    aiport serve --config %s\n", quote(envPath), quote(*path), quote(*path))
	for _, name := range agentNames {
		fmt.Fprintf(out, "\n  In another terminal, source the same credential file, then:\n    aiport agent %s --config %s --model YOUR_MODEL\n", name, quote(*path))
	}
	fmt.Fprintln(out, "\n  Presets are experimental. Model and client compatibility require live checks.\n  Dashboard is served at the gateway URL. Activity history is saved in SQLite next to the config.")
	return nil
}
func token() (string, error) {
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", errors.New("cannot generate secure local credential")
	}
	return base64.RawURLEncoding.EncodeToString(b[:]), nil
}
func quote(s string) string { return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'" }
func exclusiveWrite(path string, data []byte) error {
	// Parent must already exist: do not silently change permissions on user directories.
	if _, err := os.Stat(filepath.Dir(path)); err != nil {
		return errors.New("configuration parent directory must exist")
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return errors.New("cannot create private setup file")
	}
	ok := false
	defer func() {
		_ = f.Close()
		if !ok {
			_ = os.Remove(path)
		}
	}()
	if _, err = f.Write(data); err != nil {
		return errors.New("cannot write setup file")
	}
	if err = f.Sync(); err != nil {
		return errors.New("cannot sync setup file")
	}
	if err = f.Close(); err != nil {
		return errors.New("cannot close setup file")
	}
	ok = true
	return nil
}
