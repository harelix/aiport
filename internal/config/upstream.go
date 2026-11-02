package config

import (
	"errors"
	"regexp"
)

// Upstream is the GitHub repository of aiport itself, where the Owner may
// file "Suggest to aiport" issues and share Extensions. aiport only prefills
// drafts for it; it keeps no GitHub credential.
type Upstream struct {
	// Repository is owner/name on GitHub. Empty means the one aiport was
	// built with (BuiltInUpstream), if any.
	Repository string `yaml:"repository,omitempty" json:"repository,omitempty"`
}

// BuiltInUpstream is the upstream repository set at build time:
// -ldflags "-X aiport/internal/config.BuiltInUpstream=owner/name".
var BuiltInUpstream = ""

var repositoryPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9-]{0,38}/[A-Za-z0-9._-]{1,100}$`)

// Repo is the configured repository, or BuiltInUpstream when none is
// configured; empty when neither names a valid GitHub owner/name.
func (u Upstream) Repo() string {
	r := u.Repository
	if r == "" {
		r = BuiltInUpstream
	}
	if !repositoryPattern.MatchString(r) {
		return ""
	}
	return r
}

func (u Upstream) validate() error {
	if u.Repository != "" && !repositoryPattern.MatchString(u.Repository) {
		return errors.New("upstream.repository must be a GitHub owner/name, such as owner/aiport")
	}
	return nil
}
