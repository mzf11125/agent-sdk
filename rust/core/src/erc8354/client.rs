use crate::DataProvider;
use crate::erc8354::recompute::Verdict;
use alloc::vec::Vec;
use alloy_core::sol;
use alloy_core::sol_types::SolValue;
use alloy_primitives::FixedBytes;

sol! {
    /// Solidity struct matching the on-chain Verdict encoding (no type hash).
    struct VerdictEncoding {
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
}

fn encode_verdict(v: &Verdict) -> Vec<u8> {
    let enc = VerdictEncoding {
        agentId: v.agent_id,
        domainId: v.domain_id,
        policyRoot: v.policy_root,
        actionCommitment: v.action_commitment,
        executor: v.executor,
        expiry: v.expiry,
        nullifier: v.nullifier,
        decision: v.decision,
        policyKind: v.policy_kind,
    };
    enc.abi_encode()
}

/// ERC-8354 ConfidentialPolicyVerdict client, read-only side.
pub struct ConfidentialPolicyVerdictClient<D: DataProvider> {
    provider: D,
}

impl<D: DataProvider> ConfidentialPolicyVerdictClient<D> {
    pub fn new(provider: D) -> Self {
        Self { provider }
    }

    pub fn verify(&self, verdict: &Verdict, proof: &[u8]) -> Result<bool, alloc::string::String> {
        let key = [
            b"erc8354:guard:verify:".as_slice(),
            encode_verdict(verdict).as_slice(),
            proof,
        ]
        .concat();
        let data = self.provider.fetch(&key);
        decode_bool(&data)
    }

    pub fn verdict_digest(&self, verdict: &Verdict) -> Result<FixedBytes<32>, alloc::string::String> {
        let key = [b"erc8354:guard:verdictDigest:".as_slice(), encode_verdict(verdict).as_slice()].concat();
        let data = self.provider.fetch(&key);
        if data.len() != 32 {
            return Err("verdict digest returned wrong length".into());
        }
        Ok(FixedBytes::<32>::from_slice(&data))
    }

    pub fn is_consumed(
        &self,
        domain_id: FixedBytes<32>,
        nullifier: FixedBytes<32>,
    ) -> Result<bool, alloc::string::String> {
        let key = [
            b"erc8354:guard:isConsumed:".as_slice(),
            domain_id.as_slice(),
            nullifier.as_slice(),
        ]
        .concat();
        let data = self.provider.fetch(&key);
        decode_bool(&data)
    }
}

/// ERC-8354 IPolicyDomainRegistry client, read-only side.
pub struct PolicyDomainRegistryClient<D: DataProvider> {
    provider: D,
}

impl<D: DataProvider> PolicyDomainRegistryClient<D> {
    pub fn new(provider: D) -> Self {
        Self { provider }
    }

    pub fn is_root_acceptable(
        &self,
        domain_id: FixedBytes<32>,
        root: FixedBytes<32>,
    ) -> Result<bool, alloc::string::String> {
        let key = [
            b"erc8354:registry:isRootAcceptable:".as_slice(),
            domain_id.as_slice(),
            root.as_slice(),
        ]
        .concat();
        let data = self.provider.fetch(&key);
        decode_bool(&data)
    }
}

/// Decode an ABI encoded bool. A standard encoded true is 32 bytes with a
/// final byte of 0x01; reading the first byte treats that as false.
fn decode_bool(data: &[u8]) -> Result<bool, alloc::string::String> {
    if data.len() != 32 {
        return Err("bool return encoded with wrong length".into());
    }
    Ok(data[31] == 1)
}
