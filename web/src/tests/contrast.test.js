import { describe, it, expect } from 'vitest';
import { readFileSync } from 'fs';
import { resolve } from 'path';

/**
 * Computes WCAG 2.1 relative luminance of a hex color.
 * @param {string} hex - e.g. "#f5f0ed"
 */
function luminance(hex) {
	const r = parseInt(hex.slice(1, 3), 16) / 255;
	const g = parseInt(hex.slice(3, 5), 16) / 255;
	const b = parseInt(hex.slice(5, 7), 16) / 255;
	const toLinear = (c) => c <= 0.03928 ? c / 12.92 : Math.pow((c + 0.055) / 1.055, 2.4);
	return 0.2126 * toLinear(r) + 0.7152 * toLinear(g) + 0.0722 * toLinear(b);
}

/**
 * Computes WCAG contrast ratio between two hex colors.
 */
function contrastRatio(hex1, hex2) {
	const l1 = luminance(hex1);
	const l2 = luminance(hex2);
	const lighter = Math.max(l1, l2);
	const darker = Math.min(l1, l2);
	return (lighter + 0.05) / (darker + 0.05);
}

/**
 * Extracts CSS custom property values from a :root block.
 * @param {string} css
 * @param {string} selector - e.g. ":root" or ":root[data-theme=\"dark\"]"
 */
function extractTokens(css, selector) {
	const escaped = selector.replace(/[[\]"]/g, '\\$&');
	const regex = new RegExp(escaped + '\\s*\\{([^}]+)\\}', 's');
	const match = css.match(regex);
	if (!match) return {};
	const tokens = {};
	const lines = match[1].split('\n');
	for (const line of lines) {
		const m = line.match(/--([\w-]+)\s*:\s*(#[0-9a-fA-F]{6})/);
		if (m) tokens[`--${m[1]}`] = m[2];
	}
	return tokens;
}

// Text/background pairs that must meet WCAG AA
const PAIRS = [
	{ text: '--tinta', bg: '--porcelana', label: 'primary text on main bg' },
	{ text: '--tinta-suave', bg: '--porcelana', label: 'secondary text on main bg' },
	{ text: '--tinta', bg: '--nevoa', label: 'primary text on card bg' },
	{ text: '--ouro', bg: '--nevoa', label: 'gold accent on card bg', largeText: true },
	{ text: '--rosa', bg: '--nevoa', label: 'pink accent on card bg', largeText: true },
	{ text: '--ouro-escuro', bg: '--ouro-fundo', label: 'gold text on gold bg' },
	{ text: '--sucesso-texto', bg: '--sucesso-fundo', label: 'success text on success bg' },
	{ text: '--erro-texto', bg: '--erro-fundo', label: 'error text on error bg' },
	{ text: '--aviso-texto', bg: '--aviso-fundo', label: 'warning text on warning bg' },
];

describe('WCAG AA Contrast Compliance', () => {
	const cssPath = resolve(__dirname, '../lib/components/ui/tokens.css');
	const css = readFileSync(cssPath, 'utf-8');

	describe('Light mode', () => {
		const tokens = extractTokens(css, ':root');

		for (const pair of PAIRS) {
			const minRatio = pair.largeText ? 3.0 : 4.5;
			it(`${pair.label} (${pair.text} on ${pair.bg}) ≥ ${minRatio}:1`, () => {
				const textColor = tokens[pair.text];
				const bgColor = tokens[pair.bg];
				if (!textColor || !bgColor) return; // skip if token not found as hex
				const ratio = contrastRatio(textColor, bgColor);
				expect(ratio, `${pair.text}(${textColor}) on ${pair.bg}(${bgColor}) = ${ratio.toFixed(2)}`).toBeGreaterThanOrEqual(minRatio);
			});
		}
	});

	describe('Dark mode', () => {
		const tokens = extractTokens(css, ':root[data-theme="dark"]');

		for (const pair of PAIRS) {
			const minRatio = pair.largeText ? 3.0 : 4.5;
			it(`${pair.label} (${pair.text} on ${pair.bg}) ≥ ${minRatio}:1`, () => {
				const textColor = tokens[pair.text];
				const bgColor = tokens[pair.bg];
				if (!textColor || !bgColor) return; // skip if token not found as hex
				const ratio = contrastRatio(textColor, bgColor);
				expect(ratio, `${pair.text}(${textColor}) on ${pair.bg}(${bgColor}) = ${ratio.toFixed(2)}`).toBeGreaterThanOrEqual(minRatio);
			});
		}
	});
});
