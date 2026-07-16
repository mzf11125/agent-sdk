from eth_account import Account
from web3 import Web3

from agent_sdk.identity.erc8004.client import IdentityRegistryClient
from agent_sdk.identity.erc8323.client import SOURCE_BINDING_VIEW_INTERFACE_ID, SourceBindingViewClient

# registerWithSource / mint are not part of SourceBindingViewClient (deliberately
# read-only per spec — source binding is an on-chain fact, no Layer-2 recompute).
# Setup here writes directly against the deployed testkit fixtures, mirroring what
# a real agent-minting flow does before a verifier ever touches the view client.
_SOURCE_ABI = [
    {
        "type": "function",
        "name": "mint",
        "stateMutability": "nonpayable",
        "inputs": [{"name": "to", "type": "address"}],
        "outputs": [{"name": "tokenId", "type": "uint256"}],
    },
    {
        "type": "function",
        "name": "transferFrom",
        "stateMutability": "nonpayable",
        "inputs": [
            {"name": "from", "type": "address"},
            {"name": "to", "type": "address"},
            {"name": "tokenId", "type": "uint256"},
        ],
        "outputs": [],
    },
]
_REGISTRY_ABI = [
    {
        "type": "function",
        "name": "registerWithSource",
        "stateMutability": "payable",
        "inputs": [{"name": "sourceTokenId", "type": "uint256"}],
        "outputs": [{"name": "agentId", "type": "uint256"}],
    },
]


def _send(w3: Web3, account, address: str, abi: list, fn: str, args: list) -> None:
    contract = w3.eth.contract(address=Web3.to_checksum_address(address), abi=abi)
    tx = getattr(contract.functions, fn)(*args).build_transaction(
        {"from": account.address, "nonce": w3.eth.get_transaction_count(account.address)}
    )
    signed = account.sign_transaction(tx)
    tx_hash = w3.eth.send_raw_transaction(signed.raw_transaction)
    w3.eth.wait_for_transaction_receipt(tx_hash)


def _deploy(deploy_contracts):
    source_address, registry_address = deploy_contracts("identity/ERC8323", "DeployERC8323")
    return Web3.to_checksum_address(source_address), Web3.to_checksum_address(registry_address)


def test_reflects_a_live_register_with_source_mint(deploy_contracts, anvil_rpc_url, anvil_account):
    source_address, registry_address = _deploy(deploy_contracts)
    w3 = Web3(Web3.HTTPProvider(anvil_rpc_url))
    holder = Account.from_key(anvil_account(1)["privateKey"])

    _send(w3, holder, source_address, _SOURCE_ABI, "mint", [holder.address])
    source_token_id = 1
    _send(w3, holder, registry_address, _REGISTRY_ABI, "registerWithSource", [source_token_id])
    agent_id = 1

    client = SourceBindingViewClient(anvil_rpc_url, registry_address)

    assert client.has_source_nft(agent_id) is True
    source = client.get_source_nft(agent_id)
    assert source.source_contract.lower() == source_address.lower()
    assert source.source_token_id == source_token_id
    assert client.is_source_nft_ownership_valid(agent_id) is True
    assert client.supports_source_binding_view() is True


def test_ownership_validity_flips_false_on_resale_while_provenance_stays(
    deploy_contracts, anvil_rpc_url, anvil_account
):
    source_address, registry_address = _deploy(deploy_contracts)
    w3 = Web3(Web3.HTTPProvider(anvil_rpc_url))
    stranger = Account.from_key(anvil_account(2)["privateKey"])
    holder = Account.from_key(anvil_account(1)["privateKey"])

    _send(w3, stranger, source_address, _SOURCE_ABI, "mint", [stranger.address])
    source_token_id = 1
    _send(w3, stranger, registry_address, _REGISTRY_ABI, "registerWithSource", [source_token_id])
    agent_id = 1

    client = SourceBindingViewClient(anvil_rpc_url, registry_address)
    assert client.is_source_nft_ownership_valid(agent_id) is True

    _send(w3, stranger, source_address, _SOURCE_ABI, "transferFrom", [stranger.address, holder.address, source_token_id])

    assert client.has_source_nft(agent_id) is True  # provenance unchanged
    assert client.is_source_nft_ownership_valid(agent_id) is False  # live check flips


def test_interface_id_matches_independently_recomputed_spec_id():
    assert SOURCE_BINDING_VIEW_INTERFACE_ID == "0x8b3597c9"


def test_identity_registry_client_register_with_source(deploy_contracts, anvil_rpc_url, anvil_account):
    source_address, registry_address = _deploy(deploy_contracts)
    w3 = Web3(Web3.HTTPProvider(anvil_rpc_url))
    holder = Account.from_key(anvil_account(1)["privateKey"])

    _send(w3, holder, source_address, _SOURCE_ABI, "mint", [holder.address])
    source_token_id = 1

    identity_client = IdentityRegistryClient(anvil_rpc_url, registry_address, holder)
    agent_id = identity_client.register_with_source(source_token_id)

    view_client = SourceBindingViewClient(anvil_rpc_url, registry_address)
    assert view_client.has_source_nft(agent_id) is True
    source = view_client.get_source_nft(agent_id)
    assert source.source_token_id == source_token_id
    assert identity_client.owner_of(agent_id) == holder.address
