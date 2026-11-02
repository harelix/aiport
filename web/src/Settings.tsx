import React, {FormEvent, useState} from 'react';
import {Activity, Check, ChevronRight, Cloud, Copy, HardDrive, Plus} from 'lucide-react';
import {AuthenticationSettings} from './AuthGate';
import {ProviderLogo} from './ProviderLogo';
import type {Reason} from './Reason';
import {ApplicationsPanel} from './Applications';
import KeyField, {KeySource, keyBadge} from './KeyField';

export type Location = 'local' | 'remote';
export type Target = {id: string; provider: string; protocol: string; base_url?: string; url: string; endpoint?: string; region?: string; api_key_env?: string; disabled: boolean; location: Location};
type EndpointSpec = {id: string; name: string; urls: Record<string, string>; note?: string};
export type ProviderSpec = {id: string; name: string; protocol: string; api_key_env?: string; location: Location; regional: boolean; endpoints: EndpointSpec[]; note?: string};
/* The host:port aiport hands out: the first configured hostname on the
   listen port, otherwise the listen address (as config.Address in Go). */
export function gatewayAddress(config?: {listen: string; hostnames?: string[]} | null): string | undefined {
  if (!config) return undefined;
  const name = config.hostnames?.[0];
  return name ? name + ':' + config.listen.slice(config.listen.lastIndexOf(':') + 1) : config.listen;
}
export type SettingsConfig = {
  auth: {username?: string; disabled: boolean}; listen: string; hostnames?: string[]; targets: Target[];
  origins: {id: string; name: string; target: string; token_env: string; aliases?: {alias: string; target: string; model: string}[]}[];
  defaults: {openai_target?: string; anthropic_target?: string};
  evaluation?: {connection?: string; model?: string};
};
type KeyResult = {secret_status: Record<string, boolean>; secret_source: Record<string, KeySource>};

type Props = {
  config: SettingsConfig | null; providers: ProviderSpec[];
  secrets: Record<string, boolean>; sources: Record<string, KeySource>;
  onKeys: (r: KeyResult) => void; onChange: () => void;
  queueDepth?: number;
};

const GROUPS: Record<Location, {title: string; icon: typeof Cloud; blurb: string}> = {
  local: {title: 'Local', icon: HardDrive, blurb: 'Model servers on this machine. No API key; requests never leave your computer.'},
  remote: {title: 'Remote APIs', icon: Cloud, blurb: 'Hosted providers. Each needs an API key; requests are sent to the provider over HTTPS.'},
};

function CopyBlock({text, label}: {text: string; label: string}) {
  const [copied, setCopied] = useState(false);
  return <div className="copy-block">
    <pre>{text}</pre>
    <button type="button" className="icon-button" aria-label={'Copy ' + label} title="Copy" onClick={() =>
      void navigator.clipboard?.writeText(text).then(() => { setCopied(true); setTimeout(() => setCopied(false), 1200); })}>
      {copied ? <Check size={14} aria-hidden/> : <Copy size={14} aria-hidden/>}
    </button>
  </div>;
}

const REGIONS = ['us-east-1', 'us-east-2', 'us-west-2', 'ca-central-1', 'sa-east-1', 'eu-central-1', 'eu-west-1', 'eu-west-2', 'eu-west-3', 'eu-north-1', 'eu-south-1', 'ap-south-1', 'ap-northeast-1', 'ap-northeast-2', 'ap-southeast-1', 'ap-southeast-2'];
const protoName = (p: string) => p === 'openai' ? 'OpenAI' : 'Anthropic';
const protocolsOf = (p: ProviderSpec) => [...new Set(p.endpoints.flatMap(e => Object.keys(e.urls)))];

/** True when a preset has anything to choose: region, endpoint or protocol. */
const configurable = (spec?: ProviderSpec) => !!spec && (spec.regional || spec.endpoints.length > 1 || protocolsOf(spec).length > 1);

