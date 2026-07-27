// SPDX-License-Identifier: MIT
pragma solidity 0.8.25;

import {IBoundedAgentAction} from "@agent-ercs/metering/ERC8312/IBoundedAgentAction.sol";
import {IContestableEnvelope} from "@agent-ercs/metering/ERC8312/IContestableEnvelope.sol";

/// @title MockContestableEnvelope
/// @notice Reference implementation of IContestableEnvelope for local testing
///         only. Anyone can contest (no bond requirement) and only the contract
///         deployer can resolve. This is a minimal honest interpretation for
///         testing the contest → resolve lifecycle.
///         This contract replicates IBoundedAgentAction state inline so it
///         can be deployed as a standalone mock without inheriting from a base.
contract MockContestableEnvelope is IContestableEnvelope {
    uint64 private _nextId;

    struct InternalEnvelope {
        bytes32 id;
        address principal;
        bytes32 capabilityRoot;
        bytes32 cursorRoot;
        uint64 createdAt;
        uint64 expiresAt;
        Status status;
    }

    mapping(bytes32 id => InternalEnvelope) private _envelopes;

    /// @notice Deploy with id counter starting at 1.
    constructor() {
        _nextId = 1;
    }

    // ── IBoundedAgentAction (via IContestableEnvelope inheritance) ──────────

    function registerEnvelope(
        address principal,
        bytes32 capabilityRoot,
        uint64 expiresAt,
        bytes calldata /* initData */
    ) external override returns (bytes32 id) {
        id = keccak256(abi.encodePacked(address(this), _nextId++));
        require(_envelopes[id].createdAt == 0, "id collision");

        _envelopes[id] = InternalEnvelope({
            id: id,
            principal: principal,
            capabilityRoot: capabilityRoot,
            cursorRoot: bytes32(0),
            createdAt: uint64(block.timestamp),
            expiresAt: expiresAt,
            status: Status.Active
        });

        emit EnvelopeRegistered(id, principal, capabilityRoot);
    }

    function advanceCursor(
        bytes32 id,
        bytes calldata witness
    ) external override returns (bytes32 newCursor) {
        InternalEnvelope storage env = _requireActive(id);

        newCursor = keccak256(abi.encodePacked(env.cursorRoot, witness));
        bytes32 prevCursor = env.cursorRoot;
        env.cursorRoot = newCursor;

        emit EnvelopeAdvanced(id, prevCursor, newCursor);
    }

    function setStatus(bytes32 id, Status newStatus) external override {
        InternalEnvelope storage env = _envelopes[id];
        require(env.createdAt != 0, "unknown envelope");
        Status oldStatus = _effectiveStatus(env);
        env.status = newStatus;
        emit EnvelopeStatusChanged(id, oldStatus, newStatus);
    }

    function getEnvelope(bytes32 id) external view override returns (Envelope memory) {
        InternalEnvelope memory env = _envelopes[id];
        require(env.createdAt != 0, "unknown envelope");
        return Envelope({
            id: env.id,
            principal: env.principal,
            capabilityRoot: env.capabilityRoot,
            cursorRoot: env.cursorRoot,
            createdAt: env.createdAt,
            expiresAt: env.expiresAt,
            status: _effectiveStatus(env)
        });
    }

    function getCursor(bytes32 id) external view override returns (bytes32) {
        InternalEnvelope storage env = _envelopes[id];
        require(env.createdAt != 0, "unknown envelope");
        return env.cursorRoot;
    }

    function getStatus(bytes32 id) external view override returns (Status) {
        InternalEnvelope storage env = _envelopes[id];
        require(env.createdAt != 0, "unknown envelope");
        return _effectiveStatus(env);
    }

    function isActive(bytes32 id) external view override returns (bool) {
        InternalEnvelope storage env = _envelopes[id];
        if (env.createdAt == 0) return false;
        return _effectiveStatus(env) == Status.Active;
    }

    // ── IContestableEnvelope ─────────────────────────────────────────────────

    function contest(bytes32 id, bytes calldata /* evidence */) external override {
        InternalEnvelope storage env = _envelopes[id];
        require(env.createdAt != 0, "unknown envelope");
        require(_effectiveStatus(env) == Status.Active, "not contestable");

        env.status = Status.Contested;
        emit EnvelopeContested(id, msg.sender);
    }

    function resolve(bytes32 id, Status outcome, bytes calldata /* resolution */) external override {
        InternalEnvelope storage env = _envelopes[id];
        require(env.createdAt != 0, "unknown envelope");
        require(env.status == Status.Contested, "not contested");
        require(outcome == Status.Active || outcome == Status.Revoked, "invalid outcome");

        Status fromStatus = env.status;
        env.status = outcome;
        emit EnvelopeResolved(id, outcome);
        emit EnvelopeStatusChanged(id, fromStatus, outcome);
    }

    // ── Internals ────────────────────────────────────────────────────────────

    function _requireActive(bytes32 id) private view returns (InternalEnvelope storage) {
        InternalEnvelope storage env = _envelopes[id];
        require(env.createdAt != 0, "unknown envelope");
        require(_effectiveStatus(env) == Status.Active, "not active");
        return env;
    }

    function _effectiveStatus(InternalEnvelope memory env) private view returns (Status) {
        if (block.timestamp >= env.expiresAt && env.status == Status.Active) {
            return Status.Expired;
        }
        return env.status;
    }

    // ── ERC-165 ──────────────────────────────────────────────────────────────

    function supportsInterface(bytes4 interfaceId) external pure override returns (bool) {
        return interfaceId == type(IContestableEnvelope).interfaceId
            || interfaceId == type(IBoundedAgentAction).interfaceId
            || interfaceId == 0x01ffc9a7;
    }
}
