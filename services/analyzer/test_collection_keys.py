"""Tests for collection_keys derivation against shared fixtures."""

import json
from pathlib import Path

import pytest

from collection_keys import derive_collection_keys

FIXTURES_PATH = Path(__file__).parent.parent.parent / "fixtures" / "buscas.json"


def load_fixtures():
    with open(FIXTURES_PATH) as f:
        return json.load(f)


FIXTURES = load_fixtures()


@pytest.mark.parametrize(
    "fixture",
    FIXTURES,
    ids=[f["id"] for f in FIXTURES],
)
def test_derive_collection_keys_fixtures(fixture):
    shop_ids = fixture.get("shop_ids") or []
    keywords = fixture.get("keywords") or []
    categorias = fixture.get("categorias") or []
    expected = fixture["collection_keys"]

    result = derive_collection_keys(shop_ids, keywords, categorias)

    assert result == expected, f"For {fixture['id']}: got {result}, want {expected}"


def test_derive_collection_keys_sorted():
    result = derive_collection_keys([999, 111, 555], [])
    assert result == sorted(result)


def test_derive_collection_keys_no_duplicates():
    # "42" appears as both shop_id and keyword
    result = derive_collection_keys([42], ["42"])
    assert result == ["42"]


def test_derive_collection_keys_empty_keywords_ignored():
    result = derive_collection_keys([], ["  ", "", "valid"])
    assert result == ["valid"]


def test_derive_collection_keys_lowercase_trim():
    result = derive_collection_keys([], ["  HELLO  ", "World"])
    assert result == ["hello", "world"]
