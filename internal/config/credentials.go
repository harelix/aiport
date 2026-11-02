package config

import (
	"errors"
	"io"
	"os"
	"sort"
	"strings"
)

// CredentialFile is the private credential file kept next to a config
// (aiport.yaml -> aiport.yaml.env). It holds random local tokens and provider
// API keys entered in the dashboard. It is never part of the YAML.
func CredentialFile(configPath string) string { return configPath + ".env" }

const credentialHeader = "# aiport credentials: local tokens and provider API keys entered in the dashboard.\n# Private (0600). Never commit or share this file.\n"

// ReadCredentialFile parses `export NAME='value'` lines. A missing file yields an
// empty map; a file readable by group/others is refused.
func ReadCredentialFile(path string) (map[string]string, error) {
	values := map[string]string{}
	f, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return values, nil
	}
	if err != nil {
		return nil, errors.New("cannot open credential file " + path)
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil || !st.Mode().IsRegular() || st.Mode().Perm()&0077 != 0 {
		return nil, errors.New("credential file must be private (chmod 600 " + path + ")")
	}
	raw, err := io.ReadAll(io.LimitReader(f, MaxConfigBytes+1))
	if err != nil || len(raw) > MaxConfigBytes {
		return nil, errors.New("cannot read credential file " + path)
	}
	for _, line := range strings.Split(string(raw), "\n") {
		rest, ok := strings.CutPrefix(strings.TrimSpace(line), "export ")
		if !ok {
			continue
		}
		name, value, ok := strings.Cut(rest, "='")
		if !ok || !strings.HasSuffix(value, "'") || !envName.MatchString(name) {
			continue
		}
		value = strings.TrimSuffix(value, "'")
		if strings.ContainsAny(value, "'\r\n\x00") {
			continue
		}
		values[name] = value
	}
	return values, nil
}

// ValidCredentialValue reports whether value can be stored in the credential
// file and sent as an HTTP header value.
func ValidCredentialValue(value string) bool {
	if value == "" || len(value) > 8192 {
		return false
	}
	for _, r := range value {
		if r < 33 || r > 126 || r == '\'' {
			return false
		}
	}
	return true
}

// SetCredential stores (or, with an empty value, removes) one variable in the
// private credential file, atomically and with 0600 permissions. Other entries
// are preserved.
func SetCredential(path, name, value string) error {
	if !envName.MatchString(name) {
		return errors.New("invalid credential name")
	}
	if value != "" && !ValidCredentialValue(value) {
		return errors.New("API key must be printable text without spaces or quotes")
	}
	values, err := ReadCredentialFile(path)
	if err != nil {
		return err
	}
	if value == "" {
		delete(values, name)
	} else {
		values[name] = value
	}
	names := make([]string, 0, len(values))
	for n := range values {
		names = append(names, n)
	}
	sort.Strings(names)
	var b strings.Builder
	b.WriteString(credentialHeader)
	for _, n := range names {
		b.WriteString("export " + n + "='" + values[n] + "'\n")
	}
	if st, err := os.Lstat(path); err == nil && !st.Mode().IsRegular() {
		return errors.New("credential file must be a regular file")
	}
	if err := replacePrivate(path, []byte(b.String())); err != nil {
		return errors.New("cannot replace credential file")
	}
	return nil
}
