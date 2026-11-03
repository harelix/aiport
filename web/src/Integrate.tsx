import React, {useEffect, useRef, useState} from 'react';
import {HelpLink} from './help';
import {AppWindow, ArrowUpRight, Bot, Check, ChevronRight, CircleCheck, CircleDashed, Copy, ExternalLink, KeyRound, Plus, Terminal, Trash2} from 'lucide-react';
import {AddApplicationForm, forgetRejected, type IssuedToken, issueToken, MayRequestToggle, Origin, Pairing, PairingList, Rejected, RemoveApplication, Target, TokenReveal, useMissingTokens, usePairings, useRejected} from './Applications';
import {ApplicationFix, CatalogChange, isCatalogChange, type ExplainStatus, explainOf, type FixStatus, type Issue, IssueExplanation, IssueFix, issueAnchor} from './Changes';
import {Inline} from './markdown';
import {AppCatalog} from './RoutingMap';
import type {ProviderSpec} from './Settings';
import {AgentSettings, agentConnected, agentName, type CodingAgents, effectiveAgent, useCodingAgents} from './CodingAgents';
import {AppLevels, appLevelsLabel, type LevelsDoc, useLevels} from './Levels';

/* The Integrate page. Main column: the Applications (waiting pairings,
   approved and, behind a filter, rejected ones) and the onboarding wizard,
   which follows the Application selected in the list. Right column: the two
   briefs aiport serves at /_connect, as collapsed tiles. Everything shown
   comes from live state, so the page always matches this gateway. */

type Snippet = {id: string; title: string; language: string; sdk?: string; protocol?: string; available: boolean; code: string};
type Brief = {for: string; title: string; intro: string; steps: string[]; snippets: Snippet[]; note: string};
type BriefLink = {for: string; title: string; audience: string; url: string};
type ConnectDoc = {docs: string; agent_prompt: string; protocols: {name: string; available: boolean}[]; integration: Brief; agent_integration: Brief; briefs: BriefLink[]; adapting: boolean};
type Language = 'python' | 'typescript' | 'curl';
type Filter = 'approved' | 'rejected' | 'all';

const NEW = '';
const STEPS = 6;
const LANGUAGES: [Language, string][] = [['python', 'Python'], ['typescript', 'TypeScript'], ['curl', 'curl']];
const FILTERS: [Filter, string][] = [['approved', 'Approved'], ['rejected', 'Rejected'], ['all', 'All']];
const LANGUAGE_KEY = 'aiport.integrate.language';
const COPIED_KEY = 'aiport.integrate.mcp-copied.';

function remember(key: string, value?: string): string {
  try {
    if (value !== undefined) localStorage.setItem(key, value);
    return localStorage.getItem(key) || '';
  } catch { return value || ''; }
}

const local = (url: string) => url.replace(/^https?:\/\/[^/]+/, '');

export function CopyButton({text, label, onCopied}: {text: string; label: string; onCopied?: () => void}) {
  const [copied, setCopied] = useState(false);
  const copy = () => void navigator.clipboard?.writeText(text).then(() => { setCopied(true); onCopied?.(); setTimeout(() => setCopied(false), 1500); });
  return <button type="button" className="secondary with-icon copy-button" disabled={!text} onClick={copy} aria-label={copied ? 'Copied' : label}>
    {copied ? <><Check size={14} aria-hidden/> Copied</> : <><Copy size={14} aria-hidden/> {label}</>}
  </button>;
}

function SnippetBlock({s, onCopied}: {s: Snippet; onCopied?: () => void}) {
  return <figure className="snippet">
    <figcaption><span>{s.title}</span>
      {s.protocol && !s.available && <span className="badge warn" title="Ask the Owner to configure this protocol first">{s.protocol} not configured</span>}
      <CopyButton text={s.code} label="Copy" onCopied={onCopied}/></figcaption>
    <pre className="snippet-code"><code>{s.code}</code></pre>
  </figure>;
}

/** The latest call (last seven days) of each Application: a time, null for
    none, or undefined while loading. off: activity recording is off. */
type Calls = {off: boolean; at: Record<string, number | null | undefined>};

function useLastCalls(ids: string[], tick: number): Calls {
  const [calls, setCalls] = useState<Calls>({off: false, at: {}});
  const key = ids.join('\n');
  useEffect(() => {
    if (!ids.length) { setCalls({off: false, at: {}}); return; }
    const controller = new AbortController();
    const opts = {credentials: 'same-origin' as const, signal: controller.signal};
    Promise.all([
      fetch('/api/v1/stats?hours=1', opts).then(r => r.ok ? r.json() : {}) as Promise<{telemetry?: {state?: string}}>,
      ...ids.map(id => fetch('/api/v1/requests?' + new URLSearchParams({origin: id, hours: '168', limit: '1'}), opts)
        .then(r => r.ok ? r.json() : {requests: []}).then(d => [id, d.requests?.[0]?.started_at_ms ?? null] as const)),
    ]).then(([stats, ...latest]) => setCalls({off: stats.telemetry?.state === 'disabled', at: Object.fromEntries(latest)})).catch(() => {});
    return () => controller.abort();
  }, [key, tick]);
  return calls;
}

