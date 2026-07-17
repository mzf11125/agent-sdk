// SPDX-License-Identifier: MIT
pragma solidity 0.8.25;

import {IOnChainProof} from "@agent-ercs/anchor/ERC8263/IOnChainProof.sol";

/// @title MockOnChainProof
/// @notice Reference implementation of IOnChainProof for local testing only.
///         Enforces the canonical-form guards specified in ERC-8263.
///         Not audited, not for production use.
contract MockOnChainProof is IOnChainProof {
    /// @inheritdoc IOnChainProof
    function anchor(uint8 agentIdScheme, bytes32 agentId, bytes32 proofHash) external {
        _validate(agentIdScheme, agentId, proofHash);
        emit AnchorProof(agentIdScheme, agentId, proofHash, msg.sender, "");
    }

    /// @inheritdoc IOnChainProof
    function anchorWithAux(
        uint8 agentIdScheme,
        bytes32 agentId,
        bytes32 proofHash,
        bytes calldata aux
    ) external {
        _validate(agentIdScheme, agentId, proofHash);
        emit AnchorProof(agentIdScheme, agentId, proofHash, msg.sender, aux);
    }

    /// @dev Canonical-form guards per ERC-8263 §"Canonical-form guards (write-time invariants)":
    ///      - proofHash != 0
    ///      - scheme 0x00 requires agentId == 0
    ///      - schemes 0x01/0x02 require agentId != 0
    ///      - schemes 0x03+ revert
    function _validate(uint8 agentIdScheme, bytes32 agentId, bytes32 proofHash) internal pure {
        require(proofHash != bytes32(0), "MockOnChainProof: proofHash must be non-zero");

        if (agentIdScheme == 0x00) {
            require(agentId == bytes32(0), "MockOnChainProof: ANONYMOUS scheme requires agentId == 0");
        } else if (agentIdScheme == 0x01 || agentIdScheme == 0x02) {
            require(agentId != bytes32(0), "MockOnChainProof: registered scheme requires non-zero agentId");
        } else {
            revert("MockOnChainProof: reserved agentIdScheme");
        }
    }
}
