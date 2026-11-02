import React, {useEffect, useState} from 'react';
import {BookOpen, Check, CircleCheck, CircleDashed, Copy, ExternalLink} from 'lucide-react';
import {MayRequestToggle, Origin, PairingList, Target, usePairings} from './Applications';
import {Inline} from './markdown';

/* The Integrate page: the two briefs aiport serves at /_connect, made
   interactive. Everything shown comes from the live /_connect document, so
   the page always matches this gateway. */

type Snippet = {id: string; title: string; language: string; sdk?: string; protocol?: string; available: boolean; code: string};
type Brief = {for: string; title: string; intro: string; steps: string[]; snippets: Snippet[]; note: string};
type BriefLink = {for: string; title: string; audience: string; url: string};
type ConnectDoc = {protocols: {name: string; available: boolean}[]; integration: Brief; agent_integration: Brief; briefs: BriefLink[]; adapting: boolean};
type Language = 'python' | 'typescript' | 'curl';
type Tab = 'application' | 'agent';

const NEW = '';
const LANGUAGES: [Language, string][] = [['python', 'Python'], ['typescript', 'TypeScript'], ['curl', 'curl']];
const LANGUAGE_KEY = 'aiport.integrate.language';
const COPIED_KEY = 'aiport.integrate.mcp-copied.';

function remember(key: string, value?: string): string {
  try {
    if (value !== undefined) localStorage.setItem(key, value);
    return localStorage.getItem(key) || '';
  } catch { return value || ''; }
}

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

type Seen = 'loading' | 'off' | number | null;

/** Has aiport seen a call from this Application in the last seven days? */
function useFirstCall(app: string, tick: number): Seen {
  const [seen, setSeen] = useState<Seen>('loading');
  useEffect(() => {
    if (!app) { setSeen(null); return; }
    const controller = new AbortController();
    const q = new URLSearchParams({origin: app, hours: '168', limit: '1'});
    Promise.all([
      fetch('/api/v1/requests?' + q, {credentials: 'same-origin', signal: controller.signal}).then(r => r.ok ? r.json() : {requests: []}),
      fetch('/api/v1/stats?hours=1', {credentials: 'same-origin', signal: controller.signal}).then(r => r.ok ? r.json() : {}) as Promise<{telemetry?: {state?: string}}>,
    ]).then(([calls, stats]) => {
      if (stats.telemetry?.state === 'disabled') setSeen('off');
      else setSeen(calls.requests?.[0]?.started_at_ms ?? null);
    }).catch(() => {});
    return () => controller.abort();
  }, [app, tick]);
  return seen;
}

function Step({done, title, children}: {done: boolean; title: string; children: React.ReactNode}) {
  return <li className={'step' + (done ? ' done' : '')}>
    {done ? <CircleCheck size={18} strokeWidth={1.75} aria-hidden className="step-icon"/> : <CircleDashed size={18} strokeWidth={1.75} aria-hidden className="step-icon"/>}
    <div><strong>{title}</strong><span className="sr-only">{done ? ' (done)' : ' (not yet)'}</span><div className="muted-small">{children}</div></div>
  </li>;
}

function Checklist({origins, targets, app, setApp, adapting, mcpCopied, setMcpCopied, tick, onChange}: {
  origins: Origin[]; targets: Target[]; app: string; setApp: (id: string) => void; adapting: boolean;
  mcpCopied: boolean; setMcpCopied: (v: boolean) => void; tick: number; onChange: () => void;
}) {
  const [pairings, refresh] = usePairings();
  const seen = useFirstCall(app, tick);
  const o = origins.find(x => x.id === app);
  const done = () => { refresh(); onChange(); };
  return <section className="panel" aria-labelledby="checklist-heading">
    <div className="panel-heading"><h2 id="checklist-heading">Checklist</h2>
      <label className="inline-select">Application
        <select value={app} onChange={e => setApp(e.target.value)}>
          <option value={NEW}>A new application</option>
          {origins.map(x => <option key={x.id} value={x.id}>{x.name}</option>)}
        </select>
      </label>
    </div>
    {pairings.length > 0 && <><p className="muted handover-intro"><strong>Waiting for your approval.</strong> Check the application shows the same code.</p><PairingList pairings={pairings} targets={targets} onDone={done}/></>}
    <ol className="steps">
      <Step done={!!o} title="Paired">{o
        ? <>Connected as <code>{o.id}</code>{o.target ? <>, bound to <code>{o.target}</code></> : ', using the protocol defaults'}.</>
        : <>Give the Application brief below to the coding agent working on the application. It asks aiport to pair; the request appears here for you to approve, and the application is listed above once approved.</>}</Step>
      <Step done={typeof seen === 'number'} title="First call seen">{!o ? 'After pairing, the application makes one real call.'
        : seen === 'loading' ? 'Checking…'
        : seen === 'off' ? <>Activity recording is off (<code>telemetry.enabled: false</code>), so aiport cannot tell.</>
        : typeof seen === 'number' ? <>Latest call {new Date(seen).toLocaleString()}.</>
        : 'No call from this application in the last 7 days. If its calls show as unknown in Activity, it is not sending its own token.'}</Step>
      <Step done={!!o?.may_request_changes && adapting} title="May request changes">{!o ? 'Optional. You can allow it while approving the pairing, or here afterwards.'
        : !adapting ? <>Adapting is turned off in the configuration (<code>adapting.disabled</code>), so no Application can request changes now.</>
        : <><MayRequestToggle o={o} onDone={done}/> Lets coding agents in its repository ask you for Changes; you decide each one in Changes.</>}</Step>
      <Step done={mcpCopied} title="MCP entry copied">
        <label className="check-label compact"><input type="checkbox" checked={mcpCopied} onChange={e => setMcpCopied(e.target.checked)}/>Copy the project MCP entry from the Coding agent tab into the application's repository (marked here when you copy it).</label>
      </Step>
    </ol>
  </section>;
}

