// SPDX-License-Identifier: MIT
pragma solidity 0.8.25;

import {IBoundedAgentAction} from "@agent-ercs/metering/ERC8312/IBoundedAgentAction.sol";

/// @title MockBoundedAgentAction
/// @notice Reference implementation of IBoundedAgentAction for local testing only.
///         Uses sequential IDs prefixed by deployer address. A real registry
///         would verify caller authorization and use a proper addressing scheme.
///         This mock emits every event the interface defines.
contract MockBoundedAgentAction is IBoundedAgentAction {
    uint64 private _nextId;

    mapping(bytes32 id => Envelope) private _envelopes;
    mapping(bytes32 id => Status) private _statusOverrides; // stored override

    /// @notice Deploy with id counter starting at 1.
    constructor() {
        _nextId = 1;
    }

    // ── Lifecycle ───────────────────────────────────────────────────────────

    function registerEnvelope(
        address principal,
        bytes32 capabilityRoot,
        uint64 expiresAt,
        bytes calldata /* initData */
    ) external override returns (bytes32 id) {
        id = keccak256(abi.encodePacked(address(this), _nextId++));
        require(_envelopes[id].createdAt == 0, "id collision");

        _envelopes[id] = Envelope({
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
        Envelope storage env = _requireActive(id);

        // "Advance" by hashing the witness with the previous cursor,
        // producing a deterministic new cursor — a valid public
        // commitment update for testing purposes.
        newCursor = keccak256(abi.encodePacked(env.cursorRoot, witness));
        bytes32 prevCursor = env.cursorRoot;
        env.cursorRoot = newCursor;

        emit EnvelopeAdvanced(id, prevCursor, newCursor);
    }

    function setStatus(bytes32 id, Status newStatus) external override {
        Envelope storage env = _envelopes[id];
        require(env.createdAt != 0, "unknown envelope");

        Status oldStatus = _effectiveStatus(env);
        // Minimal auth: anyone can setStatus for testing.
        env.status = newStatus;
        _statusOverrides[id] = newStatus;

        emit EnvelopeStatusChanged(id, oldStatus, newStatus);
    }

    // ── Views ───────────────────────────────────────────────────────────────

    function getEnvelope(bytes32 id) external view override returns (Envelope memory) {
        Envelope memory env = _envelopes[id];
        require(env.createdAt != 0, "unknown envelope");

        // Return with effective (not stored) status
        env.status = _effectiveStatus(env);
        return env;
    }

    function getCursor(bytes32 id) external view override returns (bytes32) {
        Envelope storage env = _envelopes[id];
        require(env.createdAt != 0, "unknown envelope");
        return env.cursorRoot;
    }

    function getStatus(bytes32 id) external view override returns (Status) {
        Envelope storage env = _envelopes[id];
        require(env.createdAt != 0, "unknown envelope");
        return _effectiveStatus(env);
    }

    function isActive(bytes32 id) external view override returns (bool) {
        Envelope storage env = _envelopes[id];
        if (env.createdAt == 0) return false;
        return _effectiveStatus(env) == Status.Active;
    }

    // ── Internals ────────────────────────────────────────────────────────────

    function _requireActive(bytes32 id) private view returns (Envelope storage) {
        Envelope storage env = _envelopes[id];
        require(env.createdAt != 0, "unknown envelope");
        require(_effectiveStatus(env) == Status.Active, "not active");
        return env;
    }

    function _effectiveStatus(Envelope memory env) private view returns (Status) {
        if (block.timestamp >= env.expiresAt && env.status == Status.Active) {
            return Status.Expired;
        }
        return env.status;
    }

    // ── ERC-165 ──────────────────────────────────────────────────────────────

    function supportsInterface(bytes4 interfaceId) external pure override returns (bool) {
        return interfaceId == type(IBoundedAgentAction).interfaceId
            || interfaceId == 0x01ffc9a7; // ERC-165
    }
}
