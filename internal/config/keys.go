package config

import "os"

// Where a connection's API key comes from. Values are never shown anywhere.
const (
	KeyNotNeeded   = "not_needed"  // the connection has no api_key_env (local servers)
	KeySaved       = "saved"       // entered in the dashboard, kept in the private credential file
	KeyEnvironment = "environment" // the server's environment
	KeyMissing     = "missing"
)

// Keys is the one answer to "which value does a credential reference have,
// and where does it come from". Provider API keys saved in the private
// credential file win over the environment, so a key entered in the
// dashboard keeps applying after a restart. Local tokens come from the
// environment first, then the credential file.
type Keys struct {
	File string              // private credential file; "" = none
	Env  func(string) string // the process environment (os.Getenv in production)
}

// EnvKeys reads the real environment and the credential file next to configPath.
func EnvKeys(configPath string) Keys {
	k := Keys{Env: os.Getenv}
	if configPath != "" {
		k.File = CredentialFile(configPath)
	}
	return k
}

func (k Keys) saved() map[string]string {
	if k.File == "" {
		return map[string]string{}
	}
	s, err := ReadCredentialFile(k.File)
	if err != nil {
		return map[string]string{}
	}
	return s
}

func (k Keys) env(name string) string {
	if k.Env == nil {
		return ""
	}
	return k.Env(name)
}

// Lookup returns the reference resolver Config.Resolve needs for c.
func (k Keys) Lookup(c Config) func(string) string {
	saved := k.saved()
	provider := map[string]bool{}
	for _, t := range c.Targets {
		if t.APIKeyEnv != "" {
			provider[t.APIKeyEnv] = true
		}
	}
	return func(name string) string {
		if provider[name] {
			if v := saved[name]; v != "" {
				return v
			}
			return k.env(name)
		}
		if v := k.env(name); v != "" {
			return v
		}
		return saved[name]
	}
}

// Provider returns the key a connection uses and its source. Values that
// cannot be sent as a header are treated as missing.
func (k Keys) Provider(t Target) (value, source string) {
	if t.APIKeyEnv == "" {
		return "", KeyNotNeeded
	}
	if v := k.saved()[t.APIKeyEnv]; ValidCredentialValue(v) {
		return v, KeySaved
	}
	if v := k.env(t.APIKeyEnv); ValidCredentialValue(v) {
		return v, KeyEnvironment
	}
	return "", KeyMissing
}