/** Sends a connection change; the server edits the YAML file and applies it. */
async function mutate(method: string, path: string, body?: object) {
  const response = await fetch('/api/v1/' + path, {
    method, credentials: 'same-origin',
    headers: {'Content-Type': 'application/json', 'X-Aiport-CSRF': '1'},
    body: body ? JSON.stringify(body) : undefined,
  });
  const data = await response.json().catch(() => ({}));
  if (!response.ok) throw new Error(data.error?.message || 'The change was not saved (' + response.status + ').');
  return data;
}

/** Result of POST /api/v1/targets/{id}/check (a free GET /models probe). */
type Check = {ok: boolean; state: string; http_status?: number; models?: number; latency_ms: number; message: string; checked_at_ms: number; connection_error?: string; upstream_error_type?: string; reason?: Reason};
type CheckState = Check | 'running' | {error: string};

const checkTone = (c: Check) => c.state === 'ok' ? 'good' : c.ok ? 'warn' : 'bad';
const checkLabel = (c: Check) => c.state === 'ok' ? 'Healthy' : c.state === 'no_model_list' ? 'Reachable' : c.state === 'no_key' ? 'Needs a key' : c.state === 'disabled' ? 'Disabled' : 'Unhealthy';

/** Shows the last connection check for one card. */
function CheckResult({state}: {state?: CheckState}) {
  if (!state) return null;
  if (state === 'running') return <p className="check-result" role="status">Checking…</p>;
  if ('error' in state) return <p className="check-result" role="alert"><span className="badge bad">Check failed</span> {state.error}</p>;
  return <p className="check-result" role="status"><span className={'badge ' + checkTone(state)}>{checkLabel(state)}</span> {state.message}{state.models !== undefined && ' ' + state.models + ' model' + (state.models === 1 ? '' : 's') + ' listed.'} <small>{state.latency_ms} ms · {new Date(state.checked_at_ms).toLocaleTimeString()}</small>{state.reason && <span className="check-reason">{state.reason.explanation} {state.reason.fix && <><b>What to do:</b> {state.reason.fix}</>}</span>}</p>;
}

/** Result of GET /api/v1/targets/{id}/models (the same free GET /models listing, cached briefly). */
type ModelList = {connection: string; listed: boolean; state: string; models: string[]; listed_at_ms: number};

/** "N models" on a Connection card, with the list on request. */
function ConnectionModels({t, check}: {t: Target; check?: CheckState}) {
  const [list, setList] = useState<ModelList | null>(null), [open, setOpen] = useState(false), [busy, setBusy] = useState(false), [error, setError] = useState('');
  async function load(refresh: boolean) {
    setBusy(true); setError('');
    try { setList(await mutate('GET', 'targets/' + encodeURIComponent(t.id) + '/models' + (refresh ? '?refresh=1' : '')) as ModelList); setOpen(true); }
    catch (err) { setError((err as Error).message); } finally { setBusy(false); }
  }
  if (t.disabled) return null;
  const checked = check && typeof check === 'object' && !('error' in check) ? check.models : undefined;
  const count = list?.listed ? list.models.length : checked;
  const label = count === undefined ? 'Models' : count + ' model' + (count === 1 ? '' : 's');
  return <div className="conn-models">
    <button type="button" className="text-button" aria-expanded={open} disabled={busy} onClick={() => open ? setOpen(false) : load(false)} title="Lists this connection's models with GET /models. Generates nothing and costs nothing.">{busy ? 'Listing models…' : open ? 'Hide models' : list || count !== undefined ? 'Show ' + label.toLowerCase() : 'Show models'}</button>
    {error && <p className="auth-error" role="alert">{error}</p>}
    {open && list && (list.listed
      ? <div className="model-list">
          <p className="caption">{label} · listed {new Date(list.listed_at_ms).toLocaleTimeString()} <button type="button" className="text-button" disabled={busy} onClick={() => load(true)}>Refresh</button></p>
          {list.models.length ? <ul>{list.models.map(m => <li key={m}><code>{m}</code></li>)}</ul> : <p className="caption">This connection lists no models.</p>}
        </div>
      : <p className="caption" role="status">The model list could not be read ({list.state.replaceAll('_', ' ')}). aiport treats its models as unknown. Use <b>Check connection</b> to see why. <button type="button" className="text-button" disabled={busy} onClick={() => load(true)}>Try again</button></p>)}
  </div>;
}

