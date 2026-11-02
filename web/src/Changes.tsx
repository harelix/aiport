import React, {useEffect, useState} from 'react';
import {ArrowUpRight, History, Lightbulb, Puzzle, Undo2} from 'lucide-react';
import {type Upstream, UpstreamAction} from './Upstream';

/** One applied Change to aiport's configuration, as GET /api/v1/changes lists it. */
type Change = {
  id: string; at_ms: number; what: string; by: string;
  revert_of?: string; reverted_by?: string; reverted_at_ms?: number; reverted: boolean;
};

/** A Suggestion or Change request waiting for the Owner, as GET /api/v1/decisions lists it. */
export type Decision = {
  id: string; source: string; kind: string; application: string; subject?: string; need: string; scope: string;
  details?: string; why?: string; scope_case?: string; filed_by?: string; repeats?: number;
  asked_by?: Asker[];
  evidence?: {count: number; from_ms: number; to_ms: number; window_minutes: number};
  proposal: {alias?: string; target?: string; model?: string};
  /** Informational Suggestions have nothing to apply: only Dismiss. */
  informational?: boolean;
  /** The panel where the Owner acts on it ('applications': the Applications panel, where Pairing is). */
  opens?: string;
};

/** One Application that asked for a Change for every Application, with its own argument. */
type Asker = {application: string; why: string; scope_case: string; filed_by?: string; repeats?: number; asked_at_ms: number};

/** Demand for a kind of Change aiport cannot apply yet: anonymous counts only. */
export type Demand = {kind: string; requests: number; applications: number; last_at_ms: number};

const times = (n: number) => n + ' ' + (n === 1 ? 'time' : 'times');

const scopeLabel = (s: string, app: string) => s === 'every_application' ? 'every Application' : 'only ' + app;

async function send(method: string, path: string, body?: unknown) {
  const r = await fetch('/api/v1/' + path, {method, credentials: 'same-origin', headers: {'Content-Type': 'application/json', 'X-Aiport-CSRF': '1'}, body: body === undefined ? undefined : JSON.stringify(body)});
  const d = await r.json().catch(() => ({}));
  if (!r.ok) throw Object.assign(new Error(d.error?.message || 'Request failed (' + r.status + ').'), {code: d.error?.code, extension: d.error?.extension, reason: d.error?.reason});
  return d;
}

/** A guardrail Extension's block of Approve: which Extension and why. */
type Blocked = {extension: string; reason: string};

const when = (ms: number) => new Date(ms).toLocaleString();

/**
 * One pending Suggestion or Change request: the evidence or the request's own text, the Change it would make,
 * the Scope to grant, and Approve / Dismiss (Suggestions) or Approve / Reject (Change requests, whose note goes
 * back to the requester). Request text comes from Applications: it is shown as plain text only, never as markup.
 */
