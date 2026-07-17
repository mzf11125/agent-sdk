import pytest
from eth_account import Account
from web3 import Web3
from web3.exceptions import ContractLogicError

from agent_sdk.anchor.erc8263.client import OnChainProofClient


def _make_client(deploy_contract, anvil_rpc_url, anvil_account):
    contract_address = Web3.to_checksum_address(
        deploy_contract("anchor/ERC8263", "DeployERC8263")
    )
    account = Account.from_key(anvil_account(1)["privateKey"])
    client = OnChainProofClient(anvil_rpc_url, contract_address, account)
    return client


class TestOnChainProofClient:
    """ERC-8263 OnChainProof anchor integration tests."""

    def test_anchor_anonymous_scheme(self, deploy_contract, anvil_rpc_url, anvil_account):
        client = _make_client(deploy_contract, anvil_rpc_url, anvil_account)
        receipt = client.anchor(
            0x00,
            "0x0000000000000000000000000000000000000000000000000000000000000000",
            "0x0000000000000000000000000000000000000000000000000000000000000001",
        )
        assert receipt is not None
        assert len(receipt) > 0

    def test_anchor_registry_scheme(self, deploy_contract, anvil_rpc_url, anvil_account):
        client = _make_client(deploy_contract, anvil_rpc_url, anvil_account)
        receipt = client.anchor(
            0x01,
            "0x0000000000000000000000000000000000000000000000000000000000000001",
            "0x0000000000000000000000000000000000000000000000000000000000000002",
        )
        assert receipt is not None

    def test_anchor_uri_hash_scheme(self, deploy_contract, anvil_rpc_url, anvil_account):
        client = _make_client(deploy_contract, anvil_rpc_url, anvil_account)
        receipt = client.anchor(
            0x02,
            "0x0000000000000000000000000000000000000000000000000000000000000001",
            "0x0000000000000000000000000000000000000000000000000000000000000002",
        )
        assert receipt is not None

    def test_anchor_with_aux(self, deploy_contract, anvil_rpc_url, anvil_account):
        client = _make_client(deploy_contract, anvil_rpc_url, anvil_account)
        receipt = client.anchor_with_aux(
            0x01,
            "0x0000000000000000000000000000000000000000000000000000000000000001",
            "0x0000000000000000000000000000000000000000000000000000000000000002",
            b"\xc0\xff\xee",
        )
        assert receipt is not None

    def test_rejects_zero_proof_hash(self, deploy_contract, anvil_rpc_url, anvil_account):
        client = _make_client(deploy_contract, anvil_rpc_url, anvil_account)
        with pytest.raises(ContractLogicError, match="proofHash must be non-zero"):
            client.anchor(
                0x01,
                "0x0000000000000000000000000000000000000000000000000000000000000001",
                "0x0000000000000000000000000000000000000000000000000000000000000000",
            )

    def test_rejects_anonymous_with_non_zero_agent_id(self, deploy_contract, anvil_rpc_url, anvil_account):
        client = _make_client(deploy_contract, anvil_rpc_url, anvil_account)
        with pytest.raises(ContractLogicError, match="ANONYMOUS scheme requires agentId == 0"):
            client.anchor(
                0x00,
                "0x0000000000000000000000000000000000000000000000000000000000000001",
                "0x0000000000000000000000000000000000000000000000000000000000000002",
            )

    def test_rejects_reserved_scheme(self, deploy_contract, anvil_rpc_url, anvil_account):
        client = _make_client(deploy_contract, anvil_rpc_url, anvil_account)
        with pytest.raises(ContractLogicError, match="reserved agentIdScheme"):
            client.anchor(
                0x03,
                "0x0000000000000000000000000000000000000000000000000000000000000000",
                "0x0000000000000000000000000000000000000000000000000000000000000001",
            )
