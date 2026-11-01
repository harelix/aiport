// Package config owns the authoritative YAML configuration and startup validation.
package config

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"go.yaml.in/yaml/v3"
)

const MaxConfigBytes = 1 << 20

var identifier = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]{0,63}$`)
var envName = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

type Config struct {
	Proxy     ProxyLimits `yaml:"proxy" json:"proxy"`
	Version   int         `yaml:"version" json:"version"`
	Listen    string      `yaml:"listen" json:"listen"`
	Auth      Auth        `yaml:"auth" json:"auth"`
	Targets   []Target    `yaml:"targets" json:"targets"`
	Origins   []Origin    `yaml:"origins" json:"origins"`
	Defaults  Defaults    `yaml:"defaults" json:"defaults"`
	Aliases   []Alias     `yaml:"aliases" json:"aliases"`
	Telemetry Telemetry   `yaml:"telemetry" json:"telemetry"`
	// Evaluation selects the connection and model Aiport uses to judge
	// compression (NLI, claims, LLM judge). Off when Connection is empty.
	Evaluation Evaluation `yaml:"evaluation,omitempty" json:"evaluation"`
}

// Evaluation is the user-chosen evaluator: any configured connection.
type Evaluation struct {
	Connection string `yaml:"connection,omitempty" json:"connection,omitempty"`
	Model      string `yaml:"model,omitempty" json:"model,omitempty"`
}
type Auth struct {
	Disabled     bool   `yaml:"disabled,omitempty" json:"disabled"`
	Username     string `yaml:"username,omitempty" json:"username,omitempty"`
	PasswordHash string `yaml:"password_hash,omitempty" json:"-"`
	TokenEnv     string `yaml:"token_env" json:"token_env"`
}
type Target struct {
	ID       string `yaml:"id" json:"id"`
	Provider string `yaml:"provider" json:"provider"`
	Protocol string `yaml:"protocol,omitempty" json:"protocol"`
	BaseURL  string `yaml:"base_url,omitempty" json:"base_url,omitempty"`
	Region   string `yaml:"region,omitempty" json:"region,omitempty"`
	// Endpoint selects a provider API surface (Bedrock: runtime or mantle).
	Endpoint  string `yaml:"endpoint,omitempty" json:"endpoint,omitempty"`
	APIKeyEnv string `yaml:"api_key_env,omitempty" json:"api_key_env,omitempty"`
	Disabled  bool   `yaml:"disabled,omitempty" json:"disabled"`
	// ResolvedURL is the API prefix actually used (preset, endpoint and
	// protocol chosen, {region} expanded). Derived at load; never written.
	ResolvedURL string   `yaml:"-" json:"url"`
	Location    Location `yaml:"-" json:"location"`
}

// URL is the API prefix requests are forwarded to.
func (t Target) URL() string {
	if t.ResolvedURL != "" {
		return t.ResolvedURL
	}
	if r, err := resolveTarget(t, 0); err == nil {
		return r.ResolvedURL
	}
	return t.BaseURL
}

type Origin struct {
	ID       string `yaml:"id" json:"id"`
	Name     string `yaml:"name" json:"name"`
	TokenEnv string `yaml:"token_env" json:"token_env"`
	Target   string `yaml:"target,omitempty" json:"target,omitempty"`
}
type Defaults struct {
	OpenAITarget    string `yaml:"openai_target,omitempty" json:"openai_target,omitempty"`
	AnthropicTarget string `yaml:"anthropic_target,omitempty" json:"anthropic_target,omitempty"`
}
type Alias struct {
	Target string `yaml:"target" json:"target"`
	Alias  string `yaml:"alias" json:"alias"`
	Model  string `yaml:"model" json:"model"`
}
type Telemetry struct {
	Enabled         bool `yaml:"enabled" json:"enabled"`
	RetentionDays   int  `yaml:"retention_days" json:"retention_days"`
	QueueCapacity   int  `yaml:"queue_capacity" json:"queue_capacity"`
	BatchSize       int  `yaml:"batch_size" json:"batch_size"`
	FlushIntervalMS int  `yaml:"flush_interval_ms" json:"flush_interval_ms"`
	ShutdownDrainMS int  `yaml:"shutdown_drain_ms" json:"shutdown_drain_ms"`
	MaxEventBytes   int  `yaml:"max_event_bytes" json:"max_event_bytes"`
}

func Parse(data []byte) (Config, error) {
	c := Config{Proxy: DefaultProxyLimits(), Listen: "127.0.0.1:8787", Telemetry: Telemetry{Enabled: true, RetentionDays: 7, QueueCapacity: 4096, BatchSize: 100, FlushIntervalMS: 250, ShutdownDrainMS: 2000, MaxEventBytes: 16384}}
	if len(data) > MaxConfigBytes {
		return Config{}, errors.New("configuration exceeds 1 MiB")
	}
	d := yaml.NewDecoder(bytes.NewReader(data))
	d.KnownFields(true)
	// YAML diagnostics can echo offending values, which may contain accidentally pasted secrets.
	if err := d.Decode(&c); err != nil {
		return Config{}, errors.New("invalid configuration YAML: check field names, types and duplicate keys")
	}
	var extra any
	if err := d.Decode(&extra); err != io.EOF {
		return Config{}, errors.New("configuration must contain exactly one YAML document")
	}
	for i := range c.Targets {
		t, err := resolveTarget(c.Targets[i], i+1)
		if err != nil {
			return Config{}, err
		}
		t.Location = LocationOf(t)
		c.Targets[i] = t
	}
	if err := c.Validate(); err != nil {
		return Config{}, err
	}
	return c, nil
}

func Load(path string) (Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return Config{}, errors.New("cannot open configuration file")
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil || !st.Mode().IsRegular() {
		return Config{}, errors.New("configuration must be a regular file")
	}
	if runtime.GOOS != "windows" && st.Mode().Perm()&0077 != 0 {
		return Config{}, errors.New("configuration permissions must be 0600 or stricter (run chmod 600)")
	}
	data, err := io.ReadAll(io.LimitReader(f, MaxConfigBytes+1))
	if err != nil {
		return Config{}, errors.New("cannot read configuration file")
	}
	return Parse(data)
}

func (c Config) Validate() error {
	if c.Version != 1 {
		return errors.New("configuration version must be 1")
	}
	host, port, err := net.SplitHostPort(c.Listen)
	if err != nil {
		return errors.New("listen must be a literal loopback IP and port")
	}
	ip := net.ParseIP(host)
	n, e := strconv.Atoi(port)
	if ip == nil || !ip.IsLoopback() || e != nil || n < 1 || n > 65535 {
		return errors.New("listen must be a literal loopback IP and port 1–65535; remote serving is not available")
	}
	if !envName.MatchString(c.Auth.TokenEnv) {
		return errors.New("auth.token_env must name an environment variable")
	}
	if len(c.Auth.Username) > 64 || strings.ContainsAny(c.Auth.Username, "\r\n\x00") || !validPasswordHash(c.Auth.PasswordHash) {
		return errors.New("invalid local username or password hash")
	}
	targets := map[string]Target{}
	for i, t := range c.Targets {
		if !identifier.MatchString(t.ID) {
			return fmt.Errorf("target %d has an invalid id", i+1)
		}
		if _, ok := targets[t.ID]; ok {
			return errors.New("duplicate target id")
		}
		if !identifier.MatchString(t.Provider) {
			return fmt.Errorf("target %d has an invalid provider", i+1)
		}
		t, err := resolveTarget(t, i+1)
		if err != nil {
			return err
		}
		if t.Protocol != "openai" && t.Protocol != "anthropic" {
			return fmt.Errorf("target %d protocol must be openai or anthropic", i+1)
		}
		u, err := url.Parse(t.ResolvedURL)
		if err != nil || u.Hostname() == "" || u.User != nil || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" || u.Opaque != "" || (u.Scheme != "http" && u.Scheme != "https") {
			return fmt.Errorf("target %d base_url must be an HTTP(S) API prefix without credentials, query or fragment", i+1)
		}
		if u.Port() != "" {
			port, err := strconv.Atoi(u.Port())
			if err != nil || port < 1 || port > 65535 {
				return fmt.Errorf("target %d base_url port must be 1–65535", i+1)
			}
		}
		if u.Scheme == "http" && !loopbackHost(u.Hostname()) {
			return fmt.Errorf("target %d requires HTTPS for remote destinations", i+1)
		}
		if t.APIKeyEnv != "" && !envName.MatchString(t.APIKeyEnv) {
			return fmt.Errorf("target %d api_key_env must name an environment variable", i+1)
		}
		targets[t.ID] = t
	}
	for _, ref := range []struct{ id, protocol string }{{c.Defaults.OpenAITarget, "openai"}, {c.Defaults.AnthropicTarget, "anthropic"}} {
		if ref.id == "" {
			continue
		}
		t, ok := targets[ref.id]
		if !ok || t.Protocol != ref.protocol {
			return errors.New("default target is missing or has an incompatible protocol")
		}
	}
	origins := map[string]bool{}
	refs := map[string]bool{c.Auth.TokenEnv: true}
	for i, o := range c.Origins {
		if !identifier.MatchString(o.ID) || origins[o.ID] {
			return fmt.Errorf("origin %d has an invalid or duplicate id", i+1)
		}
		origins[o.ID] = true
		if strings.TrimSpace(o.Name) == "" || len(o.Name) > 128 || strings.ContainsAny(o.Name, "\r\n\x00") {
			return fmt.Errorf("origin %d name must contain 1–128 bytes of text", i+1)
		}
		if !envName.MatchString(o.TokenEnv) || refs[o.TokenEnv] {
			return fmt.Errorf("origin %d needs a distinct token_env reference", i+1)
		}
		refs[o.TokenEnv] = true
		if o.Target != "" {
			if _, ok := targets[o.Target]; !ok {
				return errors.New("origin references a missing target")
			}
		}
	}
	aliases := map[string]bool{}
	for _, a := range c.Aliases {
		if _, ok := targets[a.Target]; !ok {
			return errors.New("alias references a missing target")
		}
		if !boundedModel(a.Alias) || !boundedModel(a.Model) {
			return errors.New("alias and model must contain 1–256 non-control bytes")
		}
		key := a.Target + "\x00" + a.Alias
		if aliases[key] {
			return errors.New("duplicate target/model alias")
		}
		aliases[key] = true
	}
	p := c.Proxy
	if p.MaxConcurrent < 1 || p.MaxConcurrent > 1024 || p.MaxRequestBytes < 1 || p.MaxRequestBytes > 64<<20 || p.ConnectTimeoutMS < 1 || p.ConnectTimeoutMS > 300000 || p.ResponseHeaderTimeoutMS < 1 || p.ResponseHeaderTimeoutMS > 600000 || p.StreamIdleTimeoutMS < 1 || p.StreamIdleTimeoutMS > 3600000 || p.RequestReadTimeoutMS < 1 || p.RequestReadTimeoutMS > 600000 {
		return errors.New("proxy limits out of range")
	}
	if e := c.Evaluation; e.Connection != "" || e.Model != "" {
		if _, ok := targets[e.Connection]; !ok {
			return errors.New("evaluation.connection must name a configured connection")
		}
		if !boundedModel(e.Model) {
			return errors.New("evaluation.model must name the model to use (1–256 characters)")
		}
	}
	t := c.Telemetry
	if t.RetentionDays < 1 || t.RetentionDays > 3650 || t.QueueCapacity < 1 || t.QueueCapacity > 65536 || t.BatchSize < 1 || t.BatchSize > t.QueueCapacity || t.FlushIntervalMS < 1 || t.FlushIntervalMS > 60000 || t.ShutdownDrainMS < 1 || t.ShutdownDrainMS > 60000 || t.MaxEventBytes < 1024 || t.MaxEventBytes > 65536 {
		return errors.New("telemetry limits out of range")
	}
	return nil
}
func loopbackHost(host string) bool {
	return host == "localhost" || net.ParseIP(host) != nil && net.ParseIP(host).IsLoopback()
}
func boundedModel(s string) bool {
	if len(s) < 1 || len(s) > 256 {
		return false
	}
	for _, r := range s {
		if r < 32 || r == 127 {
			return false
		}
	}
	return true
}

// Secrets are resolved once at startup and are never part of serialized config.
type Credential struct {
	Digest     [32]byte
	OriginID   string
	Management bool
}
type Secrets struct {
	Credentials []Credential
	Upstream    map[string]string
	// MissingKeys lists enabled targets whose API key is not set yet. The server
	// still starts so the key can be entered in the dashboard.
	MissingKeys []string
}

func (c Config) Resolve(lookup func(string) string) (Secrets, error) {
	s := Secrets{Upstream: map[string]string{}}
	seen := map[[32]byte]bool{}
	add := func(ref, origin string, management bool) error {
		token := lookup(ref)
		if token == "" && c.Auth.Disabled {
			return nil
		}
		if len(token) < 32 || len(token) > 512 {
			return errors.New("local token must contain 32–512 bytes; generate with aiport token")
		}
		for _, r := range token {
			if r < 33 || r > 126 {
				return errors.New("local token must contain printable ASCII without spaces")
			}
		}
		digest := sha256.Sum256([]byte(token))
		if seen[digest] {
			return errors.New("local credentials must have distinct values")
		}
		seen[digest] = true
		s.Credentials = append(s.Credentials, Credential{digest, origin, management})
		return nil
	}
	if err := add(c.Auth.TokenEnv, "unknown", true); err != nil {
		return Secrets{}, err
	}
	for _, o := range c.Origins {
		if err := add(o.TokenEnv, o.ID, false); err != nil {
			return Secrets{}, err
		}
	}
	for _, t := range c.Targets {
		if t.APIKeyEnv == "" || t.Disabled {
			continue
		}
		value := lookup(t.APIKeyEnv)
		if value == "" {
			s.MissingKeys = append(s.MissingKeys, t.ID)
			continue
		}
		if len(value) > 8192 || strings.ContainsAny(value, "\r\n\x00") {
			return Secrets{}, errors.New("invalid upstream credential value")
		}
		if seen[sha256.Sum256([]byte(value))] {
			return Secrets{}, errors.New("upstream and local credentials must be distinct")
		}
		s.Upstream[t.ID] = value
	}
	return s, nil
}
