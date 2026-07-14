// SPDX-License-Identifier: MIT
pragma solidity 0.8.25;

import {Script} from "forge-std/Script.sol";
import {ConsultEscrow} from "@agent-ercs/settlement/ConsultEscrow/ConsultEscrow.sol";

/// @notice Deploys the real ConsultEscrow from agent-ercs. The contract is
///         already Etherscan-verified on Ethereum mainnet at
///         0x7057fbA75Ca88B8eF43564be3244bdd7163De04D. This script deploys
///         a fresh instance for local anvil-based testing.
contract DeployERC8203 is Script {
    function run() external returns (address) {
        vm.startBroadcast();
        ConsultEscrow escrow = new ConsultEscrow();
        vm.stopBroadcast();
        return address(escrow);
    }
}