/** Region / endpoint / protocol fields shared by the add and edit forms. */
function useConnectionFields(spec: ProviderSpec, idPrefix: string, initial?: {region?: string; endpoint?: string; protocol?: string}) {
  const [region, setRegion] = useState(initial?.region || ''), [endpoint, setEndpoint] = useState(initial?.endpoint || spec.endpoints[0].id), [protocol, setProtocol] = useState(initial?.protocol || spec.protocol);
  const ep = spec.endpoints.find(e => e.id === endpoint) || spec.endpoints[0];
  const protocols = Object.keys(ep.urls);
  const effective = protocols.includes(protocol) ? protocol : protocols[0];
  const url = (ep.urls[effective] || '').replaceAll('{region}', region || '{region}');
  // Preset defaults are sent as empty so the YAML keeps only what differs.
  const value = {region, endpoint: ep.id === spec.endpoints[0].id ? '' : ep.id, protocol: effective === spec.protocol ? '' : effective};
  const fields = <>
    {spec.regional && <label>Region
      <input list={idPrefix + '-regions'} value={region} onChange={e => setRegion(e.target.value.trim())} placeholder="us-east-1" required pattern="[a-z]{2,4}(-[a-z0-9]+){1,3}" maxLength={32} spellCheck={false} autoComplete="off"/>
      <datalist id={idPrefix + '-regions'}>{REGIONS.map(r => <option key={r} value={r}/>)}</datalist>
    </label>}
    {spec.endpoints.length > 1 && <fieldset>
      <legend>Endpoint</legend>
      <div className="segmented" role="radiogroup" style={{gridTemplateColumns: `repeat(${spec.endpoints.length},1fr)`}}>
        {spec.endpoints.map((x, i) => <button type="button" key={x.id} role="radio" aria-checked={ep.id === x.id} className={ep.id === x.id ? 'on' : ''} onClick={() => setEndpoint(x.id)}>{x.name}{i === 0 ? ' · default' : ''}</button>)}
      </div>
      {ep.note && <small>{ep.note}</small>}
    </fieldset>}
    {protocols.length > 1 && <label>Protocol
      <select value={effective} onChange={e => setProtocol(e.target.value)}>{protocols.map(p => <option key={p} value={p}>{protoName(p)}{p === 'anthropic' ? ' (Messages API · Claude Code)' : ' (Chat Completions, Responses)'}</option>)}</select>
    </label>}
    <div className="conn-preview"><span>Requests go to</span><code>{url}</code></div>
  </>;
  return {fields, value, protocol: effective, ready: !spec.regional || !!region};
}

function Status({error, message}: {error: string; message: string}) {
  return <>{error && <p className="auth-error" role="alert">{error}</p>}{message && <p className="auth-success" role="status">{message}</p>}</>;
}

function EditConnection({t, spec, onSaved}: {t: Target; spec: ProviderSpec; onSaved: () => void}) {
  const c = useConnectionFields(spec, 'edit-' + t.id, {region: t.region, endpoint: t.endpoint, protocol: t.protocol});
  const [busy, setBusy] = useState(false), [error, setError] = useState(''), [message, setMessage] = useState('');
  const dirty = c.value.region !== (t.region || '') || c.value.endpoint !== (t.endpoint || '') || c.protocol !== t.protocol;
  async function save(e: FormEvent) {
    e.preventDefault(); setBusy(true); setError(''); setMessage('');
    try { await mutate('PATCH', 'targets/' + encodeURIComponent(t.id), c.value); setMessage('Saved. New requests use it now.'); onSaved(); }
    catch (err) { setError((err as Error).message); } finally { setBusy(false); }
  }
  if (t.base_url) return <p className="caption conn-note">This connection uses a custom <code>base_url</code>; edit it in the YAML file.</p>;
  return <form className="conn-form" onSubmit={save}>
    {c.fields}<Status error={error} message={message}/>
    <button className="secondary" disabled={busy || !dirty || !c.ready}>{busy ? 'Saving…' : 'Save connection'}</button>
  </form>;
}

