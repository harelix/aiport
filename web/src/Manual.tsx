import React from 'react';
import manual from '../../docs/manual.md?raw';
import {Markdown} from './markdown';

/* The user manual (docs/manual.md), built into the dashboard so it is
   available offline and always matches this aiport. */
export default function ManualPage() {
  return <section className="panel manual" aria-label="aiport manual"><div className="brief"><Markdown source={manual}/></div></section>;
}
