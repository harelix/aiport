import React, {useEffect, useMemo, useState} from 'react';
import {Search} from 'lucide-react';
import manual from '../../docs/manual.md?raw';
import {headingsOf, Markdown, slug} from './markdown';

/*
 * The user manual (docs/manual.md), built into the dashboard so it is available offline and always matches this
 * aiport. Laid out like current documentation sites: a sticky table of contents with a filter and the section you
 * are reading highlighted, headings with anchor links, and a comfortable reading width. anchor, when set, is a
 * heading to open at (from an info button next to a page or panel).
 */
export default function ManualPage({anchor = ''}: {anchor?: string}) {
  const toc = useMemo(() => headingsOf(manual, 2), []);
  const [filter, setFilter] = useState(''), [current, setCurrent] = useState('');
  useEffect(() => {
    const el = anchor ? document.getElementById(slug(anchor)) : null;
    if (el) el.scrollIntoView({block: 'start'}); else window.scrollTo(0, 0);
  }, [anchor]);
  useEffect(() => {
    // The section being read: the last heading above the top quarter of the window.
    const ids = toc.map(([, t]) => slug(t));
    const spy = () => {
      let at = ids[0] || '';
      for (const id of ids) { const el = document.getElementById(id); if (el && el.getBoundingClientRect().top < window.innerHeight / 4) at = id; }
      setCurrent(at);
    };
    spy(); window.addEventListener('scroll', spy, {passive: true});
    return () => window.removeEventListener('scroll', spy);
  }, [toc]);
  const shown = filter.trim() ? toc.filter(([, t]) => t.toLowerCase().includes(filter.trim().toLowerCase())) : toc;
  return <div className="docs">
    <nav className="docs-toc" aria-label="Manual contents">
      <label className="docs-filter"><Search size={14} strokeWidth={1.75} aria-hidden/><input type="search" value={filter} onChange={e => setFilter(e.target.value)} placeholder="Filter sections" aria-label="Filter sections"/></label>
      <ol>{shown.map(([level, text]) => { const id = slug(text); return <li key={id} className={'level-' + level}>
        <a href={'#' + id} className={current === id ? 'current' : ''} aria-current={current === id ? 'location' : undefined}
          onClick={e => { e.preventDefault(); document.getElementById(id)?.scrollIntoView({block: 'start'}); }}>{text.replace(/[`*]/g, '')}</a></li>; })}
        {!shown.length && <li className="docs-none">No section matches.</li>}</ol>
    </nav>
    <article className="docs-body brief" aria-label="aiport manual"><Markdown source={manual} anchors/></article>
  </div>;
}
