// SPDX-License-Identifier: MIT
pragma solidity 0.8.25;

import {Script} from "forge-std/Script.sol";
import {SourceCollectionMock} from "../../../contracts/mocks/identity/ERC8323/SourceCollectionMock.sol";
import {MockSourceBindingRegistry} from "../../../contracts/mocks/identity/ERC8323/MockSourceBindingRegistry.sol";

/// @dev Deploys the source collection first, then the registry bound to it —
///      broadcast order matters, deploy.ts's deployContracts() returns
///      addresses in that order (source collection, then registry).
contract DeployERC8323 is Script {
    function run() external returns (address, address) {
        vm.startBroadcast();
        SourceCollectionMock source = new SourceCollectionMock();
        MockSourceBindingRegistry registry = new MockSourceBindingRegistry(address(source));
        vm.stopBroadcast();
        return (address(source), address(registry));
    }
}
