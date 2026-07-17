// SPDX-License-Identifier: MIT
pragma solidity 0.8.25;

import {Script} from "forge-std/Script.sol";
import {MockIdentityRegistry} from "../../../contracts/mocks/identity/ERC8004/MockIdentityRegistry.sol";

contract DeployERC8004 is Script {
    function run() external returns (address) {
        vm.startBroadcast();
        MockIdentityRegistry registry = new MockIdentityRegistry();
        vm.stopBroadcast();
        return address(registry);
    }
}
