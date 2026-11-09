// Every icon on the site comes from Lucide (lucide-react, the same package and
// version the dashboard uses). This Vite plugin replaces Starlight's built-in
// icon map with Lucide drawings, so Starlight's own interface (sidebar, search,
// theme, pagination, edit link) and our pages use one icon set.
//
// Starlight names map to the closest Lucide icon; any Lucide name (for example
// `plug` on a <Card>) can also be used directly if it is listed in EXTRA.
import { readFileSync } from 'node:fs';
import { createRequire } from 'node:module';
import { dirname, join } from 'node:path';

const STARLIGHT_TO_LUCIDE = {
	'up-caret': 'chevron-up',
	'down-caret': 'chevron-down',
	'right-caret': 'chevron-right',
	'left-caret': 'chevron-left',
	'up-arrow': 'arrow-up',
	'down-arrow': 'arrow-down',
	'right-arrow': 'arrow-right',
	'left-arrow': 'arrow-left',
	bars: 'menu',
	close: 'x',
	translate: 'languages',
	pencil: 'pencil',
	pen: 'pen',
	document: 'file-text',
	'add-document': 'file-plus',
	setting: 'settings',
	link: 'link',
	'link-alt': 'link-2',
	external: 'external-link',
	download: 'download',
	'cloud-download': 'cloud-download',
	moon: 'moon',
	sun: 'sun',
	laptop: 'laptop',
	desktop: 'monitor',
	'mobile-android': 'smartphone',
	window: 'app-window',
	clock: 'clock',
	database: 'database',
	server: 'server',
	'code-branch': 'git-branch',
	'open-book': 'book-open',
	notes: 'notebook-pen',
	information: 'info',
	magnifier: 'search',
	'forward-slash': 'slash',
	error: 'circle-x',
	warning: 'triangle-alert',
	'approve-check-circle': 'circle-check',
	'approve-check': 'check',
	'question-circle': 'circle-question-mark',
	question: 'circle-question-mark',
	rocket: 'rocket',
	star: 'star',
	puzzle: 'puzzle',
	'list-format': 'list',
	analytics: 'chart-line',
	random: 'shuffle',
	padlock: 'lock',
	comment: 'message-square',
	'comment-alt': 'message-square-text',
	heart: 'heart',
	rss: 'rss',
	email: 'mail',
	sidebar: 'panel-left',
	'right-sidebar': 'panel-right',
	header: 'panel-top',
	starlight: 'sparkles',
	// Lucide has no brand icons: the GitHub link uses a git branch.
	github: 'git-branch',
};

// Lucide icons used by name in our own pages (src/pages-src).
const EXTRA = ['plug', 'key-round', 'bot', 'gauge', 'shrink', 'thumbs-up', 'wrench', 'eye-off', 'book-open', 'sliders-horizontal', 'file-code', 'rocket'];

const require = createRequire(import.meta.url);
const iconsDir = join(dirname(require.resolve('lucide-react/package.json')), 'dist/esm/icons');

/** The inner SVG markup of one Lucide icon, drawn the Lucide way (2px strokes). */
function lucide(name) {
	const src = readFileSync(join(iconsDir, `${name}.mjs`), 'utf8');
	const literal = src.match(/const __iconData = (\{[\s\S]*?\n\});/)?.[1];
	if (!literal) throw new Error(`lucide-icons: cannot read lucide-react icon "${name}"`);
	const { node } = Function(`return (${literal})`)();
	const inner = node
		.map(([tag, attrs]) => {
			const a = Object.entries(attrs)
				.filter(([k]) => k !== 'key')
				.map(([k, v]) => `${k}="${String(v).replace(/"/g, '&quot;')}"`)
				.join(' ');
			return `<${tag} ${a}/>`;
		})
		.join('');
	// Starlight's <svg> fills with currentColor; Lucide icons are strokes.
	return `<g fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">${inner}</g>`;
}

function iconsModule() {
	const icons = {};
	for (const [starlight, name] of Object.entries(STARLIGHT_TO_LUCIDE)) icons[starlight] = lucide(name);
	for (const name of EXTRA) icons[name] = lucide(name);
	return `export const Icons = ${JSON.stringify(icons)};\n`;
}

/** The Lucide `link` icon as a data URI, for Starlight's heading anchor links (drawn in CSS). */
export function lucideLinkDataUri() {
	const svg = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24">${lucide('link').replace('currentColor', 'black')}</svg>`;
	return `url("data:image/svg+xml,${encodeURIComponent(svg)}")`;
}

export function lucideIcons() {
	let code;
	return {
		name: 'aiport-lucide-icons',
		enforce: 'pre',
		load(id) {
			if (!/@astrojs[\\/]starlight[\\/].*components-internals[\\/]Icons\.(js|ts)$/.test(id.split('?')[0])) return null;
			code ??= iconsModule();
			return code;
		},
	};
}
