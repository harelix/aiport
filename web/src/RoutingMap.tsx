import React, {FormEvent, useEffect, useState} from 'react';
import {ArrowRight, ChevronRight, Pencil, Route as RouteIcon, X} from 'lucide-react';
import {HelpLink} from './help';
import {ProviderLogo} from './ProviderLogo';
import {catalogChangeText, type CatalogRecord, type TaskName} from './Changes';
import type {ProviderSpec} from './Settings';

/** What the map needs of a Connection: its provider, for the logo and name. */
type Target = {id: string; provider?: string; location?: string};

/** One row of the Routing map (server.RoutingRow): a name an Application can send on a protocol and what serves it.
 *  kind "default" is any other model name, sent unchanged to that Connection (name and model empty). Whatever
 *  else the Catalog gives a task name (such as per-effort models) arrives on the row as it is. */
export type Route = TaskName & {kind: 'task_name' | 'level' | 'default'; protocol: string};
/** One Application in the Routing map (server.RoutingEntry): its Catalog, routes and, alone, recent Catalog changes. */
export type RoutingEntry = {
  application: string; name: string; version: string; binding?: string; defaults?: Record<string, string>;
  connections: string[]; models: {id: string; connection: string}[]; routes: Route[]; changes?: CatalogRecord[];
};

/** The Routing map (GET /api/v1/routing), or one Application's entry with its recent Catalog changes. */
export function useRoutingMap(application: string | null, reload: unknown): {entries: RoutingEntry[] | null; error: string} {
  const [entries, setEntries] = useState<RoutingEntry[] | null>(null), [error, setError] = useState('');
  useEffect(() => {
    let live = true;
    const q = application ? '?application=' + encodeURIComponent(application) : '';
    fetch('/api/v1/routing' + q, {credentials: 'same-origin'}).then(r => { if (!r.ok) throw new Error(); return r.json(); })
      .then(b => { if (live) { setEntries(b.applications || []); setError(''); } })
      .catch(() => { if (live) setError('Could not load the Routing map.'); });
    return () => { live = false; };
  }, [application, reload]);
  return {entries, error};
}

const protocolName = (p: string) => p === 'anthropic' ? 'Anthropic' : 'OpenAI';
const kindName = (r: Route) => r.kind === 'level' ? 'Level' : r.kind === 'default' ? 'Protocol default' : r.shared ? 'Shared task name' : 'Own task name';

const EFFORTS = ['low', 'medium', 'high'] as const;
type Efforts = {low?: string; medium?: string; high?: string};

/** The per-effort models a row carries (config.EffortModels: low, medium, high on the same Connection), in effort order. */
function effortOf(r: Route): [string, string][] {
  const e = r.effort || {};
  return EFFORTS.filter(k => e[k]).map(k => [k, e[k]!]);
}

/* The per-effort models of one task name (PUT /api/v1/aliases/effort): a
   model for low, medium and high on the task name's own Connection; an
   empty one uses the task name's model. Each save is one revertable Change.
   A shared task name's models apply to every Application. */
function EffortEditor({application, r, onDone, onCancel}: {application: string; r: Route; onDone: (message: string) => void; onCancel: () => void}) {
  const [effort, setEffort] = useState<Efforts>(r.effort || {});
  const [busy, setBusy] = useState(false), [error, setError] = useState('');
  async function put(value: Efforts) {
    setBusy(true); setError('');
    try {
      const body = {application: r.shared ? '' : application, alias: r.name, connection: r.connection,
        effort: Object.fromEntries(EFFORTS.map(k => [k, (value[k] || '').trim()]).filter(([, v]) => v))};
      const response = await fetch('/api/v1/aliases/effort', {method: 'PUT', credentials: 'same-origin', headers: {'Content-Type': 'application/json', 'X-Aiport-CSRF': '1'}, body: JSON.stringify(body)});
      const data = await response.json().catch(() => ({}));
      if (!response.ok) throw new Error(data.error?.message || 'The change was not saved (' + response.status + ').');
      onDone((Object.keys(body.effort).length ? 'Saved the per-effort models of ' : 'Cleared the per-effort models of ') + r.name + ' as a Change. Revert it in Changes.');
    } catch (err) { setError((err as Error).message); } finally { setBusy(false); }
  }
  const save = (e: FormEvent) => { e.preventDefault(); void put(effort); };
  return <form className="conn-form effort-editor" onSubmit={save} aria-label={'Per-effort models of ' + r.name}>
    <p className="muted">{r.shared ? 'Shared task name: these apply to every Application that sends ' : 'Own task name of ' + application + ': '}<code>{r.name}</code>. Models of {r.connection}; an empty one uses <code>{r.model}</code>.</p>
    <div className="level-efforts">{EFFORTS.map(k => <label key={k}>{k[0].toUpperCase() + k.slice(1)}<input value={effort[k] || ''} onChange={x => setEffort(v => ({...v, [k]: x.target.value}))} maxLength={256} placeholder={r.model} spellCheck={false} autoComplete="off"/></label>)}</div>
    {error && <p className="auth-error" role="alert">{error}</p>}
    <div className="form-actions">
      <button className="primary" disabled={busy}>{busy ? 'Saving…' : 'Save'}</button>
      {effortOf(r).length > 0 && <button type="button" className="secondary" disabled={busy} onClick={() => void put({})}>Clear</button>}
      <button type="button" className="text-button" onClick={onCancel}>Cancel</button>
    </div>
  </form>;
}

