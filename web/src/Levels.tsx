import React, {FormEvent, useEffect, useState} from 'react';
import {Layers, Pencil, Plus, X} from 'lucide-react';
import {HelpLink} from './help';
import {ProviderLogo} from './ProviderLogo';
import {ModelPicker, useConnectionModels} from './ModelPicker';

/* Levels in the dashboard (harelix/aiport#38). The Levels panel on Routing
   edits the shared mapping of each Level per protocol; AppLevels, on an
   Application's row on Integrate, edits that Application's own mapping,
   which comes before the shared one. Every save is one revertable Change
   (PUT/DELETE /api/v1/levels/{protocol}/{level}). */

type EffortModels = {low?: string; medium?: string; high?: string};
export type LevelMapping = {level: string; protocol: string; connection: string; model: string; effort: EffortModels};
type Connection = {id: string; provider: string; protocol: string; location?: string; disabled: boolean};
export type LevelsDoc = {levels: string[]; protocols: string[]; efforts: string[]; shared: LevelMapping[]; applications: Record<string, LevelMapping[]>; connections: Connection[]};

const PROTOCOL: Record<string, string> = {openai: 'OpenAI', anthropic: 'Anthropic'};
/** What each Level is for, in a few words, next to its name. */
const LEVEL_HINT: Record<string, string> = {fast: 'Quick, low-cost calls', balanced: 'Everyday work', best: 'The hardest work'};
/** Protocol default Connections (defaults.openai_target / anthropic_target), so the editor starts on the one calls already reach. */
export type ProtocolDefaults = {openai_target?: string; anthropic_target?: string};
const EFFORTS: (keyof EffortModels)[] = ['low', 'medium', 'high'];

/** GET /api/v1/levels, again whenever refresh changes and after reload. */
export function useLevels(refresh: unknown): [LevelsDoc | null, () => void] {
  const [doc, setDoc] = useState<LevelsDoc | null>(null), [n, setN] = useState(0);
  useEffect(() => {
    const controller = new AbortController();
    fetch('/api/v1/levels', {credentials: 'same-origin', signal: controller.signal}).then(r => r.ok ? r.json() : null).then(d => d && setDoc(d)).catch(() => {});
    return () => controller.abort();
  }, [refresh, n]);
  return [doc, () => setN(x => x + 1)];
}

async function send(method: string, path: string, body: object) {
  const response = await fetch('/api/v1/' + path, {method, credentials: 'same-origin', headers: {'Content-Type': 'application/json', 'X-Aiport-CSRF': '1'}, body: JSON.stringify(body)});
  const data = await response.json().catch(() => ({}));
  if (!response.ok) throw new Error(data.error?.message || 'The change was not saved (' + response.status + ').');
  return data;
}

const find = (list: LevelMapping[] | undefined, level: string, protocol: string) => list?.find(m => m.level === level && m.protocol === protocol);

/** One mapping, short: the provider's mark, the model and the Connection, and its per-effort models. */
function Mapping({m, doc}: {m: LevelMapping; doc: LevelsDoc}) {
  const c = doc.connections.find(x => x.id === m.connection);
  const efforts = EFFORTS.filter(e => m.effort?.[e]);
  return <span className="level-mapping">
    <span className="level-model">{c && <ProviderLogo provider={c.provider} location={c.location} size={16}/>}<code>{m.model}</code></span>
    <small className="muted">on {m.connection}</small>
    {efforts.length > 0 && <span className="level-efforts-chips">{efforts.map(e => <span key={e} className="effort-chip" title={'Calls at ' + e + ' effort get ' + m.effort[e]}>{e}: <code>{m.effort[e]}</code></span>)}</span>}
  </span>;
}

/* The form for one Level on one protocol: a Connection of that protocol, a
   model, and optional per-effort models. Clear removes the mapping (for an
   Application: back to the shared one). */
