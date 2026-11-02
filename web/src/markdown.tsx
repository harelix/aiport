import React from 'react';

/* A small markdown renderer for text aiport itself ships (the briefs from
   /_connect and the user manual): headings, paragraphs, lists, tables, code
   fences, and inline code, bold and links. It builds React elements, never
   HTML strings, so everything is escaped. */

const inlinePattern = /(`[^`]+`|\*\*[^*]+\*\*|\[[^\]]+\]\([^)]+\))/g;

/** Inline markdown: `code`, **bold**, [text](link). Links to other docs files become their path. */
export function Inline({text}: {text: string}) {
  return <>{text.split(inlinePattern).map((part, i) => {
    if (part.startsWith('`') && part.endsWith('`') && part.length > 1) return <code key={i}>{part.slice(1, -1)}</code>;
    if (part.startsWith('**') && part.endsWith('**') && part.length > 3) return <strong key={i}><Inline text={part.slice(2, -2)}/></strong>;
    const link = /^\[([^\]]+)\]\(([^)]+)\)$/.exec(part);
    if (link) {
      const [, label, href] = link;
      if (href.startsWith('#')) return <a key={i} href={href}>{label}</a>;
      if (/^https?:\/\/(127\.0\.0\.1|localhost)[:/]/.test(href) || href.startsWith('/')) return <a key={i} href={href} target="_blank" rel="noreferrer">{label}</a>;
      // A link to another file of the repository: name it, there is nothing to open here.
      return <span key={i}>{label} (<code>{href.startsWith('docs/') || href.includes('/') ? href : 'docs/' + href}</code>)</span>;
    }
    return <React.Fragment key={i}>{part}</React.Fragment>;
  })}</>;
}

export const slug = (heading: string) => heading.toLowerCase().replace(/[`*]/g, '').replace(/[^a-z0-9 -]/g, '').trim().replace(/\s+/g, '-');

/** One section of a markdown document: the text under heading, up to the next heading of the same or a higher level. */
export function sectionOf(source: string, heading: string): string {
  const lines = source.replace(/\r\n/g, '\n').split('\n');
  const start = lines.findIndex(l => /^#{1,4}\s/.test(l) && slug(l.replace(/^#+\s+/, '')) === slug(heading));
  if (start < 0) return '';
  const level = /^#+/.exec(lines[start])![0].length;
  const body: string[] = [];
  let fence = false;
  for (const l of lines.slice(start + 1)) {
    if (l.startsWith('```')) fence = !fence;
    const h = !fence && /^(#{1,4})\s/.exec(l);
    if (h && h[1].length <= level) break;
    body.push(l);
  }
  return body.join('\n').trim();
}

/** The opening of a section: its first blocks before any sub-heading, at most max of them. */
export function openingOf(section: string, max = 2): string {
  const blocks: string[] = [];
  for (const b of section.split(/\n\s*\n/)) {
    if (/^#{1,4}\s/.test(b.trim()) || blocks.length >= max) break;
    if (b.trim()) blocks.push(b.trim());
  }
  return blocks.join('\n\n');
}

/** The headings of a document, for a table of contents: [level (1 = ##), text]. */
export function headingsOf(source: string, deepest = 2): [number, string][] {
  const out: [number, string][] = [];
  let fence = false;
  for (const l of source.split('\n')) {
    if (l.startsWith('```')) fence = !fence;
    const h = !fence && /^(#{2,4})\s+(.*)$/.exec(l);
    if (h && h[1].length - 1 <= deepest) out.push([h[1].length - 1, h[2]]);
  }
  return out;
}

/** Block markdown. */
export function Markdown({source, anchors = false}: {source: string; anchors?: boolean}) {
  const lines = source.replace(/\r\n/g, '\n').split('\n');
  const out: React.ReactNode[] = [];
  let i = 0;
  while (i < lines.length) {
    const line = lines[i];
    if (line.startsWith('```')) {
      const body: string[] = [];
      for (i++; i < lines.length && !lines[i].startsWith('```'); i++) body.push(lines[i]);
      i++;
      out.push(<pre key={out.length} className="snippet-code"><code>{body.join('\n')}</code></pre>);
      continue;
    }
    const heading = /^(#{1,4})\s+(.*)$/.exec(line);
    if (heading) {
      const level = Math.min(heading[1].length + 1, 5);
      const Tag = ('h' + level) as 'h2';
      const id = slug(heading[2]);
      out.push(<Tag key={out.length} id={id}><Inline text={heading[2]}/>{anchors && <a className="heading-anchor" href={'#' + id} aria-label={'Link to ' + heading[2]} onClick={e => { e.preventDefault(); document.getElementById(id)?.scrollIntoView({block: 'start'}); }}>#</a>}</Tag>);
      i++;
      continue;
    }
    if (/^\|/.test(line)) {
      const rows: string[][] = [];
      for (; i < lines.length && /^\|/.test(lines[i]); i++) {
        if (/^\|[\s|:-]+\|$/.test(lines[i])) continue; // separator row
        rows.push(lines[i].replace(/^\||\|$/g, '').split('|').map(c => c.trim()));
      }
      const [head, ...body] = rows;
      out.push(<div key={out.length} className="table-scroll"><table><thead><tr>{head.map((c, k) => <th key={k}><Inline text={c}/></th>)}</tr></thead><tbody>{body.map((r, k) => <tr key={k}>{r.map((c, n) => <td key={n}><Inline text={c}/></td>)}</tr>)}</tbody></table></div>);
      continue;
    }
    const item = /^(\s*)([-*]|\d+\.)\s+(.*)$/.exec(line);
    if (item) {
      const ordered = /\d/.test(item[2]);
      const items: string[] = [];
      for (; i < lines.length; i++) {
        const m = /^(\s*)([-*]|\d+\.)\s+(.*)$/.exec(lines[i]);
        if (m) items.push(m[3]);
        else if (/^\s{2,}\S/.test(lines[i]) && items.length) items[items.length - 1] += ' ' + lines[i].trim();
        else break;
      }
      const List = ordered ? 'ol' : 'ul';
      out.push(<List key={out.length}>{items.map((t, k) => <li key={k}><Inline text={t}/></li>)}</List>);
      continue;
    }
    if (!line.trim() || line.trim() === '---') { i++; continue; }
    const para: string[] = [];
    for (; i < lines.length && lines[i].trim() && !/^(#{1,4}\s|```|\||\s*([-*]|\d+\.)\s)/.test(lines[i]); i++) para.push(lines[i].trim());
    out.push(<p key={out.length}><Inline text={para.join(' ')}/></p>);
  }
  return <>{out}</>;
}
