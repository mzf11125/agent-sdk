from __future__ import annotations

from dataclasses import dataclass

from web3 import Web3

from .abi import AGENT_SOURCE_BINDING_VIEW_ABI

# ERC-165 id of IAgentSourceBindingView
# = getSourceNFT ^ hasSourceNFT ^ isSourceNFTOwnershipValid.
# The query-only subset a self-sourced ("genesis") agent honestly advertises —
# NOT the full IAgentSourceBinding (0x27eba962), which also carries the bridge
# methods (boundCollection / registerWithSource).
SOURCE_BINDING_VIEW_INTERFACE_ID = "0x8b3597c9"


@dataclass(frozen=True)
class SourceNFT:
    source_contract: str
    source_token_id: int


class SourceBindingViewClient:
    """ERC-8323 Source-Token Agent Binding — read side.

    View-only client over ``IAgentSourceBindingView``. There is no Layer-2
    recompute function: source binding is an on-chain fact (the registry maps
    agentId -> the source NFT and validates current ownership), so verification
    is a direct read, not a hash re-derivation. No signing account is required.
    """

    def __init__(self, rpc_url: str, address: str):
        self._w3 = Web3(Web3.HTTPProvider(rpc_url))
        self._contract = self._w3.eth.contract(
            address=Web3.to_checksum_address(address), abi=AGENT_SOURCE_BINDING_VIEW_ABI
        )

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

    def supports_source_binding_view(self) -> bool:
        """ERC-165: does this contract advertise IAgentSourceBindingView (0x8b3597c9)?"""
        return self._contract.functions.supportsInterface(
            bytes.fromhex(SOURCE_BINDING_VIEW_INTERFACE_ID[2:])
        ).call()
