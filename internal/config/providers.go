package config

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// RegionPlaceholder is replaced by a target's region in a URL template.
const RegionPlaceholder = "{region}"

// Location tells where a provider runs: on this machine or as a remote API.
type Location string

const (
	Local  Location = "local"
	Remote Location = "remote"
)

// EndpointSpec is one API surface of a provider. URLs maps each supported
// protocol (openai, anthropic) to its API prefix, which may contain {region}.
type EndpointSpec struct {
	ID   string            `json:"id"`
	Name string            `json:"name"`
	URLs map[string]string `json:"urls"`
	Note string            `json:"note,omitempty"`
}

// ProviderSpec is a built-in provider preset. A target naming a known provider
// may omit protocol, base_url and api_key_env. Regional providers need a
// region; providers with several endpoints accept `endpoint` (first is default).
type ProviderSpec struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Protocol  string         `json:"protocol"` // default protocol
	KeyEnv    string         `json:"api_key_env,omitempty"`
	Location  Location       `json:"location"`
	Regional  bool           `json:"regional"`
	Endpoints []EndpointSpec `json:"endpoints"`
	Note      string         `json:"note,omitempty"`
}

// URLFor returns the URL template for an endpoint id ("" = default) and protocol.
func (p ProviderSpec) URLFor(endpoint, protocol string) (string, bool) {
	for i, e := range p.Endpoints {
		if e.ID == endpoint || endpoint == "" && i == 0 {
			u, ok := e.URLs[protocol]
			return u, ok
		}
	}
	return "", false
}

func (p ProviderSpec) endpoint(id string) (EndpointSpec, bool) {
	for i, e := range p.Endpoints {
		if e.ID == id || id == "" && i == 0 {
			return e, true
		}
	}
	return EndpointSpec{}, false
}

func single(protocol, u string) []EndpointSpec {
	return []EndpointSpec{{ID: "default", Name: "Default", URLs: map[string]string{protocol: u}}}
}

const rg = RegionPlaceholder

var catalog = []ProviderSpec{
	{ID: "openai", Name: "OpenAI", Protocol: "openai", KeyEnv: "OPENAI_API_KEY", Location: Remote, Endpoints: single("openai", "https://api.openai.com/v1")},
	{ID: "anthropic", Name: "Anthropic", Protocol: "anthropic", KeyEnv: "ANTHROPIC_API_KEY", Location: Remote, Endpoints: single("anthropic", "https://api.anthropic.com/v1")},
	{ID: "bedrock", Name: "Amazon Bedrock", Protocol: "openai", KeyEnv: "AWS_BEARER_TOKEN_BEDROCK", Location: Remote, Regional: true,
		Note: "Uses a Bedrock API key. Model ids are Bedrock ids (for example us.anthropic.claude-sonnet-5).",
		Endpoints: []EndpointSpec{
			{ID: "runtime", Name: "Runtime", Note: "Recommended by AWS: Guardrails, prompt routing, cross-Region inference.", URLs: map[string]string{
				"openai":    "https://bedrock-runtime." + rg + ".amazonaws.com/openai/v1",
				"anthropic": "https://bedrock-runtime." + rg + ".amazonaws.com/anthropic/v1",
			}},
			{ID: "mantle", Name: "Mantle", Note: "Server-side tools (web search), background inference, Projects and Workspaces. A few OpenAI-protocol models are served under /openai/v1 instead; override base_url for those.", URLs: map[string]string{
				"openai":    "https://bedrock-mantle." + rg + ".api.aws/v1",
				"anthropic": "https://bedrock-mantle." + rg + ".api.aws/anthropic/v1",
			}},
		}},
	{ID: "meta", Name: "Meta Model API", Protocol: "openai", KeyEnv: "MODEL_API_KEY", Location: Remote, Endpoints: single("openai", "https://api.meta.ai/v1")},
	{ID: "ollama", Name: "Ollama", Protocol: "openai", Location: Local, Endpoints: single("openai", "http://127.0.0.1:11434/v1")},
	{ID: "llamacpp", Name: "llama.cpp", Protocol: "openai", Location: Local, Endpoints: single("openai", "http://127.0.0.1:8080/v1")},
}

