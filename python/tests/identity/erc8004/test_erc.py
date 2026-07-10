import time

import pytest
from eth_account import Account
from web3 import Web3
from web3.exceptions import ContractLogicError

from agent_sdk.identity.erc8004.client import IdentityRegistryClient


def _make_client(deploy_contract, anvil_rpc_url, anvil_account):
    contract_address = Web3.to_checksum_address(deploy_contract("identity/ERC8004", "DeployERC8004"))
    # Account #1 acts as the agent owner.
    account = Account.from_key(anvil_account(1)["privateKey"])
    client = IdentityRegistryClient(anvil_rpc_url, contract_address, account)
    return client, contract_address


def test_register_and_read_back_uri_and_metadata(deploy_contract, anvil_rpc_url, anvil_account):
    client, _ = _make_client(deploy_contract, anvil_rpc_url, anvil_account)

    agent_id = client.register("ipfs://agent-1", [])
    client.set_metadata(agent_id, "role", b"validator")

    assert client.get_agent_uri(agent_id) == "ipfs://agent-1"
    assert client.get_metadata(agent_id, "role") == b"validator"


def test_set_agent_wallet_with_valid_signature(deploy_contract, anvil_rpc_url, anvil_account):
    client, contract_address = _make_client(deploy_contract, anvil_rpc_url, anvil_account)
    account = Account.from_key(anvil_account(1)["privateKey"])
    agent_id = client.register()
    deadline = int(time.time()) + 3600
    # Account #0's address (not its key), reused here only as an arbitrary
    # target address for this "set wallet" call.
    new_wallet = anvil_account(0)["address"]

    signed = account.sign_typed_data(
        domain_data={
            "name": "MockIdentityRegistry",
            "version": "1",
            "chainId": 31337,
            "verifyingContract": contract_address,
        },
        message_types={
            "SetAgentWallet": [
                {"name": "agentId", "type": "uint256"},
                {"name": "newWallet", "type": "address"},
                {"name": "deadline", "type": "uint256"},
            ],
        },
        message_data={
            "agentId": agent_id,
            "newWallet": new_wallet,
            "deadline": deadline,
        },
    )

    client.set_agent_wallet(agent_id, new_wallet, deadline, signed.signature)

    assert client.get_agent_wallet(agent_id).lower() == new_wallet.lower()


def test_set_agent_wallet_rejects_signature_from_wrong_account(deploy_contract, anvil_rpc_url, anvil_account):
    client, contract_address = _make_client(deploy_contract, anvil_rpc_url, anvil_account)
    agent_id = client.register()
    deadline = int(time.time()) + 3600
    new_wallet = anvil_account(0)["address"]
    wrong_account = Account.from_key(anvil_account(2)["privateKey"])

    signed = wrong_account.sign_typed_data(
        domain_data={
            "name": "MockIdentityRegistry",
            "version": "1",
            "chainId": 31337,
            "verifyingContract": contract_address,
        },
        message_types={
            "SetAgentWallet": [
                {"name": "agentId", "type": "uint256"},
                {"name": "newWallet", "type": "address"},
                {"name": "deadline", "type": "uint256"},
            ],
        },
        message_data={
            "agentId": agent_id,
            "newWallet": new_wallet,
            "deadline": deadline,
        },
    )

    with pytest.raises(ContractLogicError, match="invalid wallet signature"):
        client.set_agent_wallet(agent_id, new_wallet, deadline, signed.signature)
