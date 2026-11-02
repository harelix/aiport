import React, {useState} from 'react';
import {ArrowUpRight, Copy, Send} from 'lucide-react';

/** A prefilled GitHub issue for aiport itself, as GET /api/v1/upstream lists it: kinds, hooks and anonymous counts only. */
export type Draft = {title: string; body: string; url?: string};
export type Upstream = {
  repository: string; gh: boolean;
  suggestions: (Draft & {kind: string})[];
  shares: (Draft & {extension: string; hook: string})[];
};

async function fileWithGh(what: {suggest: string} | {share: string}) {
  const r = await fetch('/api/v1/upstream/issues', {method: 'POST', credentials: 'same-origin', headers: {'Content-Type': 'application/json', 'X-Aiport-CSRF': '1'}, body: JSON.stringify({...what, confirm: true})});
  const d = await r.json().catch(() => ({}));
  if (!r.ok) throw new Error(d.error?.message || 'Request failed (' + r.status + ').');
  return d.url as string;
}

/**
 * "Suggest to aiport" or "Share upstream" for one draft. Nothing is sent on its own: the Owner first sees the exact
 * text, then opens it on GitHub in the browser, copies it, or confirms that aiport runs gh with the Owner's own login.
 */
export function UpstreamAction({label, draft, what, upstream}: {label: string; draft: Draft; what: {suggest: string} | {share: string}; upstream: Upstream}) {
  const [open, setOpen] = useState(false), [confirming, setConfirming] = useState(false);
  const [busy, setBusy] = useState(false), [error, setError] = useState(''), [filed, setFiled] = useState(''), [copied, setCopied] = useState(false);
  const file = async () => {
    setBusy(true); setError('');
    try {
      const url = await fileWithGh(what);
      setFiled(url || 'Filed.'); setConfirming(false);
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setBusy(false);
    }
  };
  const copy = () => navigator.clipboard?.writeText(draft.title + '\n\n' + draft.body).then(() => setCopied(true)).catch(() => setError('Copy failed: select the text instead.'));
  if (!open) return <button className="text-button" onClick={() => setOpen(true)}>{label}</button>;
  return <div className="upstream-draft">
    <p className="caption">This is everything that would leave aiport: kinds and anonymous counts only, no configuration, keys, paths or Application names. Nothing is sent until you choose below.</p>
    <div className="draft-title">{draft.title}</div>
    <pre className="draft-body">{draft.body}</pre>
    {error && <div className="notice error" role="alert">{error}</div>}
    {filed && <div className="notice" role="status">Filed with gh{filed.startsWith('https://') ? <>: <a href={filed} target="_blank" rel="noopener noreferrer">{filed}</a></> : '.'}</div>}
    {confirming && !filed && <div className="notice" role="alertdialog" aria-label="Confirm filing with gh">
      <div>aiport will run <code>gh issue create</code> on {upstream.repository} with your own gh login, sending exactly the text above. File it now?</div>
      <div className="actions">
        <button className="primary" disabled={busy} onClick={file}>{busy ? 'Filing…' : 'Yes, file it'}</button>
        <button className="secondary" disabled={busy} onClick={() => setConfirming(false)}>Cancel</button>
      </div>
    </div>}
    {!confirming && !filed && <div className="actions">
      {draft.url && <a className="button primary with-icon" href={draft.url} target="_blank" rel="noopener noreferrer"><ArrowUpRight size={14} aria-hidden/> Open on GitHub</a>}
      {upstream.gh && <button className="secondary with-icon" onClick={() => setConfirming(true)}><Send size={14} aria-hidden/> File with gh…</button>}
      <button className="secondary with-icon" onClick={copy}><Copy size={14} aria-hidden/> {copied ? 'Copied' : 'Copy text'}</button>
      <button className="text-button" onClick={() => { setOpen(false); setError(''); }}>Close</button>
    </div>}
    {!draft.url && <p className="caption">No upstream repository is set: set <code>upstream.repository</code> in aiport.yaml to open drafts on GitHub, or copy the text.</p>}
  </div>;
}
