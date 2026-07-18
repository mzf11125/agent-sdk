// SPDX-License-Identifier: MIT
pragma solidity 0.8.25;

import {Script} from "forge-std/Script.sol";
import {MockBoundedAgentAction} from "../../../contracts/mocks/metering/ERC8312/MockBoundedAgentAction.sol";
import {MockBudgetSubstrate} from "../../../contracts/mocks/metering/ERC8312/MockBudgetSubstrate.sol";
import {MockContestableEnvelope} from "../../../contracts/mocks/metering/ERC8312/MockContestableEnvelope.sol";

/// @notice Deploys all three ERC-8312 mock contracts, each standalone.
///         testkit/scripts/deploy.sh prints one address per line in broadcast
///         order: boundedAgentAction, then budgetSubstrate, then
///         contestableEnvelope.
contract DeployERC8312 is Script {
    function run()
        external
        returns (
            address boundedAgentAction,
            address budgetSubstrate,
            address contestableEnvelope
        )
    {
        vm.startBroadcast();
        MockBoundedAgentAction base = new MockBoundedAgentAction();
        MockBudgetSubstrate budget = new MockBudgetSubstrate();
        MockContestableEnvelope contestable = new MockContestableEnvelope();
        vm.stopBroadcast();

        boundedAgentAction = address(base);
        budgetSubstrate = address(budget);
        contestableEnvelope = address(contestable);
    }
}
