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
type ConnectDoc = {docs: string; agent_prompt: string; protocols: {name: string; available: boolean}[]; integration: Brief; agent_integration: Brief; briefs: BriefLink[]; adapting: boolean};
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

function Step({n, done, current, title, children}: {n: number; done: boolean; current: boolean; title: string; children: React.ReactNode}) {
  return <li className={'step' + (done ? ' done' : '') + (current ? ' current' : '')} aria-current={current ? 'step' : undefined}>
    <span className="step-number" aria-hidden>{done ? <CircleCheck size={20} strokeWidth={1.75}/> : current ? n : <CircleDashed size={20} strokeWidth={1.75}/>}</span>
    <div><strong>{n}. {title}</strong><span className="sr-only">{done ? ' (done)' : current ? ' (do this now)' : ' (not yet)'}</span><div className="step-body">{children}</div></div>
  </li>;
}

/* The onboarding wizard: what the Owner does, in order, to connect one
   application. Each step ticks itself from live state; the first step not
   done is the one to do now. */
function Wizard({origins, targets, app, setApp, adapting, prompt, docs, mcpCopied, setMcpCopied, tick, onChange, onAgentTab}: {
  origins: Origin[]; targets: Target[]; app: string; setApp: (id: string) => void; adapting: boolean; prompt: string; docs: string;
  mcpCopied: boolean; setMcpCopied: (v: boolean) => void; tick: number; onChange: () => void; onAgentTab: () => void;
}) {
  const [pairings, refresh] = usePairings();
  const [handed, setHanded] = useState(false);
  const known = React.useRef(new Set(origins.map(x => x.id)));
  // An application approved while onboarding "a new application" becomes the one followed.
  useEffect(() => {
    const added = origins.find(x => !known.current.has(x.id));
    origins.forEach(x => known.current.add(x.id));
    if (added && app === NEW) setApp(added.id);
  }, [origins]);
  const seen = useFirstCall(app, tick);
  const o = origins.find(x => x.id === app);
  const done = () => { refresh(); onChange(); };
  const steps = [!!o || handed || pairings.length > 0, !!o, typeof seen === 'number', !!o?.may_request_changes && adapting, mcpCopied];
  const current = steps.findIndex(x => !x);
  return <section className="panel wizard" aria-labelledby="wizard-heading">
    <div className="panel-heading"><h2 id="wizard-heading">Onboard an application</h2>
      <label className="inline-select">Application
        <select value={app} onChange={e => setApp(e.target.value)}>
          <option value={NEW}>A new application</option>
          {origins.map(x => <option key={x.id} value={x.id}>{x.name}</option>)}
        </select>
      </label>
    </div>
    <ol className="steps">
      <Step n={1} done={steps[0]} current={current === 0} title="Hand this to the coding agent">
        <p>Open the coding agent (Claude Code, Codex, Cursor…) <strong>in the application's repository</strong> and paste this instruction. It tells the agent to read <a href={docs.replace(/^https?:\/\/[^/]+/, '')} target="_blank" rel="noreferrer"><code>{docs}</code></a>, which aiport writes from this gateway and which holds everything for both parts: wiring the application's LLM calls through aiport, and letting later sessions ask aiport for changes.</p>
        <blockquote className="agent-prompt">{prompt}</blockquote>
        <div className="button-row"><CopyButton text={prompt} label="Copy instruction" onCopied={() => setHanded(true)}/><CopyButton text={docs} label="Copy URL only"/></div>
      </Step>
      <Step n={2} done={steps[1]} current={current === 1} title="Approve the pairing">
        {o ? <>Connected as <code>{o.id}</code>{o.target ? <>, bound to <code>{o.target}</code></> : ', using the protocol defaults'}.</>
          : pairings.length > 0 ? <><p>The agent asked to pair. Check the code matches the one the agent shows you, then approve. You can bind a Connection and allow change requests here.</p><PairingList pairings={pairings} targets={targets} onDone={done}/></>
          : <>Waiting for the agent to ask. Its request appears here within a few seconds, with a short code; requests expire after 10 minutes.</>}
      </Step>
      <Step n={3} done={steps[2]} current={current === 2} title="First call arrives">
        {!o ? 'After pairing, the agent makes one real call through aiport.'
          : seen === 'loading' ? 'Checking…'
          : seen === 'off' ? <>Activity recording is off (<code>telemetry.enabled: false</code>), so aiport cannot tell. Ask the agent whether its test call succeeded.</>
          : typeof seen === 'number' ? <>Latest call {new Date(seen).toLocaleString()}. The application now goes through aiport.</>
          : 'No call from this application in the last 7 days yet. If its calls show as unknown in Activity, it is not sending its own token.'}
      </Step>
      <Step n={4} done={steps[3]} current={current === 3} title="Allow change requests (optional)">
        {!o ? 'Once paired, decide whether coding agents in its repository may ask you for changes.'
          : !adapting ? <>Adapting is turned off in the configuration (<code>adapting.disabled</code>), so no Application can request changes now.</>
          : <><MayRequestToggle o={o} onDone={done}/> You still decide each request on the Changes page.</>}
      </Step>
      <Step n={5} done={steps[4]} current={current === 4} title="MCP entry in the repository (optional)">
        <label className="check-label compact"><input type="checkbox" checked={mcpCopied} onChange={e => setMcpCopied(e.target.checked)}/>The agent adds it as part of step 1. To do it by hand, copy it from the <button type="button" className="text-button" onClick={onAgentTab}>Coding agent tab</button>.</label>
      </Step>
    </ol>
    {current < 0 && <p className="notice" role="status"><CircleCheck size={16} aria-hidden/> This application is onboarded.</p>}
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
    <Wizard origins={origins} targets={targets} app={app} setApp={setApp} adapting={doc.adapting} prompt={doc.agent_prompt} docs={doc.docs} mcpCopied={copied} setMcpCopied={markCopied} tick={tick} onChange={onChange} onAgentTab={() => setTab('agent')}/>
    <p className="muted integrate-intro"><BookOpen size={14} aria-hidden/> The two briefs below are what the agent reads, shown here so you can check them or copy parts by hand. aiport writes them from this gateway's live address and protocols (<code>{doc.protocols.map(p => p.name + (p.available ? ' configured' : ' not configured')).join(', ')}</code>). <button type="button" className="text-button" onClick={onManual}>Read the manual</button></p>
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
