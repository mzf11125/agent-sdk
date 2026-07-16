"""Integration tests for ERC-8323 Source Binding client (Layer 1)."""

from eth_account import Account

from agent_sdk.identity.erc8323.client import SourceBindingClient


def test_bound_collection(deploy_erc8323, anvil_rpc_url, anvil_account):
    """Returns the bound collection address."""
    dummy_addr, binding_addr = deploy_erc8323
    account = Account.from_key(anvil_account(1)["privateKey"])
    client = SourceBindingClient(anvil_rpc_url, binding_addr, account)
    collection = client.bound_collection()
    assert collection.lower() == dummy_addr.lower()


def test_register_and_get_source_nft(deploy_erc8323, anvil_rpc_url, anvil_account):
    """Registers an agent and reads back its source NFT."""
    _, binding_addr = deploy_erc8323
    account = Account.from_key(anvil_account(1)["privateKey"])
    client = SourceBindingClient(anvil_rpc_url, binding_addr, account)
    agent_id = client.register(42)
    assert agent_id > 0

    source_nft = client.get_source_nft(agent_id)
    assert source_nft.source_token_id == 42


def test_has_source_nft(deploy_erc8323, anvil_rpc_url, anvil_account):
    """Returns true for hasSourceNFT after registration."""
    _, binding_addr = deploy_erc8323
    account = Account.from_key(anvil_account(1)["privateKey"])
    client = SourceBindingClient(anvil_rpc_url, binding_addr, account)
    agent_id = client.register(99)
    assert client.has_source_nft(agent_id) is True


def test_is_source_nft_ownership_valid(deploy_erc8323, anvil_rpc_url, anvil_account):
    """Validates source NFT ownership."""
    _, binding_addr = deploy_erc8323
    account = Account.from_key(anvil_account(1)["privateKey"])
    client = SourceBindingClient(anvil_rpc_url, binding_addr, account)
    agent_id = client.register(7)
    assert client.is_source_nft_ownership_valid(agent_id) is True


def test_supports_source_binding(deploy_erc8323, anvil_rpc_url, anvil_account):
    """Reports supportsInterface correctly."""
    _, binding_addr = deploy_erc8323
    account = Account.from_key(anvil_account(1)["privateKey"])
    client = SourceBindingClient(anvil_rpc_url, binding_addr, account)
    assert client.supports_source_binding() is True