/** Picks a free id: bedrock, bedrock-eu-west-1, bedrock-2 … */
function freeId(base: string, taken: Set<string>) {
  if (!taken.has(base)) return base;
  for (let i = 2; ; i++) if (!taken.has(base + '-' + i)) return base + '-' + i;
}

function AddPreset({spec, taken, onDone}: {spec: ProviderSpec; taken: Set<string>; onDone: () => void}) {
  const c = useConnectionFields(spec, 'add-' + spec.id);
  const suggested = freeId(spec.regional && c.value.region && taken.has(spec.id) ? spec.id + '-' + c.value.region : spec.id, taken);
  const [id, setId] = useState(''), [busy, setBusy] = useState(false), [error, setError] = useState('');
  async function add(e: FormEvent) {
    e.preventDefault(); setBusy(true); setError('');
    try { await mutate('POST', 'targets', {id: id || suggested, provider: spec.id, ...c.value}); onDone(); }
    catch (err) { setError((err as Error).message); } finally { setBusy(false); }
  }
  return <form className="conn-form add-form" onSubmit={add}>
    {spec.note && <small>{spec.note}</small>}
    <label>Connection id <small>Shown in activity and used in routing</small>
      <input value={id} onChange={e => setId(e.target.value.trim())} placeholder={suggested} pattern="[a-zA-Z0-9][a-zA-Z0-9_-]{0,63}" maxLength={64} spellCheck={false} autoComplete="off"/>
    </label>
    {c.fields}<Status error={error} message=""/>
    <button className="primary" disabled={busy || !c.ready}>{busy ? 'Adding…' : 'Add ' + spec.name}</button>
    {spec.api_key_env && <small>After adding, enter the API key on its card.</small>}
  </form>;
}

function AddCustom({location, taken, onDone}: {location: Location; taken: Set<string>; onDone: () => void}) {
  const [id, setId] = useState(''), [provider, setProvider] = useState(''), [protocol, setProtocol] = useState('openai');
  const [url, setUrl] = useState(location === 'local' ? 'http://127.0.0.1:8000/v1' : 'https://'), [keyEnv, setKeyEnv] = useState('');
  const [busy, setBusy] = useState(false), [error, setError] = useState('');
  async function add(e: FormEvent) {
    e.preventDefault(); setBusy(true); setError('');
    try { await mutate('POST', 'targets', {id: id || provider, provider, protocol, base_url: url, api_key_env: keyEnv}); onDone(); }
    catch (err) { setError((err as Error).message); } finally { setBusy(false); }
  }
  return <form className="conn-form add-form" onSubmit={add}>
    <small>Any server that speaks the OpenAI or Anthropic API natively. aiport does not translate between them.</small>
    <div className="form-row">
      <label>Provider name<input value={provider} onChange={e => setProvider(e.target.value.trim())} placeholder={location === 'local' ? 'vllm' : 'together'} required pattern="[a-zA-Z0-9][a-zA-Z0-9_-]{0,63}" spellCheck={false}/></label>
      <label>Connection id<input value={id} onChange={e => setId(e.target.value.trim())} placeholder={provider ? freeId(provider, taken) : 'same as name'} pattern="[a-zA-Z0-9][a-zA-Z0-9_-]{0,63}" spellCheck={false}/></label>
    </div>
    <label>Base URL <small>API prefix, usually ending in /v1. Remote URLs must use HTTPS.</small><input value={url} onChange={e => setUrl(e.target.value.trim())} required type="url" spellCheck={false}/></label>
    <div className="form-row">
      <label>Protocol<select value={protocol} onChange={e => setProtocol(e.target.value)}><option value="openai">OpenAI</option><option value="anthropic">Anthropic</option></select></label>
      <label>Key variable <small>Optional; name only, e.g. TOGETHER_API_KEY</small><input value={keyEnv} onChange={e => setKeyEnv(e.target.value.trim().toUpperCase())} pattern="[A-Za-z_][A-Za-z0-9_]*" spellCheck={false}/></label>
    </div>
    <Status error={error} message=""/>
    <button className="primary" disabled={busy}>{busy ? 'Adding…' : 'Add connection'}</button>
  </form>;
}