function Pending({item, targets, onDecided, onOpenApplications}: {item: Decision; targets: string[]; onDecided: (message: string) => void; onOpenApplications?: () => void}) {
  const request = item.source === 'request', alias = item.kind === 'application_alias';
  // A failing-Connection Suggestion moves the Application away from its subject: never offer it back.
  const failing = !request && item.kind === 'bind_application' ? item.subject : undefined;
  const choices = targets.filter(t => t !== failing);
  const [target, setTarget] = useState(item.proposal.target || choices[0] || ''), [model, setModel] = useState(item.proposal.model || '');
  const [name, setName] = useState(item.proposal.alias || ''), [scope, setScope] = useState(item.scope || 'application');
  const [note, setNote] = useState(''), [busy, setBusy] = useState(''), [error, setError] = useState('');
  const [blocked, setBlocked] = useState<Blocked | null>(null);
  const no = request ? 'reject' : 'dismiss';
  const decide = async (action: 'accept' | 'reject' | 'dismiss') => {
    setBusy(action); setError(''); setBlocked(null);
    try {
      await send('POST', 'decisions/' + encodeURIComponent(item.id), action === 'accept' ? {action, scope, alias: alias ? name : undefined, target, model: alias ? model : undefined, note} : {action, note});
      const who = scopeLabel(scope, item.application);
      onDecided(action === 'accept'
        ? (alias ? 'Approved for ' + who + ': the alias “' + name + '” → ' + target + '/' + model + '.' : scope === 'every_application' ? 'Approved: ' + target + ' is now the default Connection for its protocol.' : 'Approved: ' + item.application + ' now sends its calls to ' + target + '.') + ' You can revert it below.'
        : action === 'reject' ? 'Rejected. ' + ((item.asked_by?.length || 0) > 1 ? 'Each Application that asked sees' : item.application + ' sees') + ' your note when it lists its requests.'
        : 'Dismissed. aiport stays quiet about this for 7 days unless the evidence doubles.');
    } catch (e) {
      const err = e as Error & {code?: string; extension?: string; reason?: string};
      if (err.code === 'guardrail_blocked') setBlocked({extension: err.extension || 'a guardrail', reason: err.reason || ''});
      else setError(err.message);
    } finally {
      setBusy('');
    }
  };
  const ev = item.evidence;
  const askers = item.asked_by || [], merged = askers.length > 1;
  const ready = !!target && (!alias || (!!name.trim() && !!model.trim()));
  return <div className="decision">
    <div className="need">{request && <span className="badge">{merged ? 'Asked by ' + askers.length + ' Applications' : 'Change request from ' + item.application}</span>} {item.need}</div>
    {ev && item.kind === 'pair_application' && <div className="evidence">Evidence: {ev.count} refused requests between {when(ev.from_ms)} and {when(ev.to_ms)} (looking back {ev.window_minutes} minutes). Only counts are kept, never the token.</div>}
    {ev && item.kind !== 'pair_application' && <div className="evidence">Evidence: {ev.count} requests between {when(ev.from_ms)} and {when(ev.to_ms)} (looking back {ev.window_minutes} minutes). Only request metadata was used.</div>}
    {request && !merged && <dl className="request-text">
      <dt>Kind</dt><dd>{item.kind}{item.kind !== 'application_alias' && item.kind !== 'bind_application' ? ' (not supported)' : ''}</dd>
      <dt>Details</dt><dd>{item.details}</dd>
      <dt>Why</dt><dd>{item.why}</dd>
      <dt>Asked for {scopeLabel(item.scope, item.application)}</dt><dd>{item.scope_case}</dd>
      {item.filed_by && <><dt>Filed by (as stated)</dt><dd>{item.filed_by}</dd></>}
      {!!item.repeats && <><dt>Asked again</dt><dd>{item.repeats} more {item.repeats === 1 ? 'time' : 'times'}</dd></>}
    </dl>}
    {merged && <>
      <dl className="request-text">
        <dt>Kind</dt><dd>{item.kind}</dd>
        <dt>Details</dt><dd>{item.details}</dd>
        <dt>Asked for</dt><dd>every Application</dd>
      </dl>
      <ul className="askers" aria-label="Each Application's argument">{askers.map(a => <li key={a.application}>
        <strong>{a.application}</strong>{a.filed_by ? ' (filed by ' + a.filed_by + ', as stated)' : ''}{a.repeats ? ' · asked again ' + times(a.repeats) : ''}
        <dl className="request-text"><dt>Why</dt><dd>{a.why}</dd><dt>Argument for every Application</dt><dd>{a.scope_case}</dd></dl>
      </li>)}</ul>
    </>}
    {error && <div className="notice error" role="alert">{error}</div>}
    {blocked && <div className="notice error" role="alert"><strong>Blocked by the guardrail {blocked.extension}:</strong> {blocked.reason} Nothing was changed; this item still waits. Reject it, or change what the guardrail checks and approve again.</div>}
    {(alias || item.kind === 'bind_application') && <form className="decision-form" onSubmit={e => { e.preventDefault(); decide('accept'); }}>
      <label>Scope<select value={scope} onChange={e => setScope(e.target.value)}>
        <option value="application">Only {item.application}</option>
        <option value="every_application">Every Application</option>
      </select></label>
      {alias && <label>Alias<input value={name} onChange={e => setName(e.target.value)} readOnly={!request} required maxLength={256} spellCheck={false} autoComplete="off"/></label>}
      <label>{alias ? 'Connection' : scope === 'every_application' ? 'Default Connection' : 'Bind to Connection'}<select value={target} onChange={e => setTarget(e.target.value)}>{choices.map(t => <option key={t}>{t}</option>)}</select></label>
      {alias && <label>Model to use<input value={model} onChange={e => setModel(e.target.value)} required maxLength={256} placeholder="model name on that connection" spellCheck={false} autoComplete="off"/></label>}
      <label>Note {request ? 'for ' + item.application : '(optional)'}<input value={note} onChange={e => setNote(e.target.value)} maxLength={2000}/></label>
      <div className="actions">
        <button className="primary" disabled={!!busy || !ready}>{busy === 'accept' ? 'Approving…' : 'Approve'}</button>
        <button type="button" className="secondary" disabled={!!busy} onClick={() => decide(no)}>{busy === no ? (request ? 'Rejecting…' : 'Dismissing…') : (request ? 'Reject' : 'Dismiss')}</button>
      </div>
    </form>}
    {item.informational && <div className="actions">
      {item.opens === 'applications' && onOpenApplications && <button type="button" className="primary" onClick={onOpenApplications}>Open Applications to pair it <ArrowUpRight size={14} aria-hidden/></button>}
      <button type="button" className="secondary" disabled={!!busy} onClick={() => decide('dismiss')}>{busy === 'dismiss' ? 'Dismissing…' : 'Dismiss'}</button>
    </div>}
  </div>;
}

