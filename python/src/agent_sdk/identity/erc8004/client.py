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

    def register_with_source(self, source_token_id: int, value: int = 0) -> int:
        """ERC-8323 (Source-Token Agent Binding) — mints an agent bound to
        ``source_token_id`` on this registry's fixed source collection. This
        is the single spec-defined overload (``registerWithSource(uint256)``);
        a registry MAY require payment (the ABI marks it ``payable``) --
        pass ``value`` (wei) if the deployment enforces a mint price, omit
        it (defaults to 0) for a free registry like this repo's own
        MockSourceBindingRegistry. Two deployment-specific extension
        overloads exist below (``register_with_source_and_uri`` /
        ``register_with_source_and_metadata``) for registries known to
        expose them (confirmed on Merlini's AgentIdentityRegistry,
        2026-07-16) -- NOT part of the ERC-8323 base interface, don't
        assume a given registry has them.
        """
        tx_hash = self._contract.functions.registerWithSource(source_token_id).transact({"value": value})
        receipt = self._w3.eth.wait_for_transaction_receipt(tx_hash)
        linked_events = self._contract.events.SourceNFTLinked().process_receipt(receipt, errors=DISCARD)
        if not linked_events:
            raise RuntimeError("register_with_source: SourceNFTLinked event not found in transaction receipt")
        return linked_events[0]["args"]["agentId"]

    def register_with_source_and_uri(
        self, agent_uri: str, source_token_id: int, value: int = 0
    ) -> int:
        """Deployment-specific extension of register_with_source (NOT
        ERC-8323 base spec) that lets the caller override the minted
        agent's URI instead of the registry's baseAgentURI template.
        Confirmed live on Merlini's AgentIdentityRegistry (mainnet
        0xe0454dfa17a57a84c3e0e2dbfda5318cbbe91e2c, 2026-07-16 Telegram) --
        only call this against a registry known to implement the exact same
        overload; a bare ERC-8323 registry does not have to expose it.
        """
        fn = self._contract.get_function_by_signature("registerWithSource(string,uint256)")
        tx_hash = fn(agent_uri, source_token_id).transact({"value": value})
        receipt = self._w3.eth.wait_for_transaction_receipt(tx_hash)
        linked_events = self._contract.events.SourceNFTLinked().process_receipt(receipt, errors=DISCARD)
        if not linked_events:
            raise RuntimeError(
                "register_with_source_and_uri: SourceNFTLinked event not found in transaction receipt"
            )
        return linked_events[0]["args"]["agentId"]

    def register_with_source_and_metadata(
        self,
        agent_uri: str,
        source_token_id: int,
        metadata: list[MetadataEntry] | None = None,
        value: int = 0,
    ) -> int:
        """Deployment-specific extension of register_with_source (NOT
        ERC-8323 base spec) that seeds initial metadata entries at
        registration time, on top of the custom-URI overload above. Same
        provenance/caveat as register_with_source_and_uri.
        """
        metadata_tuples = [(entry.metadata_key, entry.metadata_value) for entry in (metadata or [])]
        fn = self._contract.get_function_by_signature(
            "registerWithSource(string,uint256,(string,bytes)[])"
        )
        tx_hash = fn(agent_uri, source_token_id, metadata_tuples).transact({"value": value})
        receipt = self._w3.eth.wait_for_transaction_receipt(tx_hash)
        linked_events = self._contract.events.SourceNFTLinked().process_receipt(receipt, errors=DISCARD)
        if not linked_events:
            raise RuntimeError(
                "register_with_source_and_metadata: SourceNFTLinked event not found in transaction receipt"
            )
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
