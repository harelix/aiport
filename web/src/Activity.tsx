/* Activity's request table. Details open in place, under their row, grouped
   for reading: what the Application asked for, what served it, the outcome
   and why it failed. A failed request from an Application can be reported to
   that Application's coding agent (POST /api/v1/issues): it becomes an Issue
   the agent reads at the start of its session, its next failing call of that
   kind names it in X-Aiport-Hint, and Fix it in Changes can hand it to a
   coding agent aiport runs. Request history keeps metadata only, so the
   Owner's note is where the error text the Application showed goes. */
import React, {useState} from 'react';
import {ArrowUpRight, ChevronDown, Copy, Inbox, Send, ThumbsDown, ThumbsUp} from 'lucide-react';
import {Reason, ReasonPanel} from './Reason';
import {issueAnchor} from './Changes';

export type Call = {id: string; started_at_ms: number; origin_id: string; target_id: string; provider: string; protocol?: string; route: string; method?: string; requested_model: string | null; effort?: string; effort_source?: string; resolved_model?: string; module?: string; feedback?: number | null; http_status: number; outcome: string; latency_ms: number; ttfb_ms: number | null; compression?: string; compression_status?: string; refusal_code?: string; upstream_error_type?: string; connection_error?: string; reason?: Reason};

// Where a request's effort came from, as request details say it.
const effortSources: Record<string, string> = {header: 'from the X-Aiport-Effort header', field: "from the provider's effort field", default: 'none stated, so the default'};
const fmt = (n: number | null | undefined, suffix = '') => n == null ? '—' : new Intl.NumberFormat(undefined, {maximumFractionDigits: 0}).format(n) + suffix;
export const failedCall = (r: Call) => r.outcome !== 'success' || r.http_status >= 400;
const MAX_NOTE = 2000;

type Props = {
  calls: Call[]; selected: string; onSelect: (id: string) => void; onPage: (page: string) => void;
  applications: Record<string, string>; recording: 'off' | 'on'; onSetup: () => void;
};

export default function RequestTable({calls, selected, onSelect, onPage, applications, recording, onSetup}: Props) {
  if (!calls.length) return <div className="empty"><span className="empty-icon"><Inbox size={20} strokeWidth={1.75} aria-hidden/></span>
    <h3>{recording === 'off' ? 'Activity recording is off' : 'Your gateway is ready for traffic'}</h3>
    <p>{recording === 'off' ? 'Enable telemetry in your YAML configuration and restart to see session metadata.' : 'Connect a coding agent or send an API request. Metadata will appear here; prompts and responses are never captured.'}</p>
    <button className="secondary" onClick={onSetup}>View connection setup</button></div>;
  return <div className="table-scroll"><table className="activity-table"><thead><tr><th>Time</th><th>Origin / target</th><th>Model</th><th>Outcome</th><th className="numeric">Latency</th><th></th></tr></thead>
    <tbody>{calls.map(r => { const open = selected === r.id; return <React.Fragment key={r.id}>
      <tr className={open ? 'row-open' : undefined}>
        <td title={new Date(r.started_at_ms).toISOString()}>{new Date(r.started_at_ms).toLocaleTimeString()}<small>{new Date(r.started_at_ms).toLocaleDateString()}</small></td>
        <td>{r.origin_id || 'Unknown'}<small>{r.target_id || 'No target'}</small></td>
        <td>{r.requested_model || 'Unknown'}<small>{r.route.replace('/v1/', '')}{r.resolved_model && r.resolved_model !== r.requested_model ? ' → ' + r.resolved_model : ''}{r.effort ? ' · ' + r.effort : ''}{r.module ? ' · ' + r.module : ''}</small></td>
        <td><span className={'badge ' + (r.outcome === 'success' ? 'good' : '')}>{r.outcome.replaceAll('_', ' ')}</span><small className={r.reason ? 'reason-line' : undefined} title={r.reason?.explanation}>{r.reason ? r.http_status + ' · ' + r.reason.summary : 'HTTP ' + r.http_status}</small></td>
        <td className="numeric">{fmt(r.latency_ms, ' ms')}</td>
        <td><button className={'text-button details-toggle' + (open ? ' on' : '')} onClick={() => onSelect(open ? '' : r.id)} aria-expanded={open} aria-controls={'detail-' + r.id} aria-label={(open ? 'Hide' : 'Show') + ' details of request ' + r.id}>Details <ChevronDown size={14} aria-hidden/></button></td>
      </tr>
      {open && <tr className="detail-row" id={'detail-' + r.id}><td colSpan={6}><RequestDetail call={r} onPage={onPage} application={applications[r.origin_id]}/></td></tr>}
    </React.Fragment>; })}</tbody></table></div>;
}

