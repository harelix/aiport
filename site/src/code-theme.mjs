// Code blocks in the site's two colours: ink on white, with the windsock's
// orange (a shade dark enough to read) for strings, keywords and numbers.
const ink = '#0F1116';
const muted = '#5C5F66';
const orange = '#C93A1F';
export default {
	name: 'aiport',
	type: 'light',
	colors: {
		'editor.background': '#FFFFFF',
		'editor.foreground': ink,
		'editor.selectionBackground': '#FEF2E5',
		'editorLineNumber.foreground': muted,
		'terminal.background': '#FFFFFF',
		'terminal.foreground': ink,
	},
	tokenColors: [
		{ settings: { foreground: ink } },
		{ scope: ['comment', 'punctuation.definition.comment'], settings: { foreground: muted, fontStyle: 'italic' } },
		{
			scope: ['string', 'constant', 'constant.numeric', 'keyword', 'storage', 'entity.name.tag', 'support.type.property-name', 'markup.heading', 'variable.parameter'],
			settings: { foreground: orange },
		},
		{ scope: ['entity.name.function', 'support.function'], settings: { foreground: ink, fontStyle: 'bold' } },
	],
};
