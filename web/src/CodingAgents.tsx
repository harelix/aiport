import React, {useEffect, useState} from 'react';
import {Bot} from 'lucide-react';
import type {Origin} from './Applications';

/* Coding agents aiport can start for an Application's Agent runs (Claude
   Code, Codex, Gemini CLI), as aiport found them on this machine, and the
   controls that connect one to an Application. aiport never installs
   anything: a missing agent or ACP adapter shows the command to run. */

export type CodingAgent = {id: string; name: string; program?: string; acp?: {path: string; args?: string[]}; adapter?: string; installed: boolean; ready: boolean; install?: string};
export type CodingAgents = {agents: CodingAgent[]; default_agent?: string; agent_runs?: {connection?: string; model?: string; explain_minutes?: number}};

async function send(path: string, body: object, method: string) {
  const r = await fetch('/api/v1/' + path, {method, credentials: 'same-origin', headers: {'Content-Type': 'application/json', 'X-Aiport-CSRF': '1'}, body: JSON.stringify(body)});
  const data = await r.json().catch(() => ({}));
  if (!r.ok) throw new Error(data.error?.message || 'Request failed (' + r.status + ').');
  return data;
}

/** GET /api/v1/agents, refetched when tick changes. */
export function useCodingAgents(tick: number): [CodingAgents, () => void] {
  const [list, setList] = useState<CodingAgents>({agents: []}), [own, setOwn] = useState(0);
  useEffect(() => {
    const controller = new AbortController();
    fetch('/api/v1/agents', {credentials: 'same-origin', signal: controller.signal}).then(r => r.ok ? r.json() : {agents: []}).then(setList).catch(() => {});
    return () => controller.abort();
  }, [tick, own]);
  return [list, () => setOwn(x => x + 1)];
}

/** The agent an Application uses: its own, otherwise the default. */
export const effectiveAgent = (o: Origin | undefined, a: CodingAgents) => o?.agent || a.default_agent || '';
export const agentName = (id: string, a: CodingAgents) => a.agents.find(x => x.id === id)?.name || id;
/** Whether an Application has a coding agent and a repository path. */
export const agentConnected = (o: Origin | undefined, a: CodingAgents) => !!(o?.repository && effectiveAgent(o, a));

/** What is missing for an agent, with the command to install it. */
export function InstallHint({agent}: {agent?: CodingAgent}) {
  if (!agent || agent.ready) return null;
  return <p className="install-hint" role="note">
    {agent.installed ? <>{agent.name} is installed, but aiport talks to it through its ACP adapter, which is not. Install it yourself with:</> : <>{agent.name} is not installed on this machine. Install it with:</>}
    <code className="install-command">{agent.install}</code>
    <small className="muted">aiport never downloads or installs anything. Refresh this page afterwards.</small>
  </p>;
}

/** Choose an agent: installed ones, plus the default or none. */
function AgentSelect({id, value, agents, emptyLabel, onChange, disabled}: {id: string; value: string; agents: CodingAgent[]; emptyLabel: string; onChange: (v: string) => void; disabled?: boolean}) {
  return <select id={id} value={value} disabled={disabled} onChange={e => onChange(e.target.value)}>
    <option value="">{emptyLabel}</option>
    {agents.map(a => <option key={a.id} value={a.id} disabled={!a.installed && a.id !== value}>{a.name}{a.installed ? a.ready ? '' : ' (adapter missing)' : ' (not installed)'}</option>)}
  </select>;
}

/* "Connect a coding agent": the Application's agent and repository path,
   saved together as one revertable Change. */
export function AgentSettings({o, agents, onDone}: {o: Origin; agents: CodingAgents; onDone: () => void}) {
  const [agent, setAgent] = useState(o.agent || ''), [repository, setRepository] = useState(o.repository || '');
  const [busy, setBusy] = useState(false), [error, setError] = useState(''), [saved, setSaved] = useState(false);
  useEffect(() => { setAgent(o.agent || ''); setRepository(o.repository || ''); }, [o.id, o.agent, o.repository]);
  useEffect(() => { setError(''); setSaved(false); }, [o.id]);
  const fallback = agents.default_agent ? 'Default (' + agentName(agents.default_agent, agents) + ')' : 'None (no default set)';
  const chosen = agents.agents.find(a => a.id === (agent || agents.default_agent));
  const changed = agent !== (o.agent || '') || repository.trim() !== (o.repository || '');
  async function save() {
    setBusy(true); setError(''); setSaved(false);
    try { await send('origins/' + encodeURIComponent(o.id), {agent, repository: repository.trim()}, 'PATCH'); setSaved(true); onDone(); }
    catch (e) { setError((e as Error).message); } finally { setBusy(false); }
  }
  return <div className="agent-settings">
    <div className="agent-fields">
      <label htmlFor={'agent-' + o.id}>Coding agent</label>
      <AgentSelect id={'agent-' + o.id} value={agent} agents={agents.agents} emptyLabel={fallback} onChange={setAgent} disabled={busy}/>
      <label htmlFor={'repo-' + o.id}>Repository</label>
      <input id={'repo-' + o.id} value={repository} disabled={busy} spellCheck={false} placeholder="/home/you/src/docs-bot" onChange={e => setRepository(e.target.value)} aria-describedby={'repo-help-' + o.id}/>
    </div>
    <small id={'repo-help-' + o.id} className="muted">The top directory of the application's git repository (where <code>.git</code> is). Agent runs work only there.</small>
    <InstallHint agent={chosen}/>
    <div className="button-row">
      <button type="button" className="secondary" disabled={busy || !changed} onClick={save}>{busy ? 'Saving…' : 'Save'}</button>
      {saved && !changed && <span className="muted-small" role="status">Saved as a Change you can revert.</span>}
    </div>
    {error && <p className="auth-error" role="alert">{error}</p>}
  </div>;
}

