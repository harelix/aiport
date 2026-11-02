import React, {useEffect, useState} from 'react';
import {HelpLink} from './help';
import {AppWindow} from 'lucide-react';

/* Applications on this machine that use aiport for LLM calls: approve the
   ones asking to pair and see the ones connected. The integration
   instructions for coding agents are on the Integrate page. */

/** repository: the path the Application reported, not yet confirmed; repository_problem: what is wrong with it. */
export type Pairing = {code: string; name: string; description?: string; repository?: string; repository_problem?: string; requested_at_ms: number; expires_at_ms: number};
/** agent and repository: its Coding agent (empty = the default) and confirmed repository path. */
export type Origin = {id: string; name: string; target?: string; aliases?: {alias: string; target: string; model: string}[]; may_request_changes?: boolean; agent?: string; repository?: string};
export type Target = {id: string; disabled?: boolean};
/** A pairing request the Owner rejected, kept so it can still be seen (no secrets). */
export type Rejected = {id: string; code: string; name: string; description?: string; rejected_at_ms: number};

async function post(path: string, body?: object, method = 'POST') {
  const r = await fetch('/api/v1/' + path, {method, credentials: 'same-origin', headers: {'Content-Type': 'application/json', 'X-Aiport-CSRF': '1'}, body: body ? JSON.stringify(body) : undefined});
  const data = await r.json().catch(() => ({}));
  if (!r.ok) throw new Error(data.error?.message || 'Request failed (' + r.status + ').');
  return data;
}

const minutesLeft = (ms: number) => Math.max(0, Math.ceil((ms - Date.now()) / 60000));

function PairingRow({p, targets, onDone}: {p: Pairing; targets: Target[]; onDone: () => void}) {
  const [target, setTarget] = useState(''), [mayRequest, setMayRequest] = useState(false), [busy, setBusy] = useState(false), [error, setError] = useState('');
  // The reported repository path is only a proposal: what is in this field when the Owner approves is what aiport keeps.
  const [repository, setRepository] = useState(p.repository || '');
  async function act(action: 'approve' | 'reject') {
    setBusy(true); setError('');
    try { await post('pairings/' + encodeURIComponent(p.code) + '/' + action, action === 'approve' ? {target, may_request_changes: mayRequest, repository: repository.trim()} : undefined); onDone(); }
    catch (e) { setError((e as Error).message); } finally { setBusy(false); }
  }
  return <li className="pairing">
    <div>
      <strong>{p.name}</strong> <span className="pair-code" aria-label={'Code ' + p.code}>{p.code}</span>
      {p.description && <p className="muted-small">{p.description}</p>}
      <small>Check the application shows this same code. Expires in {minutesLeft(p.expires_at_ms)} min.</small>
      {p.repository && <div className="pairing-repository">
        <label htmlFor={'repo-' + p.code}>Repository <small className="muted">(reported by the application; confirm, edit or clear it)</small></label>
        <input id={'repo-' + p.code} value={repository} spellCheck={false} disabled={busy} onChange={e => setRepository(e.target.value)}/>
        {p.repository_problem && repository.trim() === p.repository && <small className="auth-error">{p.repository_problem}</small>}
      </div>}
    </div>
    <div className="pairing-actions">
      <label className="sr-only" htmlFor={'bind-' + p.code}>Connection</label>
      <select id={'bind-' + p.code} value={target} onChange={e => setTarget(e.target.value)} title="Optional: always send this application's calls to one connection">
        <option value="">Protocol default</option>
        {targets.filter(t => !t.disabled).map(t => <option key={t.id} value={t.id}>{t.id}</option>)}
      </select>
      <label className="check-label" title="Lets this application ask you for changes to aiport. You decide on each request.">
        <input type="checkbox" checked={mayRequest} onChange={e => setMayRequest(e.target.checked)}/>May request changes
      </label>
      <button type="button" className="primary" disabled={busy} onClick={() => act('approve')}>Approve</button>
      <button type="button" className="text-button" disabled={busy} onClick={() => act('reject')}>Reject</button>
    </div>
    {error && <p className="auth-error" role="alert">{error}</p>}
  </li>;
}

