use crate::DataProvider;
use alloy_primitives::FixedBytes;

/// ERC-8263 OnChainProof anchor client.
///
/// Anchors a cryptographic commitment (`proofHash`) together with an
/// identity-scheme byte and 32-byte agent identifier, producing a verifiable,
/// immutable timeline of agent activity.
///
/// Generic over `DataProvider` — compiles in host (RPC-backed) and guest
/// (preimage-backed) contexts.
pub struct OnChainProofClient<D: DataProvider> {
    provider: D,
}

impl<D: DataProvider> OnChainProofClient<D> {
    pub fn new(provider: D) -> Self {
        Self { provider }
    }

    /// Minimal anchor: empty aux, lowest calldata cost.
    ///
    /// `agent_id_scheme` — identity scheme byte (0x00 ANONYMOUS, 0x01 REGISTRY, 0x02 URI_HASH).
    /// `agent_id` — 32-byte agent identifier per the scheme registry.
    /// `proof_hash` — non-zero 32-byte commitment to the action.
    pub fn anchor(
        &self,
        agent_id_scheme: u8,
        agent_id: FixedBytes<32>,
        proof_hash: FixedBytes<32>,
    ) -> Result<Vec<u8>, String> {
        let mut key = Vec::from(&b"erc8263:onChainProof:anchor:"[..]);
        key.push(agent_id_scheme);
        key.extend_from_slice(agent_id.as_slice());
        key.extend_from_slice(proof_hash.as_slice());
        let data = self.provider.fetch(&key);
        if data.is_empty() {
            return Err("anchor transaction failed".into());
        }
        Ok(data)
    }

    /// Extended anchor with opaque aux bytes for adjacent protocols.
    ///
    /// `agent_id_scheme` — identity scheme byte (0x00 ANONYMOUS, 0x01 REGISTRY, 0x02 URI_HASH).
    /// `agent_id` — 32-byte agent identifier per the scheme registry.
    /// `proof_hash` — non-zero 32-byte commitment to the action.
    /// `aux` — opaque extension bytes (non-normative; e.g. OCP digest commitments,
    /// session ids, parent-proof references).
    pub fn anchor_with_aux(
        &self,
        agent_id_scheme: u8,
        agent_id: FixedBytes<32>,
        proof_hash: FixedBytes<32>,
        aux: &[u8],
    ) -> Result<Vec<u8>, String> {
        let mut key = Vec::from(&b"erc8263:onChainProof:anchorWithAux:"[..]);
        key.push(agent_id_scheme);
        key.extend_from_slice(agent_id.as_slice());
        key.extend_from_slice(proof_hash.as_slice());
        key.extend_from_slice(aux);
        let data = self.provider.fetch(&key);
        if data.is_empty() {
            return Err("anchorWithAux transaction failed".into());
        }
        Ok(data)
    }
}
