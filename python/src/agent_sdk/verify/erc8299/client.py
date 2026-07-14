from __future__ import annotations

from eth_account.signers.local import LocalAccount
from web3 import Web3
from web3.middleware import SignAndSendRawMiddlewareBuilder

from .abi import JUDGMENT_EXECUTION_ABI, WYRIWE_ATTESTATION_ABI


class WyriweAttestationClient:
    def __init__(self, rpc_url: str, address: str, account: LocalAccount):
        self._w3 = Web3(Web3.HTTPProvider(rpc_url))
        self._w3.middleware_onion.add(SignAndSendRawMiddlewareBuilder.build(account))
        self._w3.eth.default_account = account.address
        self._contract = self._w3.eth.contract(
            address=Web3.to_checksum_address(address), abi=WYRIWE_ATTESTATION_ABI
        )

    # A read-only simulated call, not a broadcast transaction: this checks
    # an EIP-712 signature against the known attestor. Anyone can re-derive
    # the answer without spending gas or holding a funded key — that's the
    # whole point of exposing it this way.
    def verify(self, attestation: dict, signature: bytes) -> bool:
        return self._contract.functions.verify(attestation, signature).call()

    def proof_system(self) -> str:
        return self._contract.functions.proofSystem().call()


class JudgmentExecutionClient:
    def __init__(self, rpc_url: str, address: str, account: LocalAccount):
        self._w3 = Web3(Web3.HTTPProvider(rpc_url))
        self._w3.middleware_onion.add(SignAndSendRawMiddlewareBuilder.build(account))
        self._w3.eth.default_account = account.address
        self._contract = self._w3.eth.contract(
            address=Web3.to_checksum_address(address), abi=JUDGMENT_EXECUTION_ABI
        )

    # A read-only simulated call, not a broadcast transaction: this checks
    # an EIP-712 signature against the known attestor. Anyone can re-derive
    # the answer without spending gas or holding a funded key — that's the
    # whole point of exposing it this way.
    def verify(self, attestation: dict, signature: bytes) -> bool:
        return self._contract.functions.verify(attestation, signature).call()

    def proof_system(self) -> str:
        return self._contract.functions.proofSystem().call()
