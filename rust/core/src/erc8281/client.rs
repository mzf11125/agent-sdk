use crate::DataProvider;
use alloc::string::String;
use alloy_primitives::FixedBytes;

/// ERC-8281 Observation Commitment Protocol client.
///
/// Generic over `DataProvider` — compiles in host (RPC-backed) and guest
/// (preimage-backed) contexts. Write methods (e.g. `record`) are not
/// included in core — they live in the host-only `providers` crate.
pub struct ObservationCommitmentClient<D: DataProvider> {
    provider: D,
}

impl<D: DataProvider> ObservationCommitmentClient<D> {
    pub fn new(provider: D) -> Self {
        Self { provider }
    }

    /// Verify that a digest was committed by checking the event log.
    /// Key encodes the digest to look up.
    pub fn check_recorded(&self, digest: FixedBytes<32>) -> Result<bool, String> {
        let key = [b"erc8281:recorded:", digest.as_slice()].concat();
        let data = self.provider.fetch(&key);
        Ok(!data.is_empty())
    }
}
