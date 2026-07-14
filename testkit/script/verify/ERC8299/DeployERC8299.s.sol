// SPDX-License-Identifier: MIT
pragma solidity 0.8.25;

import {Script} from "forge-std/Script.sol";
import {MockWyriweAttestation} from "../../../contracts/mocks/verify/ERC8299/MockWyriweAttestation.sol";
import {MockJudgmentExecutionAttestation} from "../../../contracts/mocks/verify/ERC8299/MockJudgmentExecutionAttestation.sol";

/// @notice Deploys both ERC-8299 mock contracts, wired in order. The deployer
///         account (msg.sender within the broadcast) is used as the attestor
///         for both mocks. testkit/scripts/deploy.sh prints one address per
///         line in broadcast order: wyriweAttestation, then
///         judgmentExecutionAttestation.
contract DeployERC8299 is Script {
    function run() external returns (address wyriweAttestation, address judgmentExecutionAttestation) {
        vm.startBroadcast();
        MockWyriweAttestation wyriweAttestationContract = new MockWyriweAttestation();
        MockJudgmentExecutionAttestation judgmentExecutionAttestationContract =
            new MockJudgmentExecutionAttestation();
        vm.stopBroadcast();

        wyriweAttestation = address(wyriweAttestationContract);
        judgmentExecutionAttestation = address(judgmentExecutionAttestationContract);
    }
}
