import React, {useEffect, useState} from 'react';
import {ArrowUpRight, CircleAlert, History, Lightbulb, Puzzle, Undo2} from 'lucide-react';
import {type Upstream, UpstreamAction} from './Upstream';
import {DecisionSoundToggle} from './decisionSound';
import {CopyButton} from './Integrate';

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
  /** The Issue this Suggestion is the aiport fix for: shown with that Issue, not on its own. */
  issue?: string;
};

/** An Issue: a problem aiport keeps seeing for one Application, as GET /api/v1/issues lists it. */
export type Issue = {
  id: string; application: string; cause: string; subject: string;
  state: 'open' | 'resolving' | 'resolved' | 'dismissed';
  summary: string; needs_owner: boolean; application_fix: string;
  evidence: {count: number; from_ms: number; to_ms: number; window_minutes: number; status?: number; error?: string; examples?: string[]; connections?: string[]};
  diagnosis?: {protocol: string; answer: 'one' | 'none' | 'several' | 'unknown'; serving: string[]; unknown: string[]; available?: string[]};
  fix?: {suggestion: string; kind: string; scope: string; proposal: {alias?: string; target?: string; model?: string}; state: 'proposed' | 'applied' | 'rejected' | 'reverted'; change_id?: string; by?: string};
  resolving_by?: string; resolution?: {how: string; at_ms: number; change_id?: string}; reopened?: number;
};

const diagnosisText = (d: NonNullable<Issue['diagnosis']>, model: string) =>
  d.answer === 'one' ? 'Only ' + d.serving[0] + ' serves ' + model + '.'
  : d.answer === 'several' ? d.serving.join(', ') + ' all serve ' + model + ': choose one.'
  : d.answer === 'none' ? 'No enabled ' + d.protocol + ' Connection lists ' + model + '.'
  : 'aiport could not list the models of ' + (d.unknown.join(', ') || 'every Connection') + ', so it is not certain which serves ' + model + '.';

/**
 * One open or resolving Issue: what is wrong, the evidence, which Connections serve the model, the aiport fix
 * (an Alias for this Application only) with Approve / Dismiss, and the fix in the Application for its Coding agent.
 */
