"""Pytest fixtures for ERC-8281 integration tests."""

from pathlib import Path

import pytest

TESTKIT_DIR = Path(__file__).resolve().parent.parent.parent.parent.parent / "testkit"


@pytest.fixture(scope="module")
def deploy_erc8281():
    """Deploy MockObservationCommitment via Foundry.

    Returns the contract address as a string.
    """
    import subprocess

    result = subprocess.run(
        [
            str(TESTKIT_DIR / "scripts" / "deploy.sh"),
            "verify/ERC8281",
            "DeployERC8281",
        ],
        capture_output=True,
        text=True,
        check=True,
        cwd=str(TESTKIT_DIR),
    )
    return result.stdout.strip()
