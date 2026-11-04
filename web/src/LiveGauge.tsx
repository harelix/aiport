/* The live gauge at the top of every page: whether the gateway is listening,
   which Applications are calling it, and a level meter that lights with the
   calls as they pass, like a channel meter on a console (fast attack, slow
   release, a peak that holds and falls; a failure lights the top red).

   It reads request metadata only (GET /api/v1/requests, the last hour) on
   its own short poll and replays each new call at the moment it started, so
   the meter moves with the traffic rather than in poll-sized jumps. */
import React, {useEffect, useRef, useState} from 'react';

type Row = {id: string; started_at_ms: number; origin_id: string; outcome: string; http_status: number};
type App = {id: string; name: string};

const POLL_MS = 3000;
const LIVE_MS = 2 * 60_000; // an Application counts as live for two minutes after a call
const SEGMENTS = 18;
const ATTACK = 12, RELEASE = 900, HIT = 160, PEAK_HOLD = 900, PEAK_FALL = 700, FAIL_GLOW = 1400;
const clamp = (x: number, lo: number, hi: number) => Math.max(lo, Math.min(hi, x));
const ease = (dt: number, tau: number) => 1 - Math.exp(-Math.max(0, dt) / tau);
const failed = (r: Row) => r.outcome !== 'success' || r.http_status >= 400;

type Meter = {level: number; hit: number; peak: number; peakAt: number; fail: number; queue: {t: number; v: number; failed: boolean}[]; at: number | null};

function step(m: Meter, now: number) {
  const dt = m.at == null ? 16 : clamp(now - m.at, 0, 250);
  m.at = now;
  while (m.queue.length && m.queue[0].t <= now) {
    const p = m.queue.shift()!;
    m.hit = Math.max(m.hit, p.v);
    if (p.failed) m.fail = 1;
  }
  m.level = m.level < m.hit ? m.level + (m.hit - m.level) * ease(dt, ATTACK) : m.level * (1 - ease(dt, RELEASE));
  m.hit *= 1 - ease(dt, HIT);
  if (m.level >= m.peak) { m.peak = m.level; m.peakAt = now; } else if (now - m.peakAt > PEAK_HOLD) m.peak *= 1 - ease(dt, PEAK_FALL);
  m.fail *= 1 - ease(dt, FAIL_GLOW);
  if (m.level < 0.003) m.level = 0;
  if (m.peak < 0.004) m.peak = 0;
  if (m.fail < 0.01) m.fail = 0;
  return m.queue.length > 0 || m.level > 0 || m.peak > 0 || m.fail > 0;
}

function css(el: Element, name: string, fallback: string) {
  return getComputedStyle(el).getPropertyValue(name).trim() || fallback;
}

function draw(c: HTMLCanvasElement, m: Meter) {
  const g = c.getContext('2d');
  if (!g) return;
  const dpr = window.devicePixelRatio || 1, w = c.clientWidth, h = c.clientHeight;
  if (c.width !== Math.round(w * dpr) || c.height !== Math.round(h * dpr)) { c.width = Math.round(w * dpr); c.height = Math.round(h * dpr); }
  g.setTransform(dpr, 0, 0, dpr, 0, 0);
  g.clearRect(0, 0, w, h);
  const good = css(c, '--good', '#a3be8c'), warn = css(c, '--warn', '#ebcb8b'), bad = css(c, '--bad', '#bf616a'), off = css(c, '--line', '#3b4252');
  const gap = 2, sw = (w - gap * (SEGMENTS - 1)) / SEGMENTS;
  const lit = m.level * SEGMENTS, peakAt = Math.round(m.peak * SEGMENTS) - 1;
  for (let i = 0; i < SEGMENTS; i++) {
    const p = i / (SEGMENTS - 1);
    const color = p > 0.82 ? bad : p > 0.6 ? warn : good;
    const on = clamp(lit - i, 0, 1);
    const x = i * (sw + gap);
    g.globalAlpha = 1;
    g.fillStyle = off;
    g.beginPath(); g.roundRect(x, 1, sw, h - 2, 1.5); g.fill();
    if (on > 0 || i === peakAt) {
      g.globalAlpha = i === peakAt && on === 0 ? 0.75 : 0.25 + 0.75 * on;
      g.fillStyle = color;
      g.beginPath(); g.roundRect(x, 1, sw, h - 2, 1.5); g.fill();
    }
  }
  if (m.fail > 0) { // a failure glows over the top segments
    g.globalAlpha = 0.8 * m.fail;
    g.fillStyle = bad;
    const from = Math.floor(SEGMENTS * 0.83);
    for (let i = from; i < SEGMENTS; i++) { g.beginPath(); g.roundRect(i * (sw + gap), 1, sw, h - 2, 1.5); g.fill(); }
  }
  g.globalAlpha = 1;
}