function CardActions({t, defaults, onChange}: {t: Target; defaults: SettingsConfig['defaults']; onChange: () => void}) {
  const [confirm, setConfirm] = useState(false), [busy, setBusy] = useState(false), [error, setError] = useState('');
  const key = t.protocol === 'openai' ? 'openai_target' : 'anthropic_target';
  const isDefault = defaults[key] === t.id;
  async function run(fn: () => Promise<unknown>) {
    setBusy(true); setError('');
    try { await fn(); onChange(); } catch (err) { setError((err as Error).message); } finally { setBusy(false); setConfirm(false); }
  }
  return <div className="card-actions">
    {!isDefault && !t.disabled && <button type="button" className="text-button" disabled={busy} onClick={() => run(() => mutate('PUT', 'defaults', {...defaults, [key]: t.id}))}>Make {protoName(t.protocol)} default</button>}
    {confirm
      ? <span className="confirm">Remove {t.id}? <button type="button" className="danger" disabled={busy} onClick={() => run(() => mutate('DELETE', 'targets/' + encodeURIComponent(t.id)))}>Remove</button><button type="button" className="text-button" onClick={() => setConfirm(false)}>Cancel</button></span>
      : <button type="button" className="text-button subtle" onClick={() => setConfirm(true)}>Remove</button>}
    {error && <p className="auth-error" role="alert">{error}</p>}
  </div>;
}

function ProviderCard({t, spec, src, role, defaults, onKeys, onChange, check, onCheck}: {t: Target; spec?: ProviderSpec; src: KeySource; role: string; defaults: SettingsConfig['defaults']; onKeys: (r: KeyResult) => void; onChange: () => void; check?: CheckState; onCheck: () => void}) {
  const [editing, setEditing] = useState(false);
  const name = spec?.name || t.provider;
  const [text, tone] = t.disabled ? ['Disabled', ''] : keyBadge[src];
  return <article className={'provider' + (t.disabled ? ' off' : '')}>
    <header className="provider-title">
      <span className="provider-icon" aria-hidden><ProviderLogo provider={t.provider} location={t.location} size={20}/></span>
      <div><h3>{name}</h3><small>{t.id}</small></div>
      <span className={'badge ' + tone}>{text}</span>
    </header>
    {role && <span className="chip">Default · {role}</span>}
    <dl className="facts">
      <dt>URL</dt><dd><code>{t.url}</code></dd>
      <dt>Protocol</dt><dd>{protoName(t.protocol)}</dd>
      {t.region && <><dt>Region</dt><dd>{t.region}</dd></>}
      {spec && spec.endpoints.length > 1 && <><dt>Endpoint</dt><dd>{(spec.endpoints.find(e => e.id === t.endpoint) || spec.endpoints[0]).name}</dd></>}
      {t.api_key_env && <><dt>Key variable</dt><dd><code>{t.api_key_env}</code></dd></>}
      {!spec && <><dt>Preset</dt><dd>Custom provider</dd></>}
    </dl>
    {configurable(spec) && <div className="conn-edit">
      <button type="button" className="text-button" aria-expanded={editing} onClick={() => setEditing(v => !v)}>{editing ? 'Hide connection settings' : 'Edit region, endpoint and protocol'}</button>
      {editing && <EditConnection t={t} spec={spec!} onSaved={onChange}/>}
    </div>}
    {t.api_key_env && <KeyField target={t.id} envName={t.api_key_env} source={src} onChange={onKeys}/>}
    <div className="conn-check">
      <button type="button" className="secondary with-icon" disabled={check === 'running' || t.disabled} onClick={onCheck} title="Sends GET /models with this connection's key. Generates nothing and costs nothing."><Activity size={14} aria-hidden/>Check connection</button>
      <CheckResult state={check}/>
      <ConnectionModels t={t} check={check}/>
    </div>
    <CardActions t={t} defaults={defaults} onChange={onChange}/>
  </article>;
}

