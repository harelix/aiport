import React, {useEffect, useState} from 'react';
import {AppWindow, Check, Copy} from 'lucide-react';

/* Applications on this machine that use aiport for LLM calls: approve the
   ones asking to pair, see the ones connected, and hand integration
   instructions to a coding agent. */

type Pairing = {code: string; name: string; description?: string; requested_at_ms: number; expires_at_ms: number};
type Origin = {id: string; name: string; target?: string; aliases?: {alias: string; target: string; model: string}[]; may_request_changes?: boolean};
type Target = {id: string; disabled?: boolean};

async function post(path: string, body?: object, method = 'POST') {
  const r = await fetch('/api/v1/' + path, {method, credentials: 'same-origin', headers: {'Content-Type': 'application/json', 'X-Aiport-CSRF': '1'}, body: body ? JSON.stringify(body) : undefined});
  const data = await r.json().catch(() => ({}));
  if (!r.ok) throw new Error(data.error?.message || 'Request failed (' + r.status + ').');
  return data;
}

const minutesLeft = (ms: number) => Math.max(0, Math.ceil((ms - Date.now()) / 60000));

function PairingRow({p, targets, onDone}: {p: Pairing; targets: Target[]; onDone: () => void}) {
  const [target, setTarget] = useState(''), [mayRequest, setMayRequest] = useState(false), [busy, setBusy] = useState(false), [error, setError] = useState('');
  async function act(action: 'approve' | 'reject') {
    setBusy(true); setError('');
    try { await post('pairings/' + encodeURIComponent(p.code) + '/' + action, action === 'approve' ? {target, may_request_changes: mayRequest} : undefined); onDone(); }
    catch (e) { setError((e as Error).message); } finally { setBusy(false); }
  }
  return <li className="pairing">
    <div>
      <strong>{p.name}</strong> <span className="pair-code" aria-label={'Code ' + p.code}>{p.code}</span>
      {p.description && <p className="muted-small">{p.description}</p>}
      <small>Check the application shows this same code. Expires in {minutesLeft(p.expires_at_ms)} min.</small>
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

function MayRequestToggle({o, onDone}: {o: Origin; onDone: () => void}) {
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

export function ApplicationsPanel({origins, targets, onChange}: {origins: Origin[]; targets: Target[]; onChange: () => void}) {
  const [pairings, setPairings] = useState<Pairing[]>([]), [tick, setTick] = useState(0);
  useEffect(() => {
    const controller = new AbortController();
    fetch('/api/v1/pairings', {credentials: 'same-origin', signal: controller.signal}).then(r => r.ok ? r.json() : {pairings: []}).then(d => setPairings(d.pairings || [])).catch(() => {});
    const timer = setTimeout(() => setTick(x => x + 1), 4000); // new requests show up while the page is open
    return () => { controller.abort(); clearTimeout(timer); };
  }, [tick]);
  const done = () => { setTick(x => x + 1); onChange(); };
  return <section className="panel" aria-labelledby="apps-heading">
    <div className="panel-heading"><h2 id="apps-heading"><AppWindow size={16} aria-hidden/> Applications</h2><span>{pairings.length ? pairings.length + ' waiting for approval' : origins.length + ' connected'}</span></div>
    {pairings.length > 0 && <ul className="pairings">{pairings.map(p => <PairingRow key={p.code} p={p} targets={targets} onDone={done}/>)}</ul>}
    {origins.length > 0
      ? <div className="table-scroll"><table>
          <thead><tr><th>Application</th><th>Client id</th><th>Connection</th><th>Own aliases</th><th>May request changes</th></tr></thead>
          <tbody>{origins.map(o => <tr key={o.id}><td>{o.name}</td><td><code>{o.id}</code></td><td>{o.target || 'Protocol default'}</td><td>{o.aliases?.length ? o.aliases.map(a => a.alias + ' → ' + a.model).join(', ') : '—'}</td><td><MayRequestToggle o={o} onDone={done}/></td></tr>)}</tbody>
        </table></div>
      : !pairings.length && <p className="quiet-empty">No applications yet. Hand the instructions below to the coding agent of an application; when it pairs, its request appears here for you to approve.</p>}
  </section>;
}

export function HandoverPanel() {
  const [text, setText] = useState(''), [copied, setCopied] = useState(false), [error, setError] = useState('');
  useEffect(() => {
    fetch('/_connect?format=handover', {credentials: 'same-origin'}).then(r => { if (!r.ok) throw new Error(); return r.text(); }).then(setText).catch(() => setError('Could not load the instructions from /_connect.'));
  }, []);
  const copy = () => void navigator.clipboard?.writeText(text).then(() => { setCopied(true); setTimeout(() => setCopied(false), 1500); });
  return <section className="panel" aria-labelledby="handover-heading">
    <div className="panel-heading"><h2 id="handover-heading">Hand over to a coding agent</h2>
      <button type="button" className="secondary with-icon" disabled={!text} onClick={copy}>{copied ? <><Check size={14} aria-hidden/> Copied</> : <><Copy size={14} aria-hidden/> Copy instructions</>}</button></div>
    <p className="muted handover-intro">Paste these into the coding agent working on an application (or into that repository's <code>AGENTS.md</code>). The agent pairs the application, you approve it above, and its LLM calls then go through aiport with its own token. They always match this gateway: <a href="/_connect?format=handover" target="_blank" rel="noreferrer">/_connect?format=handover</a>.</p>
    {error ? <p className="auth-error" role="alert">{error}</p> : <pre className="handover">{text || 'Loading…'}</pre>}
  </section>;
}