/** The open and resolving Issues of every Application (GET /api/v1/issues), grouped by Application, whether Explain is available, and reload. */
function useIssues(tick: number): {issues: Record<string, Issue[]>; explain: ExplainStatus | null; reload: () => void} {
  const [issues, setIssues] = useState<Record<string, Issue[]>>({}), [explain, setExplain] = useState<ExplainStatus | null>(null), [again, setAgain] = useState(0);
  useEffect(() => {
    const controller = new AbortController();
    fetch('/api/v1/issues', {credentials: 'same-origin', signal: controller.signal}).then(r => r.ok ? r.json() : {issues: []})
      .then((d: {issues?: Issue[]; explain?: ExplainStatus; explain_by?: Record<string, ExplainStatus>; fix_by?: Record<string, FixStatus>}) => {
        const by: Record<string, Issue[]> = {};
        for (const i of d.issues || []) (by[i.application] ||= []).push(i);
        setIssues(by); setExplain(explainOf(d));
      }).catch(() => {});
    return () => controller.abort();
  }, [tick, again]);
  return {issues, explain, reload: () => setAgain(x => x + 1)};
}

/** aiport's guidance version for coding agents and, per Application, the
    version its coding agent last confirmed (GET /api/v1/guidance). */
export type Guidance = {current: number; behind: number; never_confirmed?: number; instruction: string; applications: Record<string, {confirmed: number | null; confirmed_at_ms: number | null; behind: boolean; sends_levels: boolean}>};

/* The Overview line: Applications whose coding agent confirmed an older
   guidance version ("behind"), and those whose agent never confirmed any
   (integrated before guidance versions, say), told apart. */
export function guidanceLine(g: Guidance): string {
  const never = Math.min(g.never_confirmed || 0, g.behind), older = g.behind - never;
  const apps = (n: number) => n === 1 ? '1 Application\'s coding agent' : n + ' Applications\' coding agents';
  const parts: string[] = [];
  if (older) parts.push(apps(older) + (older === 1 ? ' is' : ' are') + ' behind: ' + (older === 1 ? 'it confirmed' : 'they confirmed') + ' an older guidance version than ' + g.current);
  if (never) parts.push(apps(never) + (never === 1 ? ' has' : ' have') + ' never confirmed any guidance version');
  return parts.join('; ') + '.';
}

export function useGuidance(tick: number): Guidance | null {
  const [g, setG] = useState<Guidance | null>(null);
  useEffect(() => {
    const controller = new AbortController();
    fetch('/api/v1/guidance', {credentials: 'same-origin', signal: controller.signal}).then(r => r.ok ? r.json() : null).then(setG).catch(() => {});
    return () => controller.abort();
  }, [tick]);
  return g;
}

/* An Application's guidance state in the list: up to date, behind vN → vM,
   or unknown (never confirmed), with the refresh instruction to copy while
   it is not up to date. Sending Levels is a hint only. */
function GuidanceCell({id, guidance}: {id: string; guidance: Guidance | null}) {
  const a = guidance?.applications[id];
  if (!guidance || !a) return <span className="muted">…</span>;
  if (!a.behind) return <span className="badge good" title={a.confirmed_at_ms ? 'Confirmed ' + new Date(a.confirmed_at_ms).toLocaleString() : undefined}>Up to date</span>;
  const label = a.confirmed ? 'Behind v' + a.confirmed + ' → v' + guidance.current : 'Unknown';
  const why = a.confirmed ? 'Its coding agent confirmed guidance version ' + a.confirmed + '; aiport\'s is ' + guidance.current + '.' : 'Its coding agent has never confirmed a guidance version; aiport\'s is ' + guidance.current + '.';
  return <div className="guidance-cell">
    <span className={'badge' + (a.confirmed ? ' warn' : ' info')} title={why + (a.sends_levels ? ' Its calls sent Levels in the last 6 hours, a hint that its code may follow version 2.' : '')}>{label}</span>
    {a.sends_levels && <small className="muted">Sends Levels</small>}
    <CopyButton text={guidance.instruction} label="Copy refresh instruction"/>
  </div>;
}

/** A short cause for an Issue, for the Applications list. */
const shortCause = (i: Issue) => i.cause === 'unknown_model' ? 'Unknown model ' + i.subject : i.cause === 'failing_connection' ? 'Failing Connection ' + i.subject : i.cause === 'level_unmapped' ? 'Unmapped Level ' + i.subject : isCatalogChange(i) ? 'Catalog change ' + i.subject : i.summary;

/* An Application's Issues, under its row: each with its short cause, a link
   to the Issue on the Changes page, Explain and "Copy fix for the Coding agent". */
function AppIssues({issues, explain, onOpen, onExplained}: {issues: Issue[]; explain: ExplainStatus | null; onOpen: (i: Issue) => void; onExplained: () => void}) {
  return <ul className="app-issues">{issues.map(i => <li key={i.id}>
    <div className="app-issue-head">
      <span className={'badge' + (i.state === 'resolving' ? ' good' : ' warn')}>{i.state === 'resolving' ? 'Resolving' : isCatalogChange(i) ? 'Catalog change' : 'Issue'}</span>
      <span>{i.summary}</span>
      <button type="button" className="text-button" onClick={() => onOpen(i)} aria-label={'Open the Issue ' + shortCause(i) + ' in Changes'}>Open in Changes <ArrowUpRight size={14} aria-hidden/></button>
    </div>
    {isCatalogChange(i) ? <CatalogChange issue={i}/> : <IssueExplanation issue={i} explain={explain} onExplained={onExplained}/>}
    <IssueFix issue={i} explain={explain} onChanged={onExplained}/>
    <ApplicationFix issue={i}/>
  </li>)}</ul>;
}