function ProviderGroup({location, targets, presets, providers, taken, sources, secrets, defaults, onKeys, onChange}: {
  location: Location; targets: Target[]; presets: ProviderSpec[]; providers: ProviderSpec[]; taken: Set<string>;
  sources: Record<string, KeySource>; secrets: Record<string, boolean>;
  defaults: SettingsConfig['defaults']; onKeys: (r: KeyResult) => void; onChange: () => void;
}) {
  const g = GROUPS[location], Icon = g.icon;
  const [adding, setAdding] = useState('');
  // Remote APIs start collapsed; Local starts open.
  const [open, setOpen] = useState(location !== 'remote');
  const [checks, setChecks] = useState<Record<string, CheckState>>({});
  async function runCheck(id: string) {
    setChecks(c => ({...c, [id]: 'running'}));
    try { const r = await mutate('POST', 'targets/' + encodeURIComponent(id) + '/check') as Check; setChecks(c => ({...c, [id]: r})); }
    catch (err) { setChecks(c => ({...c, [id]: {error: (err as Error).message}})); }
  }
  const checkable = targets.filter(t => !t.disabled);
  const checkAll = () => { setOpen(true); checkable.forEach(t => runCheck(t.id)); };
  const unhealthy = Object.values(checks).filter(c => typeof c === 'object' && ('error' in c || !c.ok)).length;
  const source = (t: Target): KeySource => sources[t.id] || (t.api_key_env ? (secrets[t.id] ? 'environment' : 'missing') : 'not_needed');
  const missing = targets.filter(t => !t.disabled && source(t) === 'missing').length;
  const role = (t: Target) => [defaults.openai_target === t.id && 'OpenAI protocol', defaults.anthropic_target === t.id && 'Anthropic protocol'].filter(Boolean).join(' · ');
  const count = (id: string) => targets.filter(t => t.provider === id).length;
  const done = () => { setAdding(''); onChange(); };
  const body = 'group-body-' + location;
  return <section className={'panel provider-group' + (open ? '' : ' collapsed')} aria-labelledby={'group-' + location}>
    <div className="group-heading">
      <span className="group-icon" aria-hidden><Icon size={18} strokeWidth={1.75}/></span>
      <div>
        <h2 id={'group-' + location}><button type="button" className="group-toggle" aria-expanded={open} aria-controls={body} onClick={() => setOpen(o => !o)}><ChevronRight size={16} strokeWidth={2} aria-hidden className="chevron"/>{g.title} <span className="count">{targets.length}</span></button></h2>
        <p>{g.blurb}</p>
      </div>
      <div className="group-actions">
        {missing > 0 && <span className="badge warn">{missing} need{missing === 1 ? 's' : ''} a key</span>}
        {unhealthy > 0 && <span className="badge bad">{unhealthy} unhealthy</span>}
        {checkable.length > 0 && <button type="button" className="text-button" onClick={checkAll} disabled={Object.values(checks).includes('running')}>Check all</button>}
      </div>
    </div>
    {open && <div id={body}>
    {targets.length
      ? <div className="provider-grid">{targets.map(t => <ProviderCard key={t.id} t={t} spec={providers.find(p => p.id === t.provider)} src={source(t)} role={role(t)} defaults={defaults} onKeys={onKeys} onChange={onChange} check={checks[t.id]} onCheck={() => runCheck(t.id)}/>)}</div>
      : <p className="group-empty">No {location === 'local' ? 'local model servers' : 'remote providers'} yet. Add one below.</p>}
    <div className="available">
      <h3>Add a connection</h3>
      <ul>
        {presets.map(p => <li key={p.id}>
          <button type="button" className={'add-row' + (adding === p.id ? ' open' : '')} aria-expanded={adding === p.id} onClick={() => setAdding(a => a === p.id ? '' : p.id)}>
            <Plus size={14} aria-hidden/><ProviderLogo provider={p.id} location={location} size={16}/><strong>{p.name}</strong>{count(p.id) > 0 && <em className="count">{count(p.id)} added</em>}
            <span>{[p.regional && 'Region', p.endpoints.length > 1 && p.endpoints.length + ' endpoints', p.api_key_env ? 'API key' : 'No key', protocolsOf(p).map(protoName).join(' or ') + ' protocol'].filter(Boolean).join(' · ')}</span>
          </button>
          {adding === p.id && <AddPreset spec={p} taken={taken} onDone={done}/>}
        </li>)}
        <li>
          <button type="button" className={'add-row' + (adding === 'custom' ? ' open' : '')} aria-expanded={adding === 'custom'} onClick={() => setAdding(a => a === 'custom' ? '' : 'custom')}>
            <Plus size={14} aria-hidden/><strong>Custom {location === 'local' ? 'local server' : 'provider'}</strong><span>Your own URL · OpenAI or Anthropic protocol</span>
          </button>
          {adding === 'custom' && <AddCustom location={location} taken={taken} onDone={done}/>}
        </li>
      </ul>
    </div>
    </div>}
  </section>;
}

