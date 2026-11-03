package main

import (
	"aiport/internal/agents"
	"aiport/internal/onboarding"
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"aiport/internal/config"
	"aiport/internal/server"
	"aiport/internal/store"
	"aiport/internal/telemetry"
	"aiport/internal/terminal"
	"aiport/internal/worktree"
)

var version = "0.0.0-dev"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(ctx, os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "aiport:", err)
		os.Exit(1)
	}
}
func run(ctx context.Context, args []string, out, errOut io.Writer) error {
	if len(args) == 0 {
		return errors.New("usage: aiport <init|serve|check|agent|auth|task|token|version> [--config aiport.yaml]")
	}
	switch args[0] {
	case "auth":
		return runAuth(args[1:], out)
	case "init":
		return onboarding.Init(args[1:], os.Stdin, out)
	case "agent":
		return runAgent(ctx, args[1:], out, errOut)
	case "task":
		return runTask(args[1:], out)
	case "version":
		if len(args) != 1 {
			return errors.New("version takes no arguments")
		}
		fmt.Fprintln(out, "aiport", version)
		return nil
	case "token":
		if len(args) != 1 {
			return errors.New("token takes no arguments")
		}
		var b [32]byte
		if _, err := rand.Read(b[:]); err != nil {
			return errors.New("cannot generate secure token")
		}
		fmt.Fprintln(out, base64.RawURLEncoding.EncodeToString(b[:]))
		return nil
	case "serve", "check":
	default:
		return errors.New("unknown command; use init, serve, check, agent, auth, task, token or version")
	}
	flags := flag.NewFlagSet(args[0], flag.ContinueOnError)
	// Do not let flag errors echo arbitrary command-line values into diagnostics.
	flags.SetOutput(io.Discard)
	path := flags.String("config", "aiport.yaml", "configuration path")
	if err := flags.Parse(args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fmt.Fprintln(out, "usage: aiport", args[0], "[--config aiport.yaml]")
			return nil
		}
		return errors.New("invalid arguments; use --config PATH")
	}
	if flags.NArg() != 0 {
		return errors.New("unexpected positional arguments")
	}
	c, err := config.Load(*path)
	if err != nil {
		return err
	}
	if err := ensureLocalCredentials(*path, c, out); err != nil {
		return err
	}
	lookup, err := credentialLookup(*path, c)
	if err != nil {
		return err
	}
	secrets, err := c.Resolve(lookup)
	if err != nil {
		return err
	}
	for _, id := range secrets.MissingKeys {
		fmt.Fprintf(errOut, "aiport: connection %q has no API key yet; add it in the dashboard (Providers). Requests to it return 503 until then.\n", id)
	}
	// check explains each Alias that is a Level now; serve says it once, so a
	// deliberate Level mapping does not repeat a long notice on every start.
	if notices := config.LevelNotices(c); len(notices) > 0 {
		if args[0] == "check" {
			for _, n := range notices {
				fmt.Fprintln(errOut, n)
			}
		} else {
			fmt.Fprintf(errOut, "aiport: %d Alias entries are named fast, balanced or best and act as Levels; run `aiport check` for details.\n", len(notices))
		}
	}
	if args[0] == "check" {
		fmt.Fprintln(out, "Configuration and environment references are valid.")
		return nil
	}
	listener, err := net.Listen("tcp", c.Listen)
	if err != nil {
		return errors.New("cannot bind configured loopback listener")
	}
	collector, closeDB := openTelemetry(c, *path, errOut)
	defer closeDB()
	defer collector.Close(time.Duration(c.Telemetry.ShutdownDrainMS) * time.Millisecond)
	handler := server.NewConfigured(c, secrets, collector, *path).(*server.Handler)
	defer handler.Close()
	srv := &http.Server{Handler: handler, ReadHeaderTimeout: 5 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 16 << 10, ErrorLog: log.New(io.Discard, "", 0)}
	ctx, cancelServe := context.WithCancel(ctx)
	defer cancelServe()
	// Suggestions are detected on a background ticker, never during a request.
	handler.StartAdapting(ctx.Done())
	done := make(chan struct{})
	go func() {
		defer close(done)
		<-ctx.Done()
		deadline, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(deadline); err != nil {
			_ = srv.Close()
		}
	}()
	also := ""
	if a := c.Address(); a != c.Listen {
		also = ", also http://" + a
	}
	terminal.Frost(out, "aiport listening on http://"+c.Listen+also+" (experimental native forwarding)\n")
	printStartupToken(out, c)
	// Agent runs use git by this one absolute path, resolved once.
	if git, err := worktree.Git(); err == nil {
		fmt.Fprintf(out, "Agent runs use git at %s\n", git)
	} else {
		fmt.Fprintln(errOut, "aiport: git is not installed; Explain by a Coding agent and Fix it are unavailable")
	}
	err = srv.Serve(listener)
	if errors.Is(err, http.ErrServerClosed) {
		<-done
		return nil
	}
	return errors.New("HTTP server stopped unexpectedly")
}

