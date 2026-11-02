import React, {useEffect} from 'react';
import manual from '../../docs/manual.md?raw';
import {Markdown, slug} from './markdown';

/* The user manual (docs/manual.md), built into the dashboard so it is
   available offline and always matches this aiport. anchor, when set, is a
   heading to scroll to (from an info button next to a page or panel). */
export default function ManualPage({anchor = ''}: {anchor?: string}) {
  useEffect(() => {
    const el = anchor ? document.getElementById(slug(anchor)) : null;
    if (el) el.scrollIntoView({block: 'start'}); else window.scrollTo(0, 0);
  }, [anchor]);
  return <section className="panel manual" aria-label="aiport manual"><div className="brief"><Markdown source={manual}/></div></section>;
}