function Field({label, children, wide}: {label: string; children: React.ReactNode; wide?: boolean}) {
  return <div className={'detail-field' + (wide ? ' wide' : '')}><dt>{label}</dt><dd>{children}</dd></div>;
}

const Code = ({v}: {v?: string | number | null}) => v == null || v === '' ? <span className="muted">—</span> : <code>{v}</code>;

function RequestDetail({call: r, onPage, application}: {call: Call; onPage: (page: string) => void; application?: string}) {
  const vote = r.feedback == null ? null : r.feedback > 0.5 ? 'up' : r.feedback < 0.5 ? 'down' : 'neutral';
  const at = new Date(r.started_at_ms);
  return <div className="request-detail" role="region" aria-label={'Details of request ' + r.id}>
    {r.reason && <ReasonPanel reason={r.reason} codes={[['aiport code', r.refusal_code], ['connection error', r.connection_error], ['provider error type', r.upstream_error_type], ['HTTP status', String(r.http_status)]]} onPage={onPage}/>}
    {r.requested_model && <p className="served">Asked for <b>{r.requested_model}</b>{r.effort && <> at <b>{r.effort}</b> effort ({effortSources[r.effort_source || ''] || r.effort_source})</>}{r.resolved_model ? <>; served by <b>{r.target_id || 'an unknown Connection'}</b> with <b>{r.resolved_model}</b>.</> : '; not forwarded.'}</p>}
    <div className="detail-groups">
      <section><h4>Asked</h4><dl>
        <Field label="Application">{application ? <>{application} <small className="muted">{r.origin_id}</small></> : <Code v={r.origin_id || 'Unknown'}/>}</Field>
        <Field label="Name sent"><Code v={r.requested_model}/></Field>
        <Field label="Module"><Code v={r.module}/></Field>
        <Field label="Effort">{r.effort ? <>{r.effort} <small className="muted">{effortSources[r.effort_source || ''] || r.effort_source}</small></> : <span className="muted">—</span>}</Field>
      </dl></section>
      <section><h4>Served</h4><dl>
        <Field label="Connection"><Code v={r.target_id}/></Field>
        <Field label="Model"><Code v={r.resolved_model}/></Field>
        <Field label="Provider"><Code v={r.provider}/></Field>
        <Field label="Route"><Code v={(r.method ? r.method + ' ' : '') + r.route}/></Field>
      </dl></section>
      <section><h4>Result</h4><dl>
        <Field label="Outcome"><span className={'badge ' + (r.outcome === 'success' ? 'good' : 'bad')}>{r.outcome.replaceAll('_', ' ')}</span> <small className="muted">HTTP {r.http_status}</small></Field>
        <Field label="Latency">{fmt(r.latency_ms, ' ms')}{r.ttfb_ms != null && <small className="muted"> · first byte {fmt(r.ttfb_ms, ' ms')}</small>}</Field>
        <Field label="Vote">{vote === 'up' ? <span className="vote-chip up"><ThumbsUp size={13} aria-hidden/> up</span> : vote === 'down' ? <span className="vote-chip down"><ThumbsDown size={13} aria-hidden/> down</span> : vote ? 'neutral' : <span className="muted">not rated</span>}</Field>
        {r.compression && <Field label="Compression">{r.compression} <small className="muted">{r.compression_status}</small></Field>}
      </dl></section>
      <section><h4>Request</h4><dl>
        <Field label="Id"><Code v={r.id}/></Field>
        <Field label="Time">{at.toLocaleString()} <small className="muted">{at.toISOString()}</small></Field>
      </dl></section>
    </div>
    {failedCall(r) && <ReportToAgent call={r} application={application} onPage={onPage}/>}
    <p className="caption">Prompts, tool arguments and responses are never recorded: only this metadata.</p>
  </div>;
}

