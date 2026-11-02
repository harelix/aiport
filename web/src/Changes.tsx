import React, {useEffect, useState} from 'react';
import {History, Lightbulb, Undo2} from 'lucide-react';

/** One applied Change to aiport's configuration, as GET /api/v1/changes lists it. */
type Change = {
  id: string; at_ms: number; what: string; by: string;
  revert_of?: string; reverted_by?: string; reverted_at_ms?: number; reverted: boolean;
};

/** A Suggestion (later also a Change request) waiting for the Owner, as GET /api/v1/decisions lists it. */
export type Decision = {
  id: string; source: string; kind: string; application: string; subject?: string; need: string; scope: string;
  evidence?: {count: number; from_ms: number; to_ms: number; window_minutes: number};
  proposal: {alias?: string; target?: string; model?: string};
};

async function send(method: string, path: string, body?: unknown) {
  const r = await fetch('/api/v1/' + path, {method, credentials: 'same-origin', headers: {'Content-Type': 'application/json', 'X-Aiport-CSRF': '1'}, body: body === undefined ? undefined : JSON.stringify(body)});
  const d = await r.json().catch(() => ({}));
  if (!r.ok) throw new Error(d.error?.message || 'Request failed (' + r.status + ').');
  return d;
}

const when = (ms: number) => new Date(ms).toLocaleString();

/** One pending Suggestion: the evidence, the Alias it proposes, and Approve / Dismiss. All text is shown as plain text. */
function Pending({item, targets, onDecided}: {item: Decision; targets: string[]; onDecided: (message: string) => void}) {
  const binding = item.kind === 'bind_application', choices = binding ? targets.filter(t => t !== item.subject) : targets;
  const [target, setTarget] = useState(item.proposal.target || choices[0] || ''), [model, setModel] = useState(item.proposal.model || '');
  const [note, setNote] = useState(''), [busy, setBusy] = useState(''), [error, setError] = useState('');
  const decide = async (action: 'accept' | 'dismiss') => {
    setBusy(action); setError('');
    try {
      await send('POST', 'decisions/' + encodeURIComponent(item.id), action === 'accept' ? (binding ? {action, target, note} : {action, target, model, note}) : {action, note});
      onDecided(action === 'accept'
        ? binding
          ? 'Approved: all of ' + item.application + '’s calls now go to ' + target + '. You can revert it below.'
          : 'Approved: ' + item.application + ' now has the alias “' + item.proposal.alias + '” → ' + target + '/' + model + '. You can revert it below.'
        : 'Dismissed. aiport stays quiet about this for 7 days unless the evidence doubles.');
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setBusy('');
    }
  };
  const ev = item.evidence;
  return <div className="decision">
    <div className="need">{item.need}</div>
    {ev && <div className="evidence">Evidence: {ev.count} requests between {when(ev.from_ms)} and {when(ev.to_ms)} (looking back {ev.window_minutes} minutes). Only request metadata was used.</div>}
    {error && <div className="notice error" role="alert">{error}</div>}
    {item.kind === 'application_alias' && <form className="decision-form" onSubmit={e => { e.preventDefault(); decide('accept'); }}>
      <label>Alias for {item.application}<input value={item.proposal.alias || ''} readOnly/></label>
      <label>Connection<select value={target} onChange={e => setTarget(e.target.value)}>{targets.map(t => <option key={t}>{t}</option>)}</select></label>
      <label>Model to use<input value={model} onChange={e => setModel(e.target.value)} required maxLength={256} placeholder="model name on that connection" spellCheck={false} autoComplete="off"/></label>
      <label>Note (optional)<input value={note} onChange={e => setNote(e.target.value)} maxLength={2000}/></label>
      <div className="actions">
        <button className="primary" disabled={!!busy || !model.trim() || !target}>{busy === 'accept' ? 'Approving…' : 'Approve'}</button>
        <button type="button" className="secondary" disabled={!!busy} onClick={() => decide('dismiss')}>{busy === 'dismiss' ? 'Dismissing…' : 'Dismiss'}</button>
      </div>
    </form>}
    {binding && <form className="decision-form" onSubmit={e => { e.preventDefault(); decide('accept'); }}>
      <label>Failing Connection<input value={item.subject || ''} readOnly/></label>
      <label>Bind {item.application} to<select value={target} onChange={e => setTarget(e.target.value)}>{choices.map(t => <option key={t}>{t}</option>)}</select></label>
      <label>Note (optional)<input value={note} onChange={e => setNote(e.target.value)} maxLength={2000}/></label>
      <div className="actions">
        <button className="primary" disabled={!!busy || !target}>{busy === 'accept' ? 'Approving…' : 'Approve'}</button>
        <button type="button" className="secondary" disabled={!!busy} onClick={() => decide('dismiss')}>{busy === 'dismiss' ? 'Dismissing…' : 'Dismiss'}</button>
      </div>
    </form>}
  </div>;
}

/** Changes: Suggestions waiting for a decision (only when there are any), then every configuration edit made through aiport, newest first, with Revert. */
export default function Changes({tick, onChange, targets = []}: {tick: number; onChange: () => void; targets?: string[]}) {
  const [pending, setPending] = useState<Decision[]>([]);
  const [changes, setChanges] = useState<Change[] | null>(null), [revertable, setRevertable] = useState(false);
  const [busy, setBusy] = useState(''), [error, setError] = useState(''), [message, setMessage] = useState('');
  const [reload, setReload] = useState(0);
  useEffect(() => {
    send('GET', 'changes').then(d => { setChanges(d.changes || []); setRevertable(!!d.revertable); }).catch(e => setError(e.message));
    send('GET', 'decisions').then(d => setPending(d.items || [])).catch(() => setPending([]));
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
    <div className="panel-heading"><h2><Lightbulb size={16} aria-hidden/> Waiting for your decision</h2><span>Suggestions from request history</span></div>
    {pending.map(item => <Pending key={item.id} item={item} targets={targets} onDecided={decided}/>)}
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
