package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/harelix/aiport/internal/config"
)

// runAuth reveals the local management token only on explicit CLI request.
func runAuth(args []string, out io.Writer) error {
	if len(args) == 0 || args[0] != "token" {
		return errors.New("usage: aiport auth token [--config PATH]")
	}
	fs := flag.NewFlagSet("auth token", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	path := fs.String("config", "aiport.yaml", "")
	if fs.Parse(args[1:]) != nil || fs.NArg() != 0 {
		return errors.New("invalid auth token arguments")
	}
	c, err := config.Load(*path)
	if err != nil {
		return err
	}
	if c.Auth.Disabled {
		return errors.New("authentication is disabled; no management token is needed")
	}
	// Same lookup as serve: exported variable, else the private credential file,
	// else create and save one. Notices go to stderr so $(aiport auth token)
	// captures only the token.
	if err := ensureLocalCredentials(*path, c, os.Stderr); err != nil {
		return err
	}
	token := os.Getenv(c.Auth.TokenEnv)
	if len(token) < 32 || len(token) > 512 || strings.ContainsAny(token, "\r\n\x00'") {
		return errors.New("management token is invalid; check " + c.Auth.TokenEnv + " in this shell and " + config.CredentialFile(*path))
	}
	fmt.Fprintln(out, token)
	return nil
}
