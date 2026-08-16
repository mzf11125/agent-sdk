//! ERC-8354 Confidential Policy Verdicts: full (read and write) Layer 1 client.
//!
//! Unlike [`agent_sdk_core::erc8354::client`] (which is `no_std` and read-only,
//! backed by a preimage data provider), this client lives in `agent-sdk-providers`
//! (std) and is bound to concrete addresses plus an alloy signer, so it can both
//! read chain state and broadcast the `consume` and `consumeRelayed` transactions.
//!
//! The `core` crate pins `alloy-primitives` 0.8 while this crate uses `alloy`
//! 2.x (alloy-primitives 1.x), so this client defines its own [`Verdict`] type
//! rather than reusing the `no_std` one.

use alloy::network::Ethereum;
use alloy::primitives::{keccak256, Address, Bytes, FixedBytes, U256};
use alloy::providers::Provider;
use alloy::sol;
use alloy::transports::TransportErrorKind;

fn reverted(which: &str) -> alloy::contract::Error {
    alloy::contract::Error::TransportError(TransportErrorKind::custom_str(&format!(
        "{which} reverted"
    )))
}

sol! {
    #[allow(missing_docs)]
    struct Verdict {
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

    #[allow(missing_docs)]
    #[sol(rpc)]
    interface IConfidentialPolicyVerdict {
        function verify(Verdict v, bytes proof) external view returns (bool);
        function verdictDigest(Verdict v) external view returns (bytes32);
        function isConsumed(bytes32 domainId, bytes32 nullifier) external view returns (bool);
        function supportsInterface(bytes4 interfaceId) external view returns (bool);
        function consume(Verdict v, bytes proof) external;
        function consume(Verdict v, bytes proof, bytes executorAuth) external;
    }

    #[allow(missing_docs)]
    #[sol(rpc)]
    interface IPolicyDomainRegistry {
        struct Domain {
            address registrar;
            address verifier;
            bytes32 programKey;
            uint64 maxRootAge;
            bool active;
        }

        function domain(bytes32 domainId) external view returns (Domain);
        function currentRoot(bytes32 domainId) external view returns (bytes32 root, uint64 version, uint64 updatedAt);
        function isRootAcceptable(bytes32 domainId, bytes32 root) external view returns (bool);
    }
}

/// A decoded registry `Domain`.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Domain {
    pub registrar: Address,
    pub verifier: Address,
    pub program_key: FixedBytes<32>,
    pub max_root_age: u64,
    pub active: bool,
}

/// A decoded registry `currentRoot` response.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub struct CurrentRoot {
    pub root: FixedBytes<32>,
    pub version: u64,
    pub updated_at: u64,
}

/// The verdict digest an executor signs to authorize a relayer. Mirrors ERC-8354
/// `verdictDigest` over the `Verdict` type with EIP-712 domain name
/// `ConfidentialPolicyVerdict` and version `1`.
pub fn compute_verdict_digest(
    verdict: &Verdict,
    chain_id: U256,
    verifying_contract: Address,
) -> FixedBytes<32> {
    const VERDICT_TYPE: &str =
        "Verdict(uint256 agentId,bytes32 domainId,bytes32 policyRoot,bytes32 actionCommitment,address executor,uint64 expiry,bytes32 nullifier,uint8 decision,uint8 policyKind)";
    const DOMAIN_TYPE: &str =
        "EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)";

    let verdict_typehash = keccak256(VERDICT_TYPE.as_bytes());

    let mut struct_preimage = Vec::with_capacity(32 * 10);
    struct_preimage.extend_from_slice(verdict_typehash.as_slice());
    struct_preimage.extend_from_slice(&word(&verdict.agentId));
    struct_preimage.extend_from_slice(verdict.domainId.as_slice());
    struct_preimage.extend_from_slice(verdict.policyRoot.as_slice());
    struct_preimage.extend_from_slice(verdict.actionCommitment.as_slice());
    struct_preimage.extend_from_slice(verdict.executor.into_word().as_slice());
    struct_preimage.extend_from_slice(&word(&U256::from(verdict.expiry)));
    struct_preimage.extend_from_slice(verdict.nullifier.as_slice());
    struct_preimage.extend_from_slice(&word(&U256::from(verdict.decision)));
    struct_preimage.extend_from_slice(&word(&U256::from(verdict.policyKind)));
    let hash_struct = keccak256(&struct_preimage);

    let domain_typehash = keccak256(DOMAIN_TYPE.as_bytes());
    let name_hash = keccak256(b"ConfidentialPolicyVerdict");
    let version_hash = keccak256(b"1");
    let mut domain_preimage = Vec::with_capacity(32 * 5);
    domain_preimage.extend_from_slice(domain_typehash.as_slice());
    domain_preimage.extend_from_slice(name_hash.as_slice());
    domain_preimage.extend_from_slice(version_hash.as_slice());
    domain_preimage.extend_from_slice(&word(&chain_id));
    domain_preimage.extend_from_slice(verifying_contract.into_word().as_slice());
    let domain_separator = keccak256(&domain_preimage);

    let mut digest = [0u8; 66];
    digest[0] = 0x19;
    digest[1] = 0x01;
    digest[2..34].copy_from_slice(domain_separator.as_slice());
    digest[34..66].copy_from_slice(hash_struct.as_slice());
    keccak256(&digest)
}

