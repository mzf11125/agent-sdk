// SPDX-License-Identifier: MIT
pragma solidity 0.8.25;

import {IProofVerifier} from "@agent-ercs/verify/ERC8274/IProofVerifier.sol";

/// @title MockProofVerifier
/// @notice Reference implementation of IProofVerifier for local testing only.
///         "Verification" here is a simple hash check, not real zkML/opML/TEE
///         cryptography — see agent-ercs's README on interface vs.
///         base-implementation vs. example/reference contracts.
contract MockProofVerifier is IProofVerifier {
    function verify(bytes32 inputHash, bytes32 outputHash, bytes calldata metadata, bytes calldata proof)
        external
        pure
        override
        returns (bool valid)
    {
        bytes32 expectedDigest = keccak256(abi.encodePacked(inputHash, outputHash, metadata));
        valid = keccak256(proof) == keccak256(abi.encodePacked(expectedDigest));
    }

    function proofSystem() external pure override returns (string memory) {
        return "mock-test-only";
    }

    function proofProfile() external pure override returns (bytes32) {
        return keccak256("mock-test-only-v1");
    }
}
