import React, {useEffect, useState} from 'react';
import {ArrowUpRight, CircleAlert, History, Lightbulb, Puzzle, Sparkles, Undo2} from 'lucide-react';
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
  status?: 'pending' | 'accepted' | 'rejected' | 'dismissed' | 'withdrawn' | 'not_supported';
  /** The Owner's decision, on decided items. */
  decision?: {action: string; note?: string; scope?: string; by: string; at_ms: number; change_id?: string};
};

/** An Issue: a problem aiport keeps seeing for one Application, as GET /api/v1/issues lists it. */
export type Issue = {
  id: string; application: string; cause: string; subject: string;
  state: 'open' | 'resolving' | 'resolved' | 'dismissed';
  summary: string; needs_owner: boolean; application_fix: string;
  evidence: {count: number; from_ms: number; to_ms: number; window_minutes: number; status?: number; error?: string; examples?: string[]; connections?: string[]};
  /** summary is the server's own words for what aiport found (decide.Diagnosis.Summary). */
  diagnosis?: {protocol: string; answer: 'one' | 'none' | 'several' | 'unknown'; serving: string[]; unknown: string[]; available?: string[]; summary?: string};
  fix?: {suggestion: string; kind: string; scope: string; proposal: {alias?: string; target?: string; model?: string}; state: 'proposed' | 'applied' | 'rejected' | 'reverted'; change_id?: string; by?: string};
  resolving_by?: string; resolution?: {how: string; at_ms: number; change_id?: string}; reopened?: number;
  /** The saved root cause (decide.Explanation); also at the start of application_fix. */
  explanation?: Explanation;
  /** The latest Fix it run's branch and diff summary (decide.BranchFix). */
  branch_fix?: BranchFix;
};

/** What a Fix it run left: its branch, diff summary and, once kept, the commit aiport made on that branch. */
export type BranchFix = {branch: string; base: string; worktree?: string; agent: string; run: string;
  state: 'running' | 'ready' | 'stopped_waiting' | 'timed_out' | 'empty' | 'kept' | 'discarded' | 'stopped'; summary?: string; files?: string[]; commit?: string; note?: string; started_at_ms: number; at_ms: number};
/** Whether Fix it is available for an Application (fix_by in GET /api/v1/issues): which agent and for which Issue causes, or a hint why not. */
export type FixStatus = {available: boolean; hint?: string; agent?: string; agent_name?: string; model_calls?: ModelCalls; limit_ms?: number; causes?: string[]; cause_hint?: string};
/** Branch fix states whose diff waits in its worktree for Keep or Discard. */
const REVIEWABLE = ['ready', 'stopped_waiting', 'timed_out'];

/** An Issue's explanation and who wrote it: a Connection and model the Owner chose, or a Coding agent. */
export type Explanation = {text: string; source: 'connection' | 'coding_agent'; connection?: string; model?: string; agent?: string; at_ms?: number};

/** Whether Explain is available (the "explain" field of GET /api/v1/issues): who writes, or a hint why not. */
export type ExplainStatus = {available: boolean; hint?: string; source?: string; connection?: string; model?: string; agent?: string; agent_name?: string; model_calls?: ModelCalls; limit_ms?: number;
  /** Per Application with a connected Coding agent: the agent explains its Issues instead (explain_by). */
  by?: Record<string, ExplainStatus>;
  /** Per Application: whether Fix it is available (fix_by). */
  fix?: Record<string, FixStatus>};
/** Where an Agent run's model calls go. */
export type ModelCalls = {through_aiport: boolean; connection?: string; model?: string; note?: string};
/** An Agent run, as the runner describes it ("run" event, GET /api/v1/runs). */
export type AgentRun = {id: string; application: string; issue: string; mode: string; agent: string; agent_name: string; model_calls: ModelCalls; state: string; error?: string; started_at_ms: number; limit_ms: number;
  /** The agent's session mode aiport chose, when the agent offers modes. */
  session_mode?: string;
  /** Agent settings in the run's directory that allow without asking (aiport cannot see what they allow). */
  warnings?: string[]};
/** A command a Fix it run asks to run, waiting for the Owner (GET /api/v1/approvals). The command is the agent's text: plain text only. */
export type Approval = {id: string; run: string; application: string; issue: string; command: string; title: string; asked_at_ms: number; expires_at_ms: number};