/** A Connection, as Settings lists them. */
type RunTarget = {id: string; provider: string; protocol: string; disabled: boolean};

/**
 * Where Agent runs' model calls go: Claude Code is pointed at aiport (a run-scoped credential, attributed to the
 * built-in Application "aiport" in Activity) and its calls go to an Anthropic-protocol Connection; empty means the
 * Anthropic protocol default. Codex and Gemini CLI use their own login.
 */
function RunCalls({agents, targets, onSaved}: {agents: CodingAgents; targets: RunTarget[]; onSaved: () => void}) {
  const saved = agents.agent_runs || {};
  const [connection, setConnection] = useState(saved.connection || ''), [model, setModel] = useState(saved.model || '');
  const [busy, setBusy] = useState(false), [error, setError] = useState(''), [message, setMessage] = useState('');
  useEffect(() => { setConnection(saved.connection || ''); setModel(saved.model || ''); }, [saved.connection, saved.model]);
  const anthropic = targets.filter(t => t.protocol === 'anthropic' && (!t.disabled || t.id === saved.connection));
  const dirty = connection !== (saved.connection || '') || model.trim() !== (saved.model || '');
  async function save() {
    setBusy(true); setError(''); setMessage('');
    try { await send('agents/calls', {connection, model: model.trim()}, 'PUT'); setMessage('Saved as a Change you can revert.'); onSaved(); }
    catch (e) { setError((e as Error).message); } finally { setBusy(false); }
  }
  return <div className="settings-group">
    <h3 className="subheading">Model calls of Agent runs</h3>
    <div className="form-row">
      <label>Connection
        <select id="run-connection" value={connection} disabled={busy} onChange={e => setConnection(e.target.value)}>
          <option value="">Anthropic protocol default</option>
          {anthropic.map(t => <option key={t.id} value={t.id}>{t.id} ({t.provider})</option>)}
        </select>
      </label>
      <label>Model
        <input id="run-model" value={model} disabled={busy} maxLength={256} spellCheck={false} autoComplete="off" placeholder="the agent's own choice" onChange={e => setModel(e.target.value)}/>
      </label>
    </div>
    <small>Claude Code runs call models through aiport, with a credential that lives only as long as the run, and show in Activity under the Application <code>aiport</code>. Codex and Gemini CLI runs use their own login; aiport never edits an agent's configuration files. Explain runs stop after {saved.explain_minutes || 10} minutes (<code>agent_runs.explain_minutes</code>).</small>
    {!anthropic.length && <small>No Anthropic-protocol Connection is configured, so Claude Code runs use their own login.</small>}
    {error && <p className="auth-error" role="alert">{error}</p>}
    <div className="form-actions">
      <button type="button" className="primary" disabled={busy || !dirty} onClick={save}>{busy ? 'Saving…' : 'Save'}</button>
      {message && !dirty && <span className="muted-small" role="status">{message}</span>}
    </div>
  </div>;
}

/** Settings: the coding agent of every Application that has none of its own, and where runs' model calls go. */
export function DefaultAgentPanel({tick, targets, onChange}: {tick: number; targets: RunTarget[]; onChange: () => void}) {
  const [agents, refresh] = useCodingAgents(tick);
  const [busy, setBusy] = useState(false), [error, setError] = useState('');
  async function choose(agent: string) {
    setBusy(true); setError('');
    try { await send('agents/default', {agent}, 'PUT'); refresh(); onChange(); }
    catch (e) { setError((e as Error).message); } finally { setBusy(false); }
  }
  const current = agents.agents.find(a => a.id === agents.default_agent);
  return <section className="panel" aria-labelledby="default-agent-heading">
    <div className="panel-heading"><h2 id="default-agent-heading"><Bot size={16} aria-hidden/> Default coding agent</h2><span>For Agent runs</span></div>
    <div className="conn-form settings-body">
      <div className="settings-group">
        <div className="form-row">
          <label>Default agent
            <AgentSelect id="default-agent" value={agents.default_agent || ''} agents={agents.agents} emptyLabel="None" onChange={choose} disabled={busy}/>
          </label>
        </div>
        <small>The coding agent aiport starts to explain or fix an Application's Issues, for every Application that has none of its own (set on the Integrate page). aiport only starts agents it found installed, by absolute path. Saved as soon as you choose.</small>
        <InstallHint agent={current}/>
        {error && <p className="auth-error" role="alert">{error}</p>}
      </div>
      <RunCalls agents={agents} targets={targets} onSaved={() => { refresh(); onChange(); }}/>
    </div>
  </section>;
}
