"""governing_convention_hash -- pin-at-issuance, resolve-at-verification for ERC-8275 winRateBps.

A computation rule (formula + representation + rounding mode) is content-addressed:

    governing_convention_hash = "0x" + sha256(JCS(convention_spec))

because number formatting and rounding are exactly where two honest implementations diverge
without disagreeing on the inputs (1 win / 31 losses = 312.5 bps -> 313 half-up vs 312 half-even).
A producer pins the hash at issuance; a verifier resolves it and recomputes under *that* convention.
Tri-state, fail-closed: an unknown hash is ``unverifiable``, never a silent pass.

Binds the convention this SDK produces -- ``win_rate.bps.v0`` (the bps cutover, reputation/erc8275).
The spec object and derived hash are byte-identical to trustless-ai/recompute-kit
conformance/convention-hash-v0; the hash is DERIVED from the spec here (reproduce, don't trust) and
self-checked against the locked identity on import.
"""
from __future__ import annotations

import hashlib
import json
from typing import Any, Callable, Dict, TypedDict

from .recompute import compute_win_rate

# Byte-identical to recompute-kit convention-hash-v0. The hash is derived from THIS object below.
WIN_RATE_BPS_V0_SPEC: Dict[str, str] = {
    "id": "win_rate.bps.v0",
    "quantity": "erc8275.win_rate",
    "formula": "winRateBps = (gated_wins*20000 + total) // (2*total), total = wins+losses",
    "representation": "integer basis points, 0..10000",
    "rounding_mode": "round-half-up (half-away-from-zero), exact integer division — never a float round()",
    "erc": "ERC-8275",
    "source": "agent-sdk#5 @87b08f3 reputation/erc8275 — Python/Rust/TS identical; winRateBps live on babyblueviper /ledger",
}


def _canon(o: Any) -> bytes:
    # RFC-8785 JCS for a flat string map: sorted keys, compact separators, raw UTF-8.
    return json.dumps(o, sort_keys=True, separators=(",", ":"), ensure_ascii=False).encode("utf-8")


def governing_convention_hash(spec: Dict[str, str]) -> str:
    """Content-address a convention spec: ``"0x" + sha256(JCS(spec))``."""
    return "0x" + hashlib.sha256(_canon(spec)).hexdigest()


# Locked identity (recompute-kit convention-hash-v0) -- reproduced from the spec, not trusted.
WIN_RATE_BPS_V0_HASH = "0x0501b75db8e9ef4ef67c74efcfbe2a200b0a7e5aea5ca62f778c91c119e68daf"

assert governing_convention_hash(WIN_RATE_BPS_V0_SPEC) == WIN_RATE_BPS_V0_HASH, (
    "win_rate.bps.v0 convention-hash drift: derived hash does not match the locked identity"
)

# Conventions this SDK produces (and can therefore resolve + recompute).
_COMPUTE_BY_HASH: Dict[str, Callable[[int, int], int]] = {WIN_RATE_BPS_V0_HASH: compute_win_rate}


class PinnedWinRate(TypedDict):
    value: int
    governing_convention_hash: str


def pin_win_rate_bps(wins: int, losses: int) -> PinnedWinRate:
    """Pin-at-issuance: compute winRateBps and stamp the convention hash that produced it."""
    return {"value": compute_win_rate(wins, losses), "governing_convention_hash": WIN_RATE_BPS_V0_HASH}


class Verdict(TypedDict, total=False):
    status: str  # "verified" | "rejected" | "unverifiable"
    convention: str


def verify_win_rate(value: Any, convention_hash: str, wins: int, losses: int) -> Verdict:
    """Resolve-at-verification, tri-state, fail-closed.

    verified     -- value equals the recompute under the pinned convention;
    rejected     -- convention resolves but the value disagrees;
    unverifiable -- the hash is not one this SDK produces (never a silent pass).
    """
    compute = _COMPUTE_BY_HASH.get(convention_hash)
    if compute is None:
        return {"status": "unverifiable"}
    recomputed = compute(wins, losses)
    return {"status": "verified" if recomputed == value else "rejected", "convention": "win_rate.bps.v0"}