/** A waiting command with the Owner's three answers; the agent runs it itself, in its worktree, once allowed. */
function ApprovalCard({approval, onAnswered}: {approval: Approval; onAnswered: () => void}) {
  const [busy, setBusy] = useState(''), [error, setError] = useState('');
  const answer = async (a: 'allow_once' | 'allow_run' | 'deny') => {
    setBusy(a); setError('');
    try { await send('POST', 'approvals/' + encodeURIComponent(approval.id), {answer: a}); onAnswered(); }
    catch (e) { setError((e as Error).message); }
    finally { setBusy(''); }
  };
  const minutes = Math.max(0, Math.round((approval.expires_at_ms - approval.asked_at_ms) / 60000));
  return <div className="approval" role="group" aria-label="Command waiting for your answer">
    <p><strong>The agent asks to run a command</strong> in its worktree (the whole command; invisible characters are shown as escapes such as {'\\u202e'}):</p>
    <pre className="approval-command" tabIndex={0} aria-label="The whole command">{approval.command}</pre>
    <div className="actions">
      <button type="button" className="primary small" disabled={!!busy} onClick={() => answer('allow_once')}>Allow once</button>
      <button type="button" className="secondary small" disabled={!!busy} onClick={() => answer('allow_run')} title="Allow this exact command again in this run, until the agent changes a file">Allow until the next edit</button>
      <button type="button" className="secondary small" disabled={!!busy} onClick={() => answer('deny')}>Deny</button>
      <span className="caption">Unanswered for {minutes} min, the run stops and keeps what it changed for your review.</span>
    </div>
    {error && <p className="auth-error" role="alert">{error}</p>}
  </div>;
}

/** One step of a run besides its words: a thought, a tool call, a plan, a permission aiport answered. */
type RunStep = {kind: string; text: string; tool?: string; status?: string; verdict?: string; reason?: string};

/** Explain status from GET /api/v1/issues: explain (the chosen Connection) with explain_by (Coding agents) as by, and fix_by as fix. */
export const explainOf = (d: {explain?: ExplainStatus; explain_by?: Record<string, ExplainStatus>; fix_by?: Record<string, FixStatus>}): ExplainStatus | null =>
  d.explain ? {...d.explain, by: d.explain_by, fix: d.fix_by} : null;

const AGENTS: Record<string, string> = {claude: 'Claude Code', codex: 'Codex', gemini: 'Gemini CLI'};
const writer = (e: {source?: string; connection?: string; model?: string; agent?: string; agent_name?: string}) =>
  e.source === 'coding_agent' ? (e.agent_name || AGENTS[e.agent || ''] || e.agent) + ', after reading the code' : e.connection + ' (' + e.model + ')';

const callsLine = (c?: ModelCalls) => !c ? '' : c.through_aiport
  ? 'Model calls go through aiport to ' + (c.connection || 'the Anthropic protocol default') + (c.model ? ' (' + c.model + ')' : '') + ', under the Application aiport.'
  : c.note || 'The agent uses its own login.';

/** The run panel: which agent reads the code (read-only), where its calls go, its steps, and Stop. */
function RunPanel({run, steps, busy, branch}: {run: AgentRun; steps: RunStep[]; busy: boolean; branch?: string}) {
  const [stopping, setStopping] = useState(false), [error, setError] = useState('');
  const stop = async () => {
    setStopping(true); setError('');
    const r = await fetch('/api/v1/runs/' + encodeURIComponent(run.id) + '/stop', {method: 'POST', credentials: 'same-origin', headers: {'X-Aiport-CSRF': '1'}}).catch(() => null);
    if (!r || (!r.ok && r.status !== 404)) { setError('Could not stop the run.'); setStopping(false); }
  };
  const shown = steps.slice(-30);
  return <div className="run-panel" aria-label={'Agent run of ' + run.agent_name}>
    <div className="run-head">
      <strong>{run.agent_name}</strong> <span className="caption">{run.mode === 'fix' ? 'changes the code in its own worktree' + (branch ? ' on ' + branch : '') + ' · commands wait for your approval' :'reads the code, read-only'} · stops after {Math.round(run.limit_ms / 60000)} min</span>
      {busy && <button type="button" className="secondary small" disabled={stopping} onClick={stop}>{stopping ? 'Stopping…' : 'Stop'}</button>}
    </div>
    <p className={'caption' + (run.model_calls.through_aiport ? '' : ' run-own-login')}>{callsLine(run.model_calls)}</p>
    {run.session_mode && <p className="caption run-mode">Session mode: {run.session_mode}</p>}
    {!!run.warnings?.length && <div className="notice-inline" role="note">
      <strong>The repository's agent settings allow some things without asking.</strong> aiport never sees those, so it cannot ask you about them:
      <ul className="run-warnings">{run.warnings.map((w, i) => <li key={i}>{w}</li>)}</ul>
    </div>}
    {shown.length > 0 && <ol className="run-steps" aria-live="polite">
      {shown.map((s, i) => <li key={i} className={'run-step run-' + s.kind + (s.verdict ? ' run-' + s.verdict : '')}>
        {s.kind === 'permission' ? <>{s.verdict === 'allowed' ? 'Allowed' : 'Denied'}: {s.text}{s.reason ? <small> · {s.reason}</small> : null}</>
          : s.kind === 'tool' ? <>{s.text}{s.status && s.status !== 'completed' ? <small> · {s.status}</small> : null}</>
          : s.kind === 'thought' ? <em>{s.text}</em> : s.text}
      </li>)}
    </ol>}
    {error && <p className="auth-error" role="alert">{error}</p>}
  </div>;
}

/**
 * Explain: the root cause of an Issue in words. POST /api/v1/issues/{id}/explain streams JSON lines (start, text…,
 * then done or error); the text is shown as it arrives and, once saved, comes back with the Issue (and in its
 * Application fix). Explain again replaces it. The text comes from a model: shown as plain text only.
 */
