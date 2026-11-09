// Builds the site's pages from the repository's Markdown, so docs/ stays the
// only place documentation is written. The manual is split into one page per
// section; every link and anchor is rewritten to point at its new page.
//
// Output (generated, gitignored): src/content/docs/ and public/images/.
import { cpSync, mkdirSync, readFileSync, rmSync, writeFileSync, existsSync } from 'node:fs';
import { dirname, join, posix, relative, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import GithubSlugger from 'github-slugger';

const site = resolve(dirname(fileURLToPath(import.meta.url)), '..');
const repo = resolve(site, '..');
const out = join(site, 'src/content/docs');
const BASE = process.env.DOCS_BASE ?? '/aiport';
const REPO_URL = 'https://github.com/harelix/aiport';
const BLOB = `${REPO_URL}/blob/main/`;

// How each source file becomes pages. `split` lists the ## sections of the
// manual and where they go; sections listed in `perSubsection` get one page per
// ### (their intro text becomes the directory's index page).
const MANUAL = {
	src: 'docs/manual.md',
	sections: {
		'': { route: 'start/introduction', title: 'Introduction', order: 1 },
		'At a glance': { route: 'start/introduction' },
		Quickstart: { route: 'start/quickstart', order: 2 },
		'Core concepts': { route: 'start/concepts', order: 3 },
		Guides: { dir: 'guides' },
		'A tour of the dashboard': { dir: 'dashboard' },
		Reference: { dir: 'reference' },
		Troubleshooting: { route: 'help/troubleshooting', order: 1 },
		"What's new": { route: 'help/whats-new', order: 2 },
	},
};
const FILES = [
	{ src: 'docs/connect.md', route: 'specs/connect', order: 1 },
	{ src: 'docs/configuration.md', route: 'specs/configuration', order: 2 },
	{ src: 'docs/extensions.md', route: 'specs/extensions', order: 3 },
	{ src: 'docs/compression.md', route: 'specs/compression', order: 4 },
	{ src: 'docs/compatibility.md', route: 'specs/compatibility', order: 5 },
	{ src: 'GLOSSARY.md', route: 'project/glossary', order: 1 },
	{ src: 'docs/design.md', route: 'project/design', order: 2 },
	{ src: 'docs/implementation.md', route: 'project/implementation', order: 3 },
	{ src: 'docs/compression-evaluators.md', route: 'project/compression-evaluators', order: 4 },
];

// ---------- Markdown helpers ----------

/** Splits text into lines tagged with whether they sit inside a code fence. */
function lines(text) {
	let fence = null;
	return text.split('\n').map((line) => {
		const m = line.match(/^\s*(`{3,}|~{3,})/);
		if (m) {
			if (!fence) fence = m[1];
			else if (m[1][0] === fence[0] && m[1].length >= fence.length) {
				fence = null;
				return { line, code: true };
			}
			return { line, code: true };
		}
		return { line, code: !!fence };
	});
}

/** Heading text as plain text, the way slugs and titles see it. */
function plain(md) {
	return md
		.replace(/!\[([^\]]*)\]\([^)]*\)/g, '$1')
		.replace(/\[([^\]]*)\]\([^)]*\)/g, '$1')
		.replace(/`([^`]*)`/g, '$1')
		.replace(/(\*\*|__)(.*?)\1/g, '$2')
		.replace(/(\*|_)(.*?)\1/g, '$2')
		.replace(/<[^>]+>/g, '')
		.trim();
}

function headings(text) {
	return lines(text).flatMap(({ line, code }, i) => {
		const m = !code && line.match(/^(#{1,6})\s+(.*?)\s*#*\s*$/);
		return m ? [{ i, level: m[1].length, text: m[2] }] : [];
	});
}

function description(body) {
	const ls = lines(body);
	for (let i = 0; i < ls.length; i++) {
		const t = ls[i].line.trim();
		if (ls[i].code || !t || /^[#|>!<\-*\d]/.test(t)) continue;
		let para = t;
		while (i + 1 < ls.length && !ls[i + 1].code && ls[i + 1].line.trim()) para += ` ${ls[++i].line.trim()}`;
		const s = plain(para).replace(/\s+/g, ' ');
		const sentence = s.match(/^.*?[.:](\s|$)/)?.[0].trim() ?? s;
		return sentence.length > 200 ? `${sentence.slice(0, 197)}...` : sentence;
	}
	return undefined;
}

// ---------- Build the page list ----------

/** pages: { route, title, order, body, src, anchors: Map(oldSlug -> newSlug|'') } */
const pages = [];
/** linkMap: `${src}#${oldSlug}` -> `${route}#${newSlug}`, and `${src}` -> route */
const linkMap = new Map();

function addPage({ route, title, order, body, src, oldSlugs }) {
	const page = { route, title, order, body, src };
	pages.push(page);
	// New slugs as Starlight will make them, per page.
	const slugger = new GithubSlugger();
	const newSlugs = headings(body).map((h) => slugger.slug(plain(h.text)));
	oldSlugs.forEach((old, k) => {
		if (old.page) linkMap.set(`${src}#${old.slug}`, route);
		else linkMap.set(`${src}#${old.slug}`, `${route}#${newSlugs[old.index]}`);
	});
	if (!linkMap.has(src)) linkMap.set(src, route);
	return page;
}

/** Shifts headings so the page's own heading level becomes the title (level 1). */
function shift(text, by) {
	return lines(text)
		.map(({ line, code }) => {
			if (code || by === 0) return line;
			const m = line.match(/^(#{1,6})(\s.*)$/);
			return m ? '#'.repeat(Math.max(2, m[1].length - by)) + m[2] : line;
		})
		.join('\n');
}

/** Splits a document into blocks at headings of `level`, keeping old slugs. */
function sliceAt(text, level, start, end, globalSlugs) {
	const all = headings(text);
	const textLines = text.split('\n');
	const inRange = all.filter((h) => h.i >= start && h.i < end);
	const cuts = inRange.filter((h) => h.level === level);
	const blocks = [];
	const first = cuts[0]?.i ?? end;
	blocks.push({ heading: null, from: start, to: first });
	cuts.forEach((h, k) => blocks.push({ heading: h, from: h.i, to: cuts[k + 1]?.i ?? end }));
	return blocks.map((b) => {
		const hs = all.filter((h) => h.i > b.from - (b.heading ? 0 : 1) && h.i < b.to && h !== b.heading);
		const bodyLines = textLines.slice(b.heading ? b.from + 1 : b.from, b.to);
		const offset = b.heading ? b.from + 1 : b.from;
		return {
			heading: b.heading,
			slug: b.heading ? globalSlugs.get(b.heading.i) : null,
			body: bodyLines.join('\n'),
			inner: hs.filter((h) => h.i >= offset).map((h) => ({ slug: globalSlugs.get(h.i), lineInBody: h.i - offset })),
		};
	});
}

function globalSlugsOf(text) {
	const slugger = new GithubSlugger();
	return new Map(headings(text).map((h) => [h.i, slugger.slug(plain(h.text))]));
}

function pageFromBlock(block, { route, title, order, src, level, extraOldSlugs = [] }) {
	const body = shift(block.body, level - 1).replace(/^\s*\n/, '');
	const oldSlugs = [
		...(block.slug ? [{ slug: block.slug, page: true }] : []),
		...extraOldSlugs.map((slug) => ({ slug, page: true })),
		...block.inner.map((h, index) => ({ slug: h.slug, index })),
	];
	return addPage({ route, title, order, body, src, oldSlugs });
}

// The manual.
{
	const src = MANUAL.src;
	const text = readFileSync(join(repo, src), 'utf8');
	const slugs = globalSlugsOf(text);
	const all = headings(text);
	const h1 = all.find((h) => h.level === 1);
	const top = sliceAt(text, 2, h1 ? h1.i + 1 : 0, text.split('\n').length, slugs);
	const merged = new Map(); // route -> accumulated blocks (Introduction + At a glance)
	for (const block of top) {
		const name = block.heading ? plain(block.heading.text) : '';
		const rule = MANUAL.sections[name];
		if (!rule) throw new Error(`manual.md: section "## ${name}" has no place on the site; add it to MANUAL.sections`);
		if (rule.route) {
			const prev = merged.get(rule.route);
			if (prev) {
				// Append this section to the previous page as a ## section.
				// ### so that it becomes ## once the page's headings shift up a level.
				prev.body += `\n### ${block.heading.text}\n${block.body}`;
				prev.inner.push({ slug: block.slug }, ...block.inner);
				continue;
			}
			merged.set(rule.route, { ...block, rule, name, level: 2 });
			continue;
		}
		// Split section: intro is the directory index, each ### its own page.
		const start = block.heading.i + 1;
		const subs = sliceAt(text, 3, start, block.heading.i + 1 + block.body.split('\n').length, slugs);
		const [intro, ...rest] = subs;
		pageFromBlock(intro, {
			route: `${rule.dir}/index`,
			title: plain(block.heading.text),
			order: 0,
			src,
			level: 2,
			extraOldSlugs: [block.slug],
		});
		rest.forEach((sub, k) => {
			pageFromBlock(sub, {
				route: `${rule.dir}/${sub.slug}`,
				title: plain(sub.heading.text),
				order: k + 1,
				src,
				level: 3,
			});
		});
	}
	for (const [route, b] of merged) {
		// Recompute inner headings for merged pages so slugs line up with the new body.
		const body = shift(b.body, 1);
		const slugger = new GithubSlugger();
		const newSlugs = headings(body).map((h) => slugger.slug(plain(h.text)));
		const oldInner = b.inner.map((h) => h.slug);
		const title = b.rule.title ?? plain(b.heading.text);
		pages.push({ route, title, order: b.rule.order, body: body.replace(/^\s*\n/, ''), src });
		if (b.slug) linkMap.set(`${src}#${b.slug}`, route);
		oldInner.forEach((slug, k) => linkMap.set(`${src}#${slug}`, `${route}#${newSlugs[k]}`));
	}
	linkMap.set(src, 'start/introduction');
}

// Whole files.
for (const f of FILES) {
	const text = readFileSync(join(repo, f.src), 'utf8');
	const slugs = globalSlugsOf(text);
	const all = headings(text);
	const h1 = all.find((h) => h.level === 1);
	const block = {
		slug: h1 ? slugs.get(h1.i) : null,
		body: text.split('\n').slice(h1 ? h1.i + 1 : 0).join('\n'),
		inner: all.filter((h) => h !== h1).map((h) => ({ slug: slugs.get(h.i) })),
	};
	pageFromBlock(block, { route: f.route, title: h1 ? plain(h1.text) : f.route, order: f.order, src: f.src, level: 1 });
}

// ---------- Rewrite links ----------

/** Route to URL path: guides/index -> guides/, start/quickstart -> start/quickstart/. */
const url = (route) => `${route.replace(/(^|\/)index$/, '')}/`.replace(/^\/$/, '');

function target(src, href) {
	if (/^[a-z][a-z0-9+.-]*:|^\/\//i.test(href)) return href; // absolute URL
	const [path, hash = ''] = href.split('#');
	const file = path ? posix.normalize(posix.join(posix.dirname(src), path)) : src;
	if (file.startsWith('docs/images/')) return `${BASE}/images/${file.slice('docs/images/'.length)}`;
	const key = hash ? `${file}#${decodeURIComponent(hash)}` : file;
	const hit = linkMap.get(key) ?? (hash ? undefined : linkMap.get(file));
	if (hit) {
		const [r, a] = hit.split('#');
		return `${BASE}/${url(r)}${a ? `#${a}` : ''}`;
	}
	if (hash && linkMap.has(file)) {
		console.warn(`warning: ${src}: link to missing heading ${href}`);
		const [r] = linkMap.get(file).split('#');
		return `${BASE}/${url(r)}#${hash}`;
	}
	if (existsSync(join(repo, file))) return `${BLOB}${file}${hash ? `#${hash}` : ''}`;
	console.warn(`warning: ${src}: link to missing file ${href}`);
	return `${BLOB}${file}`;
}

function rewrite(body, src) {
	return lines(body)
		.map(({ line, code }) => {
			if (code) return line;
			// Leave inline code alone: rewrite only the parts between backticks.
			return line
				.split(/(`[^`]*`)/)
				.map((part) =>
					part.startsWith('`')
						? part
						: part
								.replace(/(\]\()(<?)([^)\s>]+)(>?)((?:\s+"[^"]*")?\))/g, (_, a, l, href, r, rest) => a + l + target(src, href) + r + rest)
								.replace(/^(\s*\[[^\]]+\]:\s*)(\S+)/, (_, a, href) => a + target(src, href))
								.replace(/((?:src|href)=")([^"]+)(")/g, (_, a, href, z) => a + target(src, href) + z),
				)
				.join('');
		})
		.join('\n');
}

// ---------- Write ----------

rmSync(out, { recursive: true, force: true });
mkdirSync(out, { recursive: true });
rmSync(join(site, 'public/images'), { recursive: true, force: true });
cpSync(join(repo, 'docs/images'), join(site, 'public/images'), { recursive: true });
cpSync(join(site, 'src/pages-src'), out, { recursive: true });

for (const p of pages) {
	const fm = [
		'---',
		`title: ${JSON.stringify(p.title)}`,
		...(description(p.body) ? [`description: ${JSON.stringify(description(p.body))}`] : []),
		`editUrl: ${JSON.stringify(`${REPO_URL}/edit/main/${p.src}`)}`,
		'sidebar:',
		`  order: ${p.order ?? 99}`,
		'---',
		'',
		`<!-- Generated from ${p.src} by site/scripts/sync-docs.mjs. Edit that file, not this one. -->`,
	];
	const file = join(out, `${p.route}.md`);
	mkdirSync(dirname(file), { recursive: true });
	writeFileSync(file, `${fm.join('\n')}\n\n${rewrite(p.body, p.src).trim()}\n`);
}
console.log(`sync-docs: ${pages.length} pages from docs/ into ${relative(site, out)}`);
