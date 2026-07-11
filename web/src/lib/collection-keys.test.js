import { describe, expect, it } from 'vitest';
import { readFileSync } from 'fs';
import { resolve } from 'path';
import { deriveCollectionKeys } from './collection-keys.js';

const fixturesPath = resolve(__dirname, '../../../fixtures/buscas.json');
const fixtures = JSON.parse(readFileSync(fixturesPath, 'utf-8'));

describe('deriveCollectionKeys', () => {
	describe('fixtures', () => {
		for (const fx of fixtures) {
			it(`matches expected for ${fx.id}`, () => {
				const result = deriveCollectionKeys(fx.shop_ids ?? [], fx.keywords ?? [], fx.categorias ?? []);
				expect(result).toEqual(fx.collection_keys);
			});
		}
	});

	it('returns sorted array', () => {
		const result = deriveCollectionKeys([999, 111, 555], []);
		expect(result).toEqual([...result].sort());
	});

	it('deduplicates shop_id and keyword with same string', () => {
		const result = deriveCollectionKeys([42], ['42']);
		expect(result).toEqual(['42']);
	});

	it('ignores empty keywords after trim', () => {
		const result = deriveCollectionKeys([], ['  ', '', 'valid']);
		expect(result).toEqual(['valid']);
	});

	it('lowercases and trims keywords', () => {
		const result = deriveCollectionKeys([], ['  HELLO  ', 'World']);
		expect(result).toEqual(['hello', 'world']);
	});

	it('returns empty array for no inputs', () => {
		const result = deriveCollectionKeys([], []);
		expect(result).toEqual([]);
	});
});
