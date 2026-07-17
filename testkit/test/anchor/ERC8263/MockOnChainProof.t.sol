// SPDX-License-Identifier: MIT
pragma solidity 0.8.25;

import "forge-std/Test.sol";
import "../../../contracts/mocks/anchor/ERC8263/MockOnChainProof.sol";

/// @notice Local-testing-only Foundry test for MockOnChainProof.
///         Validates canonical-form guards and event emission.
contract MockOnChainProofTest is Test {
    /// Spec constant: keccak256("AnchorProof(uint8,bytes32,bytes32,address,bytes)")
    bytes32 constant TOPIC0 =
        0x9fe832d83a52f83bd7d54181e4cc7ff8b4e227cc1d3a0144376894b5df6c23cc;

    MockOnChainProof harness;

    function setUp() public {
        harness = new MockOnChainProof();
    }

    /// Compilation guard: the interface compiles and the mock implements it.
    function test_interfaceCompiles() public view {
        IOnChainProof iface = IOnChainProof(address(harness));
        assertTrue(address(iface) != address(0));
    }

    /// topic0 of the canonical event matches the spec constant.
    function test_topic0_matchesSpecConstant() public pure {
        assertEq(
            keccak256("AnchorProof(uint8,bytes32,bytes32,address,bytes)"),
            TOPIC0
        );
    }

    /// ANONYMOUS scheme (0x00) requires agentId == 0.
    function test_anonymousSchemeRequiresZeroAgentId() public {
        vm.expectRevert("MockOnChainProof: ANONYMOUS scheme requires agentId == 0");
        harness.anchor(0x00, keccak256("non-zero"), keccak256("proof"));
    }

    /// ANONYMOUS scheme succeeds with agentId == 0.
    function test_anonymousSchemeSucceeds() public {
        harness.anchor(0x00, bytes32(0), keccak256("proof"));
        // No revert expected
    }

    /// REGISTRY scheme (0x01) requires non-zero agentId.
    function test_registrySchemeRequiresNonZeroAgentId() public {
        vm.expectRevert("MockOnChainProof: registered scheme requires non-zero agentId");
        harness.anchor(0x01, bytes32(0), keccak256("proof"));
    }

    /// REGISTRY scheme succeeds with non-zero agentId.
    function test_registrySchemeSucceeds() public {
        bytes32 agentId = keccak256("agent");
        harness.anchor(0x01, agentId, keccak256("proof"));
    }

    /// URI_HASH scheme (0x02) requires non-zero agentId.
    function test_uriHashSchemeRequiresNonZeroAgentId() public {
        vm.expectRevert("MockOnChainProof: registered scheme requires non-zero agentId");
        harness.anchor(0x02, bytes32(0), keccak256("proof"));
    }

    /// URI_HASH scheme succeeds with non-zero agentId.
    function test_uriHashSchemeSucceeds() public {
        bytes32 agentId = keccak256("agent");
        harness.anchor(0x02, agentId, keccak256("proof"));
    }

    /// Reserved schemes (0x03+) revert.
    function test_reservedSchemesRevert() public {
        vm.expectRevert("MockOnChainProof: reserved agentIdScheme");
        harness.anchor(0x03, bytes32(0), keccak256("proof"));
    }

    /// proofHash must be non-zero.
    function test_proofHashMustBeNonZero() public {
        vm.expectRevert("MockOnChainProof: proofHash must be non-zero");
        harness.anchor(0x01, keccak256("agent"), bytes32(0));
    }

    /// anchor() emits AnchorProof with correct topic layout (empty aux).
    function test_anchor_eventTopicLayout_emptyAux() public {
        bytes32 agentId = keccak256("agent");
        bytes32 proofHash = keccak256("proof");

        vm.recordLogs();
        harness.anchor(0x02, agentId, proofHash);
        Vm.Log[] memory logs = vm.getRecordedLogs();

        assertEq(logs.length, 1);
        assertEq(logs[0].topics.length, 4);
        assertEq(logs[0].topics[0], TOPIC0);
        assertEq(logs[0].topics[1], agentId);
        assertEq(logs[0].topics[2], proofHash);
        assertEq(logs[0].topics[3], bytes32(uint256(uint160(address(this)))));

        (uint8 scheme, bytes memory aux) = abi.decode(logs[0].data, (uint8, bytes));
        assertEq(scheme, 0x02);
        assertEq(aux.length, 0);
    }

    /// anchorWithAux() emits AnchorProof with correct aux bytes.
    function test_anchorWithAux_eventTopicLayout_withAux() public {
        bytes32 agentId = keccak256("agent");
        bytes32 proofHash = keccak256("proof");
        bytes memory auxIn = hex"c0ffee";

        vm.recordLogs();
        harness.anchorWithAux(0x01, agentId, proofHash, auxIn);
        Vm.Log[] memory logs = vm.getRecordedLogs();

        assertEq(logs.length, 1);
        assertEq(logs[0].topics.length, 4);
        assertEq(logs[0].topics[0], TOPIC0);
        assertEq(logs[0].topics[1], agentId);
        assertEq(logs[0].topics[2], proofHash);
        assertEq(logs[0].topics[3], bytes32(uint256(uint160(address(this)))));

        (uint8 scheme, bytes memory aux) = abi.decode(logs[0].data, (uint8, bytes));
        assertEq(scheme, 0x01);
        assertEq(aux, auxIn);
    }
}