function Served({r, targets, providers}: {r: Route; targets: Target[]; providers: ProviderSpec[]}) {
  const t = targets.find(x => x.id === r.connection);
  const provider = t?.provider ? providers.find(p => p.id === t.provider)?.name || t.provider : '';
  return <span className="route-served">
    {t?.provider && <ProviderLogo provider={t.provider} location={t.location} size={16}/>}
    <code>{r.connection}</code>{provider && <small className="muted">{provider}</small>}
  </span>;
}

/** The rows of one Application: task names, Levels and where any other name goes, each with its Connection and model.
 *  With onChange, a task name's per-effort models can be edited on its row (Levels have their own panel). */
export function RouteTable({entry, targets, providers, provider, onChange}: {entry: RoutingEntry; targets: Target[]; providers: ProviderSpec[]; provider?: string; onChange?: () => void}) {
  const [editing, setEditing] = useState(-1), [message, setMessage] = useState('');
  const providerOf = (id: string) => targets.find(t => t.id === id)?.provider || '';
  const rows = entry.routes.filter(r => !provider || providerOf(r.connection) === provider);
  if (!rows.length) return <p className="muted route-none">{entry.routes.length ? 'Nothing on this provider.' : 'No name reaches a Connection: no binding, no protocol default and no task name.'}</p>;
  const listed = (r: Route) => entry.models.filter(m => m.connection === r.connection).length;
  return <>
  <table className="route-table" aria-label={'Names ' + entry.name + ' can send'}>
    <thead><tr><th>Name sent</th><th>Kind</th><th>Protocol</th><th>Connection</th><th>Model sent upstream</th>{onChange && <th><span className="sr-only">Edit</span></th>}</tr></thead>
    <tbody>{rows.map((r, i) => {
      const effort = effortOf(r), on = editing === i;
      return <React.Fragment key={i}><tr className={'route-' + r.kind}>
        <td>{r.kind === 'default' ? <span className="muted">Any other model name</span> : <code>{r.name}</code>}</td>
        <td>{kindName(r)}</td>
        <td>{protocolName(r.protocol)}</td>
        <td><Served r={r} targets={targets} providers={providers}/></td>
        <td>{r.kind === 'default'
          ? <span className="muted">As sent{listed(r) ? ' (' + listed(r) + ' listed)' : ''}</span>
          : <><ArrowRight size={13} aria-hidden className="muted"/> <code>{r.model}</code>
            {effort.length > 0 && <small className="route-effort">{effort.map(([k, m]) => k + ': ' + m).join(' · ')}</small>}</>}</td>
        {onChange && <td className="route-edit">{r.kind === 'task_name' && <button type="button" className="icon-button" aria-pressed={on}
          onClick={() => { setMessage(''); setEditing(on ? -1 : i); }}
          title={on ? 'Close' : 'Per-effort models of ' + r.name} aria-label={(on ? 'Close the per-effort models of ' : 'Edit the per-effort models of ') + r.name}>
          {on ? <X size={14} aria-hidden/> : <Pencil size={14} aria-hidden/>}</button>}</td>}
      </tr>
      {on && onChange && <tr className="route-editing"><td colSpan={6}>
        <EffortEditor application={entry.application} r={r} onCancel={() => setEditing(-1)} onDone={text => { setEditing(-1); setMessage(text); onChange(); }}/>
      </td></tr>}</React.Fragment>;
    })}</tbody>
  </table>
  {message && <p className="auth-success" role="status">{message}</p>}
  </>;
}

