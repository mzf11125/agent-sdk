// SPDX-License-Identifier: MIT
pragma solidity 0.8.25;

import {Script} from "forge-std/Script.sol";
import {MockAgentWorkflow} from "../../../contracts/mocks/execution/ERC8301/MockAgentWorkflow.sol";

contract DeployERC8301 is Script {
    function run() external returns (address) {
        vm.startBroadcast();
        MockAgentWorkflow workflow = new MockAgentWorkflow();
        vm.stopBroadcast();
        return address(workflow);
    }
}
