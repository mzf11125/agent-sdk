// SPDX-License-Identifier: MIT
pragma solidity 0.8.25;

import {Test} from "forge-std/Test.sol";
import {MockIdentityRegistry} from "../../../contracts/mocks/identity/ERC8004/MockIdentityRegistry.sol";
import {IIdentityRegistry} from "@agent-ercs/identity/ERC8004/IIdentityRegistry.sol";

contract MockIdentityRegistryTest is Test {
    bytes32 private constant EIP712_DOMAIN_TYPEHASH =
        keccak256("EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)");

    bytes32 private constant SET_AGENT_WALLET_TYPEHASH =
        keccak256("SetAgentWallet(uint256 agentId,address newWallet,uint256 deadline)");

    MockIdentityRegistry internal registry;
    address internal owner = address(0xA11CE);

    function setUp() public {
        registry = new MockIdentityRegistry();
    }

    function test_registerAssignsIncrementingAgentId() public {
        vm.prank(owner);
        uint256 firstId = registry.register("ipfs://agent-1", new IIdentityRegistry.MetadataEntry[](0));
        assertEq(firstId, 1);
        assertEq(registry.ownerOf(firstId), owner);
        assertEq(registry.tokenURI(firstId), "ipfs://agent-1");

        vm.prank(owner);
        uint256 secondId = registry.register();
        assertEq(secondId, 2);
    }

    function test_setAndGetMetadata() public {
        vm.prank(owner);
        uint256 agentId = registry.register();

        vm.prank(owner);
        registry.setMetadata(agentId, "role", bytes("validator"));

        assertEq(registry.getMetadata(agentId, "role"), bytes("validator"));
    }

    function test_setMetadataRevertsForNonOwner() public {
        vm.prank(owner);
        uint256 agentId = registry.register();

        vm.prank(address(0xBEEF));
        vm.expectRevert("MockIdentityRegistry: not agent owner");
        registry.setMetadata(agentId, "role", bytes("validator"));
    }

    function test_setAgentWalletWithValidSignatureSucceeds() public {
        (address signer, uint256 signerKey) = makeAddrAndKey("agent-owner");
        vm.prank(signer);
        uint256 agentId = registry.register();

        address newWallet = address(0xCAFE);
        uint256 deadline = block.timestamp + 1 hours;
        bytes32 digest = _walletDigest(agentId, newWallet, deadline);
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(signerKey, digest);
        bytes memory signature = abi.encodePacked(r, s, v);

        registry.setAgentWallet(agentId, newWallet, deadline, signature);

        assertEq(registry.getAgentWallet(agentId), newWallet);
    }

    function test_setAgentWalletWithWrongSignerReverts() public {
        vm.prank(owner);
        uint256 agentId = registry.register();

        (, uint256 wrongKey) = makeAddrAndKey("not-the-owner");
        address newWallet = address(0xCAFE);
        uint256 deadline = block.timestamp + 1 hours;
        bytes32 digest = _walletDigest(agentId, newWallet, deadline);
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(wrongKey, digest);
        bytes memory signature = abi.encodePacked(r, s, v);

        vm.expectRevert("MockIdentityRegistry: invalid wallet signature");
        registry.setAgentWallet(agentId, newWallet, deadline, signature);
    }

    function _walletDigest(uint256 agentId, address newWallet, uint256 deadline) internal view returns (bytes32) {
        bytes32 domainSeparator = keccak256(
            abi.encode(
                EIP712_DOMAIN_TYPEHASH,
                keccak256(bytes("MockIdentityRegistry")),
                keccak256(bytes("1")),
                block.chainid,
                address(registry)
            )
        );
        bytes32 structHash = keccak256(abi.encode(SET_AGENT_WALLET_TYPEHASH, agentId, newWallet, deadline));
        return keccak256(abi.encodePacked("\x19\x01", domainSeparator, structHash));
    }
}
