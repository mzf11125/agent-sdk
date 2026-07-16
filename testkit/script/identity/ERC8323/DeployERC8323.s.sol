// SPDX-License-Identifier: MIT
pragma solidity 0.8.25;

import {Script} from "forge-std/Script.sol";
import {ERC721} from "@openzeppelin/contracts/token/ERC721/ERC721.sol";
import {MockAgentSourceBinding} from "../../../contracts/mocks/identity/ERC8323/MockAgentSourceBinding.sol";

/// @notice Deploys a dummy ERC-721 (to serve as the source collection) and
///         then the MockAgentSourceBinding pointing to it. deploy.sh prints
///         one address per line in broadcast order: dummy collection, then
///         source binding.
contract DeployERC8323 is Script {
    function run() external returns (address dummyCollection, address binding) {
        vm.startBroadcast();

        DummyERC721 dummy = new DummyERC721();
        MockAgentSourceBinding sourceBinding = new MockAgentSourceBinding(address(dummy));

        vm.stopBroadcast();

        dummyCollection = address(dummy);
        binding = address(sourceBinding);
    }
}

/// @notice Minimal ERC-721 to serve as the bound source collection for
///         MockAgentSourceBinding testing.
contract DummyERC721 is ERC721 {
    uint256 private _nextId = 1;

    constructor() ERC721("Dummy Source", "DSRC") {}

    function mint(address to) external returns (uint256) {
        uint256 id = _nextId++;
        _safeMint(to, id);
        return id;
    }
}
