import {ConnectionModelPicker} from './ModelPicker';
import React, {FormEvent, useEffect, useState} from 'react';
import {Shrink} from 'lucide-react';
import type {ProviderSpec, SettingsConfig} from './Settings';

type Group = {
  key: string; requests: number; errors: number; error_rate: number; avg_latency_ms: number | null;
  bytes_before: number; bytes_after: number; saved_ratio: number | null; est_tokens_saved: number;
  feedback_count: number; feedback_avg: number | null; statuses?: Record<string, number>;
  verdict?: 'insufficient_data' | 'possible_harm' | 'no_harm_seen'; verdict_reason?: string;
};
type Bucket = {start_ms: number; compressed: Group; baseline: Group};
type Method = {id: string; name: string; description: string; lossless: boolean; risk: string};
type Report = {groups: Group[]; baseline: Group; trend: Bucket[]; bucket_ms: number; methods: Method[]; header: string; feedback_path: string; telemetry: {state: string}};
type Metric = 'feedback' | 'errors' | 'saved';

const pct = (v: number | null | undefined, digits = 0) => v == null ? '—' : (v * 100).toFixed(digits) + '%';
const num = (v: number | null | undefined) => v == null ? '—' : new Intl.NumberFormat().format(Math.round(v));
const verdictText = {insufficient_data: ['Not enough data', ''], possible_harm: ['Possible harm', 'bad'], no_harm_seen: ['No harm seen', 'good']} as const;

function TrendChart({trend, metric, bucketMs}: {trend: Bucket[]; metric: Metric; bucketMs: number}) {
  const value = (g: Group) => metric === 'feedback' ? g.feedback_avg : metric === 'errors' ? (g.requests ? g.error_rate : null) : g.saved_ratio;
  const label = metric === 'feedback' ? 'Average feedback score' : metric === 'errors' ? 'Error rate' : 'Request size saved';
  const fmt = (v: number | null) => v == null ? 'no data' : metric === 'feedback' ? v.toFixed(2) : pct(v, 1);
  const max = Math.max(0.0001, ...trend.flatMap(b => [value(b.compressed), metric === 'saved' ? null : value(b.baseline)]).filter((v): v is number => v != null));
  const when = (ms: number) => new Date(ms).toLocaleString(undefined, bucketMs >= 86400000 ? {month: 'short', day: 'numeric'} : {hour: '2-digit', minute: '2-digit'});
  if (!trend.length) return <div className="quiet-empty">No inference requests in this period.</div>;
  return <>
    <div className="legend" aria-hidden><span><i className="swatch compressed"/>Compressed</span>{metric !== 'saved' && <span><i className="swatch baseline"/>Uncompressed</span>}</div>
    <div className="pair-bars" role="img" aria-label={label + ' per period, compressed versus uncompressed'}>
      {trend.map(b => {
        const c = value(b.compressed), u = value(b.baseline);
        return <div key={b.start_ms} className="pair" title={`${when(b.start_ms)}\nCompressed: ${fmt(c)} (${b.compressed.requests} requests)` + (metric === 'saved' ? '' : `\nUncompressed: ${fmt(u)} (${b.baseline.requests} requests)`)}>
          <div className="pair-cols">
            <i className="compressed" style={{height: c == null ? 0 : Math.max(2, c / max * 100) + '%'}}/>
            {metric !== 'saved' && <i className="baseline" style={{height: u == null ? 0 : Math.max(2, u / max * 100) + '%'}}/>}
          </div>
          <small>{when(b.start_ms)}</small>
        </div>;
      })}
    </div>
  </>;
}

type TestResult = {ok: boolean; passed?: number; total?: number; latency_ms?: number; error?: string; results?: {case: string; expected: string; got: string}[]};

async function send(method: string, path: string, body?: object) {
  const r = await fetch('/api/v1/' + path, {method, credentials: 'same-origin', headers: {'Content-Type': 'application/json', 'X-Aiport-CSRF': '1'}, body: body ? JSON.stringify(body) : undefined});
  const d = await r.json().catch(() => ({}));
  if (!r.ok) throw new Error(d.error?.message || 'Request failed (' + r.status + ').');
  return d;
}

