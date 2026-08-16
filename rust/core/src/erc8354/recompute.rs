use alloy_core::sol;
use alloy_primitives::hex;
use alloy_primitives::{keccak256, Address, FixedBytes, U256};

/// The source-class tag for a consumed ERC-8354 verdict: keccak256("zk-secret-policy").
pub const MECHANISM_ZK_SECRET_POLICY: [u8; 32] =
    hex!("a843829a78c66c29679817606d0c8a9fa26575b6c2ed0f9f97079d7c46577ac6");

sol! {
    /// Solidity struct matching the canonical PolicyAction encoding.
    struct PolicyActionEncoding {
        uint256 chainId;
        bytes32 domainId;
        uint256 agentId;
        address target;
        uint256 value;
        bytes32 callDataHash;
        uint256 actionNonce;
    }

    /// Solidity struct matching the EIP-712 Verdict hashStruct encoding.
    struct VerdictHashable {
        bytes32 typeHash;
        uint256 agentId;
        bytes32 domainId;
        bytes32 policyRoot;
        bytes32 actionCommitment;
        address executor;
        uint64 expiry;
        bytes32 nullifier;
        uint8 decision;
        uint8 policyKind;
    }

    /// Solidity struct matching the EIP-712 domain separator encoding.
    struct DomainHashable {
        bytes32 typeHash;
        bytes32 nameHash;
        bytes32 versionHash;
        uint256 chainId;
        address verifyingContract;
    }
}

/// The ERC-8354 verdict envelope. Every field is a public input of the
/// proving program.
pub struct Verdict {
    pub agent_id: U256,
    pub domain_id: FixedBytes<32>,
    pub policy_root: FixedBytes<32>,
    pub action_commitment: FixedBytes<32>,
    pub executor: Address,
    pub expiry: u64,
    pub nullifier: FixedBytes<32>,
    pub decision: u8,
    pub policy_kind: u8,
}

/// keccak256("Verdict(uint256 agentId,bytes32 domainId,bytes32 policyRoot,bytes32 actionCommitment,address executor,uint64 expiry,bytes32 nullifier,uint8 decision,uint8 policyKind)").
const VERDICT_TYPE_HASH: [u8; 32] =
    hex!("00922e359f6328eb71248af0b379d42b614b9bbf74553d8e3a2d28a919f97811");
/// keccak256("EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)").
const DOMAIN_TYPE_HASH: [u8; 32] =
    hex!("8b73c3c69bb8fe3d512ecc4cf759cc79239f7b179b0ffacaa9a75d522b39400f");
/// keccak256("ConfidentialPolicyVerdict").
const NAME_HASH: [u8; 32] =
    hex!("69d1fc3c9251a02e663fa90a81ff86cf5c616b34a39da5368ac3807b0b52a821");
/// keccak256("1").
const VERSION_HASH: [u8; 32] =
    hex!("c89efdaa54c0f20c7adf612882df0950f5a951637e0307cdcb4c672f298b8bc6");

/// Compute the canonical action commitment for a policy verdict.
///
/// ERC-8354 Action commitment:
///   actionCommitment = keccak256(abi.encode(chainId, domainId, agentId,
///       target, value, keccak256(callData), actionNonce))
///
/// The Guard recomputes this from the action it is about to execute and
/// compares it to `Verdict.actionCommitment`. Empty callData hashes to
/// keccak256(""), never bytes32 zero.
pub fn compute_action_commitment(
    chain_id: U256,
    domain_id: FixedBytes<32>,
    agent_id: U256,
    target: Address,
    value: U256,
    call_data: &[u8],
    action_nonce: U256,
) -> FixedBytes<32> {
    let call_data_hash = keccak256(call_data);
    let enc = PolicyActionEncoding {
        chainId: chain_id,
        domainId: domain_id,
        agentId: agent_id,
        target,
        value,
        callDataHash: call_data_hash,
        actionNonce: action_nonce,
    };
    keccak256(&alloy_core::sol_types::SolValue::abi_encode(&enc))
}