export default function LiveGauge({apps, listen}: {apps: App[]; listen?: string}) {
  const [ok, setOk] = useState<boolean | null>(null), [signedOut, setSignedOut] = useState(false);
  const [rows, setRows] = useState<Row[]>([]);
  const [now, setNow] = useState(Date.now());
  const canvas = useRef<HTMLCanvasElement | null>(null);
  const meter = useRef<Meter>({level: 0, hit: 0, peak: 0, peakAt: 0, fail: 0, queue: [], at: null});
  const seen = useRef<Set<string> | null>(null);
  const frame = useRef(0);

  const animate = () => {
    if (frame.current) return;
    const tick = (t: number) => {
      const c = canvas.current;
      const moving = step(meter.current, t);
      if (c) draw(c, meter.current);
      frame.current = moving ? requestAnimationFrame(tick) : 0;
    };
    frame.current = requestAnimationFrame(tick);
  };

  useEffect(() => {
    let stop = false, timer = 0;
    const poll = async () => {
      if (stop) return;
      if (!document.hidden) {
        try {
          const res = await fetch('/api/v1/requests?hours=1&limit=100', {credentials: 'same-origin'});
          if (res.status === 401) { setSignedOut(true); setOk(true); timer = window.setTimeout(poll, POLL_MS); return; }
          if (!res.ok) throw new Error(String(res.status));
          setSignedOut(false);
          const body = await res.json();
          const list: Row[] = body.requests || [];
          setOk(true); setRows(list); setNow(Date.now());
          if (seen.current) {
            // New calls replay in the next poll window at their real spacing.
            const fresh = list.filter(r => !seen.current!.has(r.id)).sort((a, b) => a.started_at_ms - b.started_at_ms);
            if (fresh.length) {
              const first = fresh[0].started_at_ms, span = Math.max(1, fresh[fresh.length - 1].started_at_ms - first);
              const t0 = performance.now();
              const v = clamp(0.45 + 0.12 * Math.log2(1 + fresh.length), 0.45, 1);
              fresh.forEach(r => meter.current.queue.push({t: t0 + (r.started_at_ms - first) / span * Math.min(span, POLL_MS - 300), v: failed(r) ? 1 : v, failed: failed(r)}));
              meter.current.queue.sort((a, b) => a.t - b.t);
              animate();
            }
          }
          seen.current = new Set(list.map(r => r.id));
        } catch { setOk(false); }
      }
      timer = window.setTimeout(poll, POLL_MS);
    };
    poll();
    const redraw = () => { if (canvas.current) draw(canvas.current, meter.current); };
    const media = matchMedia('(prefers-color-scheme: dark)');
    media.addEventListener('change', redraw);
    const observer = new MutationObserver(redraw);
    observer.observe(document.documentElement, {attributes: true, attributeFilter: ['data-theme', 'class']});
    redraw();
    return () => { stop = true; clearTimeout(timer); cancelAnimationFrame(frame.current); frame.current = 0; media.removeEventListener('change', redraw); observer.disconnect(); };
  }, []);

  const lastOf = new Map<string, Row>();
  for (const r of rows) { const s = lastOf.get(r.origin_id); if (!s || r.started_at_ms > s.started_at_ms) lastOf.set(r.origin_id, r); }
  const minute = rows.filter(r => now - r.started_at_ms < 60_000);
  const minuteFailed = minute.filter(failed).length;
  const states = apps.map(a => {
    const last = lastOf.get(a.id);
    const live = !!last && now - last.started_at_ms < LIVE_MS;
    return {...a, last, state: !last ? 'idle' : live ? (failed(last) ? 'failing' : 'live') : 'quiet'};
  });
  const live = states.filter(s => s.state === 'live' || s.state === 'failing').length;
  const ago = (t: number) => { const s = Math.round((now - t) / 1000); return s < 60 ? s + ' s ago' : Math.round(s / 60) + ' min ago'; };

  return <div className="live-gauge" role="group" aria-label="Gateway listener, connected applications and live traffic">
    <span className={'gauge-pill' + (ok ? ' on' : ok === false ? ' bad' : '')} role="status" title={ok === false ? 'The dashboard got no answer from the gateway.' : 'aiport is listening' + (listen ? ' on ' + listen : '')}>
      <span className="led" aria-hidden/>{ok === false ? 'No reply' : signedOut ? 'Signed out' : ok ? 'Listening' : 'Connecting…'}
    </span>
    <span className="gauge-apps" title={states.length ? states.map(s => s.name + ': ' + (s.last ? (s.state === 'failing' ? 'last call failed, ' : '') + ago(s.last.started_at_ms) : 'no call in the last hour')).join('\n') : 'No Applications configured yet'} aria-label={live + ' of ' + states.length + ' applications called in the last two minutes'}>
      {states.slice(0, 8).map(s => <span key={s.id} className={'app-led ' + s.state} aria-hidden/>)}
      <small>{states.length ? live + '/' + states.length + ' apps' : 'no apps'}</small>
    </span>
    <canvas ref={canvas} className="gauge-meter" role="img" aria-label={minute.length + ' calls in the last minute' + (minuteFailed ? ', ' + minuteFailed + ' failed' : '')} title={minute.length + ' calls in the last minute' + (minuteFailed ? ', ' + minuteFailed + ' failed' : '') + '. The lights follow each call as it passes; red at the top is a failure.'}/>
  </div>;
}