function IssueCard({issue, targets, onDecided}: {issue: Issue; targets: string[]; onDecided: (message: string) => void}) {
  const fix = issue.fix, ev = issue.evidence;
  const [target, setTarget] = useState(fix?.proposal.target || ''), [model, setModel] = useState(fix?.proposal.model || issue.subject);
  const [busy, setBusy] = useState(''), [error, setError] = useState(''), [blocked, setBlocked] = useState<Blocked | null>(null);
  const decide = async (action: 'accept' | 'dismiss') => {
    setBusy(action); setError(''); setBlocked(null);
    try {
      await send('POST', 'issues/' + encodeURIComponent(issue.id), action === 'accept' ? {action, target, model} : {action});
      onDecided(action === 'accept'
        ? 'Approved for only ' + issue.application + ': the alias “' + issue.subject + '” → ' + target + '/' + model + '. The Issue resolves itself when its calls succeed; you can revert the Change below.'
        : 'Dismissed. aiport stays quiet about this Issue for 7 days unless the evidence doubles.');
    } catch (e) {
      const err = e as Error & {code?: string; extension?: string; reason?: string};
      if (err.code === 'guardrail_blocked') setBlocked({extension: err.extension || 'a guardrail', reason: err.reason || ''});
      else setError(err.message);
    } finally {
      setBusy('');
    }
  };
  return <div className="decision issue" id={issueAnchor(issue.id)} tabIndex={-1}>
    <div className="need"><span className={'badge' + (issue.state === 'resolving' ? ' good' : '')}>{issue.state === 'resolving' ? 'Resolving' : 'Issue'}</span> {issue.summary}</div>
    <div className="evidence">Evidence: {times(ev.count)} {ev.status}{ev.error ? ' ' + ev.error : ''} between {when(ev.from_ms)} and {when(ev.to_ms)} (looking back {ev.window_minutes} minutes){ev.connections?.length ? ', sent to ' + ev.connections.join(', ') : ''}.{ev.examples?.length ? ' Example requests: ' + ev.examples.slice(0, 3).join(', ') + '.' : ''}{issue.reopened ? ' Came back ' + times(issue.reopened) + '.' : ''}</div>
    {issue.diagnosis && <div className="evidence">Diagnosis: {diagnosisText(issue.diagnosis, issue.subject)} Found from the Connections’ free model lists.</div>}
    {issue.state === 'resolving' && <div className="evidence">{issue.resolving_by === 'aiport_fix' ? 'The aiport fix is applied' + (fix?.by ? ' (by ' + fix.by + ')' : '') : 'The Application says it fixed this'}: aiport marks the Issue resolved when {issue.application}’s calls for {issue.subject} succeed.</div>}
    {error && <div className="notice error" role="alert">{error}</div>}
    {blocked && <div className="notice error" role="alert"><strong>Blocked by the guardrail {blocked.extension}:</strong> {blocked.reason} Nothing was changed; this Issue stays open.</div>}
    {issue.state === 'open' && fix?.state === 'proposed' && <form className="decision-form" onSubmit={e => { e.preventDefault(); decide('accept'); }}>
      <label>Fix in aiport: alias for only {issue.application}<input value={issue.subject} readOnly spellCheck={false}/></label>
      <label>Connection<select value={target} onChange={e => setTarget(e.target.value)}>
        {!target && <option value="">Choose a Connection</option>}
        {targets.map(t => <option key={t}>{t}</option>)}
      </select></label>
      <label>Model to use<input value={model} onChange={e => setModel(e.target.value)} required maxLength={256} spellCheck={false} autoComplete="off"/></label>
      <div className="actions">
        <button className="primary" disabled={!!busy || !target || !model.trim()}>{busy === 'accept' ? 'Approving…' : 'Approve'}</button>
        <button type="button" className="secondary" disabled={!!busy} onClick={() => decide('dismiss')}>{busy === 'dismiss' ? 'Dismissing…' : 'Dismiss'}</button>
      </div>
    </form>}
    <ApplicationFix issue={issue}/>
  </div>;
}

/** The id of an Issue's card on the Changes page, so Integrate can link to it. */
export const issueAnchor = (id: string) => 'issue-' + id;

/**
 * The fix in the Application, for its coding agent: "Copy fix for the coding agent" copies the text the server
 * writes (decide.ApplicationFix, the same text aiport_my_issues gives the agent), shown below it to read first.
 */
export function ApplicationFix({issue}: {issue: Issue}) {
  if (!issue.application_fix) return null;
  return <div className="application-fix">
    <CopyButton text={issue.application_fix} label="Copy fix for the coding agent"/>
    <details><summary>Read the fix for {issue.application}’s coding agent</summary><pre>{issue.application_fix}</pre></details>
  </div>;
}

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
      {item.opens === 'applications' && onOpenApplications && <button type="button" className="primary" onClick={onOpenApplications}>Open Integrate to pair it <ArrowUpRight size={14} aria-hidden/></button>}
      <button type="button" className="secondary" disabled={!!busy} onClick={() => decide('dismiss')}>{busy === 'dismiss' ? 'Dismissing…' : 'Dismiss'}</button>
    </div>}
  </div>;
}

