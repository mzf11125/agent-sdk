use crate::DataProvider;
use alloc::string::String;
use alloc::vec::Vec;
use alloy_primitives::FixedBytes;

/// ERC-8323 Source-Token Agent Binding client.
///
/// Generic over `DataProvider` — compiles in host (RPC-backed) and guest
/// (preimage-backed) contexts.
pub struct SourceBindingClient<D: DataProvider> {
    provider: D,
}

impl<D: DataProvider> SourceBindingClient<D> {
    pub fn new(provider: D) -> Self {
        Self { provider }
    }

    pub fn get_source_nft(&self, agent_id: FixedBytes<32>) -> Result<Vec<u8>, String> {
        let key = [b"erc8323:getSourceNFT:", agent_id.as_slice()].concat();
        let data = self.provider.fetch(&key);
        if data.is_empty() {
            return Err("source NFT not found".into());
        }
        Ok(data)
    }

    pub fn has_source_nft(&self, agent_id: FixedBytes<32>) -> Result<bool, String> {
        let key = [b"erc8323:hasSourceNFT:", agent_id.as_slice()].concat();
        let data = self.provider.fetch(&key);
        Ok(!data.is_empty() && data[0] != 0)
    }

    pub fn is_source_nft_ownership_valid(&self, agent_id: FixedBytes<32>) -> Result<bool, String> {
        let key = [b"erc8323:isSourceNFTOwnershipValid:", agent_id.as_slice()].concat();
        let data = self.provider.fetch(&key);
        Ok(!data.is_empty() && data[0] != 0)
    }
}
