// SPDX-License-Identifier: MIT
pragma solidity 0.8.25;

import {Test} from "forge-std/Test.sol";
import {IWyriweAttestation} from "@agent-ercs/verify/ERC8299/IWyriweAttestation.sol";
import {MockWyriweAttestation} from "../../../contracts/mocks/verify/ERC8299/MockWyriweAttestation.sol";

contract MockWyriweAttestationTest is Test {
    MockWyriweAttestation public wyriwe;

    bytes32 public constant AGENT_ID = keccak256("agent");
    address public constant REGISTRY = address(0x5678);
    bytes32 public constant MODEL_HASH = keccak256("model");
    bytes32 public constant RAW_INPUT_HASH = keccak256("raw_input");
    bytes32 public constant SANITIZATION_PIPELINE_HASH = keccak256("sanitization_pipeline");
    bytes32 public constant INPUT_HASH = keccak256("input");
    bytes32 public constant OUTPUT_HASH = keccak256("output");
    uint256 public constant TIMESTAMP = 1000000;

    function setUp() public {
        wyriwe = new MockWyriweAttestation();
    }

    function testProofSystem() public view {
        assertEq(wyriwe.proofSystem(), "attestation/wyriwe");
    }

    function testVerifyAcceptsValidSignature() public view {
        IWyriweAttestation.WyriweAttestation memory attestation = IWyriweAttestation.WyriweAttestation({
            agentId: AGENT_ID,
            registry: REGISTRY,
            modelHash: MODEL_HASH,
            rawInputHash: RAW_INPUT_HASH,
            sanitizationPipelineHash: SANITIZATION_PIPELINE_HASH,
            inputHash: INPUT_HASH,
            outputHash: OUTPUT_HASH,
            timestamp: TIMESTAMP
        });

        // The mock checks: keccak256(signature) == keccak256(abi.encode(attestation))
        bytes memory validSignature = abi.encodePacked(keccak256(abi.encode(attestation)));
        bool result = wyriwe.verify(attestation, validSignature);
        assertTrue(result);
    }

    function testVerifyReturnsFalseForInvalidSignature() public view {
        IWyriweAttestation.WyriweAttestation memory attestation = IWyriweAttestation.WyriweAttestation({
            agentId: AGENT_ID,
            registry: REGISTRY,
            modelHash: MODEL_HASH,
            rawInputHash: RAW_INPUT_HASH,
            sanitizationPipelineHash: SANITIZATION_PIPELINE_HASH,
            inputHash: INPUT_HASH,
            outputHash: OUTPUT_HASH,
            timestamp: TIMESTAMP
        });

        bytes memory invalidSignature = abi.encodePacked(bytes32(uint256(1)), bytes32(uint256(2)), uint8(27));
        bool result = wyriwe.verify(attestation, invalidSignature);
        assertFalse(result);
    }
}
