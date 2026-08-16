// SPDX-License-Identifier: MIT
pragma solidity 0.8.25;

import {Test} from "forge-std/Test.sol";
import {ConfidentialPolicyVerdict} from "../../../contracts/mocks/verify/ERC8354/ConfidentialPolicyVerdict.sol";
import {PolicyDomainRegistry} from "../../../contracts/mocks/verify/ERC8354/PolicyDomainRegistry.sol";
import {MockVerifier} from "../../../contracts/mocks/verify/ERC8354/MockVerifier.sol";
import {IConfidentialPolicyVerdict, Verdict} from "../../../contracts/mocks/verify/ERC8354/IConfidentialPolicyVerdict.sol";

contract ConfidentialPolicyVerdictTest is Test {
    PolicyDomainRegistry registry;
    ConfidentialPolicyVerdict guard;
    MockVerifier verifier;

    bytes32 constant DOMAIN = keccak256("test-domain");
    bytes32 constant ROOT = keccak256("root-v1");
    bytes32 constant PROGRAM = keccak256("interpreter-vkey");
    address constant EXECUTOR = address(0xE0);

    function setUp() public {
        vm.warp(1_700_000_000);
        registry = new PolicyDomainRegistry();
        verifier = new MockVerifier();
        guard = new ConfidentialPolicyVerdict(registry);
        registry.registerDomain(DOMAIN, address(0xA11CE), address(verifier), PROGRAM, 1 hours);
        registry.updateRoot(DOMAIN, ROOT);
    }

    function _verdict() internal view returns (Verdict memory v) {
        v = Verdict({
            agentId: 1,
            domainId: DOMAIN,
            policyRoot: ROOT,
            actionCommitment: keccak256("action"),
            executor: EXECUTOR,
            expiry: uint64(block.timestamp + 1 hours),
            nullifier: keccak256("nf-1"),
            decision: 1,
            policyKind: 0
        });
    }

    function test_HappyPath() public {
        Verdict memory v = _verdict();
        vm.prank(EXECUTOR);
        guard.consume(v, "proof");
        assertTrue(guard.isConsumed(DOMAIN, v.nullifier));
    }

    function test_VerifyReturnsTrue() public view {
        Verdict memory v = _verdict();
        assertTrue(guard.verify(v, "proof"));
    }

    function test_Replay() public {
        Verdict memory v = _verdict();
        vm.startPrank(EXECUTOR);
        guard.consume(v, "proof");
        vm.expectRevert(
            abi.encodeWithSelector(IConfidentialPolicyVerdict.VerdictReplayed.selector, v.nullifier)
        );
        guard.consume(v, "proof");
        vm.stopPrank();
    }

    function test_Expired() public {
        Verdict memory v = _verdict();
        vm.warp(v.expiry);
        vm.prank(EXECUTOR);
        vm.expectRevert(
            abi.encodeWithSelector(IConfidentialPolicyVerdict.VerdictExpired.selector, v.expiry)
        );
        guard.consume(v, "proof");
    }

    function test_ExecutorMismatch() public {
        Verdict memory v = _verdict();
        vm.prank(address(0xBAD));
        vm.expectRevert(
            abi.encodeWithSelector(
                IConfidentialPolicyVerdict.ExecutorMismatch.selector,
                v.executor,
                address(0xBAD)
            )
        );
        guard.consume(v, "proof");
    }

    function test_DeniedDecision() public {
        Verdict memory v = _verdict();
        v.decision = 0;
        v.policyKind = 1;
        vm.prank(EXECUTOR);
        vm.expectRevert(IConfidentialPolicyVerdict.VerdictDenied.selector);
        guard.consume(v, "proof");
    }

    function test_KindMismatch() public {
        Verdict memory v = _verdict();
        v.policyKind = 3;
        vm.prank(EXECUTOR);
        vm.expectRevert(
            abi.encodeWithSelector(
                IConfidentialPolicyVerdict.VerdictKindMismatch.selector,
                v.decision,
                v.policyKind
            )
        );
        guard.consume(v, "proof");
    }

    function test_InvalidProof() public {
        Verdict memory v = _verdict();
        verifier.setResult(false);
        vm.prank(EXECUTOR);
        vm.expectRevert(IConfidentialPolicyVerdict.InvalidProof.selector);
        guard.consume(v, "proof");
    }
}
