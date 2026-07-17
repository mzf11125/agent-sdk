// SPDX-License-Identifier: MIT
pragma solidity 0.8.25;

import {Script} from "forge-std/Script.sol";
import {MockOnChainProof} from "../../../contracts/mocks/anchor/ERC8263/MockOnChainProof.sol";

contract DeployERC8263 is Script {
    function run() external returns (address) {
        vm.startBroadcast();
        MockOnChainProof proof = new MockOnChainProof();
        vm.stopBroadcast();
        return address(proof);
    }
}
