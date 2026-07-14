// SPDX-License-Identifier: MIT
pragma solidity 0.8.25;

import {Test} from "forge-std/Test.sol";
import {IJudgmentExecutionAttestation} from "@agent-ercs/verify/ERC8299/IJudgmentExecutionAttestation.sol";
import {MockJudgmentExecutionAttestation} from "../../../contracts/mocks/verify/ERC8299/MockJudgmentExecutionAttestation.sol";

contract MockJudgmentExecutionAttestationTest is Test {
    MockJudgmentExecutionAttestation public judgment;

    address public attestor = address(0x1234);
    bytes32 public constant AGENT_ID = keccak256("agent");
    address public constant REGISTRY = address(0x5678);
    bytes32 public constant VALIDATOR_ID = keccak256("validator");
    bytes32 public constant RAW_PROPOSAL_HASH = keccak256("raw_proposal");
    bytes32 public constant VERDICT_HASH = keccak256("verdict");
    bytes32 public constant EXECUTED_ACTION_HASH = keccak256("executed_action");
    uint256 public constant VERDICT_TIMESTAMP = 1000000;
    uint256 public constant EXECUTED_TIMESTAMP = 2000000;
    string public constant RECORD_POINTER = "https://example.com/record/1";

    function setUp() public {
        judgment = new MockJudgmentExecutionAttestation();
    }

    function testProofSystem() public view {
        assertEq(judgment.proofSystem(), "attestation/judgment");
    }

    function testVerifyAcceptsValidSignature() public view {
        IJudgmentExecutionAttestation.JudgmentExecutionAttestation memory att = IJudgmentExecutionAttestation
            .JudgmentExecutionAttestation({
            agentId: AGENT_ID,
            registry: REGISTRY,
            validatorId: VALIDATOR_ID,
            rawProposalHash: RAW_PROPOSAL_HASH,
            verdictHash: VERDICT_HASH,
            executedActionHash: EXECUTED_ACTION_HASH,
            verdictTimestamp: VERDICT_TIMESTAMP,
            executedTimestamp: EXECUTED_TIMESTAMP,
            recordPointer: RECORD_POINTER
        });

        // The mock checks: keccak256(signature) == keccak256(abi.encode(attestation))
        bytes memory validSignature = abi.encodePacked(keccak256(abi.encode(att)));
        bool result = judgment.verify(att, validSignature);
        assertTrue(result);
    }

    function testVerifyReturnsFalseForInvalidSignature() public view {
        IJudgmentExecutionAttestation.JudgmentExecutionAttestation memory att = IJudgmentExecutionAttestation
            .JudgmentExecutionAttestation({
            agentId: AGENT_ID,
            registry: REGISTRY,
            validatorId: VALIDATOR_ID,
            rawProposalHash: RAW_PROPOSAL_HASH,
            verdictHash: VERDICT_HASH,
            executedActionHash: EXECUTED_ACTION_HASH,
            verdictTimestamp: VERDICT_TIMESTAMP,
            executedTimestamp: EXECUTED_TIMESTAMP,
            recordPointer: RECORD_POINTER
        });

        bytes memory invalidSignature = abi.encodePacked(bytes32(uint256(1)), bytes32(uint256(2)), uint8(27));
        bool result = judgment.verify(att, invalidSignature);
        assertFalse(result);
    }
}
