import React, {useEffect, useState} from 'react';
import {ArrowRight, ChevronRight, Route as RouteIcon} from 'lucide-react';
import {HelpLink} from './help';
import {ProviderLogo} from './ProviderLogo';
import {catalogChangeText, type CatalogRecord, type TaskName} from './Changes';
import type {ProviderSpec} from './Settings';

/** What the map needs of a Connection: its provider, for the logo and name. */
type Target = {id: string; provider?: string; location?: string};

/** One row of the Routing map (server.Route): a name an Application can send on a protocol and what serves it.
 *  kind "default" is any other model name, sent unchanged to that Connection (name and model empty). Whatever
 *  else the Catalog gives a task name (such as per-effort models) arrives on the row as it is. */
export type Route = TaskName & {kind: 'task_name' | 'level' | 'default'; protocol: string; [more: string]: unknown};
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

/** Effort-specific models a row may carry (added by per-effort mappings), shown as "low → model". */
function effortOf(r: Route): [string, string][] {
  const e = r.effort;
  if (!e || typeof e !== 'object') return [];
  return Object.entries(e as Record<string, unknown>).filter(([, v]) => typeof v === 'string' && v || v && typeof v === 'object' && 'model' in (v as object))
    .map(([k, v]) => [k, typeof v === 'string' ? v : String((v as {model: unknown}).model)]);
}

function Served({r, targets, providers}: {r: Route; targets: Target[]; providers: ProviderSpec[]}) {
  const t = targets.find(x => x.id === r.connection);
  const provider = t?.provider ? providers.find(p => p.id === t.provider)?.name || t.provider : '';
  return <span className="route-served">
    {t?.provider && <ProviderLogo provider={t.provider} location={t.location} size={16}/>}
    <code>{r.connection}</code>{provider && <small className="muted">{provider}</small>}
  </span>;
}

/** The rows of one Application: task names, Levels and where any other name goes, each with its Connection and model. */
export function RouteTable({entry, targets, providers, provider}: {entry: RoutingEntry; targets: Target[]; providers: ProviderSpec[]; provider?: string}) {
  const providerOf = (id: string) => targets.find(t => t.id === id)?.provider || '';
  const rows = entry.routes.filter(r => !provider || providerOf(r.connection) === provider);
  if (!rows.length) return <p className="muted route-none">{entry.routes.length ? 'Nothing on this provider.' : 'No name reaches a Connection: no binding, no protocol default and no task name.'}</p>;
  const listed = (r: Route) => entry.models.filter(m => m.connection === r.connection).length;
  return <table className="route-table" aria-label={'Names ' + entry.name + ' can send'}>
    <thead><tr><th>Name sent</th><th>Kind</th><th>Protocol</th><th>Connection</th><th>Model sent upstream</th></tr></thead>
    <tbody>{rows.map((r, i) => {
      const effort = effortOf(r);
      return <tr key={i} className={'route-' + r.kind}>
        <td>{r.kind === 'default' ? <span className="muted">Any other model name</span> : <code>{r.name}</code>}</td>
        <td>{kindName(r)}</td>
        <td>{protocolName(r.protocol)}</td>
        <td><Served r={r} targets={targets} providers={providers}/></td>
        <td>{r.kind === 'default'
          ? <span className="muted">As sent{listed(r) ? ' (' + listed(r) + ' listed)' : ''}</span>
          : <><ArrowRight size={13} aria-hidden className="muted"/> <code>{r.model}</code>
            {effort.length > 0 && <small className="route-effort">{effort.map(([k, m]) => k + ': ' + m).join(' · ')}</small>}</>}</td>
      </tr>;
    })}</tbody>
  </table>;
}

/** Routing page: every Application, each name it can send and the Connection and model serving it. */
export function RoutingMapPanel({targets, providers, reload}: {targets: Target[]; providers: ProviderSpec[]; reload: unknown}) {
  const {entries, error} = useRoutingMap(null, reload);
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
        <RouteTable entry={e} targets={targets} providers={providers} provider={provider}/>
      </div>)}
    </>}
  </section>;
}

/** Integrate: one Application's Catalog (the same entry the Routing map shows) and its recent Catalog changes. */
export function AppCatalog({application, targets, providers, reload}: {application: string; targets: Target[]; providers: ProviderSpec[]; reload: unknown}) {
  const {entries, error} = useRoutingMap(application, reload);
  const [all, setAll] = useState(false);
  if (error) return <p className="auth-error" role="alert">{error}</p>;
  if (!entries) return <p className="muted">Loading…</p>;
  const e = entries[0];
  if (!e) return null;
  const changes = e.changes || [], few = all ? changes : changes.slice(0, 5);
  return <div className="app-catalog">
    <h4>Catalog <code>{e.version}</code><HelpLink section="The Catalog"/></h4>
    <RouteTable entry={e} targets={targets} providers={providers}/>
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