type Reported = {id: string; application_fix: string; evidence: {count: number}};

function ReportToAgent({call: r, application, onPage}: {call: Call; application?: string; onPage: (page: string) => void}) {
  const [note, setNote] = useState(''), [busy, setBusy] = useState(false), [error, setError] = useState(''), [done, setDone] = useState<Reported | null>(null), [copied, setCopied] = useState(false);
  const app = application || r.origin_id;
  if (!application) return <div className="report-agent muted-small">This request came from no Application, so there is no coding agent to report it to.</div>;
  const report = async () => {
    setBusy(true); setError('');
    try {
      const res = await fetch('/api/v1/issues', {method: 'POST', credentials: 'same-origin', headers: {'Content-Type': 'application/json', 'X-Aiport-CSRF': '1'}, body: JSON.stringify({request_id: r.id, note})});
      const body = await res.json().catch(() => null);
      if (!res.ok) throw new Error(body?.error?.message || 'HTTP ' + res.status);
      setDone(body.issue);
    } catch (e) { setError(String((e as Error).message || e)); }
    finally { setBusy(false); }
  };
  const openIssue = () => { if (!done) return; history.replaceState(null, '', '#' + issueAnchor(done.id)); onPage('Changes'); };
  const copy = async () => { if (!done) return; try { await navigator.clipboard.writeText(done.application_fix); setCopied(true); setTimeout(() => setCopied(false), 1600); } catch { setError('The browser did not allow copying.'); } };
  return <section className="report-agent" aria-label={'Report to ' + app + "'s coding agent"}>
    <h4>Hand the fix to {app}’s coding agent</h4>
    {done ? <div className="report-done" role="status">
      <p><b>Reported as Issue <code>{done.id}</code>{done.evidence.count > 1 ? ' (' + done.evidence.count + ' requests of this kind)' : ''}.</b> {app}’s coding agent reads it at the start of its next session (<code>aiport_my_issues</code>), and its next failing call of this kind names it in <code>X-Aiport-Hint</code>. To have aiport run a coding agent on it now, use <b>Fix it</b> on the Issue.</p>
      <div className="report-actions"><button type="button" className="secondary" onClick={openIssue}>Open the Issue in Changes <ArrowUpRight size={14} aria-hidden/></button><button type="button" className="text-button" onClick={copy}><Copy size={13} aria-hidden/> {copied ? 'Copied' : 'Copy the fix for a coding agent'}</button></div>
    </div> : <>
      <p className="muted-small">aiport writes what failed, what it knows about it and what to do as an Issue for {app}. It keeps no provider messages, so paste the error the application showed if you have it.</p>
      <label className="report-note-field"><span>Note for the coding agent <small className="muted">(optional)</small></span>
        <textarea value={note} maxLength={MAX_NOTE} rows={3} placeholder={'For example: ' + (r.http_status === 400 ? 'Meta answered 400: … Missing \'visualLayout\'.' : 'what the user did and the error shown')} onChange={e => setNote(e.target.value)}/>
        <small className="muted">{note.length} / {MAX_NOTE}</small>
      </label>
      {error && <div className="notice error" role="alert">{error}</div>}
      <button type="button" className="primary report-send" disabled={busy} onClick={report}><span>{busy ? 'Reporting…' : 'Report to ' + app + '’s coding agent'}</span><Send size={15} aria-hidden/></button>
    </>}
  </section>;
}
