import React, {useEffect, useState} from 'react';
import {ArrowRight, ThumbsDown, ThumbsUp, Undo2} from 'lucide-react';

type Votes = {calls: number; rated: number; up: number; down: number; up_ratio: number | null};
type Row = Votes & {name: string; connection: string; model: string; effort: string};
type Module = Votes & {module: string; rows: Row[]};
type App = Votes & {application: string; modules: Module[]};
type Served = {connection: string; model: string};
type Suggestion = {application: string; module: string; name: string; from: Served; up_ratio: number; rated: number; to: Served; basis: 'evidence' | 'escalation' | 'cheaper'; why: string};
type ModuleRoute = {module: string; name: string; target: string; model: string};
type Report = {hours: number; applications: App[]; suggestions: Suggestion[]; module_routes: Record<string, ModuleRoute[]>; feedback_path: string; module_header: string; telemetry: {state: string}};

// Below this many ratings a share of up-votes says little: it is shown, muted.
const FEW = 5;
const PERIODS: [number, string][] = [[168, '7 days'], [720, '30 days'], [2160, '90 days']];
const num = (v: number) => new Intl.NumberFormat().format(v);

function tone(v: Votes): string {
  if (v.up_ratio == null) return 'none';
  if (v.rated < FEW) return 'few';
  return v.up_ratio >= 0.75 ? 'good' : v.up_ratio < 0.5 ? 'bad' : 'mixed';
}

function Share({v}: {v: Votes}) {
  const t = tone(v);
  if (v.up_ratio == null) return <span className="vote-share none">No votes</span>;
  const pct = Math.round(v.up_ratio * 100);
  return <span className={'vote-share ' + t} title={`${v.up} up, ${v.down} down of ${v.rated} rated` + (t === 'few' ? ` (fewer than ${FEW} votes: too few to judge)` : '')}>
    <span className="vote-track" aria-hidden><i style={{width: pct + '%'}}/></span>
    <b>{pct}%</b>{t === 'few' && <small>few votes</small>}
  </span>;
}

function Counts({v}: {v: Votes}) {
  return <span className="vote-counts">
    <span title="Thumbs up"><ThumbsUp size={13} aria-label="up"/> {num(v.up)}</span>
    <span title="Thumbs down"><ThumbsDown size={13} aria-label="down"/> {num(v.down)}</span>
  </span>;
}

const BASIS: Record<Suggestion['basis'], [string, string]> = {
  evidence: ['Better elsewhere', 'Another application or model does clearly better on a module of the same name.'],
  escalation: ['One Level up', 'No better evidence yet: the next Level\'s model, untested for this module.'],
  cheaper: ['Cheaper may do', 'Rated very highly: the Level below may do as well for less, untested for this module.'],
};

async function routeModule(app: string, module: string, name: string, body: object | null) {
  const r = await fetch(`/api/v1/module-routes/${encodeURIComponent(app)}/${encodeURIComponent(module)}/${encodeURIComponent(name)}`, {
    method: body ? 'PUT' : 'DELETE', headers: {'Content-Type': 'application/json', 'X-Aiport-CSRF': '1'}, body: body ? JSON.stringify(body) : undefined,
  });
  if (!r.ok) throw new Error((await r.json().catch(() => null))?.error?.message || 'HTTP ' + r.status);
}

function Suggestions({items, onDone}: {items: Suggestion[]; onDone: () => void}) {
  const [busy, setBusy] = useState(''), [error, setError] = useState('');
  if (!items.length) return null;
  const apply = async (s: Suggestion) => {
    const key = s.application + '/' + s.module + '/' + s.name;
    setBusy(key); setError('');
    try { await routeModule(s.application, s.module, s.name, {connection: s.to.connection, model: s.to.model, why: 'from Feedback: ' + s.why}); onDone(); }
    catch (e) { setError(String((e as Error).message || e)); }
    finally { setBusy(''); }
  };
  return <section className="panel" aria-labelledby="suggest-heading">
    <div className="panel-heading"><h2 id="suggest-heading">Suggestions from the votes</h2><span>{items.length === 1 ? '1 suggestion' : items.length + ' suggestions'}</span></div>
    <p className="muted feedback-intro">aiport never moves a module by itself. Apply routes only that module's calls for that name to the suggested model, as a Change you can revert on Changes; then watch the module's share of up-votes here.</p>
    {error && <div className="notice error feedback-notice" role="alert">{error}</div>}
    <ul className="suggestions">
      {items.map(s => { const key = s.application + '/' + s.module + '/' + s.name; return <li key={key} className={'suggestion ' + s.basis}>
        <div className="suggestion-main">
          <div className="suggestion-title"><b>{s.application}</b> · <code>{s.module}</code> sending <code>{s.name}</code>
            <span className={'badge ' + (s.basis === 'cheaper' ? 'good' : s.basis === 'evidence' ? 'info' : 'warn')} title={BASIS[s.basis][1]}>{BASIS[s.basis][0]}</span></div>
          <div className="suggestion-move">{s.from.model} <small className="muted">on {s.from.connection}</small> <ArrowRight size={14} aria-hidden/> <b>{s.to.model}</b> <small className="muted">on {s.to.connection}</small></div>
          <p className="muted-small">{s.why}.</p>
        </div>
        <button type="button" className="primary" disabled={busy === key} onClick={() => apply(s)}>{busy === key ? 'Applying…' : 'Apply'}</button>
      </li>; })}
    </ul>
  </section>;
}

