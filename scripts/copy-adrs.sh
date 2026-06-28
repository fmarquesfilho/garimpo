#!/usr/bin/env bash
set -euo pipefail
SRC="docs/decisoes"
DST="docs-site/src/content/docs/decisoes"
mkdir -p "$DST"

for f in "$SRC"/*.md; do
  base=$(basename "$f")
  title=$(head -1 "$f" | sed 's/^# //')
  printf '%s\n' "---" "title: \"$title\"" "---" "" > "$DST/$base"
  tail -n +2 "$f" >> "$DST/$base"
done
echo "Copiados $(ls "$DST"/*.md | wc -l) ADRs para $DST"