export function IssueExplanation({issue, explain: all, onExplained}: {issue: Issue; explain: ExplainStatus | null; onExplained: () => void}) {
  const explain = all?.by?.[issue.application] || all;
  const [text, setText] = useState(''), [busy, setBusy] = useState(false), [error, setError] = useState(''), [by, setBy] = useState('');
  const [agentRun, setAgentRun] = useState<AgentRun | null>(null), [steps, setSteps] = useState<RunStep[]>([]);
  // An Explain Agent run goes on without this page: follow it again from its log (GET /api/v1/issues/{id}/run)
  // when the page opens while it runs, or when the stream broke off; it saves its explanation itself.
  const follow = async (alive: () => boolean) => {
    let following = false;
    while (alive()) {
      const d = await fetch('/api/v1/issues/' + encodeURIComponent(issue.id) + '/run', {credentials: 'same-origin'}).then(r => r.ok ? r.json() : null).catch(() => null);
      if (!alive() || !d?.run || d.run.mode !== 'explain') break;
      if (!following && d.run.state !== 'running') break;
      following = true;
      const log: (RunStep & {kind: string})[] = d.log || [];
      setBusy(true); setAgentRun(d.run); setBy(writer({source: 'coding_agent', agent: d.run.agent, agent_name: d.run.agent_name}));
      setText(log.filter(u => u.kind === 'text').map(u => u.text).join(''));
      setSteps(log.filter(u => u.kind !== 'text' && u.kind !== 'status'));
      if (d.run.state !== 'running') {
        setBusy(false);
        if (d.run.state === 'done') onExplained(); else setError('The Agent run ended (' + d.run.state + ')' + (d.run.error ? ': ' + d.run.error : '') + '; nothing was saved.');
        return;
      }
      await new Promise(r => setTimeout(r, 1500));
    }
    if (following) setBusy(false);
  };
  useEffect(() => {
    if (explain?.source !== 'coding_agent') return;
    let alive = true;
    follow(() => alive);
    return () => { alive = false; };
  }, [issue.id, explain?.source]);
  const run = async () => {
    setBusy(true); setError(''); setText(''); setBy(''); setAgentRun(null); setSteps([]);
    try {
      const r = await fetch('/api/v1/issues/' + encodeURIComponent(issue.id) + '/explain', {method: 'POST', credentials: 'same-origin', headers: {'X-Aiport-CSRF': '1'}});
      if (!r.ok || !r.body) { const d = await r.json().catch(() => ({})); throw new Error(d.error?.message || 'Explain failed (' + r.status + ').'); }
      const reader = r.body.pipeThrough(new TextDecoderStream()).getReader();
      let buffer = '', done = false, failed = '';
      while (!done) {
        const chunk = await reader.read();
        if (chunk.done) break;
        buffer += chunk.value;
        let nl;
        while ((nl = buffer.indexOf('\n')) >= 0) {
          const line = buffer.slice(0, nl).trim(); buffer = buffer.slice(nl + 1);
          if (!line) continue;
          const ev = JSON.parse(line);
          if (ev.type === 'start') setBy(writer({...ev.by, agent_name: explain?.agent_name}));
          else if (ev.type === 'run') setAgentRun(ev.run);
          else if (ev.type === 'step') setSteps(s => [...s, ev.step]);
          else if (ev.type === 'text') setText(t => t + ev.text);
          else if (ev.type === 'done') done = true;
          else if (ev.type === 'error') failed = ev.error?.message || 'Explain failed.';
        }
      }
      if (failed) throw new Error(failed);
      if (!done && explain?.source === 'coding_agent') { setBusy(false); await follow(() => true); return; }
      if (!done) throw new Error('The explanation broke off; nothing was saved.');
      onExplained();
    } catch (e) {
      setError((e as Error).message);
      if (e instanceof TypeError && explain?.source === 'coding_agent') { setError(''); await follow(() => true); } // the connection broke: the run goes on
    } finally {
      setBusy(false);
    }
  };
  const saved = issue.explanation;
  const shown = busy || (text && error) ? text : saved?.text || '';
  return <div className="explanation">
    {shown && <div className="explanation-text" aria-live="polite"><strong>Root cause</strong>{(busy ? by : saved && writer(saved)) && <small> · written by {busy ? by : writer(saved!)}{!busy && saved?.source === 'connection' ? ' from the evidence only' : ''}</small>}<p>{shown}</p></div>}
    {agentRun && <RunPanel run={agentRun} steps={steps} busy={busy}/>}
    {busy && !text && !agentRun && <p className="caption" role="status">Asking {by || 'the model'}…</p>}
    {error && <div className="notice error" role="alert">{error}{saved ? ' The earlier explanation stays.' : ''}</div>}
    <div className="actions">
      <button type="button" className="secondary" disabled={busy || !explain?.available} onClick={run} title={explain?.available ? 'Explain the root cause with ' + writer(explain) : explain?.hint}>
        <Sparkles size={14} aria-hidden/> {busy ? 'Explaining…' : saved ? 'Explain again' : 'Explain'}
      </button>
      {explain && !explain.available && explain.hint && <span className="caption">{explain.hint}</span>}
    </div>
  </div>;
}

