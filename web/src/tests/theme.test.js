import { describe, it, expect, beforeEach, vi } from 'vitest';
import { getStoredTheme, resolveTheme, setTheme, getSystemTheme, applyTheme } from '$lib/theme.js';

// Mock matchMedia (jsdom doesn't implement it)
function mockMatchMedia(matches = false) {
	window.matchMedia = vi.fn().mockImplementation((query) => ({
		matches,
		media: query,
		addEventListener: vi.fn(),
		removeEventListener: vi.fn()
	}));
}

describe('theme.js', () => {
	beforeEach(() => {
		localStorage.clear();
		document.documentElement.removeAttribute('data-theme');
		mockMatchMedia(false); // default: prefers light
	});

	describe('getStoredTheme', () => {
		it('returns null when nothing stored', () => {
			expect(getStoredTheme()).toBe(null);
		});

		it('returns "dark" when stored', () => {
			localStorage.setItem('theme', 'dark');
			expect(getStoredTheme()).toBe('dark');
		});

		it('returns "light" when stored', () => {
			localStorage.setItem('theme', 'light');
			expect(getStoredTheme()).toBe('light');
		});

		it('returns "system" when stored', () => {
			localStorage.setItem('theme', 'system');
			expect(getStoredTheme()).toBe('system');
		});

		it('returns null for invalid stored value', () => {
			localStorage.setItem('theme', 'garbage');
			expect(getStoredTheme()).toBe(null);
		});
	});

	describe('resolveTheme', () => {
		it('returns "light" when stored is light', () => {
			localStorage.setItem('theme', 'light');
			expect(resolveTheme()).toBe('light');
		});

		it('returns "dark" when stored is dark', () => {
			localStorage.setItem('theme', 'dark');
			expect(resolveTheme()).toBe('dark');
		});

		it('falls back to system preference when stored is system', () => {
			localStorage.setItem('theme', 'system');
			// jsdom matchMedia defaults to not matching
			expect(resolveTheme()).toBe('light');
		});

		it('falls back to system preference when nothing stored', () => {
			expect(resolveTheme()).toBe('light');
		});
	});

	describe('setTheme', () => {
		it('persists "dark" to localStorage and applies', () => {
			setTheme('dark');
			expect(localStorage.getItem('theme')).toBe('dark');
			expect(document.documentElement.getAttribute('data-theme')).toBe('dark');
		});

		it('persists "light" to localStorage and applies', () => {
			setTheme('light');
			expect(localStorage.getItem('theme')).toBe('light');
			expect(document.documentElement.getAttribute('data-theme')).toBe('light');
		});

		it('removes localStorage for "system" and applies system preference', () => {
			localStorage.setItem('theme', 'dark');
			setTheme('system');
			expect(localStorage.getItem('theme')).toBe(null);
			// jsdom matchMedia defaults to light
			expect(document.documentElement.getAttribute('data-theme')).toBe('light');
		});
	});

	describe('applyTheme', () => {
		it('sets data-theme attribute on documentElement', () => {
			applyTheme('dark');
			expect(document.documentElement.getAttribute('data-theme')).toBe('dark');
		});
	});

	describe('getSystemTheme', () => {
		it('returns "light" when matchMedia does not match dark', () => {
			mockMatchMedia(false);
			expect(getSystemTheme()).toBe('light');
		});

		it('returns "dark" when matchMedia matches dark', () => {
			mockMatchMedia(true);
			expect(getSystemTheme()).toBe('dark');
		});
	});
});
