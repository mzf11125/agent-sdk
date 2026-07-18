// SPDX-License-Identifier: MIT
pragma solidity 0.8.25;

import {Test} from "forge-std/Test.sol";
import {IBoundedAgentAction} from "@agent-ercs/metering/ERC8312/IBoundedAgentAction.sol";
import {IBudgetSubstrate} from "@agent-ercs/metering/ERC8312/IBudgetSubstrate.sol";
import {IContestableEnvelope} from "@agent-ercs/metering/ERC8312/IContestableEnvelope.sol";
import {MockBoundedAgentAction} from "../../../contracts/mocks/metering/ERC8312/MockBoundedAgentAction.sol";
import {MockBudgetSubstrate} from "../../../contracts/mocks/metering/ERC8312/MockBudgetSubstrate.sol";
import {MockContestableEnvelope} from "../../../contracts/mocks/metering/ERC8312/MockContestableEnvelope.sol";

contract MockBoundedAgentActionTest is Test {
    MockBoundedAgentAction internal action;
    address internal constant PRINCIPAL = address(0xCAFE);

    function setUp() public {
        action = new MockBoundedAgentAction();
    }

    function test_registerEnvelope() public {
        bytes32 capRoot = keccak256("cap-root");
        uint64 expiresAt = uint64(block.timestamp + 1 days);

        bytes32 id = action.registerEnvelope(PRINCIPAL, capRoot, expiresAt, "");

        IBoundedAgentAction.Envelope memory env = action.getEnvelope(id);
        assertEq(env.principal, PRINCIPAL);
        assertEq(env.capabilityRoot, capRoot);
        assertEq(uint256(env.status), uint256(IBoundedAgentAction.Status.Active));
        assertEq(env.cursorRoot, bytes32(0));
    }

    function test_advanceCursor() public {
        bytes32 capRoot = keccak256("cap-root");
        uint64 expiresAt = uint64(block.timestamp + 1 days);
        bytes32 id = action.registerEnvelope(PRINCIPAL, capRoot, expiresAt, "");
        bytes memory witness = hex"beef";

        bytes32 newCursor = action.advanceCursor(id, witness);

        assertTrue(newCursor != bytes32(0));
        IBoundedAgentAction.Envelope memory env = action.getEnvelope(id);
        assertEq(env.cursorRoot, newCursor);
    }

    function test_setStatus() public {
        bytes32 capRoot = keccak256("cap-root");
        uint64 expiresAt = uint64(block.timestamp + 1 days);
        bytes32 id = action.registerEnvelope(PRINCIPAL, capRoot, expiresAt, "");

        action.setStatus(id, IBoundedAgentAction.Status.Completed);

        assertEq(uint256(action.getStatus(id)), uint256(IBoundedAgentAction.Status.Completed));
    }

    function test_isActiveReturnsTrueAfterRegistration() public {
        bytes32 capRoot = keccak256("cap-root");
        uint64 expiresAt = uint64(block.timestamp + 1 days);
        bytes32 id = action.registerEnvelope(PRINCIPAL, capRoot, expiresAt, "");

        assertTrue(action.isActive(id));
    }

    function test_isActiveReturnsFalseAfterRevoke() public {
        bytes32 capRoot = keccak256("cap-root");
        uint64 expiresAt = uint64(block.timestamp + 1 days);
        bytes32 id = action.registerEnvelope(PRINCIPAL, capRoot, expiresAt, "");

        action.setStatus(id, IBoundedAgentAction.Status.Revoked);

        assertFalse(action.isActive(id));
    }

    function test_getStatusRevertsOnUnknownId() public {
        vm.expectRevert("unknown envelope");
        action.getStatus(keccak256("nonexistent"));
    }

    function test_registerEnvelopeEmitsEvent() public {
        bytes32 capRoot = keccak256("cap-root");
        uint64 expiresAt = uint64(block.timestamp + 1 days);

        // Compute the expected id (matches MockBoundedAgentAction's logic)
        bytes32 expectedId = keccak256(abi.encodePacked(address(action), uint64(1)));

        vm.expectEmit(true, true, true, true);
        emit IBoundedAgentAction.EnvelopeRegistered({id: expectedId, principal: PRINCIPAL, capabilityRoot: capRoot});
        action.registerEnvelope(PRINCIPAL, capRoot, expiresAt, "");
    }
}

