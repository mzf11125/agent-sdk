// SPDX-License-Identifier: MIT
pragma solidity 0.8.25;

import {Script} from "forge-std/Script.sol";
import {MockObservationCommitment} from "../../../contracts/mocks/verify/ERC8281/MockObservationCommitment.sol";

/// @notice Deploys MockObservationCommitment for local testing.
contract DeployERC8281 is Script {
    function run() external returns (address) {
        vm.startBroadcast();
        MockObservationCommitment commitment = new MockObservationCommitment();
        vm.stopBroadcast();
        return address(commitment);
    }
}
