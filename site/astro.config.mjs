// The aiport documentation site. Pages are generated from ../docs by
// scripts/sync-docs.mjs; edit the Markdown there, not src/content/docs.
import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';
import starlightLinksValidator from 'starlight-links-validator';
import starlightLlmsTxt from 'starlight-llms-txt';

const base = process.env.DOCS_BASE ?? '/aiport';

export default defineConfig({
	site: process.env.DOCS_SITE ?? 'https://harelix.github.io',
	base,
	trailingSlash: 'always',
	integrations: [
		starlight({
			title: 'aiport',
			description: 'A self-improving local LLM gateway for developers who code with several agents and models.',
			logo: { src: '../docs/images/logo-no-bg.png', alt: 'aiport' },
			favicon: '/images/logo-64.png',
			social: [{ icon: 'github', label: 'GitHub', href: 'https://github.com/harelix/aiport' }],
			editLink: { baseUrl: 'https://github.com/harelix/aiport/edit/main/docs/' },
			customCss: ['./src/styles/custom.css'],
			head: [
				{ tag: 'meta', attrs: { property: 'og:image', content: `https://harelix.github.io${base}/images/social-preview.png` } },
			],
			sidebar: [
				{ label: 'Start here', items: [{ autogenerate: { directory: 'start' } }] },
				{ label: 'Guides', collapsed: false, items: [{ autogenerate: { directory: 'guides' } }] },
				{ label: 'Dashboard', collapsed: true, items: [{ autogenerate: { directory: 'dashboard' } }] },
				{ label: 'Reference', collapsed: true, items: [{ autogenerate: { directory: 'reference' } }] },
				{ label: 'Specifications', collapsed: true, items: [{ autogenerate: { directory: 'specs' } }] },
				{ label: 'Help', items: [{ autogenerate: { directory: 'help' } }] },
				{ label: 'Project', collapsed: true, items: [{ autogenerate: { directory: 'project' } }] },
			],
			plugins: [
				starlightLinksValidator({ errorOnRelativeLinks: false }),
				starlightLlmsTxt({
					projectName: 'aiport',
					description:
						'aiport is a local LLM gateway: applications and coding agents call it with their own token, and it forwards each call natively to the OpenAI, Anthropic, Bedrock, Meta, Ollama or llama.cpp Connection the Owner chose.',
					// Design specs are for contributors: keep them out of the short file
					// and at the end of the full one.
					promote: ['index', 'start/**', 'guides/**'],
					demote: ['project/**'],
					exclude: ['project/design', 'project/implementation', 'project/compression-evaluators'],
				}),
			],
		}),
	],
});