/** Routing page: every Application, each name it can send and the Connection and model serving it. */
export function RoutingMapPanel({targets, providers, reload, onChange}: {targets: Target[]; providers: ProviderSpec[]; reload: unknown; onChange: () => void}) {
  const [n, setN] = useState(0);
  const {entries, error} = useRoutingMap(null, String(reload) + n);
  const changed = () => { setN(x => x + 1); onChange(); };
  const [app, setApp] = useState(''), [provider, setProvider] = useState('');
  const used = Array.from(new Set((entries || []).flatMap(e => e.routes.map(r => targets.find(t => t.id === r.connection)?.provider || '')).filter(Boolean)));
  const shown = (entries || []).filter(e => !app || e.application === app);
  return <section className="panel routing-map" aria-labelledby="routing-map-heading">
    <div className="panel-heading">
      <h2 id="routing-map-heading"><RouteIcon size={16} aria-hidden/> Routing map<HelpLink section="The Routing map"/></h2>
      <span>Each name an Application can send, and the Connection and model that serve it</span>
    </div>
    {error && <p className="auth-error" role="alert">{error}</p>}
    {!entries && !error && <p className="muted">Loading…</p>}
    {entries && entries.length === 0 && <p className="quiet-empty">No Applications yet. When one is approved, every name it can send is listed here.</p>}
    {entries && entries.length > 0 && <>
      <div className="routing-map-tools">
        <label>Application <select value={app} onChange={e => setApp(e.target.value)}>
          <option value="">All ({entries.length})</option>
          {entries.map(e => <option key={e.application} value={e.application}>{e.name || e.application}</option>)}
        </select></label>
        {used.length > 1 && <div className="task-filters" role="tablist" aria-label="Provider">
          <button type="button" role="tab" aria-selected={!provider} className={!provider ? 'active' : ''} onClick={() => setProvider('')}>All providers</button>
          {used.map(p => <button key={p} type="button" role="tab" aria-selected={provider === p} className={provider === p ? 'active' : ''} onClick={() => setProvider(p)}>
            <ProviderLogo provider={p} size={14}/> {providers.find(x => x.id === p)?.name || p}</button>)}
        </div>}
      </div>
      {shown.map(e => <div key={e.application} className="routing-map-app">
        <h3>{e.name || e.application} <small><code>{e.application}</code></small>
          <span className="muted-small">{e.binding ? 'Bound to ' + e.binding : 'Follows the protocol defaults'} · Catalog {e.version}</span></h3>
        <RouteTable entry={e} targets={targets} providers={providers} provider={provider} onChange={changed}/>
      </div>)}
    </>}
  </section>;
}

/** Integrate: one Application's Catalog (the same entry the Routing map shows) and its recent Catalog changes. */
export function AppCatalog({application, targets, providers, reload, onChange}: {application: string; targets: Target[]; providers: ProviderSpec[]; reload: unknown; onChange?: () => void}) {
  const [n, setN] = useState(0);
  const {entries, error} = useRoutingMap(application, String(reload) + n);
  const [all, setAll] = useState(false);
  if (error) return <p className="auth-error" role="alert">{error}</p>;
  if (!entries) return <p className="muted">Loading…</p>;
  const e = entries[0];
  if (!e) return null;
  const changes = e.changes || [], few = all ? changes : changes.slice(0, 5);
  return <div className="app-catalog">
    <h4>Catalog <code>{e.version}</code><HelpLink section="The Catalog"/></h4>
    <RouteTable entry={e} targets={targets} providers={providers} onChange={() => { setN(x => x + 1); onChange?.(); }}/>
    {e.models.length > 0 && <details className="app-catalog-models"><summary><ChevronRight size={14} aria-hidden className="chevron"/> {e.models.length} models listed by {e.connections.join(', ')}</summary>
      <p>{e.models.map(m => m.id).join(', ')}</p></details>}
    <h4>Recent Catalog changes</h4>
    {changes.length === 0 ? <p className="muted">None kept. A Change that adds, removes or moves a name this Application can send is listed here.</p>
      : <ul className="app-catalog-changes">{few.map(c => <li key={c.change_id + c.at_ms}>
        <span className="muted-small">{new Date(c.at_ms).toLocaleString()} · Change {c.change_id}{c.revert_of ? ' (revert of ' + c.revert_of + ')' : ''}</span>
        {catalogChangeText(c).join('; ') || 'Catalog version changed'}</li>)}</ul>}
    {changes.length > 5 && <button type="button" className="text-button" onClick={() => setAll(x => !x)}>{all ? 'Show fewer' : 'Show all ' + changes.length}</button>}
  </div>;
}
