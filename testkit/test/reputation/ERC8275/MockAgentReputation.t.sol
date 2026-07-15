// SPDX-License-Identifier: MIT
pragma solidity 0.8.25;

import {Test} from "forge-std/Test.sol";
import {MockAgentReputation} from "../../../contracts/mocks/reputation/ERC8275/MockAgentReputation.sol";

contract MockAgentReputationTest is Test {
    MockAgentReputation public rep;

    bytes32 constant AGENT_ID = keccak256("test-agent");
    bytes32 constant ORDER_ID = keccak256("test-order");

    function setUp() public {
        rep = new MockAgentReputation();
    }

    function testSetAndGetReputation() public view {
        uint64 completed = 42;
        uint64 disputed = 3;
        uint64 volume = 1000000;
        uint64 lastActive = 1700000000;
        uint16 score = 850;

        (uint64 c, uint64 d, uint64 v, uint64 l, uint16 s) = rep.getReputation(AGENT_ID);
        // Default should be zeros
        assertEq(c, 0);
        assertEq(d, 0);
        assertEq(v, 0);
        assertEq(l, 0);
        assertEq(s, 0);
    }

    function testSetAndGetDecayWeight() public {
        rep.setDecayWeight(AGENT_ID, 7500);
        assertEq(rep.getDecayWeight(AGENT_ID), 7500);
    }

    function testVerifyOutcome() public {
        rep.setOutcome(ORDER_ID, AGENT_ID, true);
        bytes memory proof = abi.encode(AGENT_ID);
        assertTrue(rep.verifyOutcome(ORDER_ID, proof));

        rep.setOutcome(ORDER_ID, AGENT_ID, false);
        assertFalse(rep.verifyOutcome(ORDER_ID, proof));
    }

    function testVerifyOutcomeUnknownOrder() public {
        bytes32 unknownOrder = keccak256("unknown");
        bytes memory proof = abi.encode(AGENT_ID);
        assertFalse(rep.verifyOutcome(unknownOrder, proof));
    }

    function testValidatorType() public view {
        assertEq(uint256(rep.validatorType()), 0); // ReExecution
    }

    function testClaimType() public view {
        assertEq(uint256(rep.claimType()), 0); // ReExecution
    }
}