/** Which of the six onboarding steps are done for an Application. */
function progress(o: Origin | undefined, opts: {handed: boolean; waiting: boolean; seen: number | null | undefined; adapting: boolean; mcp: boolean; agents: CodingAgents}): boolean[] {
  return [!!o || opts.handed || opts.waiting, !!o, typeof opts.seen === 'number', !!o?.may_request_changes && opts.adapting, opts.mcp, agentConnected(o, opts.agents)];
}

const mcpCopied = (id: string) => remember(COPIED_KEY + id) === '1';

function Step({n, done, current, title, children}: {n: number; done: boolean; current: boolean; title: string; children: React.ReactNode}) {
  return <li className={'step' + (done ? ' done' : '') + (current ? ' current' : '')} aria-current={current ? 'step' : undefined}>
    <span className="step-number" aria-hidden>{done ? <CircleCheck size={20} strokeWidth={1.75}/> : current ? n : <CircleDashed size={20} strokeWidth={1.75}/>}</span>
    <div><strong>{n}. {title}</strong><span className="sr-only">{done ? ' (done)' : current ? ' (do this now)' : ' (not yet)'}</span><div className="step-body">{children}</div></div>
  </li>;
}

const ago = (ms: number) => {
  const m = Math.round((Date.now() - ms) / 60000);
  return m < 1 ? 'just now' : m < 60 ? m + ' min ago' : m < 48 * 60 ? Math.round(m / 60) + ' h ago' : Math.round(m / 1440) + ' days ago';
};

function RejectedRow({x, onDone}: {x: Rejected; onDone: () => void}) {
  const [busy, setBusy] = useState(false), [error, setError] = useState('');
  async function forget() {
    setBusy(true); setError('');
    try { await forgetRejected(x.id); onDone(); } catch (e) { setError((e as Error).message); } finally { setBusy(false); }
  }
  return <tr className="rejected-row">
    <td><strong>{x.name}</strong>{x.description && <small className="muted">{x.description}</small>}</td>
    <td><span className="pair-code small">{x.code}</span></td>
    <td><span className="badge bad">Rejected</span></td>
    <td title={new Date(x.rejected_at_ms).toLocaleString()}>{ago(x.rejected_at_ms)}</td>
    <td className="row-actions"><button type="button" className="text-button" disabled={busy} onClick={forget} aria-label={'Forget ' + x.name}>Forget</button>{error && <small className="auth-error" role="alert">{error}</small>}</td>
  </tr>;
}

/* The Applications list: waiting pairings first (always shown), then the
   approved Applications; rejected ones only behind the filter. Selecting an
   Application makes the wizard follow it. */
