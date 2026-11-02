import React, {useState} from 'react';
import {ArrowUpRight, Copy, Send} from 'lucide-react';

/** A prefilled GitHub issue for aiport itself, as GET /api/v1/upstream lists it: kinds, hooks and anonymous counts only. */
export type Draft = {title: string; body: string; url?: string};
export type Upstream = {
  repository: string; gh: boolean;
  suggestions: (Draft & {kind: string})[];
  shares: (Draft & {extension: string; hook: string})[];
};

async function fileWithGh(what: {suggest: string} | {share: string}, title: string, body: string) {
  const r = await fetch('/api/v1/upstream/issues', {method: 'POST', credentials: 'same-origin', headers: {'Content-Type': 'application/json', 'X-Aiport-CSRF': '1'}, body: JSON.stringify({...what, confirm: true, title, body})});
  const d = await r.json().catch(() => ({}));
  if (!r.ok) throw new Error(d.error?.message || 'Request failed (' + r.status + ').');
  return d.url as string;
}

/** The GitHub link that opens a new issue on repository with this title and body. */
function issueURL(repository: string, title: string, body: string) {
  return 'https://github.com/' + repository + '/issues/new?' + new URLSearchParams({body, title}).toString();
}

/**
 * "Suggest to aiport" or "Share upstream" for one draft. Nothing is sent on its own: the Owner first checks the text
 * (a kind named by an Application is marked in it) and may edit it, then opens it on GitHub in the browser, copies it,
 * or confirms that aiport runs gh with the Owner's own login. Every way out uses the text as edited.
 */
export function UpstreamAction({label, draft, what, upstream}: {label: string; draft: Draft; what: {suggest: string} | {share: string}; upstream: Upstream}) {
  const [open, setOpen] = useState(false), [confirming, setConfirming] = useState(false);
  const [busy, setBusy] = useState(false), [error, setError] = useState(''), [filed, setFiled] = useState(''), [copied, setCopied] = useState(false);
  const [title, setTitle] = useState(draft.title), [body, setBody] = useState(draft.body);
  const ready = title.trim() !== '' && body.trim() !== '';
  const file = async () => {
    setBusy(true); setError('');
    try {
      const url = await fileWithGh(what, title.trim(), body);
      setFiled(url || 'Filed.'); setConfirming(false);
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setBusy(false);
    }
  };
  const copy = () => navigator.clipboard?.writeText(title.trim() + '\n\n' + body).then(() => setCopied(true)).catch(() => setError('Copy failed: select the text instead.'));
  if (!open) return <button className="text-button" onClick={() => setOpen(true)}>{label}</button>;
  return <div className="upstream-draft">
    <p className="caption">This is everything that would leave aiport: kinds and anonymous counts only, no configuration, keys, paths or Application names. A kind named by an Application is its own words: check it, and edit the text as you like. Nothing is sent until you choose below.</p>
    <label className="draft-field">Title<input className="draft-title" value={title} maxLength={256} disabled={confirming || !!filed} onChange={e => { setTitle(e.target.value.replace(/[\r\n]/g, ' ')); setCopied(false); }}/></label>
    <label className="draft-field">Body<textarea className="draft-body" value={body} rows={10} disabled={confirming || !!filed} onChange={e => { setBody(e.target.value); setCopied(false); }}/></label>
    {error && <div className="notice error" role="alert">{error}</div>}
    {filed && <div className="notice" role="status">Filed with gh{filed.startsWith('https://') ? <>: <a href={filed} target="_blank" rel="noopener noreferrer">{filed}</a></> : '.'}</div>}
    {confirming && !filed && <div className="notice" role="alertdialog" aria-label="Confirm filing with gh">
      <div>aiport will run <code>gh issue create</code> on {upstream.repository} with your own gh login, sending exactly the text above as you edited it. File it now?</div>
      <div className="actions">
        <button className="primary" disabled={busy} onClick={file}>{busy ? 'Filing…' : 'Yes, file it'}</button>
        <button className="secondary" disabled={busy} onClick={() => setConfirming(false)}>Cancel</button>
      </div>
    </div>}
    {!confirming && !filed && <div className="actions">
      {draft.url && ready && <a className="button primary with-icon" href={issueURL(upstream.repository, title.trim(), body)} target="_blank" rel="noopener noreferrer"><ArrowUpRight size={14} aria-hidden/> Open on GitHub</a>}
      {upstream.gh && <button className="secondary with-icon" disabled={!ready} onClick={() => setConfirming(true)}><Send size={14} aria-hidden/> File with gh…</button>}
      <button className="secondary with-icon" disabled={!ready} onClick={copy}><Copy size={14} aria-hidden/> {copied ? 'Copied' : 'Copy text'}</button>
      <button className="text-button" onClick={() => { setOpen(false); setError(''); }}>Close</button>
    </div>}
    {!draft.url && <p className="caption">No upstream repository is set: set <code>upstream.repository</code> in aiport.yaml to open drafts on GitHub, or copy the text.</p>}
  </div>;
}