function DefaultsForm({config, providers, onChange}: {config: SettingsConfig | null; providers: ProviderSpec[]; onChange: () => void}) {
  const [openai, setOpenai] = useState(config?.defaults.openai_target || ''), [anthropic, setAnthropic] = useState(config?.defaults.anthropic_target || '');
  const [busy, setBusy] = useState(false), [error, setError] = useState(''), [message, setMessage] = useState('');
  const targets = config?.targets || [];
  const label = (t: Target) => (providers.find(p => p.id === t.provider)?.name || t.provider) + ' · ' + t.id;
  const dirty = openai !== (config?.defaults.openai_target || '') || anthropic !== (config?.defaults.anthropic_target || '');
  async function save(e: FormEvent) {
    e.preventDefault(); setBusy(true); setError(''); setMessage('');
    try { await mutate('PUT', 'defaults', {openai_target: openai, anthropic_target: anthropic}); setMessage('Saved. New requests use it now.'); onChange(); }
    catch (err) { setError((err as Error).message); } finally { setBusy(false); }
  }
  const pick = (protocol: string, value: string, set: (v: string) => void) =>
    <select value={value} onChange={e => set(e.target.value)}>
      <option value="">None (requests are rejected)</option>
      {targets.filter(t => t.protocol === protocol).map(t => <option key={t.id} value={t.id}>{label(t)}</option>)}
    </select>;
  return <form className="conn-form defaults-form" onSubmit={save}>
    <label>OpenAI protocol <small>/v1/chat/completions, /v1/responses</small>{pick('openai', openai, setOpenai)}</label>
    <label>Anthropic protocol <small>/v1/messages · Claude Code</small>{pick('anthropic', anthropic, setAnthropic)}</label>
    <Status error={error} message={message}/>
    <button className="primary" disabled={busy || !dirty}>{busy ? 'Saving…' : 'Save defaults'}</button>
  </form>;
}

function useLists(config: SettingsConfig | null, providers: ProviderSpec[]) {
  const targets = config?.targets || [];
  const name = (id?: string) => { const t = targets.find(x => x.id === id); return t ? (providers.find(p => p.id === t.provider)?.name || t.provider) + ' (' + t.id + ')' : 'Not configured'; };
  return {targets, name};
}

/** Providers page: connections grouped as Remote APIs and Local. */
export function ProvidersPage({config, providers, secrets, sources, onKeys, onChange}: Props) {
  const {targets} = useLists(config, providers);
  const taken = new Set(targets.map(t => t.id));
  const group = (l: Location) => <ProviderGroup location={l} targets={targets.filter(t => t.location === l)} presets={providers.filter(p => p.location === l)} providers={providers} taken={taken} sources={sources} secrets={secrets} defaults={config?.defaults || {}} onKeys={onKeys} onChange={onChange}/>;
  return <>
    {group('remote')}
    {group('local')}
    <p className="caption settings-note">Keys entered here apply immediately and are stored only in the private credential file next to your configuration (<code>aiport.yaml.env</code>), never in YAML and never shown again. A saved key has not been verified with its provider. Adding, editing and removing connections updates your configuration file in place and applies to new requests immediately, with no restart.</p>
  </>;
}