function ApplicationsList({origins, targets, providers, pairings, rejected, calls, issues, guidance, explain, onExplained, agents, levels, onLevels, app, onSelect, onAgent, onNew, adapting, onDone, onForgot, onIssue, missing}: {
  origins: Origin[]; targets: Target[]; providers: ProviderSpec[]; pairings: Pairing[]; rejected: Rejected[]; calls: Calls; issues: Record<string, Issue[]>; guidance: Guidance | null; explain: ExplainStatus | null; onExplained: () => void; agents: CodingAgents; levels: LevelsDoc | null; onLevels: () => void; app: string; onAgent: (id: string) => void;
  onSelect: (id: string) => void; onNew: () => void; adapting: boolean; onDone: () => void; onForgot: () => void; onIssue: (i: Issue) => void; missing: string[];
}) {
  // The Application whose actions are open under its row (clicking a row opens them here, not the wizard).
  const [expanded, setExpanded] = useState('');
  const pickRow = (id: string) => { setExpanded(x => x === id ? '' : id); onSelect(id); };
  const [adding, setAdding] = useState(false), [removing, setRemoving] = useState(''), [replacing, setReplacing] = useState('');
  // A token just issued or created, shown once under its row (or under the form for a new one).
  const [issued, setIssued] = useState<(IssuedToken & {name: string; replaced: boolean}) | null>(null);
  const [tokenBusy, setTokenBusy] = useState(''), [tokenError, setTokenError] = useState<Record<string, string>>({});
  async function reissue(o: Origin) {
    setReplacing(''); setTokenBusy(o.id); setTokenError(x => ({...x, [o.id]: ''}));
    try { const r = await issueToken(o.id); setIssued({id: o.id, token: r.token, name: o.name, replaced: !missing.includes(o.id)}); onDone(); }
    catch (e) { setTokenError(x => ({...x, [o.id]: (e as Error).message})); } finally { setTokenBusy(''); }
  }
  const [filter, setFilter] = useState<Filter>('approved');
  const [shown, setShown] = useState<Record<string, boolean>>({});
  const [levelsShown, setLevelsShown] = useState<Record<string, boolean>>({});
  const [catalogShown, setCatalogShown] = useState<Record<string, boolean>>({});
  // An open Catalog is read again when the Applications or Connections change, not on every poll (each read may list models).
  const configKey = JSON.stringify([origins, targets]);
  const counts: Record<Filter, number> = {approved: origins.length, rejected: rejected.length, all: origins.length + rejected.length};
  const showApproved = filter !== 'rejected', showRejected = filter !== 'approved';
  const lastCall = (id: string) => calls.off ? 'Not recorded' : calls.at[id] === undefined ? '…' : calls.at[id] === null ? 'None in 7 days' : ago(calls.at[id] as number);
  return <section className="panel apps-list grad-bottom" aria-labelledby="apps-heading">
    <div className="panel-heading">
      <h2 id="apps-heading"><AppWindow size={16} aria-hidden/> Applications<HelpLink section="Applications"/></h2>
      <div className="apps-tools">
        <div className="segmented apps-filter" role="radiogroup" aria-label="Show applications">
          {FILTERS.map(([value, label]) => <button key={value} type="button" role="radio" aria-checked={filter === value} className={filter === value ? 'on' : ''} onClick={() => setFilter(value)}>{label} <span className="count">{counts[value]}</span></button>)}
        </div>
        <button type="button" className="secondary with-icon" onClick={onNew} aria-pressed={app === NEW}><Plus size={14} aria-hidden/> Onboard a new application</button>
        <button type="button" className="secondary with-icon" onClick={() => setAdding(x => !x)} aria-expanded={adding} title="Create an application and its token yourself, without pairing, and paste the token into its .env"><KeyRound size={14} aria-hidden/> Add with a token</button>
      </div>
    </div>
    {pairings.length > 0 && <div className="waiting" role="region" aria-label="Pairing requests waiting">
      <h3 className="waiting-heading"><span className="dot amber" aria-hidden/> {pairings.length === 1 ? '1 pairing waiting for approval' : pairings.length + ' pairings waiting for approval'}</h3>
      <PairingList pairings={pairings} targets={targets} onDone={onDone}/>
    </div>}
    {adding && <AddApplicationForm targets={targets} onCancel={() => setAdding(false)} onAdded={(t, name) => { setAdding(false); setIssued({...t, name, replaced: false}); onDone(); }}/>}
    {issued && !origins.some(o => o.id === issued.id) && <TokenReveal name={issued.name} issued={issued} replaced={issued.replaced} onClose={() => setIssued(null)}/>}
    {showApproved && origins.length > 0 && <div className="apps-table"><table aria-label="Approved applications">
      <thead><tr><th>Application</th><th>Connection</th><th title="May request changes">Changes</th><th>Last call</th><th>Issues</th><th title="The aiport guidance version its coding agent last confirmed">Guidance</th><th title="Coding agent for Agent runs">Agent</th><th>Onboarding</th></tr></thead>
      <tbody>{origins.map(o => {
        const done = progress(o, {handed: true, waiting: false, seen: calls.at[o.id], adapting, mcp: mcpCopied(o.id), agents}).filter(Boolean).length;
        const selected = app === o.id, own = issues[o.id] || [], open = !!shown[o.id] && own.length > 0, catalogOpen = !!catalogShown[o.id], levelsOpen = !!levelsShown[o.id] && !!levels;
        return <React.Fragment key={o.id}><tr className={'app-row' + (selected ? ' selected' : '')} onClick={e => { if (!(e.target as HTMLElement).closest('input,button,a,label')) pickRow(o.id); }}>
          <td><button type="button" className="row-select" aria-expanded={expanded === o.id} onClick={() => pickRow(o.id)} title={'Open ' + o.name + '\'s token and removal actions'}>{o.name}</button><small><code>{o.id}</code></small>
            <span className="app-toggles"><button type="button" className="text-button issue-count app-catalog-toggle" aria-expanded={catalogOpen} aria-controls={'app-catalog-' + o.id} onClick={() => setCatalogShown(x => ({...x, [o.id]: !x[o.id]}))}
              title={'What ' + o.name + ' can send, what serves it, and its recent Catalog changes'}><ChevronRight size={14} strokeWidth={2} aria-hidden className="chevron"/> Catalog</button>
            <button type="button" className="text-button issue-count app-catalog-toggle" aria-expanded={levelsOpen} aria-controls={'app-levels-' + o.id} disabled={!levels} onClick={() => setLevelsShown(x => ({...x, [o.id]: !x[o.id]}))}
              title={o.name + '\'s own mapping of fast, balanced and best, before the shared one: ' + appLevelsLabel(levels, o.id)}><ChevronRight size={14} strokeWidth={2} aria-hidden className="chevron"/> Levels</button>
            <button type="button" className="text-button issue-count app-catalog-toggle" disabled={tokenBusy === o.id} onClick={() => missing.includes(o.id) ? reissue(o) : setReplacing(x => x === o.id ? '' : o.id)} aria-expanded={missing.includes(o.id) ? undefined : replacing === o.id} aria-label={'Issue a new token for ' + o.name}
              title={'Issue a new token for ' + o.name + ' to paste into its .env; the current one stops working'}><KeyRound size={13} aria-hidden/> {missing.includes(o.id) ? 'Issue token' : 'New token'}</button>
            <button type="button" className="text-button issue-count app-catalog-toggle" onClick={() => setRemoving(x => x === o.id ? '' : o.id)} aria-expanded={removing === o.id} aria-label={'Remove ' + o.name}><Trash2 size={13} aria-hidden/> Remove</button></span>
            {missing.includes(o.id) && <span className="badge warn" title="Its calls are refused until you issue a token">No token</span>}
            {tokenError[o.id] && <small className="auth-error" role="alert">{tokenError[o.id]}</small>}</td>
          <td>{o.target ? <code>{o.target}</code> : <span className="muted">Protocol default</span>}</td>
          <td><MayRequestToggle o={o} onDone={onDone}/></td>
          <td title={typeof calls.at[o.id] === 'number' ? new Date(calls.at[o.id] as number).toLocaleString() : undefined}>{lastCall(o.id)}</td>
          <td>{own.length ? <button type="button" className="text-button issue-count" aria-expanded={open} aria-controls={'app-issues-' + o.id} onClick={() => setShown(x => ({...x, [o.id]: !x[o.id]}))}
              title={own.map(shortCause).join('\n')}>
              <ChevronRight size={14} strokeWidth={2} aria-hidden className="chevron"/><span className="badge warn">{own.length}</span> {own.length === 1 ? shortCause(own[0]) : own.length + ' Issues'}</button>
            : <span className="muted">None</span>}</td>
          <td><GuidanceCell id={o.id} guidance={guidance}/></td>
          <td>{(() => { const id = effectiveAgent(o, agents), connected = agentConnected(o, agents);
            return <button type="button" className="text-button" onClick={() => onAgent(o.id)} aria-label={(connected ? 'Change the coding agent of ' : 'Connect a coding agent to ') + o.name}
              title={connected ? agentName(id, agents) + (o.agent ? '' : ' (default)') + ' in ' + o.repository : 'Connect a coding agent (onboarding step 6)'}>{connected ? agentName(id, agents) : 'Connect'}</button>; })()}</td>
          <td><span className={'progress' + (done === STEPS ? ' complete' : '')} aria-hidden><i style={{width: done / STEPS * 100 + '%'}}/></span> <span className="progress-text"><span className="sr-only">Onboarding steps done: </span>{done}/{STEPS}</span></td>
        </tr>{expanded === o.id && <tr className="app-issues-row"><td colSpan={8}><div className="app-actions" role="region" aria-label={o.name + ' actions'}>
            <p><strong>{o.name}</strong> <code>{o.id}</code>{missing.includes(o.id) ? <> · <span className="badge warn">No token</span> its calls are refused until you issue one.</> : <> · its application sends its token as <code>AIPORT_APP_TOKEN</code> from its <code>.env</code>.</>}</p>
            <div className="pairing-actions">
              <button type="button" className="secondary with-icon" disabled={tokenBusy === o.id} onClick={() => missing.includes(o.id) ? reissue(o) : setReplacing(o.id)}><KeyRound size={14} aria-hidden/> {missing.includes(o.id) ? 'Issue token' : 'Issue new token'}</button>
              <button type="button" className="danger with-icon" onClick={() => setRemoving(o.id)}><Trash2 size={14} aria-hidden/> Remove application</button>
              <button type="button" className="text-button" onClick={() => setExpanded('')}>Close</button>
            </div>
          </div></td></tr>}
          {issued?.id === o.id && <tr className="app-issues-row"><td colSpan={8}><TokenReveal name={o.name} issued={issued} replaced={issued.replaced} onClose={() => setIssued(null)}/></td></tr>}
          {replacing === o.id && <tr className="app-issues-row"><td colSpan={8}><div className="token-reveal" role="region" aria-label={'New token for ' + o.name}>
            <p><strong>Issue a new token for {o.name}?</strong> Its current token stops working at once; put the new one in its <code>.env</code>.</p>
            <div className="pairing-actions"><button type="button" className="primary" disabled={tokenBusy === o.id} onClick={() => reissue(o)}>Issue new token</button><button type="button" className="text-button" onClick={() => setReplacing('')}>Cancel</button></div>
          </div></td></tr>}
          {removing === o.id && <tr className="app-issues-row"><td colSpan={8}><RemoveApplication o={o} onCancel={() => setRemoving('')} onRemoved={() => { setRemoving(''); if (issued?.id === o.id) setIssued(null); onDone(); }}/></td></tr>}{open && <tr className="app-issues-row" id={'app-issues-' + o.id}><td colSpan={8}><AppIssues issues={own} explain={explain} onOpen={onIssue} onExplained={onExplained}/></td></tr>}
          {catalogOpen && <tr className="app-issues-row" id={'app-catalog-' + o.id}><td colSpan={8}><AppCatalog application={o.id} targets={targets} providers={providers} reload={configKey}/></td></tr>}
          {levelsOpen && <tr className="app-issues-row" id={'app-levels-' + o.id}><td colSpan={8}><AppLevels doc={levels!} application={o.id} onChange={onLevels}/></td></tr>}</React.Fragment>;
      })}</tbody>
    </table></div>}
    {showRejected && rejected.length > 0 && <div className="apps-table rejected-table"><table aria-label="Rejected applications">
      <thead><tr><th>Rejected application</th><th>Code</th><th>Status</th><th>Rejected</th><th><span className="sr-only">Actions</span></th></tr></thead>
      <tbody>{rejected.map(x => <RejectedRow key={x.id} x={x} onDone={onForgot}/>)}</tbody>
    </table></div>}
    {!pairings.length && !(showApproved && origins.length) && !(showRejected && rejected.length) && <p className="quiet-empty">{filter === 'rejected'
      ? 'No rejected applications. A pairing request you reject is listed here, so you can see who asked.'
      : 'No applications yet. Onboard one below: hand the instruction to the coding agent of an application; when it pairs, its request appears here for you to approve.'}</p>}
  </section>;
}

