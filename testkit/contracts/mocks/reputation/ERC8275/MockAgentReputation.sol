// SPDX-License-Identifier: MIT
pragma solidity 0.8.25;

import {IAgentReputation} from "@agent-ercs/reputation/ERC8275/IAgentReputation.sol";

/// @title MockAgentReputation
/// @notice Minimal reference implementation of IAgentReputation for local testing only.
///         Stores reputation data in a simple mapping so tests can pre-seed known
///         return values. Not audited, not for production use.
contract MockAgentReputation is IAgentReputation {
    struct ReputationData {
        uint64 completedOrders;
        uint64 disputedOrders;
        uint64 totalVolume;
        uint64 lastActiveAt;
        uint16 score;
    }

    mapping(bytes32 => ReputationData) private _reputations;
    mapping(bytes32 => uint16) private _decayWeights;
    mapping(bytes32 => mapping(bytes32 => bool)) private _outcomes;

    function setReputation(
        bytes32 agentId,
        uint64 completedOrders,
        uint64 disputedOrders,
        uint64 totalVolume,
        uint64 lastActiveAt,
        uint16 score
    ) external {
        _reputations[agentId] = ReputationData({
            completedOrders: completedOrders,
            disputedOrders: disputedOrders,
            totalVolume: totalVolume,
            lastActiveAt: lastActiveAt,
            score: score
        });
    }

    function setDecayWeight(bytes32 agentId, uint16 weight) external {
        _decayWeights[agentId] = weight;
    }

    function setOutcome(bytes32 orderId, bytes32 agentId, bool valid) external {
        _outcomes[orderId][agentId] = valid;
    }

    function getReputation(bytes32 agentId)
        external
        view
        override
        returns (
            uint64 completedOrders,
            uint64 disputedOrders,
            uint64 totalVolume,
            uint64 lastActiveAt,
            uint16 score
        )
    {
        ReputationData memory data = _reputations[agentId];
        return (data.completedOrders, data.disputedOrders, data.totalVolume, data.lastActiveAt, data.score);
    }

    function getDecayWeight(bytes32 agentId) external view override returns (uint16 weight) {
        return _decayWeights[agentId];
    }

    function verifyOutcome(bytes32 orderId, bytes calldata proof) external view override returns (bool valid) {
        // proof is the agentId encoded as bytes32 (left-padded)
        bytes32 agentId = bytesToBytes32(proof);
        return _outcomes[orderId][agentId];
    }

    function validatorType() external pure returns (ValidatorType) {
        return ValidatorType.ReExecution;
    }

    function claimType() external pure returns (ClaimType) {
        return ClaimType.ReExecution;
    }

    /// @dev Decode the first 32 bytes of `data` as bytes32 (left-padded zeros allowed).
    function bytesToBytes32(bytes memory data) internal pure returns (bytes32 result) {
        if (data.length >= 32) {
            assembly {
                result := mload(add(data, 32))
            }
        }
    }
}
