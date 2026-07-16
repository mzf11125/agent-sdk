"""Recompute tests for ERC-8281 Observation Commitment Protocol (Layer 2)."""

from eth_utils import keccak

from agent_sdk.verify.erc8281.recompute import compute_observation_digest


def test_short_observation():
    """Computes digest for a short observation."""
    observation = b"hello"
    result = compute_observation_digest(observation)
    expected = keccak(observation)
    assert result == expected


def test_empty_observation():
    """Computes digest for empty observation (b"")."""
    observation = b""
    result = compute_observation_digest(observation)
    expected = keccak(observation)
    assert result == expected


def test_long_observation():
    """Computes digest for a longer observation."""
    observation = b"a" * 64
    result = compute_observation_digest(observation)
    expected = keccak(observation)
    assert result == expected


def test_inline_golden_vector():
    """Matches the pre-computed golden vector."""
    observation = b"observation-data"
    result = compute_observation_digest(observation)
    # keccak256(b"observation-data")
    expected = bytes.fromhex("77e50e35b524dae6d88261baf6ce7533856d32e0df1e6898135b5f31d32d5da2")
    assert result == expected


def test_empty_input_golden_vector():
    """Matches the empty-input golden vector."""
    result = compute_observation_digest(b"")
    # keccak256("")
    expected = bytes.fromhex("c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470")
    assert result == expected