/** Changes: Suggestions waiting for a decision (only when there are any), then every configuration edit made through aiport, newest first, with Revert. */
export default function Changes({tick, onChange, targets = [], onOpenApplications}: {tick: number; onChange: () => void; targets?: string[]; onOpenApplications?: () => void}) {
  const [pending, setPending] = useState<Decision[]>([]), [demand, setDemand] = useState<Demand[]>([]);
  const [changes, setChanges] = useState<Change[] | null>(null), [revertable, setRevertable] = useState(false);
  const [busy, setBusy] = useState(''), [error, setError] = useState(''), [message, setMessage] = useState('');
  const [reload, setReload] = useState(0), [upstream, setUpstream] = useState<Upstream | null>(null);
  useEffect(() => {
    send('GET', 'upstream').then(setUpstream).catch(() => setUpstream(null));
    send('GET', 'changes').then(d => { setChanges(d.changes || []); setRevertable(!!d.revertable); }).catch(e => setError(e.message));
    send('GET', 'decisions').then(d => { setPending(d.items || []); setDemand(d.demand || []); }).catch(() => { setPending([]); setDemand([]); });
  }, [tick, reload]);
  const revert = async (c: Change) => {
    setBusy(c.id); setError(''); setMessage('');
    try {
      await send('POST', 'changes/' + encodeURIComponent(c.id) + '/revert');
      setMessage('Reverted “' + c.what + '”. The previous configuration is live.');
      setReload(x => x + 1); onChange();
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setBusy('');
    }
  };
  const byId = new Map((changes || []).map(c => [c.id, c]));
  const decided = (m: string) => { setMessage(m); setError(''); setReload(x => x + 1); onChange(); };
  return <>{!!pending.length && <section className="panel" aria-label="Waiting for your decision">
    <div className="panel-heading"><h2><Lightbulb size={16} aria-hidden/> Waiting for your decision</h2><span>Suggestions from request history and Change requests from Applications</span></div>
    {pending.map(item => <Pending key={item.id} item={item} targets={targets} onDecided={decided} onOpenApplications={onOpenApplications}/>)}
  </section>}{!!demand.length && <section className="panel" aria-label="Demand for things aiport cannot do yet">
    <div className="panel-heading"><h2>Asked for, not supported yet</h2><span>Change requests aiport cannot apply · kept and counted, no decision needed</span></div>
    <div className="table-scroll"><table>
      <thead><tr><th>Kind</th><th>Requests</th><th>Applications</th><th>Last asked</th></tr></thead>
      <tbody>{demand.map(d => <tr key={d.kind}><td>{d.kind}</td><td>{d.requests}</td><td>{d.applications}</td><td>{when(d.last_at_ms)}</td></tr>)}</tbody>
    </table></div>
    {upstream && !!upstream.suggestions.length && <ul className="askers" aria-label="Suggest to aiport">{upstream.suggestions.map(s => <li key={s.kind}>
      <strong>{s.kind}</strong> <UpstreamAction label="Suggest to aiport" draft={s} what={{suggest: s.kind}} upstream={upstream}/>
    </li>)}</ul>}
  </section>}{upstream && !!upstream.shares.length && <section className="panel" aria-label="Extensions">
    <div className="panel-heading"><h2><Puzzle size={16} aria-hidden/> Extensions</h2><span>Listed in aiport.yaml · share one upstream for others to use</span></div>
    <ul className="askers">{upstream.shares.map(s => <li key={s.extension}>
      <strong>{s.extension}</strong> ({s.hook}) <UpstreamAction label="Share upstream" draft={s} what={{share: s.extension}} upstream={upstream}/>
    </li>)}</ul>
  </section>}<section className="panel">
    <div className="panel-heading"><h2>Applied changes</h2><span>Newest first · kept across restarts</span></div>
    {error && <div className="notice error" role="alert">{error}</div>}
    {message && <div className="notice" role="status">{message}</div>}
    {changes && !changes.length && <div className="empty"><span className="empty-icon"><History size={20} strokeWidth={1.75} aria-hidden/></span><h3>No changes yet</h3><p>Configuration edits made through aiport appear here, with the file before and after kept so you can revert them.</p></div>}
    {!!changes?.length && <div className="table-scroll"><table>
      <thead><tr><th>Time</th><th>Change</th><th>Made by</th><th>Status</th><th></th></tr></thead>
      <tbody>{changes.map(c => {
        const undone = c.reverted_by ? byId.get(c.reverted_by) : undefined;
        return <tr key={c.id}>
          <td title={new Date(c.at_ms).toISOString()}>{when(c.at_ms)}</td>
          <td>{c.what}</td>
          <td>{c.by}</td>
          <td>{c.reverted
            ? <span className="badge" title={undone ? 'Reverted by ' + undone.by : undefined}>Reverted{c.reverted_at_ms ? ' ' + when(c.reverted_at_ms) : ''}</span>
            : c.revert_of ? <span className="badge">Revert</span> : <span className="badge good">Applied</span>}</td>
          <td>{revertable && !c.reverted && <button className="text-button" disabled={!!busy} onClick={() => revert(c)} aria-label={'Revert ' + c.what}>
            <Undo2 size={14} aria-hidden/> {busy === c.id ? 'Reverting…' : 'Revert'}</button>}</td>
        </tr>;
      })}</tbody>
    </table></div>}
    <p className="caption">Revert restores the configuration file exactly as it was before the change and makes it live at once. If the configuration was edited after a change, revert the newer changes first.</p>
  </section></>;
}
