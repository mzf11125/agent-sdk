from __future__ import annotations

from dataclasses import dataclass

from eth_account.signers.local import LocalAccount
from web3 import Web3
from web3.logs import DISCARD
from web3.middleware import SignAndSendRawMiddlewareBuilder

from .abi import AGENT_SOURCE_BINDING_ABI

# ERC-165 interface id of IAgentSourceBinding (0x27eba962).
# XOR of the five function selectors:
#   boundCollection ^ getSourceNFT ^ hasSourceNFT
#   ^ isSourceNFTOwnershipValid ^ registerWithSource.
SOURCE_BINDING_INTERFACE_ID = "0x27eba962"


@dataclass(frozen=True)
class SourceNFT:
    source_contract: str
    source_token_id: int


class SourceBindingClient:
    """ERC-8323 Source-Token Agent Binding — read + write side.

    Client over ``IAgentSourceBinding``. There is no Layer-2 recompute
    function: source binding is an on-chain fact (the registry maps
    agentId -> the source NFT and validates current ownership), so
    verification is a direct read, not a hash re-derivation.
    """

    def __init__(self, rpc_url: str, address: str, account: LocalAccount):
        self._w3 = Web3(Web3.HTTPProvider(rpc_url))
        self._w3.middleware_onion.add(SignAndSendRawMiddlewareBuilder.build(account))
        self._w3.eth.default_account = account.address
        self._contract = self._w3.eth.contract(
            address=Web3.to_checksum_address(address), abi=AGENT_SOURCE_BINDING_ABI
        )

    def bound_collection(self) -> str:
        """The source ERC-721 collection this registry is bound to."""
        return self._contract.functions.boundCollection().call()

    def register(self, source_token_id: int, value: int = 0) -> int:
        """Register an agent derived from ``source_token_id`` of the bound collection.

        ``registerWithSource`` is ``payable`` on the interface -- a real deployed
        registry (e.g. Merlini's AgentIdentityRegistry) gates on
        ``require(msg.value == mintPrice)``. ``value`` defaults to 0 (matches a
        free/mock registry unchanged); pass the registry's actual mint price
        (in wei) for a paid one, or the call reverts with insufficient value.
        """
        tx_hash = self._contract.functions.registerWithSource(source_token_id).transact(
            {"value": value}
        )
        receipt = self._w3.eth.wait_for_transaction_receipt(tx_hash)
        events = self._contract.events.SourceNFTLinked().process_receipt(receipt, errors=DISCARD)
        if not events:
            raise RuntimeError("register: SourceNFTLinked event not found in transaction receipt")
        return events[0]["args"]["agentId"]

    def get_source_nft(self, agent_id: int) -> SourceNFT:
        """The source NFT ``(contract, token_id)`` an agent is bound to."""
        source_contract, source_token_id = self._contract.functions.getSourceNFT(agent_id).call()
        return SourceNFT(source_contract=source_contract, source_token_id=source_token_id)

    def has_source_nft(self, agent_id: int) -> bool:
        """Whether the agent claims a source NFT."""
        return self._contract.functions.hasSourceNFT(agent_id).call()

    def is_source_nft_ownership_valid(self, agent_id: int) -> bool:
        """Whether the bound source NFT is still owned by the agent's controller."""
        return self._contract.functions.isSourceNFTOwnershipValid(agent_id).call()

    def supports_source_binding(self) -> bool:
        """ERC-165: does this contract advertise IAgentSourceBinding (0x27eba962)?"""
        return self._contract.functions.supportsInterface(
            bytes.fromhex(SOURCE_BINDING_INTERFACE_ID[2:])
        ).call()
