// SPDX-License-Identifier: MIT
pragma solidity 0.8.25;

import {ERC721} from "@openzeppelin/contracts/token/ERC721/ERC721.sol";
import {IERC721} from "@openzeppelin/contracts/token/ERC721/IERC721.sol";
import {IERC165} from "@openzeppelin/contracts/utils/introspection/IERC165.sol";
import {IAgentSourceBinding, IAgentSourceBindingView} from "@agent-ercs/identity/ERC8323/IAgentSourceBinding.sol";

/// @title MockSourceBindingRegistryExtended
/// @notice Testkit fixture for the two registerWithSource overloads confirmed live on
///         Merlini's real AgentIdentityRegistry (mainnet 0xe0454dfa17a57a84c3e0e2dbfda5318cbbe91e2c,
///         2026-07-16 Telegram) that are NOT part of the ERC-8323 base interface -- kept in a
///         SEPARATE mock from MockSourceBindingRegistry so the base-spec fixture stays minimal
///         and doesn't accrue one deployment's vendor-specific extras. Not audited, testkit only.
contract MockSourceBindingRegistryExtended is ERC721, IAgentSourceBinding {
    address public immutable boundCollectionAddr;
    uint256 public immutable mintPrice;

    mapping(uint256 => address) private _sourceContract;
    mapping(uint256 => uint256) private _sourceTokenId;
    mapping(uint256 => bool) private _hasSource;
    mapping(uint256 => string) private _agentURI;
    mapping(uint256 => mapping(string => bytes)) private _metadata;
    uint256 private _nextAgentId = 1;

    struct MetadataEntry {
        string metadataKey;
        bytes metadataValue;
    }

    constructor(address boundCollection_, uint256 mintPrice_) ERC721("Mock Extended Source-Bound Registry", "AGENT8323X") {
        boundCollectionAddr = boundCollection_;
        mintPrice = mintPrice_;
    }

    function boundCollection() external view override returns (address) {
        return boundCollectionAddr;
    }

    // Base ERC-8323 overload -- still required for supportsInterface conformance.
    function registerWithSource(uint256 sourceTokenId) external payable override returns (uint256 agentId) {
        agentId = _register(sourceTokenId, "");
    }

    // Extension #1: custom agentURI, confirmed live on Merlini's real registry.
    function registerWithSource(string calldata agentURI, uint256 sourceTokenId) external payable returns (uint256 agentId) {
        agentId = _register(sourceTokenId, agentURI);
    }

    // Extension #2: custom agentURI + seeded metadata, confirmed live on Merlini's real registry.
    function registerWithSource(string calldata agentURI, uint256 sourceTokenId, MetadataEntry[] calldata metadata)
        external
        payable
        returns (uint256 agentId)
    {
        agentId = _register(sourceTokenId, agentURI);
        for (uint256 i = 0; i < metadata.length; i++) {
            _metadata[agentId][metadata[i].metadataKey] = metadata[i].metadataValue;
        }
    }

    function _register(uint256 sourceTokenId, string memory agentURI) private returns (uint256 agentId) {
        require(msg.value == mintPrice, "MockSourceBindingRegistryExtended: wrong mint price");
        require(
            IERC721(boundCollectionAddr).ownerOf(sourceTokenId) == msg.sender,
            "MockSourceBindingRegistryExtended: caller does not own source token"
        );
        agentId = _nextAgentId++;
        _safeMint(msg.sender, agentId);
        _sourceContract[agentId] = boundCollectionAddr;
        _sourceTokenId[agentId] = sourceTokenId;
        _hasSource[agentId] = true;
        _agentURI[agentId] = agentURI;
        emit SourceNFTLinked(agentId, boundCollectionAddr, sourceTokenId);
    }

    function agentURI(uint256 agentId) external view returns (string memory) {
        return _agentURI[agentId];
    }

    function getMetadata(uint256 agentId, string calldata metadataKey) external view returns (bytes memory) {
        return _metadata[agentId][metadataKey];
    }

    function getSourceNFT(uint256 agentId) external view override returns (address sourceContract, uint256 sourceTokenId) {
        require(_hasSource[agentId], "MockSourceBindingRegistryExtended: no source binding");
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
        return sourceOwner == ownerOf(agentId);
    }

    function supportsInterface(bytes4 interfaceId) public view override(ERC721, IERC165) returns (bool) {
        return interfaceId == type(IAgentSourceBinding).interfaceId
            || interfaceId == type(IAgentSourceBindingView).interfaceId
            || super.supportsInterface(interfaceId);
    }
}
