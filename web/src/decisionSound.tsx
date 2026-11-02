import React, {useEffect, useRef, useState} from 'react';
import {Volume2, VolumeX} from 'lucide-react';
import {playSoftNote} from './queueChime';

/**
 * The Owner's setting for a sound when something new waits in Changes. Off by default (Changes stay quiet),
 * remembered in this browser only.
 */
const KEY = 'aiport.changes.sound';

function readOn() {
  try { return localStorage.getItem(KEY) === 'on'; } catch { return false; }
}

/**
 * Plays a short soft note when a Suggestion or Change request starts waiting for a decision (an id not seen on the
 * previous refresh), if the Owner turned the sound on. Never on the first load.
 */
export function useDecisionSound(waiting: string[] | null) {
  const seen = useRef<Set<string> | null>(null);
  const key = waiting ? waiting.slice().sort().join(',') : null;
  useEffect(() => {
    if (!waiting) return;
    const before = seen.current;
    seen.current = new Set(waiting);
    if (before && waiting.some(id => !before.has(id)) && readOn()) playSoftNote();
  }, [key]);
}

/** The toggle on the Changes page. */
export function DecisionSoundToggle() {
  const [on, setOn] = useState(readOn);
  const change = () => {
    const v = !on;
    setOn(v);
    try { localStorage.setItem(KEY, v ? 'on' : 'off'); } catch { /* this session only */ }
    if (v) playSoftNote(); // a preview, which also unlocks audio
  };
  return <button type="button" className="secondary with-icon" aria-pressed={on} onClick={change}
    title={on ? 'A short soft sound plays when something new waits for your decision. Click to turn it off.' : 'No sound. Click to hear a short soft sound when something new waits for your decision.'}>
    {on ? <Volume2 size={14} aria-hidden/> : <VolumeX size={14} aria-hidden/>}{on ? 'Sound on' : 'Sound off'}
  </button>;
}
