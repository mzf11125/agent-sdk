// SPDX-License-Identifier: MIT
pragma solidity 0.8.25;

import {Script} from "forge-std/Script.sol";
import {SourceCollectionMock} from "../../../contracts/mocks/identity/ERC8323/SourceCollectionMock.sol";
import {MockSourceBindingRegistryExtended} from "../../../contracts/mocks/identity/ERC8323/MockSourceBindingRegistryExtended.sol";

/// @dev Same broadcast-order contract as DeployERC8323.s.sol (source collection, then
///      registry) but wires the payable-extension mock instead of the base-spec one.
///      Fixed mintPrice of 0.001 ether so tests have a real, checkable, non-zero value.
contract DeployERC8323Extended is Script {
    uint256 constant MINT_PRICE = 0.001 ether;

    function run() external returns (address, address) {
        vm.startBroadcast();
        SourceCollectionMock source = new SourceCollectionMock();
        MockSourceBindingRegistryExtended registry = new MockSourceBindingRegistryExtended(address(source), MINT_PRICE);
        vm.stopBroadcast();
        return (address(source), address(registry));
    }
}