function AppPanel({app, routes, onChange}: {app: App; routes: ModuleRoute[]; onChange: () => void}) {
  const rated = app.modules.filter(s => s.module && s.rated >= FEW && s.up_ratio != null);
  const best = rated.length > 1 ? rated.reduce((a, b) => (b.up_ratio! > a.up_ratio! ? b : a)) : null;
  const worst = rated.length > 1 ? rated.reduce((a, b) => (b.up_ratio! < a.up_ratio! ? b : a)) : null;
  return <section className="panel feedback-app" aria-label={'Votes of ' + app.application}>
    <div className="panel-heading">
      <h2>{app.application}</h2>
      <span className="feedback-summary"><Counts v={app}/><Share v={app}/><small>{num(app.rated)} of {num(app.calls)} calls rated</small></span>
    </div>
    {best && worst && best !== worst && <p className="feedback-callout">
      Best module: <b>{best.module}</b> ({Math.round(best.up_ratio! * 100)}% up). Weakest: <b>{worst.module}</b> ({Math.round(worst.up_ratio! * 100)}% up): try another model or effort for it under Routing.
    </p>}
    <table className="feedback-table">
      <thead><tr><th>Module</th><th>Sent as</th><th>Model</th><th>Effort</th><th className="n">Calls</th><th className="n">Votes</th><th>Up</th></tr></thead>
      <tbody>
        {app.modules.map(s => <React.Fragment key={s.module || '(none)'}>
          <tr className={'section-row ' + tone(s)}>
            <th scope="row" colSpan={4}>{s.module ? <code>{s.module}</code> : <span className="muted">No module named</span>}{s.rows.length > 1 && <small className="muted"> · {s.rows.length} combinations</small>}</th>
            <td className="n">{num(s.calls)}</td><td className="n"><Counts v={s}/></td><td><Share v={s}/></td>
          </tr>
          {s.rows.map(r => <tr key={r.name + '|' + r.model + '|' + r.effort} className="combo-row">
            <td/><td><code>{r.name}</code></td><td>{r.model}</td><td>{r.effort || '—'}</td>
            <td className="n">{num(r.calls)}</td><td className="n"><Counts v={r}/></td><td><Share v={r}/></td>
          </tr>)}
        </React.Fragment>)}
      </tbody>
    </table>
    {routes.length > 0 && <div className="module-routes">
      <h3>Module routes</h3>
      <ul>{routes.map(r => <li key={r.module + '/' + r.name}><code>{r.module}</code> sending <code>{r.name}</code> <ArrowRight size={13} aria-hidden/> <b>{r.model}</b> <small className="muted">on {r.target}</small>
        <button type="button" className="text-button" onClick={() => routeModule(app.application, r.module, r.name, null).then(onChange, () => onChange())}><Undo2 size={13} aria-hidden/> Remove</button></li>)}</ul>
    </div>}
    {app.modules.some(s => !s.module) && <p className="caption">Calls without a module did not send <code>X-Aiport-Module</code>. Guidance version 4 asks the application's coding agent to add it.</p>}
  </section>;
}

export default function Feedback({tick, onChange}: {tick: number; onChange?: () => void}) {
  const [hours, setHours] = useState(720);
  const [report, setReport] = useState<Report | null>(null);
  const [error, setError] = useState(''), [reload, setReload] = useState(0);
  const changed = () => { setReload(x => x + 1); onChange?.(); };
  useEffect(() => {
    fetch('/api/v1/feedback?hours=' + hours, {headers: {'X-Aiport-CSRF': '1'}})
      .then(async r => { if (!r.ok) throw new Error((await r.json().catch(() => null))?.error?.message || 'HTTP ' + r.status); return r.json(); })
      .then((r: Report) => { setReport(r); setError(''); })
      .catch(e => setError(String(e.message || e)));
  }, [hours, tick, reload]);
  const apps = report?.applications || [];
  const anyVotes = apps.some(a => a.rated > 0);
  return <>
    <section className="panel">
      <div className="panel-heading">
        <h2>How people rate the answers</h2>
        <div className="segmented" role="group" aria-label="Period">
          {PERIODS.map(([h, label]) => <button key={h} type="button" className={h === hours ? 'on' : ''} aria-pressed={h === hours} onClick={() => setHours(h)}>{label}</button>)}
        </div>
      </div>
      <p className="muted feedback-intro">Each application names the module of every call (<code>{report?.module_header || 'X-Aiport-Module'}</code>, such as tldr, classify or insights) and sends its users' thumbs up and down to <code>{report?.feedback_path || '/v1/aiport/feedback'}</code>. Read each module's share of up-votes by the name it sent, the model that served it and the effort; when a module does badly, change its Level, task name or effort under Routing. Modules with fewer than {FEW} votes are too few to judge.</p>
      {error && <div className="notice error feedback-notice" role="alert">{error}</div>}
      {report && report.telemetry.state === 'disabled' && <div className="notice"><span className="dot amber"/>Request history is off, so no calls or votes are kept. Turn it on in Settings.</div>}
      {report && !apps.length && <div className="quiet-empty feedback-empty">No model calls in this period.</div>}
      {report && apps.length > 0 && !anyVotes && <div className="quiet-empty feedback-empty">No votes yet.<br/><small>Applications start sending them once their coding agent follows guidance version 4 (Integrate → Guidance).</small></div>}
    </section>
    <Suggestions items={report?.suggestions || []} onDone={changed}/>
    {apps.map(a => <AppPanel key={a.application} app={a} routes={report?.module_routes?.[a.application] || []} onChange={changed}/>)}
  </>;
}