/* The onboarding wizard: what the Owner does, in order, to connect one
   application. Each step ticks itself from live state; the first step not
   done is the one to do now. Collapsible; its header says where things are. */
function Wizard({origins, app, setApp, pairings, seen, off, adapting, agents, prompt, docs, mcp, setMcp, open, setOpen, onDone, onAgentBrief}: {
  origins: Origin[]; app: string; agents: CodingAgents; setApp: (id: string) => void; pairings: Pairing[]; seen: number | null | undefined; off: boolean; adapting: boolean;
  prompt: string; docs: string; mcp: boolean; setMcp: (v: boolean) => void; open: boolean; setOpen: (v: boolean) => void; onDone: () => void; onAgentBrief: () => void;
}) {
  const [handed, setHanded] = useState(false);
  const o = origins.find(x => x.id === app);
  const steps = progress(o, {handed, waiting: pairings.length > 0, seen, adapting, mcp, agents});
  const current = steps.findIndex(x => !x), done = steps.filter(Boolean).length;
  const summary = pairings.length ? (pairings.length === 1 ? '1 pairing waiting' : pairings.length + ' pairings waiting')
    : o ? (current < 0 ? o.name + ' is onboarded' : o.name + ' · ' + done + '/' + STEPS)
    : 'A new application';
  return <section className="panel wizard" aria-labelledby="wizard-heading">
    <div className="panel-heading wizard-heading">
      <h2 id="wizard-heading"><button type="button" className="group-toggle" aria-expanded={open} aria-controls="wizard-body" onClick={() => setOpen(!open)}>
        <ChevronRight size={16} strokeWidth={2} aria-hidden className="chevron"/>Onboard an application</button><HelpLink section="Onboarding an application, step by step"/></h2>
      <span className={'badge' + (pairings.length ? ' warn' : current < 0 && o ? ' good' : ' info')}>{summary}</span>
    </div>
    {open && <div id="wizard-body">
      <p className="wizard-follows muted-small">Following <strong>{o ? o.name : 'a new application'}</strong>{o && <> · <button type="button" className="text-button" onClick={() => setApp(NEW)}>Onboard a new application instead</button></>}. Select an application in the list to follow it.</p>
      <ol className="steps">
        <Step n={1} done={steps[0]} current={current === 0} title="Hand this to the coding agent">
          <p>Open the coding agent (Claude Code, Codex, Cursor…) <strong>in the application's repository</strong> and paste this instruction. It tells the agent to read <a href={local(docs)} target="_blank" rel="noreferrer"><code>{docs}</code></a>, which aiport writes from this gateway and which holds everything for both parts: wiring the application's LLM calls through aiport, and letting later sessions ask aiport for changes.</p>
          <blockquote className="agent-prompt">{prompt}</blockquote>
          <div className="button-row"><CopyButton text={prompt} label="Copy instruction" onCopied={() => setHanded(true)}/><CopyButton text={docs} label="Copy URL only"/></div>
        </Step>
        <Step n={2} done={steps[1]} current={current === 1} title="Approve the pairing">
          {o ? <>Connected as <code>{o.id}</code>{o.target ? <>, bound to <code>{o.target}</code></> : ', using the protocol defaults'}.</>
            : pairings.length > 0 ? <>The agent asked to pair: the request waits at the top of <strong>Applications</strong> above, with code{pairings.length > 1 ? 's' : ''} {pairings.map(p => <span key={p.code} className="pair-code small">{p.code}</span>)}. Check it matches the code the agent shows you, then approve it there; you can bind a Connection and allow change requests.</>
            : <>Waiting for the agent to ask. Its request appears at the top of Applications within a few seconds, with a short code; requests expire after 10 minutes.</>}
        </Step>
        <Step n={3} done={steps[2]} current={current === 2} title="First call arrives">
          {!o ? 'After pairing, the agent makes one real call through aiport.'
            : off ? <>Activity recording is off (<code>telemetry.enabled: false</code>), so aiport cannot tell. Ask the agent whether its test call succeeded.</>
            : seen === undefined ? 'Checking…'
            : typeof seen === 'number' ? <>Latest call {new Date(seen).toLocaleString()}. The application now goes through aiport.</>
            : 'No call from this application in the last 7 days yet. If its calls show as unknown in Activity, it is not sending its own token.'}
        </Step>
        <Step n={4} done={steps[3]} current={current === 3} title="Allow change requests (optional)">
          {!o ? 'Once paired, decide whether coding agents in its repository may ask you for changes.'
            : !adapting ? <>Adapting is turned off in the configuration (<code>adapting.disabled</code>), so no Application can request changes now.</>
            : <><MayRequestToggle o={o} onDone={onDone}/> You still decide each request on the Changes page.</>}
        </Step>
        <Step n={5} done={steps[4]} current={current === 4} title="MCP entry in the repository (optional)">
          <label className="check-label compact"><input type="checkbox" checked={mcp} onChange={e => setMcp(e.target.checked)}/>The agent adds it as part of step 1. To do it by hand, copy it from the <button type="button" className="text-button" onClick={onAgentBrief}>Coding agent brief</button>.</label>
        </Step>
        <Step n={6} done={steps[5]} current={current === 5} title="Connect a coding agent (optional)">
          <p className="step-why"><strong>Why connect one:</strong> when this application keeps failing, aiport can show you more than the error. With a coding agent and the repository, <strong>Explain</strong> reads the application's code and tells you the root cause, and <strong>Fix it</strong> makes the change on a new branch for you to review. Nothing runs until you click, Explain only reads, and aiport never commits to your branch, merges or pushes. Without one, aiport still explains from the evidence alone, more briefly.</p>
          <p>The coding agent aiport may start, on your click, to explain an Issue of this application or fix it in its own branch, and the repository it works in. Choose among the agents installed on this machine{agents.default_agent ? <>, or keep the default from Settings</> : <>; a default for every application is in Settings</>}.</p>
          {o ? <AgentSettings o={o} agents={agents} onDone={onDone}/> : 'Once paired: the repository path the agent reported while pairing, as you confirmed it, appears here.'}
        </Step>
      </ol>
      {current < 0 && <p className="notice" role="status"><CircleCheck size={16} aria-hidden/> This application is onboarded.</p>}
    </div>}
  </section>;
}