/** Changes: Suggestions waiting for a decision (only when there are any), then every configuration edit made through aiport, newest first, with Revert. */
export default function Changes({tick, onChange, targets = [], onOpenApplications}: {tick: number; onChange: () => void; targets?: string[]; onOpenApplications?: () => void}) {
  const [pending, setPending] = useState<Decision[]>([]), [demand, setDemand] = useState<Demand[]>([]), [issues, setIssues] = useState<Issue[]>([]);
  const [changes, setChanges] = useState<Change[] | null>(null), [revertable, setRevertable] = useState(false);
  const [busy, setBusy] = useState(''), [error, setError] = useState(''), [message, setMessage] = useState('');
  const [reload, setReload] = useState(0), [upstream, setUpstream] = useState<Upstream | null>(null);
  useEffect(() => {
    send('GET', 'upstream').then(setUpstream).catch(() => setUpstream(null));
    send('GET', 'changes').then(d => { setChanges(d.changes || []); setRevertable(!!d.revertable); }).catch(e => setError(e.message));
    send('GET', 'decisions').then(d => { setPending((d.items || []).filter((i: Decision) => !i.issue)); setDemand(d.demand || []); }).catch(() => { setPending([]); setDemand([]); });
    send('GET', 'issues').then(d => setIssues(d.issues || [])).catch(() => setIssues([]));
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
  // Integrate links to one Issue with #issue-<id>: bring it into view once listed.
  useEffect(() => {
    const id = window.location.hash.slice(1);
    if (!id.startsWith('issue-') || !issues.length) return;
    const el = document.getElementById(id);
    if (el) { el.scrollIntoView({behavior: 'smooth', block: 'start'}); el.focus({preventScroll: true}); history.replaceState(null, '', window.location.pathname + window.location.search); }
  }, [issues]);
  const byId = new Map((changes || []).map(c => [c.id, c]));
  const decided = (m: string) => { setMessage(m); setError(''); setReload(x => x + 1); onChange(); };
  return <>{!!issues.length && <section className="panel" aria-label="Issues">
    <div className="panel-heading"><h2><CircleAlert size={16} aria-hidden/> Issues</h2><span>Problems aiport keeps seeing for one Application · they close themselves when its calls succeed again</span></div>
    {issues.map(i => <IssueCard key={i.id + i.state + (i.fix?.suggestion || '')} issue={i} targets={targets} onDecided={decided}/>)}
  </section>}{!!pending.length && <section className="panel" aria-label="Waiting for your decision">
    <div className="panel-heading"><h2><Lightbulb size={16} aria-hidden/> Waiting for your decision</h2><span>Suggestions from request history and Change requests from Applications</span></div>
    {pending.map(item => <Pending key={item.id} item={item} targets={targets} onDecided={decided} onOpenApplications={onOpenApplications}/>)}
  </section>}{!!demand.length && <section className="panel" aria-label="Demand for things aiport cannot do yet">
    <div className="panel-heading"><h2>Asked for, not supported yet</h2><span>Change requests aiport cannot apply · kept and counted, no decision needed</span></div>
    <div className="table-scroll"><table>
      <thead><tr><th>Kind</th><th>Requests</th><th>Applications</th><th>Last asked</th></tr></thead>
      <tbody>{demand.map(d => <tr key={d.kind}><td>{d.kind}</td><td>{d.requests}</td><td>{d.applications}</td><td>{when(d.last_at_ms)}</td></tr>)}</tbody>
    </table></div>
    {upstream && !!upstream.suggestions.length && <ul className="askers" aria-label="Suggest to aiport">{upstream.suggestions.map(s => <li key={s.kind}>
      <strong>{s.kind}</strong> <span className="caption">kind named by an Application</span> <UpstreamAction label="Suggest to aiport" draft={s} what={{suggest: s.kind}} upstream={upstream}/>
    </li>)}</ul>}
  </section>}{upstream && !!upstream.shares.length && <section className="panel" aria-label="Extensions">
    <div className="panel-heading"><h2><Puzzle size={16} aria-hidden/> Extensions</h2><span>Listed in aiport.yaml · share one upstream for others to use</span></div>
    <ul className="askers">{upstream.shares.map(s => <li key={s.extension}>
      <strong>{s.extension}</strong> ({s.hook}) <UpstreamAction label="Share upstream" draft={s} what={{share: s.extension}} upstream={upstream}/>
    </li>)}</ul>
  </section>}<section className="panel">
    <div className="panel-heading"><h2>Applied changes</h2><div className="header-actions"><span className="caption">Newest first · kept across restarts</span><DecisionSoundToggle/></div></div>
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
