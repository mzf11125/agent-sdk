"""convention-hash-v0 binding tests -- reproduces the bps convention subset of
trustless-ai/recompute-kit conformance/convention-hash-v0 (the conventions this SDK produces)."""
from agent_sdk.reputation.erc8275.conventions import (
    WIN_RATE_BPS_V0_HASH,
    WIN_RATE_BPS_V0_SPEC,
    governing_convention_hash,
    pin_win_rate_bps,
    verify_win_rate,
)

_H = WIN_RATE_BPS_V0_HASH


def test_hash_reproduces_locked_identity():
    assert governing_convention_hash(WIN_RATE_BPS_V0_SPEC) == _H


def test_pin_at_issuance():
    assert pin_win_rate_bps(1, 31) == {"value": 313, "governing_convention_hash": _H}


BPS_VECTORS = [
    ("bps_golden_0_15", 0, 15, 0, "verified"),
    ("bps_golden_16_0", 16, 0, 10000, "verified"),
    ("bps_golden_1_2", 1, 2, 3333, "verified"),
    ("bps_golden_16_15", 16, 15, 5161, "verified"),
    ("bps_golden_0_10", 0, 10, 0, "verified"),
    ("bps_golden_1_31", 1, 31, 313, "verified"),
    ("bps_golden_9_23", 9, 23, 2813, "verified"),
    ("half_even_value_rejected_under_bps", 1, 31, 312, "rejected"),
    ("old_float_value_under_bps_rejected", 19, 1, 0.95, "rejected"),
]


def test_bps_convention_vectors():
    for name, w, l, value, status in BPS_VECTORS:
        assert verify_win_rate(value, _H, w, l)["status"] == status, name


def test_unknown_convention_is_unverifiable():
    assert verify_win_rate(9500, "0x" + "de" * 32, 19, 1) == {"status": "unverifiable"}
