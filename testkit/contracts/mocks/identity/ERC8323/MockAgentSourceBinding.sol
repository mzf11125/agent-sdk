// SPDX-License-Identifier: MIT
pragma solidity 0.8.25;

import {ERC721} from "@openzeppelin/contracts/token/ERC721/ERC721.sol";
import {IERC165} from "@openzeppelin/contracts/utils/introspection/IERC165.sol";
import {IAgentSourceBinding} from "@agent-ercs/identity/ERC8323/IAgentSourceBinding.sol";

/// @title MockAgentSourceBinding
/// @notice Minimal reference implementation of IAgentSourceBinding for local
///         testing only. Not audited, not for production use — see agent-ercs's
///         README on interface vs. base-implementation vs. example/reference
///         contracts.
/// @dev This mock skips the ERC-721 source-ownership check (it mints an agent
///      to anyone who calls registerWithSource) so that integration tests can
///      exercise the full SDK client without deploying a real source collection.
contract MockAgentSourceBinding is ERC721, IAgentSourceBinding {
    address private _boundCollection;
    mapping(uint256 => address) private _sourceContracts;
    mapping(uint256 => uint256) private _sourceTokenIds;
    uint256 private _nextAgentId = 1;

    constructor(address boundCollection_) ERC721("Agent Source Binding", "SRC8323") {
        _boundCollection = boundCollection_;
    }

    function boundCollection() external view override returns (address) {
        return _boundCollection;
    }

    function registerWithSource(uint256 sourceTokenId)
        external
        payable
        override
        returns (uint256 agentId)
    {
        agentId = _nextAgentId++;
        _safeMint(msg.sender, agentId);
        _sourceContracts[agentId] = _boundCollection;
        _sourceTokenIds[agentId] = sourceTokenId;
        emit SourceNFTLinked(agentId, _boundCollection, sourceTokenId);
    }

    function getSourceNFT(uint256 agentId)
        external
        view
        override
        returns (address sourceContract, uint256 sourceTokenId)
    {
        require(_ownerOf(agentId) != address(0), "agent does not exist");
        return (_sourceContracts[agentId], _sourceTokenIds[agentId]);
    }

    function hasSourceNFT(uint256 agentId) external view override returns (bool) {
        return _sourceContracts[agentId] != address(0);
    }

    function isSourceNFTOwnershipValid(uint256 agentId) external view override returns (bool) {
        // Mock: always returns true for simplicity. A production contract
        // would check ownerOf(sourceToken) against the agent owner or its
        // canonical ERC-6551 TBA.
        return true;
    }

    function supportsInterface(bytes4 interfaceId)
        public
        view
        override(ERC721, IERC165)
        returns (bool)
    {
        return interfaceId == type(IAgentSourceBinding).interfaceId
            || super.supportsInterface(interfaceId);
    }
}
