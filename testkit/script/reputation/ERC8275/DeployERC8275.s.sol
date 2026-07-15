// SPDX-License-Identifier: MIT
pragma solidity 0.8.25;

import {Script} from "forge-std/Script.sol";
import {MockAgentReputation} from "../../../contracts/mocks/reputation/ERC8275/MockAgentReputation.sol";

contract DeployERC8275 is Script {
    function run() external returns (address) {
        vm.startBroadcast();
        MockAgentReputation rep = new MockAgentReputation();
        vm.stopBroadcast();
        return address(rep);
    }
}