function BriefPanel({brief, link, snippets, picker, onSnippetCopied}: {brief: Brief; link?: BriefLink; snippets: Snippet[]; picker?: React.ReactNode; onSnippetCopied?: (s: Snippet) => void}) {
  const [text, setText] = useState('');
  useEffect(() => {
    if (!link) return;
    fetch(link.url.replace(/^https?:\/\/[^/]+/, ''), {credentials: 'same-origin'}).then(r => r.ok ? r.text() : '').then(setText).catch(() => {});
  }, [link?.url, brief]);
  return <section className="panel" aria-labelledby={'brief-' + brief.for}>
    <div className="panel-heading"><h2 id={'brief-' + brief.for}>{link?.title || brief.title}</h2><CopyButton text={text} label="Copy whole brief"/></div>
    <p className="muted handover-intro">{link?.audience} Paste it to the coding agent, or point it at <a href={link?.url.replace(/^https?:\/\/[^/]+/, '')} target="_blank" rel="noreferrer"><code>{link?.url}</code></a>.</p>
    <div className="brief">
      <p><Inline text={brief.intro}/></p>
      <ol>{brief.steps.map((s, i) => <li key={i}><Inline text={s}/></li>)}</ol>
      <div className="snippets-heading"><h3>Snippets</h3>{picker}</div>
      {snippets.map(s => <SnippetBlock key={s.id} s={s} onCopied={() => onSnippetCopied?.(s)}/>)}
      <p className="muted-small"><Inline text={brief.note}/></p>
    </div>
  </section>;
}

export default function IntegratePage({origins, targets, tick, onChange, onManual}: {origins: Origin[]; targets: Target[]; tick: number; onChange: () => void; onManual: () => void}) {
  const [doc, setDoc] = useState<ConnectDoc | null>(null), [error, setError] = useState('');
  const [tab, setTab] = useState<Tab>('application');
  const [language, setLanguage] = useState<Language>(() => (['python', 'typescript', 'curl'].includes(remember(LANGUAGE_KEY)) ? remember(LANGUAGE_KEY) : 'python') as Language);
  const [app, setApp] = useState(NEW);
  const [copied, setCopied] = useState(() => remember(COPIED_KEY + NEW) === '1');
  useEffect(() => { setCopied(remember(COPIED_KEY + app) === '1'); }, [app]);
  const markCopied = (v: boolean) => { remember(COPIED_KEY + app, v ? '1' : '0'); setCopied(v); };
  useEffect(() => {
    fetch('/_connect', {credentials: 'same-origin'}).then(r => { if (!r.ok) throw new Error(); return r.json(); }).then(setDoc).catch(() => setError('Could not load the instructions from /_connect.'));
  }, [tick]);
  const pick = (l: Language) => { setLanguage(l); remember(LANGUAGE_KEY, l); };
  if (error) return <p className="auth-error" role="alert">{error}</p>;
  if (!doc) return <p className="muted">Loading…</p>;
  const link = (f: string) => doc.briefs.find(b => b.for === f);
  const app_ = doc.integration, agent = doc.agent_integration;
  // Configured protocols first, so the snippet that works here comes first.
  const sdkSnippets = app_.snippets.filter(s => s.sdk && s.language === language).sort((a, b) => Number(b.available) - Number(a.available));
  const picker = <div className="segmented languages" role="radiogroup" aria-label="Language">
    {LANGUAGES.map(([value, label]) => <button key={value} type="button" role="radio" aria-checked={language === value} className={language === value ? 'on' : ''} onClick={() => pick(value)}>{label}</button>)}
  </div>;
  return <>
    <div className="notice info-notice"><BookOpen size={16} aria-hidden/><div>
      aiport writes these instructions itself, from this gateway's live address and protocols (<code>{doc.protocols.map(p => p.name + (p.available ? ' configured' : ' not configured')).join(', ')}</code>). The <strong>Application brief</strong> is for the coding agent connecting an application; the <strong>Coding agent brief</strong> is for agents working in that application's repository later. <button type="button" className="text-button" onClick={onManual}>Read the manual</button>
    </div></div>
    <Checklist origins={origins} targets={targets} app={app} setApp={setApp} adapting={doc.adapting} mcpCopied={copied} setMcpCopied={markCopied} tick={tick} onChange={onChange}/>
    <div className="task-filters integrate-tabs" role="tablist" aria-label="Brief">
      <button role="tab" id="tab-application" aria-controls="panel-application" aria-selected={tab === 'application'} className={tab === 'application' ? 'active' : ''} onClick={() => setTab('application')}>Application</button>
      <button role="tab" id="tab-agent" aria-controls="panel-agent" aria-selected={tab === 'agent'} className={tab === 'agent' ? 'active' : ''} onClick={() => setTab('agent')}>Coding agent</button>
    </div>
    <div role="tabpanel" id={'panel-' + tab} aria-labelledby={'tab-' + tab}>
      {tab === 'application'
        ? <BriefPanel brief={app_} link={link('application')} picker={picker} snippets={[...app_.snippets.filter(s => !s.sdk), ...sdkSnippets]}/>
        : <BriefPanel brief={agent} link={link('agent')} snippets={agent.snippets} onSnippetCopied={s => { if (s.id === 'mcp') markCopied(true); }}/>}
    </div>
    <p className="muted-small integrate-foot">Reference: <a href="/_connect?format=markdown" target="_blank" rel="noreferrer"><code>/_connect?format=markdown</code> <ExternalLink size={12} aria-hidden/></a></p>
  </>;
}
