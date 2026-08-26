// SPDX-License-Identifier: MIT
pragma solidity 0.8.25;

import {Test} from "forge-std/Test.sol";
import {ConfidentialPolicyVerdict} from "../../../contracts/mocks/verify/ERC8354/ConfidentialPolicyVerdict.sol";
import {PolicyDomainRegistry} from "../../../contracts/mocks/verify/ERC8354/PolicyDomainRegistry.sol";
import {MockVerifier} from "../../../contracts/mocks/verify/ERC8354/MockVerifier.sol";
import {MockIdentityRegistry} from "../../../contracts/mocks/verify/ERC8354/MockIdentityRegistry.sol";
import {IConfidentialPolicyVerdict, Verdict} from "../../../contracts/mocks/verify/ERC8354/IConfidentialPolicyVerdict.sol";

contract ConfidentialPolicyVerdictTest is Test {
    PolicyDomainRegistry registry;
    ConfidentialPolicyVerdict guard;
    MockVerifier verifier;
    MockIdentityRegistry identityRegistry;

    bytes32 constant DOMAIN = keccak256("test-domain");
    bytes32 constant ROOT = keccak256("root-v1");
    bytes32 constant PROGRAM = keccak256("interpreter-vkey");
    uint256 constant EXECUTOR_KEY = 0xE0E0;
    address EXECUTOR;

    function setUp() public {
        vm.warp(1_700_000_000);
        registry = new PolicyDomainRegistry();
        verifier = new MockVerifier();
        identityRegistry = new MockIdentityRegistry();
        guard = new ConfidentialPolicyVerdict(registry);
        EXECUTOR = vm.addr(EXECUTOR_KEY);
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

    function test_InvalidProofOnVerifierRevert() public {
        // A malformed proof that makes the verifier itself revert must surface
        // as InvalidProof, not propagate the verifier's own revert.
        Verdict memory v = _verdict();
        verifier.setRevert(true);
        vm.prank(EXECUTOR);
        vm.expectRevert(IConfidentialPolicyVerdict.InvalidProof.selector);
        guard.consume(v, "proof");
    }

    function test_RejectsUnknownAgent() public {
        // Declare an identity registry and register agent 7. A verdict for a
        // different, unregistered agent id must revert with AgentUnknown, and
        // that check must fire before any later check can mask it.
        registry.setIdentityRegistry(DOMAIN, address(identityRegistry));
        identityRegistry.register(7, EXECUTOR);

        Verdict memory v = _verdict();
        v.agentId = 404; // not registered
        vm.prank(EXECUTOR);
        vm.expectRevert(
            abi.encodeWithSelector(IConfidentialPolicyVerdict.AgentUnknown.selector, v.agentId)
        );
        guard.consume(v, "proof");
    }

    function test_KnownAgentConsumes() public {
        registry.setIdentityRegistry(DOMAIN, address(identityRegistry));
        identityRegistry.register(1, EXECUTOR);

        Verdict memory v = _verdict();
        vm.prank(EXECUTOR);
        guard.consume(v, "proof");
        assertTrue(guard.isConsumed(DOMAIN, v.nullifier));
    }

    function test_StaleRootGraceMultiRotation() public {
        // A -> B -> C, with real time passing between rotations, so each
        // generation ages out on its own schedule rather than all at once.
        // This is the case the single-previous-slot implementation used to
        // reject early.
        bytes32 rootA = keccak256("root-a");
        bytes32 rootB = keccak256("root-b");
        bytes32 rootC = keccak256("root-c");

        registry.updateRoot(DOMAIN, rootA);
        vm.warp(block.timestamp + 30 minutes);
        registry.updateRoot(DOMAIN, rootB);
        uint256 rootASupersededAt = block.timestamp;
        vm.warp(block.timestamp + 30 minutes);
        registry.updateRoot(DOMAIN, rootC);

        // One hour of wall clock time has passed since rootA was superseded,
        // matching maxRootAge exactly, so all three are still acceptable.
        assertTrue(registry.isRootAcceptable(DOMAIN, rootA));
        assertTrue(registry.isRootAcceptable(DOMAIN, rootB));
        assertTrue(registry.isRootAcceptable(DOMAIN, rootC));

        // Past rootA's own window, measured from when it stopped being
        // current, rootA is rejected while rootB is still inside its own
        // window. Each generation ages out on its own schedule, not the
        // domain's most recent rotation.
        vm.warp(rootASupersededAt + 1 hours + 1);
        assertFalse(registry.isRootAcceptable(DOMAIN, rootA));
        assertTrue(registry.isRootAcceptable(DOMAIN, rootB));
        assertTrue(registry.isRootAcceptable(DOMAIN, rootC));
    }

    function test_DomainInactiveOnRevoke() public {
        // Revocation bypasses the grace window entirely and is checked ahead
        // of decision and executor authorization.
        registry.revokeDomain(DOMAIN);
        Verdict memory v = _verdict();
        vm.prank(EXECUTOR);
        vm.expectRevert(
            abi.encodeWithSelector(IConfidentialPolicyVerdict.DomainInactive.selector, DOMAIN)
        );
        guard.consume(v, "proof");
    }

    function test_PolicyRootRejected() public {
        // Rotate away from ROOT and let the grace window lapse, so a verdict
        // built against the old root is no longer acceptable.
        registry.updateRoot(DOMAIN, keccak256("root-v2"));
        vm.warp(block.timestamp + 1 hours + 1);
        Verdict memory v = _verdict();
        vm.prank(EXECUTOR);
        vm.expectRevert(
            abi.encodeWithSelector(IConfidentialPolicyVerdict.PolicyRootRejected.selector, v.policyRoot)
        );
        guard.consume(v, "proof");
    }

    function test_VerifyFalseOnVerifierRevert() public {
        // The read-only twin of test_InvalidProofOnVerifierRevert: verify
        // MUST return false, not revert, when the verifier itself reverts.
        Verdict memory v = _verdict();
        verifier.setRevert(true);
        assertFalse(guard.verify(v, "proof"));
    }

    function test_RelayedConsumeWithValidExecutorAuth() public {
        // The one branch where a signature is actually verified: any address
        // may submit consume on the executor's behalf if it carries a valid
        // EIP-712 signature by the executor over this verdict's digest.
        Verdict memory v = _verdict();
        bytes32 digest = guard.verdictDigest(v);
        (uint8 sigV, bytes32 sigR, bytes32 sigS) = vm.sign(EXECUTOR_KEY, digest);
        bytes memory executorAuth = abi.encodePacked(sigR, sigS, sigV);

        vm.prank(address(0xC0FFEE));
        guard.consume(v, "proof", executorAuth);
        assertTrue(guard.isConsumed(DOMAIN, v.nullifier));
    }

    function test_RelayedConsumeRejectsWrongSignature() public {
        // A signature from a key other than the executor's must not authorize
        // the relay, even though the submitter is otherwise unrestricted.
        Verdict memory v = _verdict();
        bytes32 digest = guard.verdictDigest(v);
        uint256 wrongKey = 0xBADBAD;
        (uint8 sigV, bytes32 sigR, bytes32 sigS) = vm.sign(wrongKey, digest);
        bytes memory executorAuth = abi.encodePacked(sigR, sigS, sigV);

        vm.prank(address(0xC0FFEE));
        vm.expectRevert(IConfidentialPolicyVerdict.ExecutorAuthInvalid.selector);
        guard.consume(v, "proof", executorAuth);
    }
}
