"""Integration tests for ERC-8323 Source Binding client (Layer 1)."""

import pytest
from eth_account import Account
from web3.exceptions import ContractLogicError

from agent_sdk.identity.erc8323.client import SourceBindingClient

# Matches MockAgentSourceBinding.MINT_PRICE (0.001 ether, in wei).
MINT_PRICE = 10**15


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
    agent_id = client.register(42, value=MINT_PRICE)
    assert agent_id > 0

    source_nft = client.get_source_nft(agent_id)
    assert source_nft.source_token_id == 42


def test_has_source_nft(deploy_erc8323, anvil_rpc_url, anvil_account):
    """Returns true for hasSourceNFT after registration."""
    _, binding_addr = deploy_erc8323
    account = Account.from_key(anvil_account(1)["privateKey"])
    client = SourceBindingClient(anvil_rpc_url, binding_addr, account)
    agent_id = client.register(99, value=MINT_PRICE)
    assert client.has_source_nft(agent_id) is True


def test_is_source_nft_ownership_valid(deploy_erc8323, anvil_rpc_url, anvil_account):
    """Validates source NFT ownership."""
    _, binding_addr = deploy_erc8323
    account = Account.from_key(anvil_account(1)["privateKey"])
    client = SourceBindingClient(anvil_rpc_url, binding_addr, account)
    agent_id = client.register(7, value=MINT_PRICE)
    assert client.is_source_nft_ownership_valid(agent_id) is True


def test_supports_source_binding(deploy_erc8323, anvil_rpc_url, anvil_account):
    """Reports supportsInterface correctly."""
    _, binding_addr = deploy_erc8323
    account = Account.from_key(anvil_account(1)["privateKey"])
    client = SourceBindingClient(anvil_rpc_url, binding_addr, account)
    assert client.supports_source_binding() is True


def test_register_reverts_on_wrong_mint_price(deploy_erc8323, anvil_rpc_url, anvil_account):
    """Real bug found 2026-07-16: register() never threaded msg.value through the
    call, so a paid registry (mintPrice > 0, e.g. a real deployed
    AgentIdentityRegistry) would always revert. Confirms the default (no value)
    now fails loudly against a price-enforcing mock, and the correct value
    succeeds -- locks the fix in rather than relying on a free mock to mask it."""
    _, binding_addr = deploy_erc8323
    account = Account.from_key(anvil_account(1)["privateKey"])
    client = SourceBindingClient(anvil_rpc_url, binding_addr, account)

    with pytest.raises(ContractLogicError, match="wrong mint price"):
        client.register(1)  # default value=0, mock requires MINT_PRICE

    agent_id = client.register(1, value=MINT_PRICE)
    assert agent_id > 0