/** A diff shown as plain text, one line each, added and removed lines marked. The text is the agent's code: never markup. */
function DiffView({diff, truncated}: {diff: string; truncated: boolean}) {
  const lines = diff.split('\n');
  return <pre className="diff" aria-label="Diff">{lines.map((l, i) => <span key={i} className={l.startsWith('+++') || l.startsWith('---') ? 'diff-file' : l.startsWith('+') ? 'diff-add' : l.startsWith('-') ? 'diff-del' : l.startsWith('@@') ? 'diff-hunk' : undefined}>{l + '\n'}</span>)}
    {truncated && <span className="diff-hunk">… the diff is longer than aiport shows; read the rest on the branch.</span>}</pre>;
}

/**
 * Fix it: the Application's Coding agent changes the code in its own git worktree, on a new branch named for the
 * Issue, and ends with a diff the Owner reviews. Keep commits it on that branch (as aiport) and leaves the branch;
 * Discard removes the worktree and the branch. aiport never commits to the Owner's branch, merges or pushes.
 * Without a Coding agent or repository path, only a one-line hint shows.
 */
export function IssueFix({issue, explain, onChanged}: {issue: Issue; explain: ExplainStatus | null; onChanged: () => void}) {
  const status = explain?.fix?.[issue.application], b = issue.branch_fix;
  const [busy, setBusy] = useState(''), [error, setError] = useState('');
  const [run, setRun] = useState<AgentRun | null>(null), [steps, setSteps] = useState<RunStep[]>([]);
  const [diff, setDiff] = useState<{diff: string; truncated: boolean} | null>(null);
  const [approvals, setApprovals] = useState<Approval[]>([]);
  const running = b?.state === 'running', ready = !!b && REVIEWABLE.includes(b.state);
  // While the run goes, follow its log; once it has ended, reload the Issue for its diff.
  useEffect(() => {
    if (!running) return;
    let stop = false;
    const follow = async () => {
      while (!stop) {
        const d = await fetch('/api/v1/issues/' + encodeURIComponent(issue.id) + '/run', {credentials: 'same-origin'}).then(r => r.ok ? r.json() : null).catch(() => null);
        if (stop) return;
        if (d?.run) { setRun(d.run); setSteps((d.log || []).filter((u: RunStep) => u.kind !== 'status')); setApprovals(d.approvals || []); }
        if (d?.run && d.run.state !== 'running') { onChanged(); }
        await new Promise(r => setTimeout(r, 1500));
      }
    };
    follow();
    return () => { stop = true; setApprovals([]); };
  }, [issue.id, running]);
  useEffect(() => {
    setDiff(null);
    if (!b || (!ready && b.state !== 'kept')) return;
    fetch('/api/v1/issues/' + encodeURIComponent(issue.id) + '/fix', {credentials: 'same-origin'}).then(r => r.ok ? r.json() : null)
      .then(d => d && setDiff({diff: d.diff || '', truncated: !!d.truncated})).catch(() => {});
  }, [issue.id, b?.state, b?.run]);
  const act = async (what: 'fix' | 'fix/keep' | 'fix/discard') => {
    setBusy(what); setError('');
    try {
      await send('POST', 'issues/' + encodeURIComponent(issue.id) + '/' + what);
      setRun(null); setSteps([]);
      onChanged();
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setBusy('');
    }
  };
  // Fix it changes the Application's code: only for causes that can be there.
  const forCause = !status?.causes || status.causes.includes(issue.cause);
  const canStart = status?.available && forCause && issue.state === 'open' && !running && !ready;
  if (!b && !status?.available) return status?.hint ? <p className="caption fix-hint">{status.hint}</p> : null;
  if (!b && !forCause) return status?.cause_hint ? <p className="caption fix-hint">{status.cause_hint}</p> : null;
  return <div className="branch-fix">
    {running && run && <RunPanel run={run} steps={steps} busy branch={b?.branch}/>}
    {running && approvals.map(a => <ApprovalCard key={a.id} approval={a} onAnswered={() => { setApprovals(l => l.filter(x => x.id !== a.id)); onChanged(); }}/>)}
    {running && !run && <p className="caption" role="status">Starting {status?.agent_name || b?.agent} in a worktree on {b?.branch}…</p>}
    {ready && <div className="fix-review">
      {b!.state !== 'ready' && b!.note && <p className="caption">{b!.note}</p>}
      <p><strong>Review the change</strong> on branch <code>{b!.branch}</code>: {b!.summary}. Nothing is committed yet; your branch and working copy are untouched.</p>
      {diff && <DiffView diff={diff.diff} truncated={diff.truncated}/>}
      <div className="actions">
        <button type="button" className="primary" disabled={!!busy} onClick={() => act('fix/keep')}>{busy === 'fix/keep' ? 'Keeping…' : 'Keep the branch'}</button>
        <button type="button" className="secondary" disabled={!!busy} onClick={() => act('fix/discard')}>{busy === 'fix/discard' ? 'Discarding…' : 'Discard'}</button>
        <span className="caption">Keep commits it on {b!.branch} as aiport; merging is yours. Discard deletes the worktree and the branch.</span>
      </div>
    </div>}
    {b?.state === 'kept' && <p className="caption">Kept on branch <code>{b.branch}</code>{b.commit ? ' (commit ' + b.commit.slice(0, 12) + ')' : ''}: {b.summary}. aiport never merges or pushes: merge it when you are ready; the Issue resolves when {issue.application}’s calls succeed.</p>}
    {b && (b.state === 'discarded' || b.state === 'stopped' || b.state === 'empty') && <p className="caption">Last Fix it run: {b.state === 'discarded' ? 'discarded; the worktree and the branch were removed.' : b.note}</p>}
    {error && <div className="notice error" role="alert">{error}</div>}
    {canStart && <div className="actions">
      <button type="button" className="secondary" disabled={!!busy} onClick={() => act('fix')} title={'Fix it with ' + (status?.agent_name || status?.agent) + ' in its own worktree; stops after ' + Math.round((status?.limit_ms || 0) / 60000) + ' min'}>
        {busy === 'fix' ? 'Starting…' : b ? 'Fix it again' : 'Fix it'}
      </button>
      <span className="caption">{status?.agent_name} edits a new branch in its own worktree; you approve each command it runs; you review the diff.</span>
    </div>}
  </div>;
}

