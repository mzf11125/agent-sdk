// SPDX-License-Identifier: MIT
pragma solidity 0.8.25;

import {Test} from "forge-std/Test.sol";
import {IAgentVerifier} from "@agent-ercs/verify/ERC8274/IAgentVerifier.sol";
import {MockProofVerifier} from "../../../contracts/mocks/verify/ERC8274/MockProofVerifier.sol";
import {MockAgentVerifier} from "../../../contracts/mocks/verify/ERC8274/MockAgentVerifier.sol";
import {MockAgentVerifiable} from "../../../contracts/mocks/verify/ERC8274/MockAgentVerifiable.sol";

contract MockVerifiersTest is Test {
    MockProofVerifier internal proofVerifier;
    MockAgentVerifier internal agentVerifier;
    MockAgentVerifiable internal agentVerifiable;

    function setUp() public {
        proofVerifier = new MockProofVerifier();
        agentVerifier = new MockAgentVerifier(proofVerifier);
        agentVerifiable = new MockAgentVerifiable(address(agentVerifier));
    }

    function test_proofVerifierAcceptsValidProof() public view {
        bytes32 inputHash = keccak256("input");
        bytes32 outputHash = keccak256("output");
        bytes memory metadata = "";
        bytes memory proof = abi.encodePacked(keccak256(abi.encodePacked(inputHash, outputHash, metadata)));

        assertTrue(proofVerifier.verify(inputHash, outputHash, metadata, proof));
    }

    function test_proofVerifierRejectsInvalidProof() public view {
        bytes32 inputHash = keccak256("input");
        bytes32 outputHash = keccak256("output");
        bytes memory metadata = "";
        bytes memory badProof = abi.encodePacked(keccak256("not-the-right-digest"));

        assertFalse(proofVerifier.verify(inputHash, outputHash, metadata, badProof));
    }

    function test_proofVerifierMetadata() public view {
        assertEq(proofVerifier.proofSystem(), "mock-test-only");
        assertEq(proofVerifier.proofProfile(), keccak256("mock-test-only-v1"));
    }

    function test_agentVerifierRoutesToProofVerifierAndEmitsDigest() public {
        bytes32 taskId = keccak256("task-1");
        bytes32 agentId = keccak256("agent-1");
        bytes32 inputHash = keccak256("input");
        bytes32 outputHash = keccak256("output");
        bytes memory proof = abi.encodePacked(keccak256(abi.encodePacked(inputHash, outputHash, bytes(""))));

        bytes32 expectedProfile = proofVerifier.proofProfile();
        bytes32 expectedDigest = keccak256(abi.encode(taskId, agentId, inputHash, outputHash, true, expectedProfile));

        vm.expectEmit(true, true, false, true);
        emit IAgentVerifier.VerificationCompleted(taskId, agentId, inputHash, outputHash, true, expectedDigest);

        (bool valid, bytes32 digest) = agentVerifier.verify(taskId, agentId, inputHash, outputHash, proof);

        assertTrue(valid);
        assertEq(digest, expectedDigest);
    }

    function test_agentVerifierReturnsInvalidForBadProof() public {
        bytes32 taskId = keccak256("task-2");
        bytes32 agentId = keccak256("agent-2");
        bytes32 inputHash = keccak256("input");
        bytes32 outputHash = keccak256("output");
        bytes memory badProof = abi.encodePacked(keccak256("garbage"));

        (bool valid,) = agentVerifier.verify(taskId, agentId, inputHash, outputHash, badProof);

        assertFalse(valid);
    }

    function test_agentVerifiableDeclaresAndUpdatesVerifier() public {
        assertEq(agentVerifiable.agentVerifier(), address(agentVerifier));

        address newVerifier = address(0xBEEF);
        agentVerifiable.setAgentVerifier(newVerifier);

        assertEq(agentVerifiable.agentVerifier(), newVerifier);
    }
}
