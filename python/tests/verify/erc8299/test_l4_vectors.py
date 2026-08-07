"""Cross-lane conformance: the same vectors every language port must reproduce."""
import json
import pathlib

import pytest

from agent_sdk.verify.erc8299 import compute_raw_proposal_hash, compute_verdict_hash

VECTORS = (
    pathlib.Path(__file__).resolve().parents[4]
    / "testkit"
    / "vectors"
    / "erc8299-l4.vectors.json"
)
SUITE = json.loads(VECTORS.read_text(encoding="utf-8"))["vectors"]


@pytest.mark.parametrize("v", SUITE, ids=[f"{i}-{v['step']}" for i, v in enumerate(SUITE)])
def test_cross_lane_vector(v):
    if v["step"] == "8299-l4/raw-proposal-hash":
        assert compute_raw_proposal_hash(v["inputs"]["artifact"]) == v["expected"]
    elif v["step"] == "8299-l4/verdict-hash":
        assert (
            compute_verdict_hash(v["inputs"]["fields"], tuple(v["inputs"]["preimage_fields"]))
            == v["expected"]
        )
    else:
        pytest.fail(f"unknown step {v['step']} — a vector exists that no function covers")
