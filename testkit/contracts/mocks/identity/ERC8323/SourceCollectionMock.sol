// SPDX-License-Identifier: MIT
pragma solidity 0.8.25;

import {ERC721} from "@openzeppelin/contracts/token/ERC721/ERC721.sol";

/// @title SourceCollectionMock
/// @notice Minimal ERC-721 "source" collection an agent identity can be derived
///         from (a PFP/membership-token stand-in for local testing only).
///         Mint-only, no other logic — not audited, not for production use.
contract SourceCollectionMock is ERC721 {
    uint256 private _nextId = 1;

    constructor() ERC721("Mock Source Collection", "SRC") {}

    function mint(address to) external returns (uint256 tokenId) {
        tokenId = _nextId++;
        _safeMint(to, tokenId);
    }
}
