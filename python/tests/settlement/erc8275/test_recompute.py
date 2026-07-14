import json
from pathlib import Path

import pytest

from agent_sdk.settlement.erc8275.recompute import compute_win_rate

# ── Inline golden vectors (primary) ──────────────────────────────────────
# These reproduce the vectors from recompute-kit/conformance/agent-flow.vectors.json
# for step "8275/reputation". They are duplicated here so tests pass even
# when recompute-kit is not present on disk.

INLINE_VECTORS = [
    {
        "id": "8275-reputation",
        "label": "computeWinRate with 16 wins, 15 losses",
        "inputs": {"wins": 16, "losses": 15},
        "expected": 0.5161,
    },
]


def _conformance_vectors():
    """Read recompute-kit golden vectors for 8275/reputation.

    Returns an empty list if the file is not present (the inline vectors in
    INLINE_VECTORS are the primary assertion; the file-based check is a
    secondary cross-check).
    """
    vectors_path = (
        Path(__file__).resolve().parents[5]
        / "recompute-kit"
        / "conformance"
        / "agent-flow.vectors.json"
    )
    if not vectors_path.exists():
        return []
    with open(vectors_path) as f:
        data = json.load(f)
    return [v for v in data["vectors"] if v["step"] == "8275/reputation"]


class TestComputeWinRate:
    """ERC-8275 recompute: winRate = gated_wins / (gated_wins + gated_losses)."""

    @pytest.mark.parametrize(
        "vec",
        INLINE_VECTORS,
        ids=[v["id"] for v in INLINE_VECTORS],
    )
    def test_inline_golden_vectors(self, vec):
        assert (
            compute_win_rate(vec["inputs"]["wins"], vec["inputs"]["losses"])
            == vec["expected"]
        )

    def test_conformance_vectors_from_file(self):
        file_vectors = _conformance_vectors()
        if not file_vectors:
            pytest.skip(
                "recompute-kit vectors not found — skipping file-based conformance check"
            )

        for vec in file_vectors:
            label = f"{vec['id']}: {vec.get('desc', vec.get('spec', '(no description)'))}"
            wins = vec["inputs"]["commit_gated_wins"]
            losses = vec["inputs"]["commit_gated_losses"]
            assert compute_win_rate(wins, losses) == vec["expected"], label

    def test_zero_wins(self):
        """Zero wins with non-zero losses produces 0.0."""
        assert compute_win_rate(0, 15) == 0.0

    def test_zero_losses(self):
        """Zero losses with non-zero wins produces 1.0."""
        assert compute_win_rate(16, 0) == 1.0

    def test_both_zero_raises(self):
        """Both wins and losses zero raises ValueError."""
        with pytest.raises(ValueError):
            compute_win_rate(0, 0)
