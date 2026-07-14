import time

from eth_abi import encode
from eth_account import Account
from eth_utils import keccak
from web3 import Web3

from agent_sdk.verify.erc8299.client import JudgmentExecutionClient, WyriweAttestationClient


def _make_clients(deploy_contracts, anvil_rpc_url, anvil_account):
    wyriwe_address, judgment_address = deploy_contracts("verify/ERC8299", "DeployERC8299")
    account = Account.from_key(anvil_account(1)["privateKey"])
    wyriwe_client = WyriweAttestationClient(anvil_rpc_url, wyriwe_address, account)
    judgment_client = JudgmentExecutionClient(anvil_rpc_url, judgment_address, account)
    return wyriwe_client, judgment_client, wyriwe_address, judgment_address


class TestWyriweAttestationClient:
    def test_exposes_proof_system_identifier(self, deploy_contracts, anvil_rpc_url, anvil_account):
        wyriwe_client, _, _, _ = _make_clients(deploy_contracts, anvil_rpc_url, anvil_account)
        assert wyriwe_client.proof_system() == "attestation/wyriwe"

    def test_accepts_valid_proof_and_rejects_invalid_one(self, deploy_contracts, anvil_rpc_url, anvil_account):
        wyriwe_client, _, _, _ = _make_clients(deploy_contracts, anvil_rpc_url, anvil_account)
        account = Account.from_key(anvil_account(1)["privateKey"])
        now = int(time.time())

        agent_id = Web3.keccak(text="agent-1")
        registry = account.address
        model_hash = Web3.keccak(text="model-v1")
        raw_input_hash = Web3.keccak(text="raw input")
        sanitization_pipeline_hash = Web3.keccak(text="sanitization pipeline")
        input_hash = Web3.keccak(text="sanitized input")
        output_hash = Web3.keccak(text="model output")
        timestamp = now

        # Encode as a tuple to match Solidity's abi.encode(struct)
        encoded = encode(
            ["(bytes32,address,bytes32,bytes32,bytes32,bytes32,bytes32,uint256)"],
            [(agent_id, registry, model_hash, raw_input_hash, sanitization_pipeline_hash, input_hash, output_hash, timestamp)],
        )
        valid_signature = keccak(encoded)

        attestation = {
            "agentId": agent_id,
            "registry": registry,
            "modelHash": model_hash,
            "rawInputHash": raw_input_hash,
            "sanitizationPipelineHash": sanitization_pipeline_hash,
            "inputHash": input_hash,
            "outputHash": output_hash,
            "timestamp": timestamp,
        }

        result = wyriwe_client.verify(attestation, valid_signature)
        assert result is True

        # Invalid signature (random bytes)
        invalid_signature = b"\xff" * 32
        invalid_result = wyriwe_client.verify(attestation, invalid_signature)
        assert invalid_result is False


class TestJudgmentExecutionClient:
    def test_exposes_proof_system_identifier(self, deploy_contracts, anvil_rpc_url, anvil_account):
        _, judgment_client, _, _ = _make_clients(deploy_contracts, anvil_rpc_url, anvil_account)
        assert judgment_client.proof_system() == "attestation/judgment"

    def test_accepts_valid_proof_and_rejects_invalid_one(self, deploy_contracts, anvil_rpc_url, anvil_account):
        _, judgment_client, _, _ = _make_clients(deploy_contracts, anvil_rpc_url, anvil_account)
        account = Account.from_key(anvil_account(1)["privateKey"])
        now = int(time.time())

        agent_id = Web3.keccak(text="executing-agent")
        registry = account.address
        validator_id = Web3.keccak(text="validator")
        raw_proposal_hash = Web3.keccak(text="proposal")
        verdict_hash = Web3.keccak(text="verdict")
        executed_action_hash = Web3.keccak(text="executed action")
        verdict_timestamp = now - 3600
        executed_timestamp = now
        record_pointer = "https://example.com/record/1"

        # Encode as a tuple to match Solidity's abi.encode(struct)
        encoded = encode(
            ["(bytes32,address,bytes32,bytes32,bytes32,bytes32,uint256,uint256,string)"],
            [(agent_id, registry, validator_id, raw_proposal_hash, verdict_hash, executed_action_hash, verdict_timestamp, executed_timestamp, record_pointer)],
        )
        valid_signature = keccak(encoded)

        attestation = {
            "agentId": agent_id,
            "registry": registry,
            "validatorId": validator_id,
            "rawProposalHash": raw_proposal_hash,
            "verdictHash": verdict_hash,
            "executedActionHash": executed_action_hash,
            "verdictTimestamp": verdict_timestamp,
            "executedTimestamp": executed_timestamp,
            "recordPointer": record_pointer,
        }

        result = judgment_client.verify(attestation, valid_signature)
        assert result is True

        # Invalid signature (random bytes)
        invalid_signature = b"\xaa" * 32
        invalid_result = judgment_client.verify(attestation, invalid_signature)
        assert invalid_result is False