/* One brief as a tile in the right strip: collapsed it shows the title,
   who it is for, Copy whole brief and the URL; expanded, its steps and
   snippets. */
function BriefTile({id, title, short, icon, brief, link, open, setOpen, snippets, picker, onSnippetCopied}: {
  id: string; title: string; short: string; icon: React.ReactNode; brief: Brief; link?: BriefLink; open: boolean; setOpen: (v: boolean) => void;
  snippets: Snippet[]; picker?: React.ReactNode; onSnippetCopied?: (s: Snippet) => void;
}) {
  const [text, setText] = useState('');
  useEffect(() => {
    if (!link) return;
    fetch(local(link.url), {credentials: 'same-origin'}).then(r => r.ok ? r.text() : '').then(setText).catch(() => {});
  }, [link?.url, brief]);
  return <section className={'tile' + (open ? ' open' : '')} id={'tile-' + id} aria-labelledby={'tile-heading-' + id}>
    <div className="tile-head">
      <span className="group-icon" aria-hidden>{icon}</span>
      <div className="tile-title">
        <h3 id={'tile-heading-' + id}><button type="button" className="group-toggle" aria-expanded={open} aria-controls={'tile-body-' + id} onClick={() => setOpen(!open)}>
          <ChevronRight size={15} strokeWidth={2} aria-hidden className="chevron"/>{title}</button></h3>
        <p className="tile-audience" title={link?.audience}>{short}</p>
      </div>
    </div>
    <div className="tile-actions">
      <CopyButton text={text} label="Copy whole brief"/>
      {link && <a className="tile-url" href={local(link.url)} target="_blank" rel="noreferrer" title="Open the brief"><code>{local(link.url)}</code><ExternalLink size={12} aria-hidden/></a>}
    </div>
    {open && <div className="brief tile-body" id={'tile-body-' + id}>
      {link?.audience && <p className="muted-small">{link.audience}</p>}
      <p><Inline text={brief.intro}/></p>
      <ol>{brief.steps.map((s, i) => <li key={i}><Inline text={s}/></li>)}</ol>
      <div className="snippets-heading"><h4>Snippets</h4>{picker}</div>
      {snippets.map(s => <SnippetBlock key={s.id} s={s} onCopied={() => onSnippetCopied?.(s)}/>)}
      <p className="muted-small"><Inline text={brief.note}/></p>
    </div>}
  </section>;
}