/** The owner picks which configured connection and model judge compression. */
function EvaluatorPanel({config, providers, onChange}: {config: SettingsConfig | null; providers: ProviderSpec[]; onChange: () => void}) {
  const saved = config?.evaluation || {};
  const [connection, setConnection] = useState(saved.connection || ''), [model, setModel] = useState(saved.model || '');
  const [busy, setBusy] = useState(''), [error, setError] = useState(''), [message, setMessage] = useState(''), [test, setTest] = useState<TestResult | null>(null);
  useEffect(() => { setConnection(saved.connection || ''); setModel(saved.model || ''); }, [saved.connection, saved.model]);
  const targets = config?.targets || [];
  const chosen = targets.find(t => t.id === connection);
  const name = (id: string) => providers.find(p => p.id === id)?.name || id;
  const dirty = connection !== (saved.connection || '') || model !== (saved.model || '');
  async function save(e: FormEvent) {
    e.preventDefault(); setBusy('save'); setError(''); setMessage(''); setTest(null);
    try { await send('PUT', 'evaluation', {connection, model: connection ? model : ''}); setMessage(connection ? 'Saved. Test it to confirm the model answers.' : 'Evaluation turned off.'); onChange(); }
    catch (err) { setError((err as Error).message); } finally { setBusy(''); }
  }
  async function runTest() {
    setBusy('test'); setError(''); setMessage(''); setTest(null);
    try { setTest(await send('POST', 'evaluation/test')); } catch (err) { setError((err as Error).message); } finally { setBusy(''); }
  }
  const group = (loc: 'remote' | 'local', label: string) => {
    const list = targets.filter(t => t.location === loc && !t.disabled);
    return list.length ? <optgroup label={label}>{list.map(t => <option key={t.id} value={t.id}>{name(t.provider)} · {t.id}</option>)}</optgroup> : null;
  };
  return <section className="panel">
    <div className="panel-heading"><h2>Evaluator</h2><span>The model that judges whether compression changed an answer</span></div>
    <form className="conn-form defaults-form" onSubmit={save}>
      <div className="form-row">
        <label>Connection<select value={connection} onChange={e => { setConnection(e.target.value); setModel(''); }}>
          <option value="">Off (no model-based evaluation)</option>
          {group('local', 'Local')}{group('remote', 'Remote APIs')}
        </select></label>
        <ConnectionModelPicker connection={connection} value={model} onChange={setModel}/>
      </div>
      {chosen?.location === 'remote' && <p className="conn-note notice-inline">Evaluation sends question and answer text to {name(chosen.provider)} and is billed to that key. Choose a local connection to keep it on this machine.</p>}
      {chosen?.location === 'local' && <small>Evaluation stays on this machine.</small>}
      {error && <p className="auth-error" role="alert">{error}</p>}
      {message && <p className="auth-success" role="status">{message}</p>}
      {test && <div className={'test-result ' + (test.ok ? 'good' : 'bad')} role="status">
        <strong>{test.error ? 'Evaluator failed: ' + test.error : test.ok ? `Passed ${test.passed}/${test.total} checks in ${test.latency_ms} ms` : `Passed ${test.passed}/${test.total}; this model may be too weak to judge reliably`}</strong>
        {test.results?.map(r => <span key={r.case}>{r.case}: expected {r.expected}, got {r.got}</span>)}
      </div>}
      <div className="form-actions">
        <button className="primary" disabled={!!busy || !dirty || (!!connection && !model.trim())}>{busy === 'save' ? 'Saving…' : 'Save evaluator'}</button>
        <button type="button" className="secondary" disabled={!!busy || dirty || !saved.connection} onClick={runTest}>{busy === 'test' ? 'Testing…' : 'Test evaluator'}</button>
      </div>
    </form>
  </section>;
}

