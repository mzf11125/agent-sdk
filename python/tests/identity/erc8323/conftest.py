"""Pytest fixtures for ERC-8323 integration tests."""

from pathlib import Path

import pytest

TESTKIT_DIR = Path(__file__).resolve().parent.parent.parent.parent.parent / "testkit"


@pytest.fixture(scope="module")
def deploy_erc8323():
    """Deploy MockAgentSourceBinding + dummy ERC-721 via Foundry.

    Returns (dummy_collection_address, binding_address) as strings.
    """
    import subprocess

    result = subprocess.run(
        [
            str(TESTKIT_DIR / "scripts" / "deploy.sh"),
            "identity/ERC8323",
            "DeployERC8323",
        ],
        capture_output=True,
        text=True,
        check=True,
        cwd=str(TESTKIT_DIR),
    )
    addresses = result.stdout.strip().split("\n")
    assert len(addresses) >= 2
    return addresses[0], addresses[1]
