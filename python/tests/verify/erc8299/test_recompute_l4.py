from agent_sdk.verify.erc8299.recompute_l4 import compute_raw_proposal_hash, compute_verdict_hash

# Golden vector cross-verified against invinoveritas's actual reference implementation
# (services/proof_signing.py, both artifact_hash and compute_decision_ref) on 2026-07-14 --
# same string in, byte-identical hash out, in both Python and the TypeScript port.
ARTIFACT = "test artifact content for cross-language verification"
EXPECTED_RAW_PROPOSAL_HASH = "0xb8f70a237da212a272ecd09370acedbce6ca1d7df90745beafcac77e39697a88"

VERDICT_FIELDS = {
    "artifact_hash": "b8f70a237da212a272ecd09370acedbce6ca1d7df90745beafcac77e39697a88",
    "artifact_type": "plan",
    "policy_version": "invinoveritas.review.v4",
    "verdict": "approve",
    "source_class": "agent_reported",
    "vantage_limitation": None,
}
PREIMAGE_FIELDS = (
    "artifact_hash",
    "artifact_type",
    "policy_version",
    "verdict",
    "source_class",
    "vantage_limitation",
)
EXPECTED_VERDICT_HASH = "sha256:2970854c035d5aedb673b8523128665712895f62dd525c91fc8e858ad588ce58"


class TestComputeRawProposalHash:
    def test_matches_reference_implementation_golden_vector(self):
        assert compute_raw_proposal_hash(ARTIFACT) == EXPECTED_RAW_PROPOSAL_HASH

    def test_handles_empty_artifact(self):
        result = compute_raw_proposal_hash("")
        assert result.startswith("0x")
        assert len(result) == 66

    def test_different_artifacts_produce_different_hashes(self):
        assert compute_raw_proposal_hash("a") != compute_raw_proposal_hash("b")


class TestComputeVerdictHash:
    def test_matches_reference_implementation_golden_vector(self):
        assert compute_verdict_hash(VERDICT_FIELDS, PREIMAGE_FIELDS) == EXPECTED_VERDICT_HASH

    def test_order_independent_in_preimage_fields(self):
        reversed_fields = tuple(reversed(PREIMAGE_FIELDS))
        assert compute_verdict_hash(VERDICT_FIELDS, reversed_fields) == EXPECTED_VERDICT_HASH

    def test_missing_field_treated_as_explicit_none(self):
        without_field = {k: v for k, v in VERDICT_FIELDS.items() if k != "vantage_limitation"}
        assert compute_verdict_hash(without_field, PREIMAGE_FIELDS) == EXPECTED_VERDICT_HASH

    def test_different_verdicts_produce_different_hashes(self):
        rejected = {**VERDICT_FIELDS, "verdict": "reject"}
        assert compute_verdict_hash(rejected, PREIMAGE_FIELDS) != EXPECTED_VERDICT_HASH
