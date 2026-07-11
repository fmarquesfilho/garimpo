"""Verify that novidades and quedas routes do NOT use LIKE for busca identification."""

from pathlib import Path


def test_novidades_no_like_for_busca_id():
    """novidades.py must not use LIKE for busca_id filtering."""
    source = Path(__file__).parent / "routes" / "novidades.py"
    content = source.read_text()

    # Should use exact match
    assert "busca_id = @busca_id" in content, "Expected exact match: busca_id = @busca_id"

    # Should NOT wrap busca_id with % for LIKE
    assert '"%{busca_id}%"' not in content, "Should not use LIKE wrapping"
    assert 'f"%{busca_id}%"' not in content, "Should not use LIKE wrapping"


def test_quedas_no_like_for_busca_id():
    """quedas.py must not use LIKE for busca_id filtering."""
    source = Path(__file__).parent / "routes" / "quedas.py"
    content = source.read_text()

    # Should NOT have LIKE @busca_id
    assert "LIKE @busca_id" not in content, "Should not use LIKE for busca_id"
    # If busca_id is used, it should be exact match
    if "busca_id" in content:
        assert "busca_id = @busca_id" in content, "Should use exact match for busca_id"