// Providers lists the built-in presets in a stable order.
func Providers() []ProviderSpec { return append([]ProviderSpec(nil), catalog...) }

// LookupProvider returns the preset for a provider id.
func LookupProvider(id string) (ProviderSpec, bool) {
	for _, p := range catalog {
		if p.ID == id {
			return p, true
		}
	}
	return ProviderSpec{}, false
}

// ProviderIDs is the preset list used in help text.
func ProviderIDs() string {
	ids := make([]string, len(catalog))
	for i, p := range catalog {
		ids[i] = p.ID
	}
	return strings.Join(ids, "|")
}

// Cloud region names such as us-east-1, eu-central-2 or us-gov-west-1.
var regionName = regexp.MustCompile(`^[a-z]{2,4}(-[a-z0-9]+){1,3}$`)

// ValidRegion reports whether s looks like a cloud region name.
func ValidRegion(s string) bool { return len(s) <= 32 && regionName.MatchString(s) }

// resolveTarget fills preset defaults, picks the endpoint/protocol URL and
// expands {region}. The YAML stays generic; ResolvedURL is what is used.
func resolveTarget(t Target, n int) (Target, error) {
	spec, known := LookupProvider(t.Provider)
	if t.Protocol == "" {
		if !known {
			return t, fmt.Errorf("target %d needs protocol (openai or anthropic) for custom provider %q", n, t.Provider)
		}
		t.Protocol = spec.Protocol
	}
	if known && t.APIKeyEnv == "" {
		t.APIKeyEnv = spec.KeyEnv
	}
	template := t.BaseURL
	if t.Endpoint != "" {
		if !known || len(spec.Endpoints) < 2 {
			return t, fmt.Errorf("target %d: provider %q has no selectable endpoints", n, t.Provider)
		}
		if _, ok := spec.endpoint(t.Endpoint); !ok {
			names := []string{}
			for _, e := range spec.Endpoints {
				names = append(names, e.ID)
			}
			return t, fmt.Errorf("target %d endpoint must be one of %s", n, strings.Join(names, ", "))
		}
		if t.BaseURL != "" {
			return t, fmt.Errorf("target %d sets both endpoint and base_url; use one", n)
		}
	}
	if template == "" {
		if !known {
			return t, fmt.Errorf("target %d needs base_url for custom provider %q", n, t.Provider)
		}
		u, ok := spec.URLFor(t.Endpoint, t.Protocol)
		if !ok {
			e, _ := spec.endpoint(t.Endpoint)
			return t, fmt.Errorf("target %d: %s %s does not offer the %s protocol", n, spec.Name, e.Name, t.Protocol)
		}
		template = u
	}
	if t.Region != "" && !ValidRegion(t.Region) {
		return t, fmt.Errorf("target %d region must be a region name such as us-east-1", n)
	}
	uses := strings.Count(template, RegionPlaceholder)
	switch {
	case uses > 0 && t.Region == "":
		return t, fmt.Errorf("target %d (%s) needs region, for example region: us-east-1", n, t.Provider)
	case uses == 0 && t.Region != "":
		return t, fmt.Errorf("target %d sets region but its URL has no %s placeholder", n, RegionPlaceholder)
	}
	t.ResolvedURL = strings.ReplaceAll(template, RegionPlaceholder, t.Region)
	return t, nil
}

// LocationOf classifies a target: built-in presets declare it; custom
// providers are local when their endpoint is a loopback address.
func LocationOf(t Target) Location {
	if spec, ok := LookupProvider(t.Provider); ok && t.BaseURL == "" {
		return spec.Location
	}
	if loopbackEndpoint(t.URL()) {
		return Local
	}
	return Remote
}

// Resolve fills derived fields (URL, location, preset defaults) for a target.
func Resolve(t Target) (Target, error) {
	t, err := resolveTarget(t, 0)
	if err != nil {
		return t, err
	}
	t.Location = LocationOf(t)
	return t, nil
}

func loopbackEndpoint(raw string) bool {
	u, err := url.Parse(raw)
	return err == nil && loopbackHost(u.Hostname())
}
