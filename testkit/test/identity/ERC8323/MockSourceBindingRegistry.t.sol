// SPDX-License-Identifier: MIT
pragma solidity 0.8.25;

import {Test} from "forge-std/Test.sol";
import {SourceCollectionMock} from "../../../contracts/mocks/identity/ERC8323/SourceCollectionMock.sol";
import {MockSourceBindingRegistry} from "../../../contracts/mocks/identity/ERC8323/MockSourceBindingRegistry.sol";
import {IAgentSourceBinding, IAgentSourceBindingView} from "@agent-ercs/identity/ERC8323/IAgentSourceBinding.sol";

contract MockSourceBindingRegistryTest is Test {
    SourceCollectionMock internal source;
    MockSourceBindingRegistry internal registry;

    address internal holder = address(0xA11CE);
    address internal stranger = address(0xBEEF);

    function setUp() public {
        source = new SourceCollectionMock();
        registry = new MockSourceBindingRegistry(address(source));
    }

    function test_boundCollectionIsImmutableAndCorrect() public view {
        assertEq(registry.boundCollection(), address(source));
    }

    function test_registerWithSourceBindsProvenanceAndMints() public {
        uint256 srcId = source.mint(holder);

        vm.prank(holder);
        uint256 agentId = registry.registerWithSource(srcId);

        assertEq(registry.ownerOf(agentId), holder);
        (address srcContract, uint256 srcTokenId) = registry.getSourceNFT(agentId);
        assertEq(srcContract, address(source));
        assertEq(srcTokenId, srcId);
        assertTrue(registry.hasSourceNFT(agentId));
    }

    function test_registerWithSourceRevertsIfCallerDoesNotOwnSource() public {
        uint256 srcId = source.mint(holder);

        vm.prank(stranger);
        vm.expectRevert("MockSourceBindingRegistry: caller does not own source token");
        registry.registerWithSource(srcId);
    }

    function test_isSourceNFTOwnershipValidTracksLiveOwnership() public {
        uint256 srcId = source.mint(holder);
        vm.prank(holder);
        uint256 agentId = registry.registerWithSource(srcId);

        assertTrue(registry.isSourceNFTOwnershipValid(agentId));

        // Source token resold — binding is still recorded (provenance, immutable),
        // but live-ownership validity must flip to false.
        vm.prank(holder);
        source.transferFrom(holder, stranger, srcId);

        assertTrue(registry.hasSourceNFT(agentId)); // provenance unchanged
        assertFalse(registry.isSourceNFTOwnershipValid(agentId)); // live check flips
    }

    function test_getSourceNFTRevertsForUnboundAgent() public {
        vm.expectRevert("MockSourceBindingRegistry: no source binding");
        registry.getSourceNFT(999);
    }

    function test_hasSourceNFTReturnsFalseForUnboundAgent() public view {
        assertFalse(registry.hasSourceNFT(999));
    }

    function test_supportsInterfaceAdvertisesBothFullAndViewIds() public view {
        assertTrue(registry.supportsInterface(type(IAgentSourceBinding).interfaceId));
        assertTrue(registry.supportsInterface(type(IAgentSourceBindingView).interfaceId));
        // Independently confirm the ids themselves match the spec-documented values,
        // not just that the contract's own type(...) reference is internally consistent.
        assertEq(type(IAgentSourceBinding).interfaceId, bytes4(0x27eba962));
        assertEq(type(IAgentSourceBindingView).interfaceId, bytes4(0x8b3597c9));
    }

    function test_sourceNFTLinkedEventEmittedExactlyOnce() public {
        uint256 srcId = source.mint(holder);

        vm.expectEmit(true, true, false, true);
        emit IAgentSourceBinding.SourceNFTLinked(1, address(source), srcId);

        vm.prank(holder);
        registry.registerWithSource(srcId);
    }
}
