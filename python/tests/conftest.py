import json
import subprocess
from pathlib import Path

import pytest

TESTKIT_DIR = Path(__file__).resolve().parents[2] / "testkit"

ANVIL_RPC_URL = "http://127.0.0.1:8545"


@pytest.fixture(scope="session", autouse=True)
def anvil_session():
    subprocess.run([str(TESTKIT_DIR / "scripts" / "start-anvil.sh")], check=True)
    yield
    subprocess.run([str(TESTKIT_DIR / "scripts" / "stop-anvil.sh")], check=True)


@pytest.fixture(scope="session")
def anvil_rpc_url(anvil_session) -> str:
    return ANVIL_RPC_URL


@pytest.fixture
def anvil_account(anvil_session):
    # Reads one of anvil's own locally-generated (freshly random per
    # process) default accounts, written by
    # testkit/scripts/start-anvil.sh. No address or private key is ever a
    # literal in this repo's source.
    def _get(index: int) -> dict:
        with open(TESTKIT_DIR / ".anvil-accounts.json") as f:
            data = json.load(f)
        return data["accounts"][index]

    return _get


@pytest.fixture
def deploy_contracts():
    # Returns every contract address deployed by the given script, in
    # broadcast order, for scripts that deploy more than one
    # wired-together contract.
    def _deploy(erc_path: str, contract_name: str) -> list[str]:
        result = subprocess.run(
            [str(TESTKIT_DIR / "scripts" / "deploy.sh"), erc_path, contract_name],
            cwd=TESTKIT_DIR,
            check=True,
            capture_output=True,
            text=True,
        )
        return [line for line in result.stdout.strip().splitlines() if line]

    return _deploy


@pytest.fixture
def deploy_contract(deploy_contracts):
    def _deploy(erc_path: str, contract_name: str) -> str:
        return deploy_contracts(erc_path, contract_name)[0]

    return _deploy
