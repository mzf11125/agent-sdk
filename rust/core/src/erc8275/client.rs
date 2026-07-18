use alloy_primitives::FixedBytes;
use crate::DataProvider;

/// ERC-8275 Agent Reputation client.
///
/// Generic over `DataProvider` — compiles in host (RPC-backed) and guest
/// (preimage-backed) contexts. All methods are read-only view calls.
pub struct AgentReputationClient<D: DataProvider> {
    provider: D,
}

impl<D: DataProvider> AgentReputationClient<D> {
    pub fn new(provider: D) -> Self {
        Self { provider }
    }

    /// Fetch the reputation snapshot for an agent.
    /// Returns raw bytes from the provider — decoded by the host layer.
    pub fn get_reputation(&self, agent_id: FixedBytes<32>) -> Result<Vec<u8>, String> {
        let key = [b"erc8275:getReputation:", agent_id.as_slice()].concat();
        let data = self.provider.fetch(&key);
        if data.is_empty() {
            return Err("reputation data not found".into());
        }
        Ok(data)
    }

    /// Fetch the decay weight for an agent.
    pub fn get_decay_weight(&self, agent_id: FixedBytes<32>) -> Result<Vec<u8>, String> {
        let key = [b"erc8275:getDecayWeight:", agent_id.as_slice()].concat();
        let data = self.provider.fetch(&key);
        if data.is_empty() {
            return Err("decay weight not found".into());
        }
        Ok(data)
    }

    /// Verify a settled order outcome proof.
    pub fn verify_outcome(&self, order_id: FixedBytes<32>, proof: &[u8]) -> Result<bool, String> {
        let key = [b"erc8275:verifyOutcome:", order_id.as_slice(), b":", proof].concat();
        let data = self.provider.fetch(&key);
        Ok(!data.is_empty() && data[0] != 0)
    }
}
