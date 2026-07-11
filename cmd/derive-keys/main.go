// Command derive-keys outputs collection_keys for each fixture in buscas.json.
// Used by .mise/tasks/check/fixtures-crosslang to validate cross-language consistency.
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/fmarquesfilho/garimpo/internal/busca"
)

type fixture struct {
	ID         string   `json:"id"`
	Keywords   []string `json:"keywords"`
	ShopIDs    []int64  `json:"shop_ids"`
	Categorias []string `json:"categorias"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: derive-keys <fixtures.json>\n")
		os.Exit(1)
	}

	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading file: %v\n", err)
		os.Exit(1)
	}

	var fixtures []fixture
	if err := json.Unmarshal(data, &fixtures); err != nil {
		fmt.Fprintf(os.Stderr, "error parsing json: %v\n", err)
		os.Exit(1)
	}

	for _, fx := range fixtures {
		keys := busca.DeriveCollectionKeys(fx.ShopIDs, fx.Keywords, fx.Categorias...)
		out, _ := json.Marshal(keys)
		fmt.Printf("%s:%s\n", fx.ID, string(out))
	}
}
