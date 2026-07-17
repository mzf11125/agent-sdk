// SPDX-License-Identifier: MIT
pragma solidity 0.8.25;

import {Script} from "forge-std/Script.sol";
import {MockProofVerifier} from "../../../contracts/mocks/verify/ERC8274/MockProofVerifier.sol";
import {MockAgentVerifier} from "../../../contracts/mocks/verify/ERC8274/MockAgentVerifier.sol";
import {MockAgentVerifiable} from "../../../contracts/mocks/verify/ERC8274/MockAgentVerifiable.sol";

/// @notice Deploys all three ERC-8274 mock contracts, wired together, in
///         order. testkit/scripts/deploy.sh prints one address per line in
///         broadcast order: proofVerifier, then agentVerifier, then
///         agentVerifiable.
contract DeployERC8274 is Script {
    function run() external returns (address proofVerifier, address agentVerifier, address agentVerifiable) {
        vm.startBroadcast();
        MockProofVerifier proofVerifierContract = new MockProofVerifier();
        MockAgentVerifier agentVerifierContract = new MockAgentVerifier(proofVerifierContract);
        MockAgentVerifiable agentVerifiableContract = new MockAgentVerifiable(address(agentVerifierContract));
        vm.stopBroadcast();

        proofVerifier = address(proofVerifierContract);
        agentVerifier = address(agentVerifierContract);
        agentVerifiable = address(agentVerifiableContract);
    }
}
