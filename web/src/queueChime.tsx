import React, {useEffect, useRef, useState} from 'react';
import {Bell, BellOff, CheckCheck, X} from 'lucide-react';

const KEY = 'aiport.tasks.chime';
type TaskLite = {status: string};

function readEnabled() {
  try { return localStorage.getItem(KEY) !== 'off'; } catch { return true; }
}

let ctx: AudioContext | null = null;
/** Browsers only allow audio after a user gesture; unlock on the first one. */
function unlockAudio() {
  try {
    ctx ??= new AudioContext();
    if (ctx.state === 'suspended') void ctx.resume();
  } catch { /* audio unavailable */ }
}

/** A soft two-note chime, synthesised (no audio file to ship). */
export function playChime() {
  unlockAudio();
  if (!ctx || ctx.state !== 'running') return;
  const now = ctx.currentTime;
  [[659.25, 0], [987.77, 0.14]].forEach(([freq, delay]) => {
    const osc = ctx!.createOscillator(), gain = ctx!.createGain();
    osc.type = 'sine';
    osc.frequency.value = freq;
    gain.gain.setValueAtTime(0.0001, now + delay);
    gain.gain.exponentialRampToValueAtTime(0.18, now + delay + 0.02);
    gain.gain.exponentialRampToValueAtTime(0.0001, now + delay + 0.9);
    osc.connect(gain).connect(ctx!.destination);
    osc.start(now + delay);
    osc.stop(now + delay + 1);
  });
}

/**
 * Watches the task queue on every dashboard refresh. When the last pending or
 * in-progress task finishes, it chimes (if enabled) and shows a notice.
 * Blocked tasks don't hold the queue open; they are mentioned in the notice.
 */
export function useQueueChime(active: boolean, tick: number) {
  const [enabled, setEnabledState] = useState(readEnabled);
  const [done, setDone] = useState<{blocked: number} | null>(null);
  const open = useRef<number | null>(null);
  useEffect(() => {
    const unlock = () => unlockAudio();
    window.addEventListener('pointerdown', unlock, {once: true});
    window.addEventListener('keydown', unlock, {once: true});
    return () => { window.removeEventListener('pointerdown', unlock); window.removeEventListener('keydown', unlock); };
  }, []);
  useEffect(() => {
    if (!active) return;
    const controller = new AbortController();
    fetch('/api/v1/tasks', {credentials: 'same-origin', signal: controller.signal})
      .then(r => r.ok ? r.json() : null)
      .then((d: {tasks: TaskLite[]} | null) => {
        if (!d) return;
        const pending = d.tasks.filter(t => t.status === 'pending' || t.status === 'in_progress').length;
        // Fire only on the transition to empty, never on first load.
        if (open.current !== null && open.current > 0 && pending === 0) {
          setDone({blocked: d.tasks.filter(t => t.status === 'blocked').length});
          if (readEnabled()) playChime();
        }
        if (pending > 0) setDone(null);
        open.current = pending;
      })
      .catch(() => { /* next refresh retries */ });
    return () => controller.abort();
  }, [active, tick]);
  const setEnabled = (v: boolean) => {
    setEnabledState(v);
    try { localStorage.setItem(KEY, v ? 'on' : 'off'); } catch { /* session only */ }
    if (v) playChime(); // preview, and unlocks audio
  };
  return {enabled, setEnabled, done, dismiss: () => setDone(null)};
}

export function ChimeToggle({enabled, onChange}: {enabled: boolean; onChange: (v: boolean) => void}) {
  return <button type="button" className="secondary with-icon" aria-pressed={enabled} onClick={() => onChange(!enabled)}
    title={enabled ? 'A chime plays when the queue is done. Click to mute.' : 'Muted. Click to play a chime when the queue is done.'}>
    {enabled ? <Bell size={14} aria-hidden/> : <BellOff size={14} aria-hidden/>}{enabled ? 'Chime on' : 'Chime off'}
  </button>;
}

export function QueueDoneToast({blocked, onOpen, onClose}: {blocked: number; onOpen: () => void; onClose: () => void}) {
  return <div className="toast" role="status">
    <CheckCheck size={18} aria-hidden/>
    <div><strong>All tasks done</strong><span>{blocked ? blocked + ' blocked task' + (blocked === 1 ? '' : 's') + ' still need you.' : 'The queue is empty.'}</span></div>
    <button type="button" className="text-button" onClick={onOpen}>Open tasks</button>
    <button type="button" className="icon-button" aria-label="Dismiss" onClick={onClose}><X size={14} aria-hidden/></button>
  </div>;
}
