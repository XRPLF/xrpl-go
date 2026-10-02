import { themes as prismThemes } from "prism-react-renderer";
import type { Config } from "@docusaurus/types";
import type * as Preset from "@docusaurus/preset-classic";
import { sidebarLabelGenerator } from "./src/theme/sidebar/sidebarLabelGenerator";

// This runs in Node.js - Don't use client-side code here (browser APIs, JSX...)

const config: Config = {
	title: "XRPL GO",
	tagline: "XRPL GO",
	favicon: "img/favicon.ico",

	// Set the production url of your site here
	url: "https://xrplf.github.io",
	// Set the /<baseUrl>/ pathname under which your site is served
	// For GitHub pages deployment, it is often '/<projectName>/'
	baseUrl: "/xrpl-go",

	// GitHub pages deployment config.
	// If you aren't using GitHub pages, you don't need these.
	organizationName: "XRPLF", // Usually your GitHub org/user name.
	projectName: "xrpl-go", // Usually your repo name.

	onBrokenLinks: "throw",
	markdown: {
		hooks: {
			onBrokenMarkdownLinks: "warn",
		},
	},

	deploymentBranch: "gh-pages",
	trailingSlash: false,

	// Even if you don't use internationalization, you can use this field to set
	// useful metadata like html lang. For example, if your site is Chinese, you
	// may want to replace "en" with "zh-Hans".
	i18n: {
		defaultLocale: "en",
		locales: ["en"],
	},

	presets: [
		[
			"classic",
			{
				docs: {
					sidebarPath: require.resolve("./sidebars.ts"),
					editUrl: "https://github.com/XRPLF/xrpl-go/tree/main/docs",
				},
				blog: false,
				theme: {
					customCss: require.resolve("./src/css/custom.css"),
				},
			} satisfies Preset.Options,
		],
	],

	plugins: [
		[
			"@easyops-cn/docusaurus-search-local",
			{
				hashed: true,
				docsRouteBasePath: "/docs",
				indexBlog: false,
			},
		],
		[
			"@docusaurus/plugin-content-docs",
			{
				id: "changelogPlugin",
				path: "changelog",
				routeBasePath: "changelog",
				sidebarPath: "./sidebarsChangelog.ts",

				// Custom sidebar items generator check /src/theme/sidebar/sidebarLabelGenerator.ts
				sidebarItemsGenerator: sidebarLabelGenerator,
			},
		],
	],

	themeConfig: {
		// Replace with your project's social card
		image: "img/logo.png",
		navbar: {
			title: "XRPL GO",
			logo: {
				alt: "XRPL GO Logo",
				src: "img/xrpl-go-logo.png",
			},
			items: [
				{
					type: "docSidebar",
					sidebarId: "docsSidebar",
					position: "left",
					label: "Docs",
				},
				{
					label: "Changelog",
					position: "left",

					items: [
						{
							type: "docSidebar",
							sidebarId: "changelogSidebar1",
							docsPluginId: "changelogPlugin",
							label: "Core v0.3.x",
						},
						{
							type: "docSidebar",
							sidebarId: "changelogSidebar2",
							docsPluginId: "changelogPlugin",
							label: "Core v0.2.x",
						},
						{
							type: "docSidebar",
							sidebarId: "changelogSidebar3",
							docsPluginId: "changelogPlugin",
							label: "Core v0.1.x",
						},
						{
							type: "docSidebar",
							sidebarId: "confidentialChangelogSidebar",
							docsPluginId: "changelogPlugin",
							label: "Confidential v0.1.x",
						},
					],
				},
				{
					href: "https://github.com/XRPLF/xrpl-go",
					label: "GitHub",
					position: "right",
				},
			],
		},
		footer: {
			style: "dark",
			links: [
				{
					title: "Learn",
					items: [
						{
							label: "Getting started",
							to: "/docs/intro",
						},
						{
							label: "Installation",
							to: "/docs/installation",
						},
						{
							label: "Wallets and signing",
							to: "/docs/xrpl/wallet",
						},
						{
							label: "Confidential transfers",
							to: "/docs/confidential",
						},
					],
				},
				{
					title: "Reference",
					items: [
						{
							label: "Go API",
							href: "https://pkg.go.dev/github.com/Peersyst/xrpl-go",
						},
						{
							label: "Core changelog",
							to: "/changelog/v0.3.x/changelog",
						},
						{
							label: "Confidential changelog",
							to: "/changelog/confidential/v0.1.x/changelog",
						},
					],
				},
				{
					title: "Project",
					items: [
						{
							label: "GitHub",
							href: "https://github.com/XRPLF/xrpl-go",
						},
						{
							label: "Contributing",
							href: "https://github.com/XRPLF/xrpl-go/blob/main/CONTRIBUTING.md",
						},
					],
				},
			],
			copyright: `Copyright © ${new Date().getFullYear()} XRPL Go.`,
		},
		colorMode: {
			defaultMode: "dark",
			disableSwitch: true,
		},
		prism: {
			theme: prismThemes.github,
			darkTheme: prismThemes.dracula,
		},
	} satisfies Preset.ThemeConfig,
};

export default config;
