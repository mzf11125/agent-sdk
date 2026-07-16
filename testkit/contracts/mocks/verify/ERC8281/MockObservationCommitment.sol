// SPDX-License-Identifier: MIT
pragma solidity 0.8.25;

import {IERC165} from "@openzeppelin/contracts/utils/introspection/IERC165.sol";
import {IObservationCommitment} from "@agent-ercs/verify/ERC8281/IObservationCommitment.sol";

/// @title MockObservationCommitment
/// @notice Minimal reference implementation of IObservationCommitment for
///         local testing only. Not audited, not for production use — see
///         agent-ercs's README on interface vs. base-implementation vs.
///         example/reference contracts.
/// @dev Implements ERC-165 per the canonical reference contract pattern.
contract MockObservationCommitment is IObservationCommitment {
    /// @dev ERC-165 interface id of IObservationCommitment.
    bytes4 private constant OBSERVATION_COMMITMENT_INTERFACE_ID = 0xb5c645bd;

    function record(bytes32 digest) external override {
        emit Recorded(digest, msg.sender);
    }

    function supportsInterface(bytes4 interfaceId) external view returns (bool) {
        return interfaceId == OBSERVATION_COMMITMENT_INTERFACE_ID;
    }
}
