import React, {useEffect, useState} from 'react';
import {ThumbsDown, ThumbsUp} from 'lucide-react';

type Votes = {calls: number; rated: number; up: number; down: number; up_ratio: number | null};
type Row = Votes & {name: string; model: string; effort: string};
type Section = Votes & {task: string; rows: Row[]};
type App = Votes & {application: string; sections: Section[]};
type Report = {hours: number; applications: App[]; feedback_path: string; task_header: string; telemetry: {state: string}};

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

function AppPanel({app}: {app: App}) {
  const rated = app.sections.filter(s => s.task && s.rated >= FEW && s.up_ratio != null);
  const best = rated.length > 1 ? rated.reduce((a, b) => (b.up_ratio! > a.up_ratio! ? b : a)) : null;
  const worst = rated.length > 1 ? rated.reduce((a, b) => (b.up_ratio! < a.up_ratio! ? b : a)) : null;
  return <section className="panel feedback-app" aria-label={'Votes of ' + app.application}>
    <div className="panel-heading">
      <h2>{app.application}</h2>
      <span className="feedback-summary"><Counts v={app}/><Share v={app}/><small>{num(app.rated)} of {num(app.calls)} calls rated</small></span>
    </div>
    {best && worst && best !== worst && <p className="feedback-callout">
      Best section: <b>{best.task}</b> ({Math.round(best.up_ratio! * 100)}% up). Weakest: <b>{worst.task}</b> ({Math.round(worst.up_ratio! * 100)}% up): try another model or effort for it under Routing.
    </p>}
    <table className="feedback-table">
      <thead><tr><th>Section</th><th>Sent as</th><th>Model</th><th>Effort</th><th className="n">Calls</th><th className="n">Votes</th><th>Up</th></tr></thead>
      <tbody>
        {app.sections.map(s => <React.Fragment key={s.task || '(none)'}>
          <tr className={'section-row ' + tone(s)}>
            <th scope="row" colSpan={4}>{s.task ? <code>{s.task}</code> : <span className="muted">No section named</span>}{s.rows.length > 1 && <small className="muted"> · {s.rows.length} combinations</small>}</th>
            <td className="n">{num(s.calls)}</td><td className="n"><Counts v={s}/></td><td><Share v={s}/></td>
          </tr>
          {s.rows.map(r => <tr key={r.name + '|' + r.model + '|' + r.effort} className="combo-row">
            <td/><td><code>{r.name}</code></td><td>{r.model}</td><td>{r.effort || '—'}</td>
            <td className="n">{num(r.calls)}</td><td className="n"><Counts v={r}/></td><td><Share v={r}/></td>
          </tr>)}
        </React.Fragment>)}
      </tbody>
    </table>
    {app.sections.some(s => !s.task) && <p className="caption">Calls without a section did not send <code>X-Aiport-Task</code>. Guidance version 4 asks the application's coding agent to add it.</p>}
  </section>;
}

export default function Feedback({tick}: {tick: number}) {
  const [hours, setHours] = useState(720);
  const [report, setReport] = useState<Report | null>(null);
  const [error, setError] = useState('');
  useEffect(() => {
    fetch('/api/v1/feedback?hours=' + hours, {headers: {'X-Aiport-CSRF': '1'}})
      .then(async r => { if (!r.ok) throw new Error((await r.json().catch(() => null))?.error?.message || 'HTTP ' + r.status); return r.json(); })
      .then((r: Report) => { setReport(r); setError(''); })
      .catch(e => setError(String(e.message || e)));
  }, [hours, tick]);
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
      <p className="muted feedback-intro">Each application names the section of every call (<code>{report?.task_header || 'X-Aiport-Task'}</code>, such as tldr, classify or insights) and sends its users' thumbs up and down to <code>{report?.feedback_path || '/v1/aiport/feedback'}</code>. Read each section's share of up-votes by the name it sent, the model that served it and the effort; when a section does badly, change its Level, task name or effort under Routing. Sections with fewer than {FEW} votes are too few to judge.</p>
      {error && <div className="notice error feedback-notice" role="alert">{error}</div>}
      {report && report.telemetry.state === 'disabled' && <div className="notice"><span className="dot amber"/>Request history is off, so no calls or votes are kept. Turn it on in Settings.</div>}
      {report && !apps.length && <div className="quiet-empty feedback-empty">No model calls in this period.</div>}
      {report && apps.length > 0 && !anyVotes && <div className="quiet-empty feedback-empty">No votes yet.<br/><small>Applications start sending them once their coding agent follows guidance version 4 (Integrate → Guidance).</small></div>}
    </section>
    {apps.map(a => <AppPanel key={a.application} app={a}/>)}
  </>;
}
