package main

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"aiport/internal/config"
)

// ensureLocalCredentials makes local tokens available without manual exports.
// For every local token reference in the config (management + origins):
//  1. a value already set in the environment wins;
//  2. otherwise the value saved in the private credential file is used;
//  3. otherwise a new random token is generated, saved (0600) and used.
//
// Tokens are therefore stable across restarts and terminals.
func ensureLocalCredentials(configPath string, c config.Config, out io.Writer) error {
	path := config.CredentialFile(configPath)
	saved, err := config.ReadCredentialFile(path)
	if err != nil {
		return err
	}
	refs := []string{}
	if !c.Auth.Disabled {
		refs = append(refs, c.Auth.TokenEnv)
	}
	for _, o := range c.Origins {
		refs = append(refs, o.TokenEnv)
	}
	var created []string
	for _, ref := range refs {
		if os.Getenv(ref) != "" {
			continue
		}
		v := saved[ref]
		if v == "" {
			if v, err = newLocalToken(); err != nil {
				return err
			}
			if err = config.SetCredential(path, ref, v); err != nil {
				return err
			}
			created = append(created, ref)
		}
		if err := os.Setenv(ref, v); err != nil {
			return errors.New("cannot load local credential " + ref)
		}
	}
	if len(created) > 0 {
		fmt.Fprintf(out, "Created local credentials (%s) and saved them to %s (private, reused on every start).\n", strings.Join(created, ", "), path)
	}
	return nil
}

// credentialLookup resolves configuration references. Provider API keys saved
// from the dashboard take precedence over the environment, so a key entered in
// Settings keeps applying after a restart. Everything else comes from the
// environment (local tokens were loaded by ensureLocalCredentials).
func credentialLookup(configPath string, c config.Config) (func(string) string, error) {
	saved, err := config.ReadCredentialFile(config.CredentialFile(configPath))
	if err != nil {
		return nil, err
	}
	provider := map[string]bool{}
	for _, t := range c.Targets {
		if t.APIKeyEnv != "" {
			provider[t.APIKeyEnv] = true
		}
	}
	return func(name string) string {
		if provider[name] && saved[name] != "" {
			return saved[name]
		}
		return os.Getenv(name)
	}, nil
}

func newLocalToken() (string, error) {
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", errors.New("cannot generate secure local credential")
	}
	return base64.RawURLEncoding.EncodeToString(b[:]), nil
}
