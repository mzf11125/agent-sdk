from __future__ import annotations

from eth_account.signers.local import LocalAccount
from web3 import Web3
from web3.logs import DISCARD
from web3.middleware import SignAndSendRawMiddlewareBuilder

from .abi import AGENT_VERIFIABLE_ABI, AGENT_VERIFIER_ABI, PROOF_VERIFIER_ABI


class ProofVerifierClient:
    def __init__(self, rpc_url: str, address: str, account: LocalAccount):
        self._w3 = Web3(Web3.HTTPProvider(rpc_url))
        self._w3.middleware_onion.add(SignAndSendRawMiddlewareBuilder.build(account))
        self._w3.eth.default_account = account.address
        self._contract = self._w3.eth.contract(address=Web3.to_checksum_address(address), abi=PROOF_VERIFIER_ABI)

    # A read-only simulated call, not a broadcast transaction: this is a
    # pure cryptographic check with no state to persist, so anyone can
    # freely re-derive the answer without spending gas or holding a key
    # with funds — that's the whole point of exposing it this way.
    def verify(self, input_hash: bytes, output_hash: bytes, metadata: bytes, proof: bytes) -> bool:
        return self._contract.functions.verify(input_hash, output_hash, metadata, proof).call()

    def proof_system(self) -> str:
        return self._contract.functions.proofSystem().call()

    def proof_profile(self) -> bytes:
        return self._contract.functions.proofProfile().call()


class AgentVerifierClient:
    def __init__(self, rpc_url: str, address: str, account: LocalAccount):
        self._w3 = Web3(Web3.HTTPProvider(rpc_url))
        self._w3.middleware_onion.add(SignAndSendRawMiddlewareBuilder.build(account))
        self._w3.eth.default_account = account.address
        self._contract = self._w3.eth.contract(address=Web3.to_checksum_address(address), abi=AGENT_VERIFIER_ABI)

    # Broadcast, not simulated: unlike ProofVerifierClient.verify(), this
    # records a VerificationCompleted event on-chain — that log is the
    # point of calling it, so a real transaction is warranted here.
    def verify(
        self, task_id: bytes, agent_id: bytes, input_hash: bytes, output_hash: bytes, proof: bytes
    ) -> tuple[bool, bytes]:
        tx_hash = self._contract.functions.verify(task_id, agent_id, input_hash, output_hash, proof).transact()
        receipt = self._w3.eth.wait_for_transaction_receipt(tx_hash)
        events = self._contract.events.VerificationCompleted().process_receipt(receipt, errors=DISCARD)
        if not events:
            raise RuntimeError("verify: VerificationCompleted event not found in transaction receipt")
        args = events[0]["args"]
        return args["valid"], args["verificationDigest"]


# A standalone function, not a client class: IAgentVerifiable is a single
# getter, so a class would be one method wrapping a constructor for no
# benefit.
def get_trusted_verifier(rpc_url: str, settlement_address: str) -> str:
    w3 = Web3(Web3.HTTPProvider(rpc_url))
    contract = w3.eth.contract(address=Web3.to_checksum_address(settlement_address), abi=AGENT_VERIFIABLE_ABI)
    return contract.functions.agentVerifier().call()