/// Compute the EIP-712 digest an executor signs to authorize a relayer.
///
/// ERC-8354 verdictDigest over the Verdict type, with EIP-712 domain name
/// "ConfidentialPolicyVerdict" and version "1".
pub fn compute_verdict_digest(
    verdict: &Verdict,
    chain_id: U256,
    verifying_contract: Address,
) -> FixedBytes<32> {
    let hashable = VerdictHashable {
        typeHash: VERDICT_TYPE_HASH.into(),
        agentId: verdict.agent_id,
        domainId: verdict.domain_id,
        policyRoot: verdict.policy_root,
        actionCommitment: verdict.action_commitment,
        executor: verdict.executor,
        expiry: verdict.expiry,
        nullifier: verdict.nullifier,
        decision: verdict.decision,
        policyKind: verdict.policy_kind,
    };
    let hash_struct = keccak256(&alloy_core::sol_types::SolValue::abi_encode(&hashable));

    let domain = DomainHashable {
        typeHash: DOMAIN_TYPE_HASH.into(),
        nameHash: NAME_HASH.into(),
        versionHash: VERSION_HASH.into(),
        chainId: chain_id,
        verifyingContract: verifying_contract,
    };
    let domain_separator = keccak256(&alloy_core::sol_types::SolValue::abi_encode(&domain));

    let mut preimage = [0u8; 66];
    preimage[0] = 0x19;
    preimage[1] = 0x01;
    preimage[2..34].copy_from_slice(domain_separator.as_slice());
    preimage[34..66].copy_from_slice(hash_struct.as_slice());
    keccak256(&preimage)
}

#[cfg(test)]
mod tests {
    use super::*;
    use alloy_primitives::hex;

    fn domain_id() -> FixedBytes<32> {
        FixedBytes::from(hex!(
            "34a63641b78652cdd53505da4f32cac6058bd148e3ff543f39f75997a89c2815"
        ))
    }

    fn address(last_byte: u8) -> Address {
        let mut bytes = [0u8; 20];
        bytes[19] = last_byte;
        Address::from(bytes)
    }

    #[test]
    fn golden_action_commitment() {
        let commitment = compute_action_commitment(
            U256::from(31337u64),
            domain_id(),
            U256::from(1u64),
            address(1),
            U256::ZERO,
            &[],
            U256::ZERO,
        );
        let expected = FixedBytes::<32>::from(hex!(
            "cc8e5dc414db5ed2340be02c3d7fdc725fe5f1463b382a7ed13f8036a4a0b7b1"
        ));
        assert_eq!(commitment, expected);
    }

    #[test]
    fn empty_call_data_hashes_keccak_empty() {
        let commitment = compute_action_commitment(
            U256::from(31337u64),
            domain_id(),
            U256::from(1u64),
            address(1),
            U256::ZERO,
            &[],
            U256::ZERO,
        );
        assert_ne!(commitment, FixedBytes::ZERO);
    }

    #[test]
    fn action_commitment_changes_with_nonce() {
        let args = |nonce: u64| {
            compute_action_commitment(
                U256::from(31337u64),
                domain_id(),
                U256::from(1u64),
                address(1),
                U256::ZERO,
                &[],
                U256::from(nonce),
            )
        };
        assert_ne!(args(0), args(1));
    }

    fn verdict() -> Verdict {
        Verdict {
            agent_id: U256::from(1u64),
            domain_id: domain_id(),
            policy_root: domain_id(),
            action_commitment: FixedBytes::from(hex!(
                "cc8e5dc414db5ed2340be02c3d7fdc725fe5f1463b382a7ed13f8036a4a0b7b1"
            )),
            executor: address(2),
            expiry: 2000000000,
            nullifier: FixedBytes::from(hex!(
                "6e47261c83f90eed41cda2b00caad094c33daa0a09fec22396b3e2bfe5e222b2"
            )),
            decision: 1,
            policy_kind: 0,
        }
    }

    #[test]
    fn golden_verdict_digest() {
        let digest = compute_verdict_digest(&verdict(), U256::from(31337u64), address(3));
        let expected = FixedBytes::<32>::from(hex!(
            "f2345f63ba9e78a068eb4f74640e6543289010540b457d8016771175ad460f32"
        ));
        assert_eq!(digest, expected);
    }