/** Routing page: protocol defaults and the local endpoint apps connect to. */
export function RoutingPage({config, providers, onChange, queueDepth, onIntegrate}: Props & {onIntegrate?: () => void}) {
  const {targets} = useLists(config, providers);
  return <>
    <section className="panel">
      <div className="panel-heading"><h2>Protocol defaults</h2><span>Where requests go when an app doesn't pick a connection</span></div>
      <DefaultsForm key={JSON.stringify(config?.defaults) + targets.length} config={config} providers={providers} onChange={onChange}/>
    </section>
    <section className="panel">
      <div className="panel-heading"><h2>Connect an application</h2><span>Point any OpenAI or Anthropic SDK here</span></div>
      <dl>
        <dt>OpenAI-protocol base URL</dt><dd><code>http://{gatewayAddress(config)}/v1</code></dd>
        <dt>Anthropic-protocol base URL</dt><dd><code>http://{gatewayAddress(config)}</code></dd>
        <dt>Protocol description</dt><dd><a href="/_connect?format=markdown" target="_blank" rel="noreferrer"><code>/_connect</code></a></dd>
        <dt>Telemetry queue</dt><dd>{queueDepth ?? 0} pending</dd>
      </dl>
    </section>
    <ApplicationsPanel origins={config?.origins || []} targets={targets} onChange={onChange}/>
    <section className="panel">
      <div className="panel-heading"><h2>Instructions for coding agents</h2>{onIntegrate && <button type="button" className="secondary" onClick={onIntegrate}>Open Integrate</button>}</div>
      <p className="muted handover-intro integrate-pointer">The Application brief and the coding-agent brief, with SDK snippets and a live checklist per Application, are on the Integrate page.</p>
    </section>
    <section className="panel planned">
      <div className="panel-heading"><h2>Per-application limits</h2><span className="badge">Planned</span></div>
      <p className="muted">Model allowlists per application and routing by reasoning effort are not built yet. Today an application's calls go to its bound connection or the protocol default; per-application model aliases are set in the YAML file (<code>origins[].aliases</code>).</p>
    </section>
  </>;
}

/** Coding agents page: per-agent origins and how to launch them. */
export function AgentsPage({config, providers}: Props) {
  const {name} = useLists(config, providers);
  return <section className="panel">
    <div className="panel-heading"><h2>Coding-agent connections</h2><span>{config?.origins.length || 0} configured</span></div>
    {config?.origins.length ? config.origins.map(o => <div className="agent" key={o.id}>
      <strong>{o.name}</strong><span>{o.target ? name(o.target) : 'Protocol default'}</span>
      <code>aiport agent {o.id} --model YOUR_MODEL</code>
    </div>) : <div className="quiet-empty">No coding agents connected.<br/><small>Create a setup with <code>aiport init --agent codex,claude</code>.</small></div>}
  </section>;
}

/** Settings page: this aiport install (sign-in, configuration file). */
export default function SettingsPage({config, onChange}: Props) {
  return <>
    <AuthenticationSettings username={config?.auth.username || ''} onChange={onChange}/>
    <section className="panel setup">
      <p className="eyebrow">CONFIGURATION</p>
      <h2>One configuration file. Clear ownership.</h2>
      <p className="muted">Providers and routing are managed from their pages; coding agents, aliases and the listen address live in your YAML file. Validate hand edits with <code>aiport check</code>, then restart. A new setup can be created with:</p>
      <CopyBlock label="setup commands" text={'aiport init --config aiport-new.yaml\naiport check --config aiport-new.yaml\naiport serve --config aiport-new.yaml'}/>
    </section>
  </>;
}
