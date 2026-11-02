import React, {useEffect, useState} from 'react';
import {Bot} from 'lucide-react';
import type {Origin} from './Applications';

/* Coding agents aiport can start for an Application's Agent runs (Claude
   Code, Codex, Gemini CLI), as aiport found them on this machine, and the
   controls that connect one to an Application. aiport never installs
   anything: a missing agent or ACP adapter shows the command to run. */

export type CodingAgent = {id: string; name: string; program?: string; acp?: {path: string; args?: string[]}; adapter?: string; installed: boolean; ready: boolean; install?: string};
export type CodingAgents = {agents: CodingAgent[]; default_agent?: string};

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
  useEffect(() => { setAgent(o.agent || ''); setRepository(o.repository || ''); setError(''); setSaved(false); }, [o.id, o.agent, o.repository]);
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

/** Settings: the coding agent of every Application that has none of its own. */
export function DefaultAgentPanel({tick, onChange}: {tick: number; onChange: () => void}) {
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
    <p className="muted">The coding agent aiport starts to explain or fix an Application's Issues, for every Application that has none of its own (set on the Integrate page). aiport only starts agents it found installed, by absolute path.</p>
    <div className="agent-fields">
      <label htmlFor="default-agent">Default agent</label>
      <AgentSelect id="default-agent" value={agents.default_agent || ''} agents={agents.agents} emptyLabel="None" onChange={choose} disabled={busy}/>
    </div>
    <InstallHint agent={current}/>
    {error && <p className="auth-error" role="alert">{error}</p>}
  </section>;
}