fn word(v: &U256) -> [u8; 32] {
    v.to_be_bytes::<32>()
}

/// Address-bound ERC-8354 client over a generic alloy provider with a signer.
#[derive(Debug, Clone)]
pub struct Erc8354Client<P> {
    guard: Address,
    registry: Address,
    provider: P,
}

impl<P> Erc8354Client<P>
where
    P: Provider<Ethereum> + Clone,
{
    /// Bind to the guard and registry addresses.
    pub fn new(guard: Address, registry: Address, provider: P) -> Self {
        Self {
            guard,
            registry,
            provider,
        }
    }

    /// Call `verify` on the guard.
    pub async fn verify(&self, verdict: &Verdict, proof: &[u8]) -> Result<bool, alloy::contract::Error> {
        IConfidentialPolicyVerdict::new(self.guard, self.provider.clone())
            .verify(verdict.clone(), Bytes::copy_from_slice(proof))
            .call()
            .await
    }

    /// Call `verdictDigest` on the guard.
    pub async fn verdict_digest(&self, verdict: &Verdict) -> Result<FixedBytes<32>, alloy::contract::Error> {
        IConfidentialPolicyVerdict::new(self.guard, self.provider.clone())
            .verdictDigest(verdict.clone())
            .call()
            .await
    }

    /// Call `isConsumed` on the guard.
    pub async fn is_consumed(
        &self,
        domain_id: FixedBytes<32>,
        nullifier: FixedBytes<32>,
    ) -> Result<bool, alloy::contract::Error> {
        IConfidentialPolicyVerdict::new(self.guard, self.provider.clone())
            .isConsumed(domain_id, nullifier)
            .call()
            .await
    }

    /// Call `supportsInterface` on the guard.
    pub async fn supports_interface(&self, interface_id: [u8; 4]) -> Result<bool, alloy::contract::Error> {
        IConfidentialPolicyVerdict::new(self.guard, self.provider.clone())
            .supportsInterface(FixedBytes::from(interface_id))
            .call()
            .await
    }

    /// Broadcast `consume` (direct executor path), wait for the receipt, and
    /// return an error if the transaction reverted.
    pub async fn consume(&self, verdict: &Verdict, proof: &[u8]) -> Result<(), alloy::contract::Error> {
        let receipt = IConfidentialPolicyVerdict::new(self.guard, self.provider.clone())
            .consume_0(verdict.clone(), Bytes::copy_from_slice(proof))
            .send()
            .await?
            .get_receipt()
            .await?;
        if !receipt.status() {
            return Err(reverted("consume"));
        }
        Ok(())
    }

    /// Broadcast the relayed (`consume(Verdict,bytes,bytes)`) path carrying the
    /// executor signature, wait for the receipt, and return an error on revert.
    pub async fn consume_relayed(
        &self,
        verdict: &Verdict,
        proof: &[u8],
        executor_auth: &[u8],
    ) -> Result<(), alloy::contract::Error> {
        let receipt = IConfidentialPolicyVerdict::new(self.guard, self.provider.clone())
            .consume_1(
                verdict.clone(),
                Bytes::copy_from_slice(proof),
                Bytes::copy_from_slice(executor_auth),
            )
            .send()
            .await?
            .get_receipt()
            .await?;
        if !receipt.status() {
            return Err(reverted("consumeRelayed"));
        }
        Ok(())
    }

    /// Read the registry `domain`.
    pub async fn domain(&self, domain_id: FixedBytes<32>) -> Result<Domain, alloy::contract::Error> {
        let d = IPolicyDomainRegistry::new(self.registry, self.provider.clone())
            .domain(domain_id)
            .call()
            .await?;
        Ok(Domain {
            registrar: d.registrar,
            verifier: d.verifier,
            program_key: d.programKey,
            max_root_age: d.maxRootAge,
            active: d.active,
        })
    }

    /// Read the registry `currentRoot`.
    pub async fn current_root(&self, domain_id: FixedBytes<32>) -> Result<CurrentRoot, alloy::contract::Error> {
        let r = IPolicyDomainRegistry::new(self.registry, self.provider.clone())
            .currentRoot(domain_id)
            .call()
            .await?;
        Ok(CurrentRoot {
            root: r.root,
            version: r.version,
            updated_at: r.updatedAt,
        })
    }

    /// Read the registry `isRootAcceptable`.
    pub async fn is_root_acceptable(
        &self,
        domain_id: FixedBytes<32>,
        root: FixedBytes<32>,
    ) -> Result<bool, alloy::contract::Error> {
        IPolicyDomainRegistry::new(self.registry, self.provider.clone())
            .isRootAcceptable(domain_id, root)
            .call()
            .await
    }
}