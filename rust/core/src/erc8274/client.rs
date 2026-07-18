use alloy_primitives::FixedBytes;
use crate::DataProvider;

/// ERC-8274 ProofVerifier client.
pub struct ProofVerifierClient<D: DataProvider> {
    provider: D,
}

impl<D: DataProvider> ProofVerifierClient<D> {
    pub fn new(provider: D) -> Self { Self { provider } }

    pub fn verify(&self, input_hash: FixedBytes<32>, output_hash: FixedBytes<32>, metadata: &[u8], proof: &[u8]) -> Result<bool, String> {
        let key = [b"erc8274:proofVerifier:verify:", input_hash.as_slice(), output_hash.as_slice(), metadata, proof].concat();
        let data = self.provider.fetch(&key);
        Ok(!data.is_empty() && data[0] != 0)
    }

    pub fn proof_system(&self) -> Result<String, String> {
        let data = self.provider.fetch(b"erc8274:proofVerifier:proofSystem");
        String::from_utf8(data).map_err(|e| e.to_string())
    }
}

/// ERC-8274 AgentVerifier client.
pub struct AgentVerifierClient<D: DataProvider> {
    provider: D,
}

impl<D: DataProvider> AgentVerifierClient<D> {
    pub fn new(provider: D) -> Self { Self { provider } }

    pub fn verify(&self, task_id: FixedBytes<32>, agent_id: FixedBytes<32>, input_hash: FixedBytes<32>, output_hash: FixedBytes<32>, proof: &[u8]) -> Result<(), String> {
        let key = [b"erc8274:agentVerifier:verify:", task_id.as_slice(), agent_id.as_slice(), input_hash.as_slice(), output_hash.as_slice(), proof].concat();
        let data = self.provider.fetch(&key);
        if data.is_empty() { return Err("verification failed".into()); }
        Ok(())
    }
}

/// ERC-8274 Verifiable agent helper.
pub struct AgentVerifiable<D: DataProvider> {
    provider: D,
}

impl<D: DataProvider> AgentVerifiable<D> {
    pub fn new(provider: D) -> Self { Self { provider } }

    pub fn get_trusted_verifier(&self, agent_id: FixedBytes<32>) -> Result<Vec<u8>, String> {
        let key = [b"erc8274:agentVerifiable:getTrustedVerifier:", agent_id.as_slice()].concat();
        let data = self.provider.fetch(&key);
        if data.is_empty() { return Err("no trusted verifier".into()); }
        Ok(data)
    }
}
