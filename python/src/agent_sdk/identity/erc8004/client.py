from __future__ import annotations

from dataclasses import dataclass

from eth_account.signers.local import LocalAccount
from web3 import Web3
from web3.logs import DISCARD
from web3.middleware import SignAndSendRawMiddlewareBuilder

from .abi import IDENTITY_REGISTRY_ABI


@dataclass(frozen=True)
class MetadataEntry:
    metadata_key: str
    metadata_value: bytes


class IdentityRegistryClient:
    def __init__(self, rpc_url: str, address: str, account: LocalAccount):
        self._w3 = Web3(Web3.HTTPProvider(rpc_url))
        self._w3.middleware_onion.add(SignAndSendRawMiddlewareBuilder.build(account))
        self._w3.eth.default_account = account.address
        self._contract = self._w3.eth.contract(
            address=Web3.to_checksum_address(address), abi=IDENTITY_REGISTRY_ABI
        )

    def register(self, agent_uri: str = "", metadata: list[MetadataEntry] | None = None) -> int:
        metadata_tuples = [(entry.metadata_key, entry.metadata_value) for entry in (metadata or [])]
        tx_hash = self._contract.functions.register(agent_uri, metadata_tuples).transact()
        receipt = self._w3.eth.wait_for_transaction_receipt(tx_hash)
        # errors=DISCARD: the same transaction also emits an ERC-721 Transfer
        # log (from the underlying _safeMint) which doesn't match this
        # event's ABI. That's expected, not an error, so skip it silently
        # instead of letting web3.py's default WARN behavior emit a
        # UserWarning for it.
        registered_events = self._contract.events.Registered().process_receipt(receipt, errors=DISCARD)
        if not registered_events:
            raise RuntimeError("register: Registered event not found in transaction receipt")
        return registered_events[0]["args"]["agentId"]

    def register_with_source(self, source_token_id: int) -> int:
        """ERC-8323 (Source-Token Agent Binding) — mints an agent bound to
        ``source_token_id`` on this registry's fixed source collection. Only
        the single spec-defined overload (``registerWithSource(uint256)``)
        is bound here; a registry MAY expose additional overloads (payable
        amounts, inline metadata, etc.) not covered by the base ERC-8323
        interface — those need their own client method once their exact
        signature is known, not guessed.
        """
        tx_hash = self._contract.functions.registerWithSource(source_token_id).transact()
        receipt = self._w3.eth.wait_for_transaction_receipt(tx_hash)
        linked_events = self._contract.events.SourceNFTLinked().process_receipt(receipt, errors=DISCARD)
        if not linked_events:
            raise RuntimeError("register_with_source: SourceNFTLinked event not found in transaction receipt")
        return linked_events[0]["args"]["agentId"]

    def set_agent_uri(self, agent_id: int, agent_uri: str) -> None:
        self._send("setAgentURI", agent_id, agent_uri)

    def get_agent_uri(self, agent_id: int) -> str:
        return self._contract.functions.tokenURI(agent_id).call()

    def get_metadata(self, agent_id: int, metadata_key: str) -> bytes:
        return self._contract.functions.getMetadata(agent_id, metadata_key).call()

    def set_metadata(self, agent_id: int, metadata_key: str, metadata_value: bytes) -> None:
        self._send("setMetadata", agent_id, metadata_key, metadata_value)

    def set_agent_wallet(self, agent_id: int, new_wallet: str, deadline: int, signature: bytes) -> None:
        self._send("setAgentWallet", agent_id, Web3.to_checksum_address(new_wallet), deadline, signature)

    def get_agent_wallet(self, agent_id: int) -> str:
        return self._contract.functions.getAgentWallet(agent_id).call()

    def unset_agent_wallet(self, agent_id: int) -> None:
        self._send("unsetAgentWallet", agent_id)

    def owner_of(self, agent_id: int) -> str:
        return self._contract.functions.ownerOf(agent_id).call()

    def _send(self, function_name: str, *args) -> None:
        tx_hash = getattr(self._contract.functions, function_name)(*args).transact()
        self._w3.eth.wait_for_transaction_receipt(tx_hash)
