// Package busca implements the canonical BuscaContract derivation logic.
package busca

import (
	"sort"
	"strconv"
	"strings"
)

// DeriveCollectionKeys computes the deterministic collection_keys from a busca's
// shop_ids, keywords, and categorias. The result is a sorted, deduplicated array of non-empty strings.
//
// Rules:
//   - Each shop_id is converted to its string representation
//   - Each keyword is trimmed and lowercased
//   - Each categoria is trimmed and lowercased (fallback for tipo=categoria)
//   - Empty strings after normalization are discarded
//   - Result is sorted lexicographically and deduplicated
func DeriveCollectionKeys(shopIDs []int64, keywords []string, categorias ...string) []string {
	seen := make(map[string]struct{})
	var keys []string

	for _, id := range shopIDs {
		s := strconv.FormatInt(id, 10)
		if _, ok := seen[s]; !ok {
			seen[s] = struct{}{}
			keys = append(keys, s)
		}
	}

	for _, kw := range keywords {
		normalized := strings.TrimSpace(strings.ToLower(kw))
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; !ok {
			seen[normalized] = struct{}{}
			keys = append(keys, normalized)
		}
	}

	// Categorias are used as collection keys ONLY when shop_ids and keywords are both empty
	if len(shopIDs) == 0 && len(keywords) == 0 {
		for _, cat := range categorias {
			normalized := strings.TrimSpace(strings.ToLower(cat))
			if normalized == "" {
				continue
			}
			if _, ok := seen[normalized]; !ok {
				seen[normalized] = struct{}{}
				keys = append(keys, normalized)
			}
		}
	}

	sort.Strings(keys)
	return keys
}
