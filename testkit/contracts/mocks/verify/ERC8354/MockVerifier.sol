// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import {IVerifier} from "./IVerifier.sol";

/// @title MockVerifier
/// @notice Test double for the domain's proof verifier. Returns `result`;
///         `shouldRevert` models a malformed proof that must return false
///         rather than reverting.
contract MockVerifier is IVerifier {
    bool public result = true;
    bool public shouldRevert;

    function setResult(bool r) external {
        result = r;
    }

    function setRevert(bool r) external {
        shouldRevert = r;
    }

    function verifyProof(bytes32, bytes calldata, bytes calldata) external view returns (bool) {
        require(!shouldRevert, "malformed proof");
        return result;
    }
}
