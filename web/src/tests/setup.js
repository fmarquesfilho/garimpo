import '@testing-library/jest-dom/vitest';

// Mock window.matchMedia (required by Svelte 5 motion/reduced-motion detection)
if (typeof window !== 'undefined' && !window.matchMedia) {
	window.matchMedia = (query) => ({
		matches: false,
		media: query,
		onchange: null,
		addListener: () => {},
		removeListener: () => {},
		addEventListener: () => {},
		removeEventListener: () => {},
		dispatchEvent: () => false
	});
}
