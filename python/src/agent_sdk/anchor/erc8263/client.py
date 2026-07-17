from __future__ import annotations

from eth_account.signers.local import LocalAccount
from web3 import Web3
from web3.middleware import SignAndSendRawMiddlewareBuilder

from .abi import ON_CHAIN_PROOF_ABI


class OnChainProofClient:
    """Client for ERC-8263 OnChainProof anchor."""

    def __init__(self, rpc_url: str, address: str, account: LocalAccount):
        self._w3 = Web3(Web3.HTTPProvider(rpc_url))
        self._w3.middleware_onion.add(SignAndSendRawMiddlewareBuilder.build(account))
        self._w3.eth.default_account = account.address
        self._contract = self._w3.eth.contract(
            address=Web3.to_checksum_address(address), abi=ON_CHAIN_PROOF_ABI
        )

    def anchor(self, agent_id_scheme: int, agent_id: str, proof_hash: str) -> bytes:
        """Anchor a proof hash with empty aux bytes.

        Args:
            agent_id_scheme: Identity scheme byte (0x00, 0x01, or 0x02).
            agent_id: 32-byte agent identifier as a 0x-prefixed hex string.
            proof_hash: Non-zero 32-byte commitment as a 0x-prefixed hex string.

        Returns:
            Transaction receipt hash.
        """
        return self._send("anchor", agent_id_scheme, Web3.to_bytes(hexstr=agent_id), Web3.to_bytes(hexstr=proof_hash))

    def anchor_with_aux(self, agent_id_scheme: int, agent_id: str, proof_hash: str, aux: bytes) -> bytes:
        """Anchor a proof hash with opaque aux bytes.

        Args:
            agent_id_scheme: Identity scheme byte (0x00, 0x01, or 0x02).
            agent_id: 32-byte agent identifier as a 0x-prefixed hex string.
            proof_hash: Non-zero 32-byte commitment as a 0x-prefixed hex string.
            aux: Opaque extension bytes.

        Returns:
            Transaction receipt hash.
        """
        return self._send("anchorWithAux", agent_id_scheme, Web3.to_bytes(hexstr=agent_id), Web3.to_bytes(hexstr=proof_hash), aux)

    def _send(self, function_name: str, *args) -> bytes:
        tx_hash = getattr(self._contract.functions, function_name)(*args).transact()
        receipt = self._w3.eth.wait_for_transaction_receipt(tx_hash)
        return receipt["transactionHash"]
