package onboarding

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"

	"aiport/internal/config"
)

// Levels at first run. init proposes, per protocol of the chosen providers,
// a Connection for each Level: a local one for fast when one was chosen,
// hosted ones for balanced and best (whatever was chosen when there is no
// other kind). The presets know no model ids, so the model is always the
// Owner's answer: nothing is mapped without it, and nothing is made up.

// levelProposal is the Connection proposed for one Level on one protocol.
type levelProposal struct {
	level, protocol, connection string
}

// protocolOf is the protocol a preset Connection written by init speaks.
func protocolOf(t config.Target) string {
	if t.Protocol != "" {
		return t.Protocol
	}
	spec, _ := config.LookupProvider(t.Provider)
	return spec.Protocol
}

func isLocal(t config.Target) bool {
	spec, _ := config.LookupProvider(t.Provider)
	return spec.Location == config.Local
}

// proposeLevels proposes a Connection for every Level on every protocol
// some chosen Connection speaks, in Level and protocol order.
func proposeLevels(targets []config.Target) []levelProposal {
	var out []levelProposal
	for _, protocol := range config.Protocols {
		local, hosted := "", ""
		for _, t := range targets {
			if protocolOf(t) != protocol {
				continue
			}
			if isLocal(t) && local == "" {
				local = t.ID
			}
			if !isLocal(t) && hosted == "" {
				hosted = t.ID
			}
		}
		if local == "" && hosted == "" {
			continue
		}
		for _, level := range config.Levels {
			first, second := hosted, local
			if level == "fast" {
				first, second = local, hosted
			}
			if first == "" {
				first = second
			}
			out = append(out, levelProposal{level, protocol, first})
		}
	}
	return out
}

// askLevels shows the proposal, asks to confirm it and asks for each
// Level's model; an empty answer leaves that Level unmapped. Input that
// ends here maps nothing (init ran with flags only).
func askLevels(targets []config.Target, scanner *bufio.Scanner, out io.Writer) ([]config.Alias, error) {
	proposal := proposeLevels(targets)
	if len(proposal) == 0 {
		return nil, nil
	}
	fmt.Fprintln(out, "\n  Levels: fast, balanced and best are model names every aiport shares; Applications send one\n  instead of naming a provider. Proposed:")
	for _, protocol := range config.Protocols {
		var parts []string
		for _, p := range proposal {
			if p.protocol == protocol {
				parts = append(parts, p.level+" → "+p.connection)
			}
		}
		if len(parts) > 0 {
			fmt.Fprintf(out, "    %s protocol: %s\n", config.ProtocolName(protocol), strings.Join(parts, " · "))
		}
	}
	fmt.Fprint(out, "  Map these Levels now? (y/n) [y]: ")
	if !scanner.Scan() {
		fmt.Fprintln(out)
		return nil, nil
	}
	if v := strings.ToLower(strings.TrimSpace(scanner.Text())); v != "" && v != "y" && v != "yes" {
		return nil, nil
	}
	var aliases []config.Alias
	for _, p := range proposal {
		fmt.Fprintf(out, "  %s on %s (%s): model id, empty leaves it unmapped: ", p.level, p.connection, config.ProtocolName(p.protocol))
		if !scanner.Scan() {
			fmt.Fprintln(out)
			break
		}
		model := strings.TrimSpace(scanner.Text())
		if model == "" {
			continue
		}
		if !validModel(model) {
			return nil, errors.New("a model id must be 1–256 characters without control characters")
		}
		aliases = append(aliases, config.Alias{Alias: p.level, Target: p.connection, Model: model})
	}
	return aliases, nil
}

// parseLevels reads --levels: comma-separated LEVEL=CONNECTION/MODEL
// entries (the model may itself hold slashes); the protocol is the
// Connection's. Each Level is mapped at most once per protocol.
func parseLevels(s string, targets []config.Target) ([]config.Alias, error) {
	byID := map[string]config.Target{}
	for _, t := range targets {
		byID[t.ID] = t
	}
	seen := map[[2]string]bool{}
	var aliases []config.Alias
	for _, entry := range strings.Split(s, ",") {
		level, rest, ok := strings.Cut(strings.TrimSpace(entry), "=")
		connection, model, ok2 := strings.Cut(rest, "/")
		if !ok || !ok2 {
			return nil, errors.New("--levels entries are LEVEL=CONNECTION/MODEL, such as fast=ollama/llama3:8b")
		}
		if !config.IsLevel(level) {
			return nil, fmt.Errorf("%q is not a Level (fast, balanced, best)", level)
		}
		t, known := byID[connection]
		if !known {
			return nil, fmt.Errorf("--levels: %q is not one of the chosen providers", connection)
		}
		if !validModel(model) {
			return nil, errors.New("--levels: a model id must be 1–256 characters without control characters")
		}
		key := [2]string{level, protocolOf(t)}
		if seen[key] {
			return nil, fmt.Errorf("--levels: %s is mapped twice on the %s protocol", level, config.ProtocolName(key[1]))
		}
		seen[key] = true
		aliases = append(aliases, config.Alias{Alias: level, Target: connection, Model: model})
	}
	return aliases, nil
}

func validModel(m string) bool {
	return len(m) >= 1 && len(m) <= 256 && !strings.ContainsFunc(m, func(r rune) bool { return r < 0x20 || r == 0x7f })
}