export function MayRequestToggle({o, onDone}: {o: Origin; onDone: () => void}) {
  const [busy, setBusy] = useState(false), [error, setError] = useState('');
  async function toggle(allowed: boolean) {
    setBusy(true); setError('');
    try { await post('origins/' + encodeURIComponent(o.id), {may_request_changes: allowed}, 'PATCH'); onDone(); }
    catch (e) { setError((e as Error).message); } finally { setBusy(false); }
  }
  return <>
    <label className="check-label"><input type="checkbox" checked={!!o.may_request_changes} disabled={busy} onChange={e => toggle(e.target.checked)} aria-label={'May ' + o.name + ' request changes'}/>{o.may_request_changes ? 'Allowed' : 'Not allowed'}</label>
    {error && <p className="auth-error" role="alert">{error}</p>}
  </>;
}

/** Pairing requests waiting for the Owner, refreshed while the page is open. */
export function usePairings(): [Pairing[], () => void] {
  const [pairings, setPairings] = useState<Pairing[]>([]), [tick, setTick] = useState(0);
  useEffect(() => {
    const controller = new AbortController();
    fetch('/api/v1/pairings', {credentials: 'same-origin', signal: controller.signal}).then(r => r.ok ? r.json() : {pairings: []}).then(d => setPairings(d.pairings || [])).catch(() => {});
    const timer = setTimeout(() => setTick(x => x + 1), 4000); // new requests show up while the page is open
    return () => { controller.abort(); clearTimeout(timer); };
  }, [tick]);
  return [pairings, () => setTick(x => x + 1)];
}

/** Rejected pairing requests, newest first; refetched when tick changes. */
export function useRejected(tick: number): [Rejected[], () => void] {
  const [list, setList] = useState<Rejected[]>([]), [own, setOwn] = useState(0);
  useEffect(() => {
    const controller = new AbortController();
    fetch('/api/v1/pairings?status=rejected', {credentials: 'same-origin', signal: controller.signal}).then(r => r.ok ? r.json() : {pairings: []}).then(d => setList(d.pairings || [])).catch(() => {});
    return () => controller.abort();
  }, [tick, own]);
  return [list, () => setOwn(x => x + 1)];
}

/** Forget one rejected Application (it may ask to pair again either way). */
export const forgetRejected = (id: string) => post('pairings/rejected/' + encodeURIComponent(id), undefined, 'DELETE');

/** Pending pairings with Approve / Reject. */
export function PairingList({pairings, targets, onDone}: {pairings: Pairing[]; targets: Target[]; onDone: () => void}) {
  return pairings.length ? <ul className="pairings">{pairings.map(p => <PairingRow key={p.code} p={p} targets={targets} onDone={onDone}/>)}</ul> : null;
}

export function ApplicationsPanel({origins, targets, onChange}: {origins: Origin[]; targets: Target[]; onChange: () => void}) {
  const [pairings, refresh] = usePairings();
  const done = () => { refresh(); onChange(); };
  return <section className="panel grad-bottom" aria-labelledby="apps-heading">
    <div className="panel-heading"><h2 id="apps-heading"><AppWindow size={16} aria-hidden/> Applications<HelpLink section="Applications and pairing"/></h2><span>{pairings.length ? pairings.length + ' waiting for approval' : origins.length + ' connected'}</span></div>
    <PairingList pairings={pairings} targets={targets} onDone={done}/>
    {origins.length > 0
      ? <div className="table-scroll"><table>
          <thead><tr><th>Application</th><th>Client id</th><th>Connection</th><th>Own aliases</th><th>May request changes</th></tr></thead>
          <tbody>{origins.map(o => <tr key={o.id}><td>{o.name}</td><td><code>{o.id}</code></td><td>{o.target || 'Protocol default'}</td><td>{o.aliases?.length ? o.aliases.map(a => a.alias + ' → ' + a.model).join(', ') : '—'}</td><td><MayRequestToggle o={o} onDone={done}/></td></tr>)}</tbody>
        </table></div>
      : !pairings.length && <p className="quiet-empty">No applications yet. Hand the Application brief on the Integrate page to the coding agent of an application; when it pairs, its request appears here and on the Integrate page for you to approve.</p>}
  </section>;
}