function LevelEditor({doc, level, protocol, application, current, defaults, onDone, onCancel}: {
  doc: LevelsDoc; level: string; protocol: string; application: string; current?: LevelMapping; defaults?: ProtocolDefaults; onDone: (message: string) => void; onCancel: () => void;
}) {
  const connections = doc.connections.filter(c => c.protocol === protocol && (!c.disabled || c.id === current?.connection));
  const preferred = protocol === 'anthropic' ? defaults?.anthropic_target : defaults?.openai_target;
  const [connection, setConnection] = useState(current?.connection || connections.find(c => c.id === preferred)?.id || connections[0]?.id || '');
  const listed = useConnectionModels(connection);
  const [model, setModel] = useState(current?.model || '');
  const [effort, setEffort] = useState<EffortModels>(current?.effort || {});
  const [busy, setBusy] = useState(false), [error, setError] = useState('');
  const chosen = connections.find(c => c.id === connection);
  const name = level + ' on ' + PROTOCOL[protocol];
  async function save(e: FormEvent) {
    e.preventDefault(); setBusy(true); setError('');
    try {
      await send('PUT', 'levels/' + protocol + '/' + level, {application, connection, model: model.trim(), effort});
      onDone('Saved ' + name + ' as a Change. Revert it in Changes.');
    } catch (err) { setError((err as Error).message); } finally { setBusy(false); }
  }
  async function clear() {
    setBusy(true); setError('');
    try {
      await send('DELETE', 'levels/' + protocol + '/' + level, {application});
      onDone(application ? 'Cleared ' + application + '\'s own ' + name + ': it uses the shared mapping. Revert it in Changes.' : 'Cleared the shared ' + name + '. Revert it in Changes.');
    } catch (err) { setError((err as Error).message); } finally { setBusy(false); }
  }
  return <form className="conn-form level-editor" onSubmit={save} aria-label={'Map ' + name}>
    <h3>{application ? application + '\'s own ' : 'Shared '}<code>{level}</code> on {PROTOCOL[protocol]}</h3>
    {connections.length === 0 ? <p className="muted">No {PROTOCOL[protocol]} Connection yet. Add one on Providers first.</p> : <>
      <div className="form-row">
        <label>Connection<select value={connection} onChange={e => { setConnection(e.target.value); setModel(''); setEffort({}); }}>
          {connections.map(c => <option key={c.id} value={c.id}>{c.id} · {c.provider}{c.id === preferred ? ' (protocol default)' : ''}</option>)}
        </select></label>
        <ModelPicker key={'model-' + connection} id={'level-model-' + level + protocol} label="Model" value={model} onChange={setModel} models={listed.models} state={listed.state} required/>
      </div>
      {listed.state === 'unlisted' && <small className="muted">{chosen?.id} does not list its models; type the id the provider expects.</small>}
      <fieldset>
        <legend>Per-effort models <span className="muted">(optional)</span></legend>
        <div className="level-efforts">{EFFORTS.map(e => <ModelPicker key={e + connection} id={'level-' + e + level + protocol} label={e[0].toUpperCase() + e.slice(1) + ' effort'} value={effort[e] || ''}
          onChange={v => setEffort(x => ({...x, [e]: v}))} models={listed.models} state={listed.state} empty={'Same as the model' + (model.trim() ? ' (' + model.trim() + ')' : '')}/>)}</div>
        <small className="muted">A call that says how hard it is (<code>X-Aiport-Effort</code>, or its provider's own effort field) gets the model of its effort.</small>
      </fieldset>
    </>}
    {error && <p className="auth-error" role="alert">{error}</p>}
    <div className="form-actions">
      {connections.length > 0 && <button className="primary" disabled={busy || !connection || !model.trim()}>{busy ? 'Saving…' : 'Save'}</button>}
      {current && <button type="button" className="secondary" disabled={busy} onClick={clear}>{application ? 'Use the shared mapping' : 'Clear'}</button>}
      <button type="button" className="text-button" onClick={onCancel}>Cancel</button>
    </div>
  </form>;
}

/* The Level × protocol grid, shared (application "") or for one
   Application, with the editor for the cell being edited under it. */
function LevelsGrid({doc, application, defaults, onChange}: {doc: LevelsDoc; application: string; defaults?: ProtocolDefaults; onChange: () => void}) {
  const [editing, setEditing] = useState<{level: string; protocol: string} | null>(null), [message, setMessage] = useState('');
  const own = application ? doc.applications[application] || [] : doc.shared;
  const overrides = (level: string, protocol: string) => Object.values(doc.applications).filter(l => find(l, level, protocol)).length;
  return <>
    <div className="levels-table"><table aria-label={application ? 'Levels of ' + application : 'Shared Levels'}>
      <thead><tr><th>Level</th>{doc.protocols.map(p => <th key={p}>{PROTOCOL[p] || p}</th>)}</tr></thead>
      <tbody>{doc.levels.map(level => <tr key={level}>
        <th scope="row"><span className="level-name"><code>{level}</code><small className="muted">{LEVEL_HINT[level] || ''}</small></span></th>
        {doc.protocols.map(protocol => {
          const m = find(own, level, protocol), shared = application ? find(doc.shared, level, protocol) : undefined;
          const on = editing?.level === level && editing.protocol === protocol, n = application ? 0 : overrides(level, protocol);
          return <td key={protocol} className={on ? 'editing' : undefined}>
            <div className="level-cell">
              {m ? <>{application && <span className="badge info">Own</span>}<Mapping m={m} doc={doc}/></>
                : shared ? <span className="level-shared"><small className="muted">Shared</small><Mapping m={shared} doc={doc}/></span>
                : <span className="level-unmapped"><span className="badge warn">Not mapped</span><small className="muted">Calls sending <code>{level}</code> are refused</small></span>}
              <button type="button" className={'row-action' + (!m && !shared && !on ? ' level-map' : '')} aria-pressed={on} onClick={() => { setMessage(''); setEditing(on ? null : {level, protocol}); }}
                title={on ? 'Close' : application && !m ? 'Give ' + application + ' its own ' + level : (m ? 'Edit ' : 'Map ') + level + ' on ' + PROTOCOL[protocol]}
                aria-label={(on ? 'Close the editor of ' : application && !m ? 'Override ' : m ? 'Edit ' : 'Map ') + level + ' on ' + PROTOCOL[protocol]}>
                {on ? <><X size={14} aria-hidden/> Close</> : m ? <><Pencil size={14} aria-hidden/> Edit</> : application && shared ? <><Pencil size={14} aria-hidden/> Override</> : <><Plus size={14} aria-hidden/> Map</>}
              </button>
            </div>
            {n > 0 && <small className="muted">{n === 1 ? '1 Application has its own' : n + ' Applications have their own'}</small>}
          </td>;
        })}
      </tr>)}</tbody>
    </table></div>
    {editing && <LevelEditor key={editing.level + editing.protocol} doc={doc} level={editing.level} protocol={editing.protocol} application={application} defaults={defaults}
      current={find(own, editing.level, editing.protocol)} onCancel={() => setEditing(null)}
      onDone={text => { setEditing(null); setMessage(text); onChange(); }}/>}
    {message && <p className="auth-success" role="status">{message}</p>}
  </>;
}

/** The Levels panel on Routing: the shared mapping of each Level per protocol. */
export function LevelsPanel({refresh, defaults, onChange}: {refresh: unknown; defaults?: ProtocolDefaults; onChange: () => void}) {
  const [doc, reload] = useLevels(refresh);
  return <section className="panel levels-panel">
    <div className="panel-heading"><h2><Layers size={16} aria-hidden/> Levels<HelpLink section="Levels"/></h2><span>What <code>fast</code>, <code>balanced</code> and <code>best</code> mean on this aiport</span></div>
    <p className="muted">Applications send a Level as their model instead of naming a provider. Map each one per protocol; an Application can have its own mapping on Integrate, which comes before this one.</p>
    {doc ? <LevelsGrid doc={doc} application="" defaults={defaults} onChange={() => { reload(); onChange(); }}/> : <p className="quiet-empty">Loading Levels…</p>}
  </section>;
}

/** How many Levels an Application maps itself, for its row on Integrate. */
export function appLevelsLabel(doc: LevelsDoc | null, application: string): string {
  const n = doc?.applications[application]?.length || 0;
  return !doc ? '…' : n === 0 ? 'Shared' : n === 1 ? '1 own' : n + ' own';
}

/** One Application's own Level mappings, under its row on Integrate. */
export function AppLevels({doc, application, onChange}: {doc: LevelsDoc; application: string; onChange: () => void}) {
  return <div className="app-levels">
    <p className="muted">{application}'s own mapping of a Level comes before the shared one, on that protocol only.<HelpLink section="Levels"/></p>
    <LevelsGrid doc={doc} application={application} onChange={onChange}/>
  </div>;
}