    #[test]
    fn verdict_digest_depends_on_verifying_contract() {
        let a = compute_verdict_digest(&verdict(), U256::from(31337u64), address(3));
        let b = compute_verdict_digest(&verdict(), U256::from(31337u64), address(4));
        assert_ne!(a, b);
    }

    #[test]
    fn mechanism_matches_golden() {
        let expected =
            hex!("a843829a78c66c29679817606d0c8a9fa26575b6c2ed0f9f97079d7c46577ac6");
        assert_eq!(MECHANISM_ZK_SECRET_POLICY, expected);
    }

    #[test]
    fn golden_vectors_json() {
        // Conformance vectors live in the testkit, alongside every other
        // language. Reading them here (rather than only inline constants) keeps
        // one source of truth. Skip gracefully when the checkout is partial.
        let path = std::path::Path::new(env!("CARGO_MANIFEST_DIR"))
            .join("../testkit/vectors/erc8354-verdict.vectors.json");
        let data = match std::fs::read_to_string(&path) {
            Ok(d) => d,
            Err(_) => return, // file absent, nothing to check
        };
        let parsed: serde_json::Value =
            serde_json::from_str(&data).expect("invalid vectors JSON");

        fn h32(s: &str) -> FixedBytes<32> {
            let bytes = alloy_primitives::hex::decode(s.strip_prefix("0x").unwrap_or(s))
                .expect("bad hex");
            FixedBytes::from_slice(&bytes)
        }
        fn u256(v: serde_json::Value) -> U256 {
            U256::from(v.as_u64().expect("numeric input"))
        }

        for vector in parsed["vectors"].as_array().unwrap() {
            let expected = h32(vector["expected"].as_str().unwrap());
            let inputs = &vector["inputs"];
            match vector["step"].as_str().unwrap() {
                "8354/action-commitment" => {
                    let got = compute_action_commitment(
                        u256(inputs["chainId"].clone()),
                        h32(inputs["domainId"].as_str().unwrap()),
                        u256(inputs["agentId"].clone()),
                        Address::from_slice(
                            &alloy_primitives::hex::decode(
                                inputs["target"].as_str().unwrap().strip_prefix("0x").unwrap(),
                            )
                            .unwrap(),
                        ),
                        u256(inputs["value"].clone()),
                        &alloy_primitives::hex::decode(
                            inputs["callData"].as_str().unwrap().strip_prefix("0x").unwrap(),
                        )
                        .unwrap(),
                        u256(inputs["actionNonce"].clone()),
                    );
                    assert_eq!(got, expected, "action-commitment vector mismatch");
                }
                "8354/verdict-digest" => {
                    let v = Verdict {
                        agent_id: u256(inputs["agentId"].clone()),
                        domain_id: h32(inputs["domainId"].as_str().unwrap()),
                        policy_root: h32(inputs["policyRoot"].as_str().unwrap()),
                        action_commitment: h32(inputs["actionCommitment"].as_str().unwrap()),
                        executor: Address::from_slice(
                            &alloy_primitives::hex::decode(
                                inputs["executor"].as_str().unwrap().strip_prefix("0x").unwrap(),
                            )
                            .unwrap(),
                        ),
                        expiry: inputs["expiry"].as_u64().unwrap(),
                        nullifier: h32(inputs["nullifier"].as_str().unwrap()),
                        decision: inputs["decision"].as_u64().unwrap() as u8,
                        policy_kind: inputs["policyKind"].as_u64().unwrap() as u8,
                    };
                    let got = compute_verdict_digest(
                        &v,
                        u256(inputs["chainId"].clone()),
                        Address::from_slice(
                            &alloy_primitives::hex::decode(
                                inputs["verifyingContract"]
                                    .as_str()
                                    .unwrap()
                                    .strip_prefix("0x")
                                    .unwrap(),
                            )
                            .unwrap(),
                        ),
                    );
                    assert_eq!(got, expected, "verdict-digest vector mismatch");
                }
                other => panic!("unknown vector step {other}"),
            }
        }
    }
}