contract MockBudgetSubstrateTest is Test {
    MockBudgetSubstrate internal budget;
    address internal constant PRINCIPAL = address(0xBABE);

    function setUp() public {
        budget = new MockBudgetSubstrate();
    }

    function test_boundReturnsDefaults() public {
        uint64 expiresAt = uint64(block.timestamp + 1 days);
        bytes32 id = budget.registerEnvelope(PRINCIPAL, keccak256("cap"), expiresAt, "");

        (uint256 cap, address asset) = budget.bound(id);
        assertEq(cap, 10_000);
        assertEq(asset, address(0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48));
    }

    function test_spentIncreasesOnAdvance() public {
        uint64 expiresAt = uint64(block.timestamp + 1 days);
        bytes32 id = budget.registerEnvelope(PRINCIPAL, keccak256("cap"), expiresAt, "");

        assertEq(budget.spent(id), 0);

        budget.advanceCursor(id, abi.encode(100));
        assertEq(budget.spent(id), 100);

        budget.advanceCursor(id, abi.encode(50));
        assertEq(budget.spent(id), 150);
    }

    function test_spentCannotExceedCap() public {
        uint64 expiresAt = uint64(block.timestamp + 1 days);
        bytes32 id = budget.registerEnvelope(PRINCIPAL, keccak256("cap"), expiresAt, "");

        vm.expectRevert("exceeds cap");
        budget.advanceCursor(id, abi.encode(10_001));
    }

    function test_remainingReturnsHeadroom() public {
        uint64 expiresAt = uint64(block.timestamp + 1 days);
        bytes32 id = budget.registerEnvelope(PRINCIPAL, keccak256("cap"), expiresAt, "");

        assertEq(budget.remaining(id), 10_000);
        budget.advanceCursor(id, abi.encode(3000));
        assertEq(budget.remaining(id), 7000);
    }

    function test_remainingReturnsZeroWhenNotActive() public {
        uint64 expiresAt = uint64(block.timestamp + 1 days);
        bytes32 id = budget.registerEnvelope(PRINCIPAL, keccak256("cap"), expiresAt, "");

        // After exhausting the cap, the envelope is no longer active conceptually
        // (remaining returns 0 for non-active).
        budget.advanceCursor(id, abi.encode(10_000));
        assertEq(budget.remaining(id), 0);
    }

    function test_emitsEnvelopeAdvancedEvent() public {
        uint64 expiresAt = uint64(block.timestamp + 1 days);
        bytes32 id = budget.registerEnvelope(PRINCIPAL, keccak256("cap"), expiresAt, "");

        vm.expectEmit(true, true, true, true);
        emit IBoundedAgentAction.EnvelopeAdvanced(id, bytes32(0), keccak256(abi.encode(500)));
        budget.advanceCursor(id, abi.encode(500));
    }
}

contract MockContestableEnvelopeTest is Test {
    MockContestableEnvelope internal contestable;
    address internal constant PRINCIPAL = address(0xDEAD);

    function setUp() public {
        contestable = new MockContestableEnvelope();
    }

    function test_contestAndResolveToActive() public {
        uint64 expiresAt = uint64(block.timestamp + 1 days);
        bytes32 id = contestable.registerEnvelope(PRINCIPAL, keccak256("cap"), expiresAt, "");

        vm.expectEmit(true, true, true, true);
        emit IContestableEnvelope.EnvelopeContested(id, address(this));
        contestable.contest(id, "evidence");

        assertEq(uint256(contestable.getStatus(id)), uint256(IBoundedAgentAction.Status.Contested));

        vm.expectEmit(true, true, true, true);
        emit IContestableEnvelope.EnvelopeResolved(id, IBoundedAgentAction.Status.Active);
        contestable.resolve(id, IBoundedAgentAction.Status.Active, "");

        assertTrue(contestable.isActive(id));
    }

    function test_resolveToRevoked() public {
        uint64 expiresAt = uint64(block.timestamp + 1 days);
        bytes32 id = contestable.registerEnvelope(PRINCIPAL, keccak256("cap"), expiresAt, "");

        contestable.contest(id, "");
        contestable.resolve(id, IBoundedAgentAction.Status.Revoked, "");

        assertEq(uint256(contestable.getStatus(id)), uint256(IBoundedAgentAction.Status.Revoked));
        assertFalse(contestable.isActive(id));
    }

    function test_cannotContestInactiveEnvelope() public {
        uint64 expiresAt = uint64(block.timestamp + 1 days);
        bytes32 id = contestable.registerEnvelope(PRINCIPAL, keccak256("cap"), expiresAt, "");

        contestable.setStatus(id, IBoundedAgentAction.Status.Completed);

        vm.expectRevert("not contestable");
        contestable.contest(id, "");
    }
}