export default function Compression({tick, config, providers, onChange}: {tick: number; config: SettingsConfig | null; providers: ProviderSpec[]; onChange: () => void}) {
  const [hours, setHours] = useState('168'), [metric, setMetric] = useState<Metric>('feedback');
  const [report, setReport] = useState<Report | null>(null), [error, setError] = useState('');
  useEffect(() => {
    const controller = new AbortController();
    fetch('/api/v1/compression?hours=' + hours, {credentials: 'same-origin', signal: controller.signal})
      .then(r => r.ok ? r.json() : Promise.reject(new Error('Could not load compression data (' + r.status + ').')))
      .then(d => { setReport(d); setError(''); })
      .catch(e => { if (e.name !== 'AbortError') setError(e.message); });
    return () => controller.abort();
  }, [hours, tick]);
  const groups = report?.groups || [], base = report?.baseline;
  const compressed = groups.reduce((n, g) => n + g.requests, 0);
  const before = groups.reduce((n, g) => n + g.bytes_before, 0), after = groups.reduce((n, g) => n + g.bytes_after, 0);
  const tokens = groups.reduce((n, g) => n + g.est_tokens_saved, 0);
  const rated = groups.reduce((n, g) => n + g.feedback_count, 0);
  return <>
    {error && <div className="notice error" role="alert">{error}</div>}
    {report?.telemetry.state === 'disabled' && <div className="notice"><span className="dot amber"/><div><strong>Request recording is off.</strong><span> Compression still works when apps ask for it, but nothing is measured. Set <code>telemetry.enabled: true</code> and restart.</span></div></div>}
    <div className="filters"><label>Period<select value={hours} onChange={e => setHours(e.target.value)}><option value="24">Last 24 hours</option><option value="168">Last 7 days</option><option value="720">Last 30 days</option></select></label></div>
    <section className="metrics" aria-label="Compression summary">
      {[['Compressed requests', num(compressed), num(base?.requests) + ' uncompressed for comparison'],
        ['Request size saved', before ? pct(1 - after / before, 1) : '—', 'Across compressed requests'],
        ['Tokens saved', tokens ? '≈ ' + num(tokens) : '—', 'Estimated from bytes (÷4)'],
        ['Responses scored', num(rated), compressed ? pct(rated / compressed) + ' of compressed requests' : 'Apps send feedback']].map(([t, v, s]) =>
        <article className="metric" key={t}><span>{t}</span><strong>{v}</strong><small>{s}</small></article>)}
    </section>

    <section className="panel">
      <div className="panel-heading"><h2>Does compression hurt responses?</h2>
        <div className="segmented metric-switch" role="radiogroup" aria-label="Trend metric">
          {([['feedback', 'Feedback'], ['errors', 'Errors'], ['saved', 'Saved']] as [Metric, string][]).map(([k, l]) =>
            <button key={k} type="button" role="radio" aria-checked={metric === k} className={metric === k ? 'on' : ''} onClick={() => setMetric(k)}>{l}</button>)}
        </div>
      </div>
      <div className="chart-body"><TrendChart trend={report?.trend || []} metric={metric} bucketMs={report?.bucket_ms || 3600000}/></div>
    </section>

    <section className="panel">
      <div className="panel-heading"><h2>By method</h2><span>Compared with uncompressed requests in the same period</span></div>
      {groups.length ? <div className="table-scroll"><table>
        <thead><tr><th>Methods</th><th className="numeric">Requests</th><th className="numeric">Saved</th><th className="numeric">Errors</th><th className="numeric">Latency</th><th className="numeric">Feedback</th><th>Verdict</th></tr></thead>
        <tbody>
          {groups.map(g => { const [text, tone] = verdictText[g.verdict || 'insufficient_data']; return <tr key={g.key}>
            <td><code>{g.key}</code></td>
            <td className="numeric">{num(g.requests)}</td>
            <td className="numeric">{pct(g.saved_ratio, 1)}<small>≈ {num(g.est_tokens_saved)} tokens</small></td>
            <td className="numeric">{pct(g.error_rate, 1)}<small>vs {pct(base?.error_rate, 1)}</small></td>
            <td className="numeric">{num(g.avg_latency_ms)} ms<small>vs {num(base?.avg_latency_ms)} ms</small></td>
            <td className="numeric">{g.feedback_avg == null ? '—' : g.feedback_avg.toFixed(2)}<small>{g.feedback_count} scored · vs {base?.feedback_avg == null ? '—' : base.feedback_avg.toFixed(2)}</small></td>
            <td><span className={'badge ' + tone} title={g.verdict_reason}>{text}</span><small className="verdict-reason">{g.verdict_reason}</small></td>
          </tr>; })}
        </tbody>
      </table></div> : <div className="empty"><span className="empty-icon"><Shrink size={20} strokeWidth={1.75} aria-hidden/></span><h3>No compressed requests yet</h3><p>Applications opt in per request with the <code>{report?.header || 'X-Aiport-Compress'}</code> header. Results appear here and are compared with uncompressed traffic.</p></div>}
    </section>

    <EvaluatorPanel config={config} providers={providers} onChange={onChange}/>

    <div className="charts">
      <section className="panel">
        <div className="panel-heading"><h2>Methods</h2><span>Applied in this order; only message text is changed</span></div>
        <ul className="method-list">{(report?.methods || []).map(m => <li key={m.id}>
          <div><strong>{m.name}</strong> <code>{m.id}</code></div>
          <p>{m.description}</p>
          <span className={'badge ' + (m.lossless ? 'good' : 'warn')}>{m.lossless ? 'Meaning preserved' : 'Lossy'}</span>
        </li>)}</ul>
      </section>
      <section className="panel setup">
        <p className="eyebrow">FOR APPLICATIONS</p>
        <h2>Turn it on per request</h2>
        <p className="muted">Send the header on any chat, responses or messages call. Tool definitions and schemas are never changed. The response tells you what happened in <code>X-Aiport-Compression</code>.</p>
        <pre>{`${report?.header || 'X-Aiport-Compress'}: whitespace, json`}</pre>
        <p className="muted">Then score the answer you got (0 = useless, 1 = perfect), using the <code>X-Aiport-Request-Id</code> from the response. Scores power the verdict above.</p>
        <pre>{`POST ${report?.feedback_path || '/v1/aiport/feedback'}\n{"request_id": "<X-Aiport-Request-Id>", "score": 0.9}`}</pre>
        <p className="caption">Only sizes, outcomes and scores are recorded. Prompts and responses are never stored. Kept in memory until SQLite storage is enabled.</p>
      </section>
    </div>
  </>;
}
