from eth_account import Account
from web3 import Web3

from agent_sdk.identity.erc8004.client import IdentityRegistryClient, MetadataEntry

# Confirms the two registerWithSource extension overloads Merlini gave us for his real
# AgentIdentityRegistry (mainnet 0xe0454dfa17a57a84c3e0e2dbfda5318cbbe91e2c, 2026-07-16
# Telegram) actually work end-to-end -- including that payment is genuinely sent now,
# not just typed. Runs against MockSourceBindingRegistryExtended (a separate fixture
# from the base-spec mock, see that contract's own header comment for why it's kept
# separate). Mirrors typescript/test/identity/ERC8323/registerWithSourceExtensions.test.ts.
MINT_PRICE = 1_000_000_000_000_000  # 0.001 ether, matches DeployERC8323Extended

_SOURCE_ABI = [
    {
        "type": "function",
        "name": "mint",
        "stateMutability": "nonpayable",
        "inputs": [{"name": "to", "type": "address"}],
        "outputs": [{"name": "tokenId", "type": "uint256"}],
    },
]
_EXTENDED_ABI = [
    {
        "type": "function",
        "name": "agentURI",
        "stateMutability": "view",
        "inputs": [{"name": "", "type": "uint256"}],
        "outputs": [{"name": "", "type": "string"}],
    },
    {
        "type": "function",
        "name": "getMetadata",
        "stateMutability": "view",
        "inputs": [{"name": "", "type": "uint256"}, {"name": "", "type": "string"}],
        "outputs": [{"name": "", "type": "bytes"}],
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
    source_address, registry_address = deploy_contracts("identity/ERC8323", "DeployERC8323Extended")
    return Web3.to_checksum_address(source_address), Web3.to_checksum_address(registry_address)


def test_register_with_source_reverts_wrong_value_succeeds_with_mint_price(
    deploy_contracts, anvil_rpc_url, anvil_account
):
    source_address, registry_address = _deploy(deploy_contracts)
    w3 = Web3(Web3.HTTPProvider(anvil_rpc_url))
    holder = Account.from_key(anvil_account(3)["privateKey"])

    _send(w3, holder, source_address, _SOURCE_ABI, "mint", [holder.address])

    client = IdentityRegistryClient(anvil_rpc_url, registry_address, holder)

    try:
        client.register_with_source(1, value=0)
        raised = False
    except Exception:
        raised = True
    assert raised, "expected registerWithSource to revert with wrong value"

    agent_id = client.register_with_source(1, value=MINT_PRICE)
    assert agent_id == 1


def test_register_with_source_and_uri(deploy_contracts, anvil_rpc_url, anvil_account):
    source_address, registry_address = _deploy(deploy_contracts)
    w3 = Web3(Web3.HTTPProvider(anvil_rpc_url))
    holder = Account.from_key(anvil_account(3)["privateKey"])

    _send(w3, holder, source_address, _SOURCE_ABI, "mint", [holder.address])

    client = IdentityRegistryClient(anvil_rpc_url, registry_address, holder)
    agent_id = client.register_with_source_and_uri("ipfs://custom-uri", 1, value=MINT_PRICE)

    extended = w3.eth.contract(address=Web3.to_checksum_address(registry_address), abi=_EXTENDED_ABI)
    assert extended.functions.agentURI(agent_id).call() == "ipfs://custom-uri"


def test_register_with_source_and_metadata(deploy_contracts, anvil_rpc_url, anvil_account):
    source_address, registry_address = _deploy(deploy_contracts)
    w3 = Web3(Web3.HTTPProvider(anvil_rpc_url))
    holder = Account.from_key(anvil_account(3)["privateKey"])

    _send(w3, holder, source_address, _SOURCE_ABI, "mint", [holder.address])

    client = IdentityRegistryClient(anvil_rpc_url, registry_address, holder)
    agent_id = client.register_with_source_and_metadata(
        "ipfs://with-metadata",
        1,
        metadata=[MetadataEntry(metadata_key="role", metadata_value=b"\x01")],
        value=MINT_PRICE,
    )

    extended = w3.eth.contract(address=Web3.to_checksum_address(registry_address), abi=_EXTENDED_ABI)
    assert extended.functions.getMetadata(agent_id, "role").call() == b"\x01"
