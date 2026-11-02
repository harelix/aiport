import React, {useEffect, useState} from 'react';
import {History, Undo2} from 'lucide-react';

/** One applied Change to aiport's configuration, as GET /api/v1/changes lists it. */
type Change = {
  id: string; at_ms: number; what: string; by: string;
  revert_of?: string; reverted_by?: string; reverted_at_ms?: number; reverted: boolean;
};

async function send(method: string, path: string) {
  const r = await fetch('/api/v1/' + path, {method, credentials: 'same-origin', headers: {'Content-Type': 'application/json', 'X-Aiport-CSRF': '1'}});
  const d = await r.json().catch(() => ({}));
  if (!r.ok) throw new Error(d.error?.message || 'Request failed (' + r.status + ').');
  return d;
}

const when = (ms: number) => new Date(ms).toLocaleString();

/** Changes: every configuration edit made through aiport, newest first, with Revert. */
export default function Changes({tick, onChange}: {tick: number; onChange: () => void}) {
  const [changes, setChanges] = useState<Change[] | null>(null), [revertable, setRevertable] = useState(false);
  const [busy, setBusy] = useState(''), [error, setError] = useState(''), [message, setMessage] = useState('');
  const [reload, setReload] = useState(0);
  useEffect(() => {
    send('GET', 'changes').then(d => { setChanges(d.changes || []); setRevertable(!!d.revertable); }).catch(e => setError(e.message));
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
  return <section className="panel">
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
  </section>;
}
