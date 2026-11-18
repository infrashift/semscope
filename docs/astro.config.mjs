// @ts-check
import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

// Deployed to GitHub Pages by .github/workflows/docs-release.yml (ADR-0011).
// Each PLAN.md phase adds or updates the pages for the features it delivers
// (see the Development workflow guide in .okfw/).
// https://astro.build/config
export default defineConfig({
	site: 'https://infrashift.github.io',
	base: '/semscope',
	integrations: [
		starlight({
			title: 'semscope',
			description:
				'semscope scans Go, C, C++, Zig, Python and Rust source code with composable CUE rulesets, and writes the findings as text, JSON or SARIF.',
			favicon: '/favicon.svg',
			customCss: ['./src/styles/custom.css'],
			social: [
				{
					icon: 'github',
					label: 'GitHub',
					href: 'https://github.com/infrashift/semscope',
				},
			],
			editLink: {
				baseUrl: 'https://github.com/infrashift/semscope/edit/main/docs/',
			},
			sidebar: [
				{
					label: 'Start Here',
					items: [
						{ label: 'Overview', slug: 'index' },
						{ label: 'Getting Started', slug: 'getting-started' },
					],
				},
				{
					label: 'Guides',
					items: [
						{ label: 'Scanning', slug: 'guides/scanning' },
						{ label: 'Rulesets & Overlays', slug: 'guides/rulesets' },
						{ label: 'Output Formats', slug: 'guides/output-formats' },
						{ label: 'Suppressions & Excludes', slug: 'guides/suppressions-and-excludes' },
						{ label: 'CI Integration', slug: 'guides/ci-integration' },
					],
				},
				{
					label: 'Reference',
					items: [
						{ label: 'CLI', slug: 'reference/cli' },
						{ label: 'nasa-p10 Rules', slug: 'reference/nasa-p10' },
						{ label: 'Ruleset Schema', slug: 'reference/ruleset-schema' },
						{ label: 'Make Targets', slug: 'reference/make-targets' },
					],
				},
				{
					label: 'Project',
					collapsed: true,
					items: [
						{ label: 'Roadmap', slug: 'project/roadmap' },
						{ label: 'Design Decisions', slug: 'project/decisions' },
						{ label: 'Contributing', slug: 'project/contributing' },
					],
				},
			],
		}),
	],
});
