// Package capability answers what aiport can do right now for one
// Application: the protocols it can reach, its own Aliases, its binding,
// whether it may request changes, and which kinds of Change it may request.
//
// The HTTP route (/v1/aiport/capabilities) and the MCP tool
// (aiport_capabilities) both answer with For, so the two doors agree.
package capability

import "aiport/internal/config"

// Kind is a kind of Change an Application may ask for.
type Kind struct {
	Kind        string `json:"kind"`
	Description string `json:"description"`
}

// The built-in Change kinds, the only ones aiport can apply today. Any other
// kind is accepted as not supported and counted as demand.
const (
	ApplicationAlias = "application_alias"
	BindApplication  = "bind_application"
)

// Kinds lists the built-in Change kinds in a stable order.
var Kinds = []Kind{
	{ApplicationAlias, "An Alias only this Application sees: a model name it sends, replaced with a model on a Connection (which may be another Connection than its usual one). For every Application it becomes a shared Alias."},
	{BindApplication, "Send all of this Application's calls to one Connection. For every Application it becomes the protocol's default Connection."},
}

// BuiltIn reports whether kind is one of the built-in Change kinds.
func BuiltIn(kind string) bool {
	for _, k := range Kinds {
		if k.Kind == kind {
			return true
		}
	}
	return false
}

// Protocols are the wire protocols aiport speaks, in a stable order.
var Protocols = []string{"openai", "anthropic"}

// Router says which Connection would serve an Application on a protocol;
// the gateway's routing module implements it.
type Router interface {
	Reachable(application, protocol string) (connection string, ok bool)
}

type Protocol struct {
	Name      string `json:"name"`
	Available bool   `json:"available"`
	// Connection serves calls on this protocol when no Alias applies.
	Connection string `json:"connection,omitempty"`
}

type Alias struct {
	Alias      string `json:"alias"`
	Connection string `json:"connection"`
	Model      string `json:"model"`
}

type Application struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Capabilities is the answer for one Application.
type Capabilities struct {
	Application Application `json:"application"`
	Protocols   []Protocol  `json:"protocols"`
	Aliases     []Alias     `json:"aliases"`
	// Binding is the Connection all this Application's calls go to, or nil
	// when it follows the protocol defaults.
	Binding           *string `json:"binding"`
	MayRequestChanges bool    `json:"may_request_changes"`
	// ChangeKinds are the kinds this Application may request now; empty
	// unless the Owner allowed it to request changes.
	ChangeKinds []Kind `json:"change_kinds"`
	Note        string `json:"note"`
}

// NotPermitted explains an empty ChangeKinds list.
const NotPermitted = "This Application may not request changes. The Owner can allow it on the Integrate page of the aiport dashboard."

// AdaptingOff explains an empty ChangeKinds list while the Owner has turned
// adapting off: no Change requests are accepted from any Application.
const AdaptingOff = "The Owner has turned adapting off in aiport: no Application may request changes now."

const permitted = "This Application may request the listed kinds of Change. The Owner approves or rejects each request."

// For answers for the Application with id application in configuration c,
// with routing from r. ok is false when c has no such Application.
func For(c config.Config, application string, r Router) (caps Capabilities, ok bool) {
	app, ok := c.Application(application)
	if !ok {
		return Capabilities{}, false
	}
	caps = Capabilities{
		Application:       Application{ID: app.ID, Name: app.Name},
		Protocols:         []Protocol{},
		Aliases:           []Alias{},
		MayRequestChanges: app.MayRequestChanges,
		ChangeKinds:       []Kind{},
		Note:              NotPermitted,
	}
	for _, name := range Protocols {
		connection, ok := r.Reachable(app.ID, name)
		caps.Protocols = append(caps.Protocols, Protocol{Name: name, Available: ok, Connection: connection})
	}
	for _, a := range app.Aliases {
		caps.Aliases = append(caps.Aliases, Alias{Alias: a.Alias, Connection: a.Target, Model: a.Model})
	}
	if app.Target != "" {
		binding := app.Target
		caps.Binding = &binding
	}
	if c.Adapting.Disabled {
		caps.MayRequestChanges, caps.Note = false, AdaptingOff
	} else if app.MayRequestChanges {
		caps.ChangeKinds = append(caps.ChangeKinds, Kinds...)
		caps.Note = permitted
	}
	return caps, true
}
