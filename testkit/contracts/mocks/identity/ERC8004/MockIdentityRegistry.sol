// SPDX-License-Identifier: MIT
pragma solidity 0.8.25;

import {ERC721} from "@openzeppelin/contracts/token/ERC721/ERC721.sol";
import {EIP712} from "@openzeppelin/contracts/utils/cryptography/EIP712.sol";
import {ECDSA} from "@openzeppelin/contracts/utils/cryptography/ECDSA.sol";
import {IIdentityRegistry} from "@agent-ercs/identity/ERC8004/IIdentityRegistry.sol";

/// @title MockIdentityRegistry
/// @notice Reference implementation of IIdentityRegistry for local testing only.
///         Not audited, not for production use — see agent-ercs's README on
///         interface vs. base-implementation vs. example/reference contracts.
contract MockIdentityRegistry is ERC721, EIP712, IIdentityRegistry {
    bytes32 private constant SET_AGENT_WALLET_TYPEHASH =
        keccak256("SetAgentWallet(uint256 agentId,address newWallet,uint256 deadline)");

    mapping(uint256 => string) private _agentURIs;
    mapping(uint256 => mapping(string => bytes)) private _metadata;
    mapping(uint256 => address) private _agentWallets;
    uint256 private _nextAgentId = 1;

    constructor() ERC721("Agent Identity Registry", "AGENT8004") EIP712("MockIdentityRegistry", "1") {}

    function register(string memory agentURI, IIdentityRegistry.MetadataEntry[] memory metadata)
        public
        override
        returns (uint256 agentId)
    {
        agentId = _nextAgentId++;
        _safeMint(msg.sender, agentId);
        if (bytes(agentURI).length > 0) {
            _agentURIs[agentId] = agentURI;
        }
        for (uint256 i = 0; i < metadata.length; i++) {
            _metadata[agentId][metadata[i].metadataKey] = metadata[i].metadataValue;
            emit MetadataSet(agentId, metadata[i].metadataKey, metadata[i].metadataKey, metadata[i].metadataValue);
        }
        emit Registered(agentId, agentURI, msg.sender);
    }

    function register(string memory agentURI) public override returns (uint256 agentId) {
        IIdentityRegistry.MetadataEntry[] memory meta = new IIdentityRegistry.MetadataEntry[](0);
        return register(agentURI, meta);
    }

    function register() public override returns (uint256 agentId) {
        IIdentityRegistry.MetadataEntry[] memory meta = new IIdentityRegistry.MetadataEntry[](0);
        return register("", meta);
    }

    function setAgentURI(uint256 agentId, string calldata agentURI) external override {
        require(ownerOf(agentId) == msg.sender, "MockIdentityRegistry: not agent owner");
        _agentURIs[agentId] = agentURI;
    }

    function tokenURI(uint256 agentId) public view override returns (string memory) {
        return _agentURIs[agentId];
    }

    function getMetadata(uint256 agentId, string calldata metadataKey)
        external
        view
        override
        returns (bytes memory)
    {
        return _metadata[agentId][metadataKey];
    }

    function setMetadata(uint256 agentId, string calldata metadataKey, bytes calldata metadataValue)
        external
        override
    {
        require(ownerOf(agentId) == msg.sender, "MockIdentityRegistry: not agent owner");
        _metadata[agentId][metadataKey] = metadataValue;
        emit MetadataSet(agentId, metadataKey, metadataKey, metadataValue);
    }

    function setAgentWallet(uint256 agentId, address newWallet, uint256 deadline, bytes calldata signature)
        external
        override
    {
        require(block.timestamp <= deadline, "MockIdentityRegistry: signature expired");
        bytes32 structHash = keccak256(abi.encode(SET_AGENT_WALLET_TYPEHASH, agentId, newWallet, deadline));
        bytes32 digest = _hashTypedDataV4(structHash);
        address signer = ECDSA.recover(digest, signature);
        require(signer == ownerOf(agentId), "MockIdentityRegistry: invalid wallet signature");
        _agentWallets[agentId] = newWallet;
    }

    function getAgentWallet(uint256 agentId) external view override returns (address) {
        return _agentWallets[agentId];
    }

    function unsetAgentWallet(uint256 agentId) external override {
        require(ownerOf(agentId) == msg.sender, "MockIdentityRegistry: not agent owner");
        delete _agentWallets[agentId];
    }
}