export default function IntegratePage({origins, targets, providers = [], tick, onChange, onManual, onChanges}: {origins: Origin[]; targets: Target[]; providers?: ProviderSpec[]; tick: number; onChange: () => void; onManual: () => void; onChanges?: () => void}) {
  const [doc, setDoc] = useState<ConnectDoc | null>(null), [error, setError] = useState('');
  const [language, setLanguage] = useState<Language>(() => (['python', 'typescript', 'curl'].includes(remember(LANGUAGE_KEY)) ? remember(LANGUAGE_KEY) : 'python') as Language);
  const [app, setApp] = useState(NEW);
  const [mcp, setMcpState] = useState(() => mcpCopied(NEW));
  const [pairings, refreshPairings] = usePairings();
  const [rejected, refreshRejected] = useRejected(tick);
  const calls = useLastCalls(origins.map(o => o.id), tick);
  const {issues, explain, reload: reloadIssues} = useIssues(tick);
  const [levels, reloadLevels] = useLevels(tick);
  const guidance = useGuidance(tick);
  const missing = useMissingTokens(tick);
  const [agents] = useCodingAgents(tick);
  // The wizard starts collapsed once an Application exists; null = not chosen yet.
  const [wizardOpen, setWizardOpen] = useState<boolean | null>(null);
  const [appTile, setAppTile] = useState(false), [agentTile, setAgentTile] = useState(false);
  const wizardRef = useRef<HTMLDivElement>(null);
  const known = useRef(new Set(origins.map(x => x.id)));
  const waiting = pairings.length > 0;

  useEffect(() => { setMcpState(mcpCopied(app)); }, [app]);
  const setMcp = (v: boolean) => { remember(COPIED_KEY + app, v ? '1' : '0'); setMcpState(v); };
  // An application approved while onboarding "a new application" becomes the one followed.
  useEffect(() => {
    const added = origins.find(x => !known.current.has(x.id));
    origins.forEach(x => known.current.add(x.id));
    if (added && app === NEW) setApp(added.id);
  }, [origins]);
  // A pairing request that arrives opens the wizard, so it is never missed.
  const wasWaiting = useRef(false);
  useEffect(() => { if (waiting && !wasWaiting.current) setWizardOpen(true); wasWaiting.current = waiting; }, [waiting]);
  useEffect(() => {
    fetch('/_connect', {credentials: 'same-origin'}).then(r => { if (!r.ok) throw new Error(); return r.json(); }).then(setDoc).catch(() => setError('Could not load the instructions from /_connect.'));
  }, [tick]);

  if (error) return <p className="auth-error" role="alert">{error}</p>;
  if (!doc) return <p className="muted">Loading…</p>;
  const open = wizardOpen ?? (origins.length === 0 || waiting);
  const follow = (id: string) => { setApp(id); setWizardOpen(true); requestAnimationFrame(() => wizardRef.current?.scrollIntoView({behavior: 'smooth', block: 'nearest'})); };
  const followAgent = (id: string) => { setApp(id); setWizardOpen(true); requestAnimationFrame(() => document.getElementById('agent-' + id)?.focus()); };
  const done = () => { refreshPairings(); refreshRejected(); onChange(); };
  const openIssue = (i: Issue) => { history.replaceState(null, '', '#' + issueAnchor(i.id)); onChanges?.(); };
  const pick = (l: Language) => { setLanguage(l); remember(LANGUAGE_KEY, l); };
  const link = (f: string) => doc.briefs.find(b => b.for === f);
  const appBrief = doc.integration, agentBrief = doc.agent_integration;
  // Configured protocols first, so the snippet that works here comes first.
  const sdkSnippets = appBrief.snippets.filter(s => s.sdk && s.language === language).sort((a, b) => Number(b.available) - Number(a.available));
  const picker = <div className="segmented languages" role="radiogroup" aria-label="Language">
    {LANGUAGES.map(([value, label]) => <button key={value} type="button" role="radio" aria-checked={language === value} className={language === value ? 'on' : ''} onClick={() => pick(value)}>{label}</button>)}
  </div>;
  const openAgentBrief = () => { setAgentTile(true); requestAnimationFrame(() => document.getElementById('tile-agent')?.scrollIntoView({behavior: 'smooth', block: 'start'})); };

  return <div className="integrate-layout">
    <div className="integrate-main">
      <ApplicationsList origins={origins} targets={targets} providers={providers} pairings={pairings} rejected={rejected} calls={calls} issues={issues} guidance={guidance} explain={explain} onExplained={reloadIssues} agents={agents} levels={levels} onLevels={() => { reloadLevels(); onChange(); }} app={app} adapting={doc.adapting}
        onSelect={setApp} onAgent={followAgent} onNew={() => follow(NEW)} onDone={done} onForgot={refreshRejected} onIssue={openIssue} missing={missing}/>
      <div ref={wizardRef}>
        <Wizard origins={origins} app={app} setApp={setApp} pairings={pairings} seen={app ? calls.at[app] : null} off={calls.off} adapting={doc.adapting} agents={agents}
          prompt={doc.agent_prompt} docs={doc.docs} mcp={mcp} setMcp={setMcp} open={open} setOpen={setWizardOpen} onDone={done} onAgentBrief={openAgentBrief}/>
      </div>
    </div>
    <aside className="integrate-strip" aria-labelledby="strip-heading">
      <div className="strip-head">
        <h2 id="strip-heading">Instructions<HelpLink section="The instructions"/></h2>
        <p className="muted-small">What the coding agent reads, written by aiport from this gateway's address and protocols ({doc.protocols.map(p => p.name + (p.available ? ' configured' : ' not configured')).join(', ')}). <button type="button" className="text-button" onClick={onManual}>Read the manual</button></p>
      </div>
      <BriefTile id="application" title="Application brief" short="For the agent wiring the application's LLM calls through aiport." icon={<Terminal size={17} strokeWidth={1.75}/>}
        brief={appBrief} link={link('application')} open={appTile} setOpen={setAppTile} picker={picker} snippets={[...appBrief.snippets.filter(s => !s.sdk), ...sdkSnippets]}/>
      <BriefTile id="agent" title="Coding agent brief" short="For agents working in its repository later: MCP entry, AGENTS.md, Change requests." icon={<Bot size={17} strokeWidth={1.75}/>}
        brief={agentBrief} link={link('agent')} open={agentTile} setOpen={setAgentTile} snippets={agentBrief.snippets} onSnippetCopied={s => { if (s.id === 'mcp') setMcp(true); }}/>
      <p className="muted-small strip-foot">Reference: <a href="/_connect?format=markdown" target="_blank" rel="noreferrer"><code>/_connect?format=markdown</code> <ExternalLink size={12} aria-hidden/></a></p>
    </aside>
  </div>;
}
