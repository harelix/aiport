import React, {createContext, useContext, useEffect, useId, useLayoutEffect, useRef, useState} from 'react';
import {ArrowUpRight, Info, X} from 'lucide-react';
import manual from '../../docs/manual.md?raw';
import {Markdown, openingOf, sectionOf} from './markdown';

/**
 * A short explanation from a manual section, to fit next to its subject: the first paragraph (or a list's first
 * item), cut at a sentence end after about 480 characters. The rest is one click away in the manual.
 */
export function briefOf(section: string, limit = 480): string {
  let text = openingOf(section, 1);
  const item = /^\s*(?:[-*]|\d+\.)\s+/;
  if (item.test(text)) {
    const lines = text.split('\n'), first: string[] = [lines[0].replace(item, '')];
    for (const l of lines.slice(1)) { if (item.test(l)) break; first.push(l.trim()); }
    text = first.join(' ');
  }
  text = text.replace(/\s*\n\s*/g, ' ').trim();
  if (text.length <= limit) return text;
  const cut = text.slice(0, limit), end = Math.max(cut.lastIndexOf('. '), cut.lastIndexOf('; '));
  return (end > limit / 3 ? cut.slice(0, end + 1) : cut.replace(/\s+\S*$/, '')) + ' …';
}

/** Opens the built-in manual at a section (its heading text, exactly as in docs/manual.md). */
export const HelpContext = createContext<(section: string) => void>(() => {});

/** The manual section that explains each page, for the info button next to the page title. */
export const pageHelp: Record<string, string> = {
  Overview: 'A tour of the dashboard', Activity: 'Activity and history', Changes: 'Changes',
  Integrate: 'The Integrate page and the two briefs', Providers: 'Connections and keys', Routing: 'Routing and protocol defaults',
  Agents: 'A tour of the dashboard', Compression: 'A tour of the dashboard', Tasks: 'A tour of the dashboard', Settings: 'A tour of the dashboard',
};

/**
 * An info button next to a subject: it opens a short explanation right there (the opening of the matching manual
 * section), with a link to read the whole section in the manual. Escape, a click outside or the close button
 * closes it. Positioned fixed so panels that clip their content cannot cut it off.
 */
export function HelpLink({section}: {section: string}) {
  const open = useContext(HelpContext);
  const [shown, setShown] = useState(false), [at, setAt] = useState<{top: number; left: number}>({top: 0, left: 0});
  const button = useRef<HTMLButtonElement>(null), pop = useRef<HTMLDivElement>(null), id = useId();
  const text = briefOf(sectionOf(manual, section));
  useEffect(() => {
    if (!shown) return;
    const close = (e: Event) => {
      if (e instanceof KeyboardEvent ? e.key === 'Escape' : !pop.current?.contains(e.target as Node) && !button.current?.contains(e.target as Node)) setShown(false);
    };
    // Scrolling the page moves the button away from the card: close it. Scrolling inside the card does not.
    const scrolled = (e: Event) => { if (!pop.current?.contains(e.target as Node)) setShown(false); };
    document.addEventListener('keydown', close); document.addEventListener('mousedown', close); window.addEventListener('scroll', scrolled, true);
    return () => { document.removeEventListener('keydown', close); document.removeEventListener('mousedown', close); window.removeEventListener('scroll', scrolled, true); };
  }, [shown]);
  // The card is as tall as its text; when it would run past the bottom of the window, it opens above the button.
  useLayoutEffect(() => {
    if (!shown || !pop.current || !button.current) return;
    const r = button.current.getBoundingClientRect(), h = pop.current.offsetHeight;
    if (r.bottom + 8 + h > window.innerHeight - 12 && r.top - 8 - h >= 12) setAt(a => ({...a, top: r.top - 8 - h}));
  }, [shown]);
  function toggle() {
    const r = button.current!.getBoundingClientRect(), width = 380;
    setAt({top: r.bottom + 8, left: Math.max(12, Math.min(r.left - 16, window.innerWidth - width - 12))});
    setShown(s => !s);
  }
  return <>
    <button ref={button} type="button" className={'help-link' + (shown ? ' on' : '')} onClick={toggle} aria-expanded={shown} aria-controls={id}
      title={'What is this? ' + section} aria-label={'Explain: ' + section}>
      <Info size={14} strokeWidth={1.75} aria-hidden/>
    </button>
    {shown && <div ref={pop} id={id} className="help-pop" role="dialog" aria-label={section} style={{top: at.top, left: at.left}} onClick={e => e.stopPropagation()}>
      <div className="help-pop-head"><strong>{section}</strong>
        <button type="button" className="icon-button" onClick={() => setShown(false)} aria-label="Close"><X size={14} strokeWidth={1.75} aria-hidden/></button></div>
      <div className="help-pop-body brief">{text ? <Markdown source={text}/> : <p>Read about this in the manual.</p>}</div>
      <button type="button" className="text-button help-pop-more" onClick={() => { setShown(false); open(section); }}>Read the whole section in the manual <ArrowUpRight size={14} aria-hidden/></button>
    </div>}
  </>;
}
