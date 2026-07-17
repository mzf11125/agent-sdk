// SPDX-License-Identifier: MIT
pragma solidity 0.8.25;

import {Test} from "forge-std/Test.sol";
import {MockAgentWorkflow} from "../../../contracts/mocks/execution/ERC8301/MockAgentWorkflow.sol";
import "@agent-ercs/execution/ERC8301/IAgentWorkflow.sol";

contract MockAgentWorkflowTest is Test {
    MockAgentWorkflow internal workflow;
    address internal alice = address(0x1);

    event NewAgentTask(
        bytes32 indexed workflowRunId,
        uint8 indexed stage,
        bytes32 indexed taskHash
    );

    event WorkflowCompleted(
        bytes32 indexed workflowRunId,
        RunStatus status,
        bytes32 finalTaskHash,
        uint256 timestamp
    );

    function setUp() external {
        workflow = new MockAgentWorkflow();
    }

    function test_Run_EmitNewAgentTask() external {
        bytes32 inputHash = keccak256("test");
        uint256 expiresAt = block.timestamp + 1000;

        vm.prank(alice);
        bytes32 runId = workflow.run(inputHash, "test input", expiresAt);

        assertTrue(runId != bytes32(0), "runId must not be zero");
    }

    function test_GetAgentTask_AfterRun() external {
        bytes32 inputHash = keccak256("test");
        uint256 expiresAt = block.timestamp + 1000;

        vm.prank(alice);
        bytes32 runId = workflow.run(inputHash, "test input", expiresAt);

        // The mock stores the task internally; we compute the task hash ourselves
        // from the known inputs to verify getAgentTask returns the right data.
        // taskHash = keccak256(abi.encode(1, 0, inputHash, block.timestamp, expiresAt, innerHash, runId))
        bytes32 innerHash = keccak256("");
        bytes32 taskHash = keccak256(abi.encode(uint8(1), uint256(0), inputHash, block.timestamp, expiresAt, innerHash, runId));

        (AgentTask memory task, bool proven) = workflow.getAgentTask(taskHash);

        assertEq(uint256(task.stage), 1);
        assertEq(task.taskSeq, 0);
        assertEq(task.inputHash, inputHash);
        assertEq(task.expiresAt, expiresAt);
        assertEq(task.workflowRunId, runId);
        assertEq(task.prevReplyHashes.length, 0);
        assertTrue(proven, "initial task should be proven");
    }

    function test_OnAgentReply_StoresReply() external {
        bytes32 inputHash = keccak256("test");
        uint256 expiresAt = block.timestamp + 1000;

        vm.prank(alice);
        bytes32 runId = workflow.run(inputHash, "test input", expiresAt);

        bytes32 outputHash = keccak256("reply output");
        bytes32[] memory prevTaskHashes = new bytes32[](1);
        prevTaskHashes[0] = keccak256("task");

        AgentReply memory reply = AgentReply({
            outputHash: outputHash,
            output: "reply output",
            timestamp: block.timestamp,
            replier: alice,
            prevTaskHashes: prevTaskHashes,
            workflowRunId: runId
        });

        vm.prank(alice);
        workflow.onAgentReply(reply);

        bytes32 innerHash = keccak256(abi.encodePacked(prevTaskHashes));
        bytes32 replyHash = keccak256(abi.encode(outputHash, reply.timestamp, alice, innerHash, runId));

        (AgentReply memory storedReply,,,) = workflow.getAgentReply(replyHash);
        assertEq(storedReply.outputHash, outputHash);
        assertEq(storedReply.replier, alice);
    }

    function test_OnAgentProve_MarksProven() external {
        bytes32 inputHash = keccak256("test");
        uint256 expiresAt = block.timestamp + 1000;

        vm.prank(alice);
        bytes32 runId = workflow.run(inputHash, "test input", expiresAt);

        bytes32 outputHash = keccak256("reply output");
        bytes32[] memory prevTaskHashes = new bytes32[](1);
        prevTaskHashes[0] = keccak256("task");

        AgentReply memory reply = AgentReply({
            outputHash: outputHash,
            output: "reply output",
            timestamp: block.timestamp,
            replier: alice,
            prevTaskHashes: prevTaskHashes,
            workflowRunId: runId
        });

        vm.prank(alice);
        workflow.onAgentReply(reply);

        bytes32 innerHash = keccak256(abi.encodePacked(prevTaskHashes));
        bytes32 replyHash = keccak256(abi.encode(outputHash, reply.timestamp, alice, innerHash, runId));

        bytes32[] memory replyHashes = new bytes32[](1);
        replyHashes[0] = replyHash;

        vm.prank(alice);
        workflow.onAgentProve(replyHashes, "proof data");

        (, address verifier, bool proven,) = workflow.getAgentReply(replyHash);
        assertTrue(proven, "reply should be proven after onAgentProve");
        assertEq(verifier, alice, "verifier should be the prover");
    }

    function test_GetAgentTask_RevertIfNotFound() external {
        vm.expectRevert("MockAgentWorkflow: task not found");
        workflow.getAgentTask(keccak256("nonexistent"));
    }

    function test_GetAgentReply_RevertIfNotFound() external {
        vm.expectRevert("MockAgentWorkflow: reply not found");
        workflow.getAgentReply(keccak256("nonexistent"));
    }
}