/** Who makes a Change aiport applies itself (adapting.auto_fix). */
const AUTOMATIC = 'aiport (automatic)';

/**
 * One open or resolving Issue: what is wrong, the evidence, which Connections serve the model, the aiport fix
 * (an Alias for this Application only, or for a failing Connection binding this Application to a working one)
 * with Approve / Dismiss, and the fix in the Application for its Coding agent. Approve stays while the Application
 * says it fixed the Issue (resolving); Dismiss is there for any Issue still open or resolving.
 */
function IssueCard({issue, targets, explain, onDecided, onExplained}: {issue: Issue; targets: string[]; explain: ExplainStatus | null; onDecided: (message: string) => void; onExplained: () => void}) {
  const fix = issue.fix, ev = issue.evidence, binds = issue.cause === 'failing_connection';
  const [target, setTarget] = useState(fix?.proposal.target || ''), [model, setModel] = useState(fix?.proposal.model || issue.subject);
  const [busy, setBusy] = useState(''), [error, setError] = useState(''), [blocked, setBlocked] = useState<Blocked | null>(null);
  const automatic = fix?.state === 'applied' && fix.by === AUTOMATIC && !!fix.change_id;
  const revert = async () => {
    setBusy('revert'); setError(''); setBlocked(null);
    try {
      await send('POST', 'changes/' + encodeURIComponent(fix!.change_id!) + '/revert');
      onDecided('Reverted the automatic fix. aiport will not apply it automatically again for ' + issue.application + '; the Issue waits for your decision.');
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setBusy('');
    }
  };
  const decide = async (action: 'accept' | 'dismiss') => {
    setBusy(action); setError(''); setBlocked(null);
    try {
      await send('POST', 'issues/' + encodeURIComponent(issue.id), action !== 'accept' ? {action} : binds ? {action, target} : {action, target, model});
      onDecided(action === 'accept'
        ? 'Approved for only ' + issue.application + (binds ? ': bound to ' + target : ': the alias “' + issue.subject + '” → ' + target + '/' + model) + '. The Issue resolves itself when its calls succeed; you can revert the Change below.'
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
    <div className="decision-head">
      <h3 className="need">{issue.summary}</h3>
      <div className="decision-meta"><span className={'badge' + (issue.state === 'resolving' ? ' good' : ' warn')}>{issue.state === 'resolving' ? 'Resolving' : 'Open'}</span><span className="badge">{issue.application}</span></div>
    </div>
    <div className="evidence">Evidence: {times(ev.count)} {ev.status}{ev.error ? ' ' + ev.error : ''} between {when(ev.from_ms)} and {when(ev.to_ms)} (looking back {ev.window_minutes} minutes){ev.connections?.length ? ', sent to ' + ev.connections.join(', ') : ''}.{ev.examples?.length ? ' Example requests: ' + ev.examples.slice(0, 3).join(', ') + '.' : ''}{issue.reopened ? ' Came back ' + times(issue.reopened) + '.' : ''}</div>
    {issue.diagnosis?.summary && <div className="evidence">Diagnosis: {issue.diagnosis.summary}</div>}
    {automatic && <div className="evidence"><strong>aiport fixed this automatically:</strong> the alias “{fix!.proposal.alias}” → {fix!.proposal.target}/{fix!.proposal.model} for only {issue.application}. <button className="text-button" disabled={!!busy} onClick={revert} aria-label={'Revert the automatic fix for ' + issue.application}><Undo2 size={14} aria-hidden/> {busy === 'revert' ? 'Reverting…' : 'Revert'}</button></div>}
    {issue.state === 'resolving' && !automatic && <div className="evidence">{issue.resolving_by === 'aiport_fix' ? 'The aiport fix is applied' + (fix?.by ? ' (by ' + fix.by + ')' : '') : 'The Application says it fixed this'}: aiport marks the Issue resolved when {issue.application}’s calls {binds ? 'succeed again' : 'for ' + issue.subject + ' succeed'}.</div>}
    {error && <div className="notice error" role="alert">{error}</div>}
    {blocked && <div className="notice error" role="alert"><strong>Blocked by the guardrail {blocked.extension}:</strong> {blocked.reason} Nothing was changed; this Issue stays open.</div>}
    {fix?.state === 'proposed' && <form className="decision-form" onSubmit={e => { e.preventDefault(); decide('accept'); }}>
      {binds
        ? <label>Fix in aiport: bind only {issue.application} to<select value={target} onChange={e => setTarget(e.target.value)}>
            {!target && <option value="">Choose a working Connection</option>}
            {targets.filter(t => t !== issue.subject).map(t => <option key={t}>{t}</option>)}
          </select></label>
        : <>
          <label>Fix in aiport: alias for only {issue.application}<input value={issue.subject} readOnly spellCheck={false}/></label>
          <label>Connection<select value={target} onChange={e => setTarget(e.target.value)}>
            {!target && <option value="">Choose a Connection</option>}
            {targets.map(t => <option key={t}>{t}</option>)}
          </select></label>
          <label>Model to use<input value={model} onChange={e => setModel(e.target.value)} required maxLength={256} spellCheck={false} autoComplete="off"/></label>
        </>}
      <div className="actions">
        <button className="primary" disabled={!!busy || !target || (!binds && !model.trim())}>{busy === 'accept' ? 'Approving…' : 'Approve'}</button>
        <button type="button" className="secondary" disabled={!!busy} onClick={() => decide('dismiss')}>{busy === 'dismiss' ? 'Dismissing…' : 'Dismiss'}</button>
      </div>
    </form>}
    {fix?.state !== 'proposed' && <div className="actions">
      <button type="button" className="secondary" disabled={!!busy} onClick={() => decide('dismiss')}>{busy === 'dismiss' ? 'Dismissing…' : 'Dismiss'}</button>
    </div>}
    <IssueExplanation issue={issue} explain={explain} onExplained={onExplained}/>
    <IssueFix issue={issue} explain={explain} onChanged={onExplained}/>
    <ApplicationFix issue={issue}/>
  </div>;
}

/** The id of an Issue's card on the Changes page, so Integrate can link to it. */
export const issueAnchor = (id: string) => 'issue-' + id;

/**
 * The fix in the Application, for its Coding agent: "Copy fix for the Coding agent" copies the text the server
 * writes (decide.ApplicationFix, the same text aiport_my_issues gives the agent), shown below it to read first.
 */
export function ApplicationFix({issue}: {issue: Issue}) {
  if (!issue.application_fix) return null;
  return <div className="application-fix">
    <CopyButton text={issue.application_fix} label="Copy fix for the Coding agent"/>
    <details><summary>Read the fix for {issue.application}’s Coding agent</summary><pre>{issue.application_fix}</pre></details>
  </div>;
}

/** One Application that asked for a Change for every Application, with its own argument. */
type Asker = {application: string; why: string; scope_case: string; filed_by?: string; repeats?: number; asked_at_ms: number};

/** Demand for a kind of Change aiport cannot apply yet: anonymous counts only. */
export type Demand = {kind: string; requests: number; applications: number; last_at_ms: number};

/** The first sentence of an Application's request text: what it asks for, shown above the form; the rest is under Details. */
const firstSentence = (text: string) => { const m = text.match(/^.*?[.!?](?=\s|$)/s); return (m ? m[0] : text).trim(); };

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
  const supported = item.kind === 'application_alias' || item.kind === 'bind_application' || !!item.informational;
  return <div className="decision">
    <div className="decision-head">
      <h3 className="need">{item.need}</h3>
      <div className="decision-meta">
        <span className="badge">{request ? (merged ? 'Asked by ' + askers.length + ' Applications' : 'Request from ' + item.application) : 'Suggestion'}</span>
        <span className="badge">For {scopeLabel(item.scope, item.application)}</span>
        {!supported && <span className="badge warn">Not supported yet</span>}
      </div>
    </div>
    {ev && item.kind === 'pair_application' && <div className="evidence">Evidence: {ev.count} refused requests between {when(ev.from_ms)} and {when(ev.to_ms)} (looking back {ev.window_minutes} minutes). Only counts are kept, never the token.</div>}
    {ev && item.kind !== 'pair_application' && <div className="evidence">Evidence: {ev.count} requests between {when(ev.from_ms)} and {when(ev.to_ms)} (looking back {ev.window_minutes} minutes). Only request metadata was used.</div>}
    {request && item.details && <p className="decision-ask"><span>Asks for</span> {firstSentence(item.details)}</p>}
    {request && <details className="decision-more"><summary>Details and reasons</summary>
      {!merged && <dl className="request-text">
        <dt>Details</dt><dd>{item.details}</dd>
        <dt>Why</dt><dd>{item.why}</dd>
        <dt>Why {scopeLabel(item.scope, item.application)}</dt><dd>{item.scope_case}</dd>
        {item.filed_by && <><dt>Filed by (as stated)</dt><dd>{item.filed_by}</dd></>}
        {!!item.repeats && <><dt>Asked again</dt><dd>{item.repeats} more {item.repeats === 1 ? 'time' : 'times'}</dd></>}
      </dl>}
      {merged && <>
        <dl className="request-text"><dt>Details</dt><dd>{item.details}</dd></dl>
        <ul className="askers" aria-label="Each Application's argument">{askers.map(a => <li key={a.application}>
          <strong>{a.application}</strong>{a.filed_by ? ' (filed by ' + a.filed_by + ', as stated)' : ''}{a.repeats ? ' · asked again ' + times(a.repeats) : ''}
          <dl className="request-text"><dt>Why</dt><dd>{a.why}</dd><dt>Argument for every Application</dt><dd>{a.scope_case}</dd></dl>
        </li>)}</ul>
      </>}
    </details>}
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

/** A decided Suggestion or Change request: what it asked, the outcome, who decided, when, and the note. Plain text only. */
function Decided({item}: {item: Decision}) {
  const request = item.source === 'request', d = item.decision;
  const outcome = item.status === 'accepted' ? 'Accepted' : item.status === 'dismissed' ? 'Dismissed' : 'Rejected';
  return <div className="decision decided">
    <div className="decision-head">
      <h3 className="need">{item.need}</h3>
      <div className="decision-meta">
        <span className={'badge' + (item.status === 'accepted' ? ' good' : '')}>{outcome}</span>
        <span className="badge">{request ? 'Request from ' + item.application : 'Suggestion'}</span>
        <span className="badge">For {scopeLabel(d?.scope || item.scope, item.application)}</span>
      </div>
    </div>
    {request && item.details && <p className="decision-ask"><span>Asked for</span> {firstSentence(item.details)}</p>}
    {d && <p className="evidence">{outcome} by {d.by} on {when(d.at_ms)}{d.note ? ': “' + d.note + '”' : '.'}{item.status === 'accepted' ? ' The Change is in Applied changes below, where you can revert it.' : ''}</p>}
    {request && <details className="decision-more"><summary>Details and reasons</summary>
      <dl className="request-text">
        {item.details && <><dt>Details</dt><dd>{item.details}</dd></>}
        {item.why && <><dt>Why</dt><dd>{item.why}</dd></>}
        {item.scope_case && <><dt>Why {scopeLabel(item.scope, item.application)}</dt><dd>{item.scope_case}</dd></>}
        {item.filed_by && <><dt>Filed by (as stated)</dt><dd>{item.filed_by}</dd></>}
      </dl>
    </details>}
  </div>;
}

/** Changes: Suggestions waiting for a decision (only when there are any), then every configuration edit made through aiport, newest first, with Revert. */
export default function Changes({tick, onChange, targets = [], onOpenApplications}: {tick: number; onChange: () => void; targets?: string[]; onOpenApplications?: () => void}) {
  const [pending, setPending] = useState<Decision[]>([]), [demand, setDemand] = useState<Demand[]>([]), [issues, setIssues] = useState<Issue[]>([]);
  const [changes, setChanges] = useState<Change[] | null>(null), [revertable, setRevertable] = useState(false);
  const [busy, setBusy] = useState(''), [error, setError] = useState(''), [message, setMessage] = useState('');
  const [reload, setReload] = useState(0), [upstream, setUpstream] = useState<Upstream | null>(null);
  const [accepted, setAccepted] = useState<Decision[]>([]), [rejected, setRejected] = useState<Decision[]>([]), [view, setView] = useState<'pending' | 'accepted' | 'rejected'>('pending');
  const [autoFix, setAutoFix] = useState<boolean | null>(null), [adapting, setAdapting] = useState(true), [explain, setExplain] = useState<ExplainStatus | null>(null);
  useEffect(() => {
    send('GET', 'upstream').then(setUpstream).catch(() => setUpstream(null));
    send('GET', 'changes').then(d => { setChanges(d.changes || []); setRevertable(!!d.revertable); }).catch(e => setError(e.message));
    send('GET', 'decisions').then(d => { setPending((d.items || []).filter((i: Decision) => !i.issue)); setDemand(d.demand || []); }).catch(() => { setPending([]); setDemand([]); });
    const own = (d: {items?: Decision[]}) => (d.items || []).filter(i => !i.issue);
    send('GET', 'decisions?status=accepted').then(d => setAccepted(own(d))).catch(() => setAccepted([]));
    send('GET', 'decisions?status=rejected').then(d => setRejected(own(d))).catch(() => setRejected([]));
    send('GET', 'issues').then(d => { setIssues(d.issues || []); setAutoFix(!!d.auto_fix); setAdapting(d.adapting !== false); setExplain(explainOf(d)); }).catch(() => { setIssues([]); setAutoFix(null); });
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
  const switchAutoFix = async (on: boolean) => {
    setBusy('auto-fix'); setError(''); setMessage('');
    try {
      await send('PUT', 'adapting/auto-fix', {enabled: on});
      setMessage(on
        ? 'Automatic fixes are on: aiport applies an Issue’s aiport fix itself when only one Connection serves the model. You can revert each one below.'
        : 'Automatic fixes are off: every aiport fix waits for your decision.');
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
  return <>{!!issues.length && <section className="panel changes" aria-label="Issues">
    <div className="panel-heading"><h2><CircleAlert size={16} aria-hidden/> Issues</h2><span>Problems aiport keeps seeing for one Application · they close themselves when its calls succeed again</span></div>
    {issues.map(i => <IssueCard key={i.id + i.state + (i.fix?.suggestion || '')} issue={i} targets={targets} explain={explain} onDecided={decided} onExplained={() => setReload(x => x + 1)}/>)}
  </section>}{(adapting || pending.length + accepted.length + rejected.length > 0) && <section className="panel changes suggestions" aria-label="Suggestions and Change requests">
    <div className="panel-heading"><h2><Lightbulb size={16} aria-hidden/> Suggestions and Change requests</h2>
      <div className="task-filters decision-filters" role="tablist" aria-label="Show">
        {([['pending', 'Waiting', pending.length], ['accepted', 'Accepted', accepted.length], ['rejected', 'Rejected', rejected.length]] as const).map(([key, label, n]) =>
          <button key={key} type="button" role="tab" aria-selected={view === key} className={view === key ? 'active' : ''} onClick={() => setView(key)}>{label}<span>{n}</span></button>)}
      </div>
    </div>
    {view === 'pending' && (pending.length
      ? pending.map(item => <Pending key={item.id} item={item} targets={targets} onDecided={decided} onOpenApplications={onOpenApplications}/>)
      : <p className="quiet-empty">Nothing waits for your decision. Suggestions from request history and Change requests from Applications appear here.</p>)}
    {view !== 'pending' && ((view === 'accepted' ? accepted : rejected).length
      ? (view === 'accepted' ? accepted : rejected).map(item => <Decided key={item.id} item={item}/>)
      : <p className="quiet-empty">No {view} items. Decided items are kept as long as request history.</p>)}
  </section>}{!!demand.length && <section className="panel changes" aria-label="Demand for things aiport cannot do yet">
    <div className="panel-heading"><h2>Asked for, not supported yet</h2><span>Change requests aiport cannot apply · kept and counted, no decision needed</span></div>
    <div className="table-scroll"><table>
      <thead><tr><th>Kind</th><th>Requests</th><th>Applications</th><th>Last asked</th></tr></thead>
      <tbody>{demand.map(d => <tr key={d.kind}><td>{d.kind}</td><td>{d.requests}</td><td>{d.applications}</td><td>{when(d.last_at_ms)}</td></tr>)}</tbody>
    </table></div>
    {upstream && !!upstream.suggestions.length && <ul className="askers" aria-label="Suggest to aiport">{upstream.suggestions.map(s => <li key={s.kind}>
      <strong>{s.kind}</strong> <span className="caption">kind named by an Application</span> <UpstreamAction label="Suggest to aiport" draft={s} what={{suggest: s.kind}} upstream={upstream}/>
    </li>)}</ul>}
  </section>}{upstream && !!upstream.shares.length && <section className="panel changes" aria-label="Extensions">
    <div className="panel-heading"><h2><Puzzle size={16} aria-hidden/> Extensions</h2><span>Listed in aiport.yaml · share one upstream for others to use</span></div>
    <ul className="askers">{upstream.shares.map(s => <li key={s.extension}>
      <strong>{s.extension}</strong> ({s.hook}) <UpstreamAction label="Share upstream" draft={s} what={{share: s.extension}} upstream={upstream}/>
    </li>)}</ul>
  </section>}<section className="panel changes">
    <div className="panel-heading"><h2>Applied changes</h2><div className="header-actions"><span className="caption">Newest first · kept across restarts</span><DecisionSoundToggle/></div></div>
    {autoFix !== null && <label className="check-label compact" title="adapting.auto_fix in aiport.yaml"><input type="checkbox" checked={autoFix} disabled={!!busy || !adapting} onChange={e => switchAutoFix(e.target.checked)}/>
      <span>Automatic fixes: aiport applies an Issue’s aiport fix itself, only when exactly one Connection serves the model, and only for that one Application. Each one is a Change below that you can revert; a fix you revert is never applied automatically again.{!adapting && ' Adapting is turned off in the configuration, so nothing is applied.'}</span></label>}
    {error && <div className="notice error" role="alert">{error}</div>}
    {message && <div className="notice" role="status">{message}</div>}
    {changes && !changes.length && <div className="empty"><span className="empty-icon"><History size={20} strokeWidth={1.75} aria-hidden/></span><h3>No changes yet</h3><p>Configuration edits made through aiport appear here, with the file before and after kept so you can revert them.</p></div>}
    {!!changes?.length && <div className="table-scroll"><table>
      <thead><tr><th>Time</th><th>Change</th><th>Made by</th><th>Status</th><th></th></tr></thead>
      <tbody>{changes.map(c => {
        const undone = c.reverted_by ? byId.get(c.reverted_by) : undefined;
        return <tr key={c.id}>
          <td title={new Date(c.at_ms).toISOString()}>{when(c.at_ms)}</td>
          <td>{c.by === AUTOMATIC ? 'aiport fixed: ' + c.what : c.what}</td>
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