// openTelemetry persists request metadata to <config>.db when telemetry is on.
// If the database cannot be used, aiport still serves: a newer schema disables
// recording (never write what this build doesn't understand); any other error
// falls back to session-only history. Either way the reason is printed.
func openTelemetry(c config.Config, configPath string, errOut io.Writer) (*telemetry.Collector, func()) {
	t := c.Telemetry
	if !t.Enabled {
		return telemetry.New(false, t.QueueCapacity), func() {}
	}
	dbPath := store.Path(configPath)
	db, err := store.Open(dbPath)
	if errors.Is(err, store.ErrNewerSchema) {
		fmt.Fprintf(errOut, "aiport: %s: %v; request recording is off\n", dbPath, err)
		return telemetry.New(false, t.QueueCapacity), func() {}
	}
	if err != nil {
		fmt.Fprintf(errOut, "aiport: %v; request history is kept for this session only\n", err)
		return telemetry.New(true, t.QueueCapacity), func() {}
	}
	fmt.Fprintf(errOut, "aiport: request history is saved to %s (kept %d days)\n", dbPath, t.RetentionDays)
	collector := telemetry.NewPersistent(true, t.QueueCapacity, telemetry.Store{DB: db, RetentionDays: t.RetentionDays, BatchSize: t.BatchSize, FlushInterval: time.Duration(t.FlushIntervalMS) * time.Millisecond})
	return collector, func() { _ = db.Close() }
}

func runAgent(ctx context.Context, args []string, out, errOut io.Writer) error {
	if len(args) == 0 {
		return errors.New("usage: aiport agent codex|claude --model MODEL [--config PATH] [-- agent-options]")
	}
	name := args[0]
	fs := flag.NewFlagSet("agent", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	path := fs.String("config", "aiport.yaml", "")
	model := fs.String("model", "", "")
	dry := fs.Bool("dry-run", false, "")
	if err := fs.Parse(args[1:]); err != nil {
		return errors.New("invalid agent arguments")
	}
	c, err := config.Load(*path)
	if err != nil {
		return err
	}
	if err := ensureLocalCredentials(*path, c, errOut); err != nil {
		return err
	}
	launch, err := agents.Prepare(c, name, *model, fs.Args(), os.Environ())
	if err != nil {
		return err
	}
	if *dry {
		fmt.Fprintf(out, "Agent: %s\nGateway: http://%s\nModel: %s\nSettings apply to this launch only; credentials are not displayed.\n", launch.Name, c.Listen, *model)
		return nil
	}
	command := exec.CommandContext(ctx, launch.Name, launch.Args...)
	command.Env = launch.Env
	command.Stdin = os.Stdin
	command.Stdout = out
	command.Stderr = errOut
	if err = command.Run(); err != nil {
		return errors.New("agent launch failed; check installation, gateway and model compatibility")
	}
	return nil
}

// printStartupToken shows the management token needed for first dashboard sign-in
// and account recovery. It is printed only to an interactive terminal so it does
// not end up in redirected logs; otherwise a retrieval hint is printed instead.
func printStartupToken(out io.Writer, c config.Config) {
	if c.Auth.Disabled {
		fmt.Fprintln(out, "Authentication is disabled; no token is needed to open the dashboard.")
		return
	}
	token := os.Getenv(c.Auth.TokenEnv)
	if token == "" {
		return
	}
	if !terminal.Interactive(out) {
		fmt.Fprintln(out, "Management token: run `aiport auth token` in an interactive terminal to display it.")
		return
	}
	terminal.Dim(out, "Dashboard:        ")
	terminal.Frost(out, "http://"+c.Address()+"/\n")
	terminal.Dim(out, "Management token: ")
	terminal.Snow(out, token+"\n")
	terminal.Dim(out, "Paste the token on first sign-in or to recover access. Keep it private.\n")
}
