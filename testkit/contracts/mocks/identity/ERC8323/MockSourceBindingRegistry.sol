// SPDX-License-Identifier: MIT
pragma solidity 0.8.25;

import {ERC721} from "@openzeppelin/contracts/token/ERC721/ERC721.sol";
import {IERC721} from "@openzeppelin/contracts/token/ERC721/IERC721.sol";
import {IERC165} from "@openzeppelin/contracts/utils/introspection/IERC165.sol";
import {IAgentSourceBinding, IAgentSourceBindingView} from "@agent-ercs/identity/ERC8323/IAgentSourceBinding.sol";

/// @title MockSourceBindingRegistry
/// @notice Reference implementation of IAgentSourceBinding for local testing only.
///         Not audited, not for production use — see agent-ercs's README on
///         interface vs. base-implementation vs. example/reference contracts.
/// @dev isSourceNFTOwnershipValid here only implements case (a) — direct holder —
///      of the spec's 3-case rule (direct / canonical ERC-6551 TBA / binding
///      contract). The TBA and binding-contract cases require pinning a specific
///      ERC-6551 registry+implementation+salt, which is out of scope for a
///      minimal testkit fixture; a real conformant registry MUST implement all
///      three cases.
contract MockSourceBindingRegistry is ERC721, IAgentSourceBinding {
    address public immutable boundCollectionAddr;

    mapping(uint256 => address) private _sourceContract;
    mapping(uint256 => uint256) private _sourceTokenId;
    mapping(uint256 => bool) private _hasSource;
    uint256 private _nextAgentId = 1;

    constructor(address boundCollection_) ERC721("Mock Source-Bound Agent Registry", "AGENT8323") {
        boundCollectionAddr = boundCollection_;
    }

    function boundCollection() external view override returns (address) {
        return boundCollectionAddr;
    }

    function registerWithSource(uint256 sourceTokenId) external payable override returns (uint256 agentId) {
        require(
            IERC721(boundCollectionAddr).ownerOf(sourceTokenId) == msg.sender,
            "MockSourceBindingRegistry: caller does not own source token"
        );
        agentId = _nextAgentId++;
        _safeMint(msg.sender, agentId);
        _sourceContract[agentId] = boundCollectionAddr;
        _sourceTokenId[agentId] = sourceTokenId;
        _hasSource[agentId] = true;
        emit SourceNFTLinked(agentId, boundCollectionAddr, sourceTokenId);
    }

    function getSourceNFT(uint256 agentId) external view override returns (address sourceContract, uint256 sourceTokenId) {
        require(_hasSource[agentId], "MockSourceBindingRegistry: no source binding");
        return (_sourceContract[agentId], _sourceTokenId[agentId]);
    }

    function hasSourceNFT(uint256 agentId) external view override returns (bool) {
        return _hasSource[agentId];
    }

    function isSourceNFTOwnershipValid(uint256 agentId) external view override returns (bool) {
        if (!_hasSource[agentId]) return false;
        address sourceOwner;
        try IERC721(_sourceContract[agentId]).ownerOf(_sourceTokenId[agentId]) returns (address o) {
            sourceOwner = o;
        } catch {
            return false;
        }
        // Case (a) only — see contract-level @dev note.
        return sourceOwner == ownerOf(agentId);
    }

    function supportsInterface(bytes4 interfaceId) public view override(ERC721, IERC165) returns (bool) {
        return interfaceId == type(IAgentSourceBinding).interfaceId
            || interfaceId == type(IAgentSourceBindingView).interfaceId
            || super.supportsInterface(interfaceId);
    }
}
