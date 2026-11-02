/* Why a request failed, as the gateway explains it. The mapping from codes to
   explanations lives in one place, the gateway's refusal catalog
   (internal/refusal); the dashboard only shows what /api/v1/requests and the
   connection check return. */
import React from 'react';
import {ArrowUpRight} from 'lucide-react';

export type Reason = {code: string; summary: string; explanation: string; fix?: string; page?: string; section?: string; config: boolean};

/** Where the fix is: "Routing → Protocol defaults". */
export const placeOf = (r: Reason) => r.page ? r.page + (r.section ? ' → ' + r.section : '') : '';

/** A link to the page that fixes it. */
export function FixLink({reason, onPage}: {reason: Reason; onPage: (page: string) => void}) {
  if (!reason.page) return null;
  return <button type="button" className="text-button" onClick={() => onPage(reason.page!)}>Open {placeOf(reason)} <ArrowUpRight size={14} aria-hidden/></button>;
}

/** The reason block at the top of the request details panel. */
export function ReasonPanel({reason, codes, onPage}: {reason: Reason; codes: [string, string | undefined][]; onPage: (page: string) => void}) {
  const shown = codes.filter(([, v]) => v);
  return <div className="reason" role="group" aria-label="Why this request failed">
    <strong>{reason.summary}</strong>
    <p>{reason.explanation}</p>
    {reason.fix && <p><b>What to do:</b> {reason.fix} <FixLink reason={reason} onPage={onPage}/></p>}
    <small>{shown.map(([k, v]) => <span key={k}>{k} <code>{v}</code></span>)}</small>
  </div>;
}

/** One Overview line when the latest request on some route was refused for a
    reason in the Owner's configuration: it disappears once a request there
    succeeds again. */
export function ReasonNotice<T extends {route: string; started_at_ms: number; reason?: Reason}>({rows, onPage}: {rows: T[]; onPage: (page: string) => void}) {
  const latest = new Map<string, T>();
  for (const r of rows) {
    const seen = latest.get(r.route);
    if (!seen || r.started_at_ms > seen.started_at_ms) latest.set(r.route, r);
  }
  const failing = [...latest.values()].filter(r => r.reason?.config).sort((a, b) => b.started_at_ms - a.started_at_ms)[0];
  if (!failing?.reason) return null;
  const r = failing.reason;
  return <div className="notice" role="status"><span className="dot amber"/><div><strong>Recent requests fail. {r.summary}.</strong> {r.fix} <FixLink reason={r} onPage={onPage}/></div></div>;
}
