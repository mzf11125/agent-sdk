// SPDX-License-Identifier: MIT
pragma solidity 0.8.25;

import "@agent-ercs/execution/ERC8301/IAgentWorkflow.sol";

/// @title MockAgentWorkflow
/// @notice Reference implementation of IAgentWorkflow for local testing only.
///         Not audited, not for production use — see agent-ercs's README on
///         interface vs. base-implementation vs. example/reference contracts.
///
///         This is a minimal storage-based mock: it records tasks and replies
///         by their computed hashes, tracks runs, and emits the required events.
///         It does NOT implement an FSM or gating logic — it always dispatches
///         the initial task and accepts any reply without gate validation.
contract MockAgentWorkflow is IAgentWorkflow {
    // ── State ────────────────────────────────────────────────────────────

    struct RunState {
        uint8 status; // RunStatus as uint8 to avoid Solidity enum naming issues
        bytes32 finalTaskHash;
        uint256 completedAt;
    }

    /// @notice taskHash => AgentTask storage (packed via storeTask)
    struct StoredTask {
        uint8 stage;
        uint256 taskSeq;
        bytes32 inputHash;
        bytes input;
        uint256 timestamp;
        uint256 expiresAt;
        bytes32[] prevReplyHashes;
        bytes32 workflowRunId;
        bool exists;
    }

    /// @notice replyHash => AgentReply storage
    struct StoredReply {
        bytes32 outputHash;
        bytes output;
        uint256 timestamp;
        address replier;
        bytes32[] prevTaskHashes;
        bytes32 workflowRunId;
        bool exists;
        address verifier;
        bool proven;
        bytes32 verificationDigest;
    }

    mapping(bytes32 => StoredTask) private _tasks;
    mapping(bytes32 => StoredReply) private _replies;
    mapping(bytes32 => RunState) private _runs;
    uint256 private _nonce;

    // ── IAgentWorkflow ───────────────────────────────────────────────────

    function run(
        bytes32 inputHash,
        bytes calldata input,
        uint256 expiresAt
    ) external returns (bytes32 workflowRunId) {
        workflowRunId = keccak256(abi.encodePacked(blockhash(block.number - 1), msg.sender, ++_nonce));

        bytes32[] memory emptyHashes;
        bytes32 taskHash = _computeTaskHash(
            1, 0, inputHash, block.timestamp, expiresAt, emptyHashes, workflowRunId
        );

        _runs[workflowRunId] = RunState({status: uint8(RunStatus.Pending), finalTaskHash: 0, completedAt: 0});

        _storeTask(taskHash, 1, 0, inputHash, input, block.timestamp, expiresAt, emptyHashes, workflowRunId);

        emit NewAgentTask(workflowRunId, 1, taskHash);
    }

    function result(bytes32 workflowRunId)
        external view
        returns (RunStatus status, bytes32 finalTaskHash, uint256 completedAt)
    {
        RunState storage rs = _runs[workflowRunId];
        return (RunStatus(rs.status), rs.finalTaskHash, rs.completedAt);
    }

    function getAgentTask(bytes32 taskHash)
        external view
        returns (AgentTask memory task, bool proven)
    {
        StoredTask storage st = _tasks[taskHash];
        require(st.exists, "MockAgentWorkflow: task not found");
        return (_toAgentTask(st), st.prevReplyHashes.length == 0);
    }

    function getAgentReply(bytes32 replyHash)
        external view
        returns (AgentReply memory reply, address verifier, bool proven, bytes32 verificationDigest)
    {
        StoredReply storage sr = _replies[replyHash];
        require(sr.exists, "MockAgentWorkflow: reply not found");
        return (_toAgentReply(sr), sr.verifier, sr.proven, sr.verificationDigest);
    }

    function onAgentReply(AgentReply calldata reply) external {
        bytes32 replyHash = _computeReplyHash(reply);

        StoredReply storage sr = _replies[replyHash];
        require(!sr.exists, "MockAgentWorkflow: duplicate reply");

        _storeReplyFromCalldata(replyHash, reply);

        emit NewAgentTask(reply.workflowRunId, 2, bytes32(0));
    }

    function onAgentProve(bytes32[] calldata replyHashes, bytes calldata proof) external {
        for (uint256 i = 0; i < replyHashes.length; i++) {
            StoredReply storage sr = _replies[replyHashes[i]];
            require(sr.exists, "MockAgentWorkflow: reply not found");
            sr.proven = true;
            sr.verifier = msg.sender;
            sr.verificationDigest = keccak256(proof);
        }
    }

    // ── Internal ─────────────────────────────────────────────────────────

    function _computeTaskHash(
        uint8 stage,
        uint256 taskSeq,
        bytes32 inputHash,
        uint256 timestamp,
        uint256 expiresAt,
        bytes32[] memory prevReplyHashes,
        bytes32 workflowRunId
    ) internal pure returns (bytes32) {
        bytes32 innerHash = keccak256(abi.encodePacked(prevReplyHashes));
        return keccak256(abi.encode(stage, taskSeq, inputHash, timestamp, expiresAt, innerHash, workflowRunId));
    }

    function _computeReplyHash(AgentReply memory reply) internal pure returns (bytes32) {
        bytes32 innerHash = keccak256(abi.encodePacked(reply.prevTaskHashes));
        return keccak256(abi.encode(reply.outputHash, reply.timestamp, reply.replier, innerHash, reply.workflowRunId));
    }

    function _storeTask(
        bytes32 taskHash,
        uint8 stage,
        uint256 taskSeq,
        bytes32 inputHash,
        bytes memory input,
        uint256 timestamp,
        uint256 expiresAt,
        bytes32[] memory prevReplyHashes,
        bytes32 workflowRunId
    ) internal {
        StoredTask storage st = _tasks[taskHash];
        st.stage = stage;
        st.taskSeq = taskSeq;
        st.inputHash = inputHash;
        st.input = input;
        st.timestamp = timestamp;
        st.expiresAt = expiresAt;
        st.prevReplyHashes = prevReplyHashes;
        st.workflowRunId = workflowRunId;
        st.exists = true;
    }

    function _storeReplyFromCalldata(bytes32 replyHash, AgentReply calldata reply) internal {
        StoredReply storage sr = _replies[replyHash];
        sr.outputHash = reply.outputHash;
        sr.output = reply.output;
        sr.timestamp = reply.timestamp;
        sr.replier = reply.replier;
        sr.prevTaskHashes = reply.prevTaskHashes;
        sr.workflowRunId = reply.workflowRunId;
        sr.exists = true;
        sr.verifier = address(0);
        sr.proven = false;
        sr.verificationDigest = bytes32(0);
    }

    function _toAgentTask(StoredTask storage st) internal view returns (AgentTask memory) {
        return AgentTask({
            stage: st.stage,
            taskSeq: st.taskSeq,
            inputHash: st.inputHash,
            input: st.input,
            timestamp: st.timestamp,
            expiresAt: st.expiresAt,
            prevReplyHashes: st.prevReplyHashes,
            workflowRunId: st.workflowRunId
        });
    }

    function _toAgentReply(StoredReply storage sr) internal view returns (AgentReply memory) {
        return AgentReply({
            outputHash: sr.outputHash,
            output: sr.output,
            timestamp: sr.timestamp,
            replier: sr.replier,
            prevTaskHashes: sr.prevTaskHashes,
            workflowRunId: sr.workflowRunId
        });
    }
}
