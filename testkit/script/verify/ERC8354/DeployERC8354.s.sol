// SPDX-License-Identifier: MIT
pragma solidity 0.8.25;

import {Script} from "forge-std/Script.sol";
import {MockVerifier} from "../../../contracts/mocks/verify/ERC8354/MockVerifier.sol";
import {PolicyDomainRegistry} from "../../../contracts/mocks/verify/ERC8354/PolicyDomainRegistry.sol";
import {ConfidentialPolicyVerdict} from "../../../contracts/mocks/verify/ERC8354/ConfidentialPolicyVerdict.sol";

/// @notice Deploys the ERC-8354 mock contracts, wired together, in order.
///         testkit/scripts/deploy.sh prints one address per line in broadcast
///         order: verifier, then registry, then guard.
contract DeployERC8354 is Script {
    function run() external returns (address verifierAddr, address registryAddr, address guardAddr) {
        vm.startBroadcast();
        MockVerifier verifier = new MockVerifier();
        PolicyDomainRegistry registry = new PolicyDomainRegistry();
        ConfidentialPolicyVerdict guard = new ConfidentialPolicyVerdict(registry);
        vm.stopBroadcast();

        verifierAddr = address(verifier);
        registryAddr = address(registry);
        guardAddr = address(guard);
    }
}