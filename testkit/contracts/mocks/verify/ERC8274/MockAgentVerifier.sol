// SPDX-License-Identifier: MIT
pragma solidity 0.8.25;

import {IAgentVerifier} from "@agent-ercs/verify/ERC8274/IAgentVerifier.sol";
import {IProofVerifier} from "@agent-ercs/verify/ERC8274/IProofVerifier.sol";

/// @title MockAgentVerifier
/// @notice Reference implementation of IAgentVerifier for local testing only.
///         Routes every agent to the same single IProofVerifier with empty
///         metadata — the interface doesn't specify how agent-to-verifier
///         bindings or per-agent metadata work, so this is a minimal,
///         honest interpretation, not a general-purpose implementation.
contract MockAgentVerifier is IAgentVerifier {
    IProofVerifier public immutable proofVerifier;

    constructor(IProofVerifier _proofVerifier) {
        proofVerifier = _proofVerifier;
    }

    /// @notice Not part of IAgentVerifier — a mock-only convenience getter.
    ///         The ERC's digest formula references `agentProofProfile` but
    ///         the interface exposes no way to obtain it generically (see
    ///         this ERC's README); this getter lets tests recompute the
    ///         digest anyway, since THIS specific mock happens to expose it.
    function agentProofProfile() public view returns (bytes32) {
        return proofVerifier.proofProfile();
    }

    function verify(bytes32 taskId, bytes32 agentId, bytes32 inputHash, bytes32 outputHash, bytes calldata proof)
        external
        override
        returns (bool valid, bytes32 verificationDigest)
    {
        valid = proofVerifier.verify(inputHash, outputHash, "", proof);
        bytes32 profile = agentProofProfile();
        verificationDigest = keccak256(abi.encode(taskId, agentId, inputHash, outputHash, valid, profile));
        emit VerificationCompleted(taskId, agentId, inputHash, outputHash, valid, verificationDigest);
    }
}
