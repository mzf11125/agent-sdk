use crate::DataProvider;
use alloy_primitives::FixedBytes;

/// ERC-8312 BoundedAgentAction client (envelope lifecycle).
///
/// Generic over `DataProvider` — compiles in host (RPC-backed) and guest
/// (preimage-backed) contexts.
pub struct BoundedAgentActionClient<D: DataProvider> {
    provider: D,
}

impl<D: DataProvider> BoundedAgentActionClient<D> {
    pub fn new(provider: D) -> Self {
        Self { provider }
    }

    /// Register a new envelope and return its id.
    pub fn register_envelope(
        &self,
        principal: &[u8],
        capability_root: FixedBytes<32>,
        expires_at: u64,
        init_data: &[u8],
    ) -> Result<FixedBytes<32>, String> {
        let key = [
            b"erc8312:bounded:registerEnvelope:",
            principal,
            capability_root.as_slice(),
            &expires_at.to_be_bytes(),
            init_data,
        ]
        .concat();
        let data = self.provider.fetch(&key);
        if data.len() < 32 {
            return Err("registerEnvelope failed".into());
        }
        let mut id = FixedBytes::ZERO;
        id.as_mut_slice().copy_from_slice(&data[..32]);
        Ok(id)
    }

    /// Advance the cursor and return the new cursor value.
    pub fn advance_cursor(
        &self,
        id: FixedBytes<32>,
        witness: &[u8],
    ) -> Result<FixedBytes<32>, String> {
        let key = [
            b"erc8312:bounded:advanceCursor:",
            id.as_slice(),
            witness,
        ]
        .concat();
        let data = self.provider.fetch(&key);
        if data.len() < 32 {
            return Err("advanceCursor failed".into());
        }
        let mut new_cursor = FixedBytes::ZERO;
        new_cursor.as_mut_slice().copy_from_slice(&data[..32]);
        Ok(new_cursor)
    }

    /// Read the current cursor commitment.
    pub fn get_cursor(&self, id: FixedBytes<32>) -> Result<FixedBytes<32>, String> {
        let key = [b"erc8312:bounded:getCursor:", id.as_slice()].concat();
        let data = self.provider.fetch(&key);
        if data.len() < 32 {
            return Err("getCursor failed".into());
        }
        let mut cursor = FixedBytes::ZERO;
        cursor.as_mut_slice().copy_from_slice(&data[..32]);
        Ok(cursor)
    }

    /// Read the effective lifecycle status (0=None … 5=Expired).
    pub fn get_status(&self, id: FixedBytes<32>) -> Result<u8, String> {
        let key = [b"erc8312:bounded:getStatus:", id.as_slice()].concat();
        let data = self.provider.fetch(&key);
        if data.is_empty() {
            return Err("getStatus failed".into());
        }
        Ok(data[0])
    }

    /// Check if the envelope is active.
    pub fn is_active(&self, id: FixedBytes<32>) -> Result<bool, String> {
        let key = [b"erc8312:bounded:isActive:", id.as_slice()].concat();
        let data = self.provider.fetch(&key);
        Ok(!data.is_empty() && data[0] != 0)
    }
}

/// ERC-8312 BudgetSubstrate client (budget profile reads).
pub struct BudgetSubstrateClient<D: DataProvider> {
    provider: D,
}

impl<D: DataProvider> BudgetSubstrateClient<D> {
    pub fn new(provider: D) -> Self {
        Self { provider }
    }

    /// Return (cap, asset) for the given envelope.
    pub fn bound(&self, id: FixedBytes<32>) -> Result<(u128, Vec<u8>), String> {
        let key = [b"erc8312:budget:bound:", id.as_slice()].concat();
        let data = self.provider.fetch(&key);
        if data.len() < 20 {
            return Err("bound failed".into());
        }
        // First 16 bytes = cap (u128 big-endian), remaining = 20-byte address
        let mut cap_bytes = [0u8; 16];
        cap_bytes.copy_from_slice(&data[..16]);
        let cap = u128::from_be_bytes(cap_bytes);
        let asset = data[16..].to_vec();
        Ok((cap, asset))
    }

    /// Return cumulative spent value.
    pub fn spent(&self, id: FixedBytes<32>) -> Result<u128, String> {
        let key = [b"erc8312:budget:spent:", id.as_slice()].concat();
        let data = self.provider.fetch(&key);
        if data.len() < 16 {
            return Err("spent failed".into());
        }
        let mut buf = [0u8; 16];
        buf.copy_from_slice(&data[..16]);
        Ok(u128::from_be_bytes(buf))
    }

    /// Return remaining headroom.
    pub fn remaining(&self, id: FixedBytes<32>) -> Result<u128, String> {
        let key = [b"erc8312:budget:remaining:", id.as_slice()].concat();
        let data = self.provider.fetch(&key);
        if data.len() < 16 {
            return Err("remaining failed".into());
        }
        let mut buf = [0u8; 16];
        buf.copy_from_slice(&data[..16]);
        Ok(u128::from_be_bytes(buf))
    }
}

/// ERC-8312 ContestableEnvelope client (contestation lifecycle).
pub struct ContestableEnvelopeClient<D: DataProvider> {
    provider: D,
}

impl<D: DataProvider> ContestableEnvelopeClient<D> {
    pub fn new(provider: D) -> Self {
        Self { provider }
    }

    /// Contest an envelope and return the challenger address.
    pub fn contest(
        &self,
        id: FixedBytes<32>,
        evidence: &[u8],
    ) -> Result<Vec<u8>, String> {
        let key = [
            b"erc8312:contestable:contest:",
            id.as_slice(),
            evidence,
        ]
        .concat();
        let data = self.provider.fetch(&key);
        if data.is_empty() {
            return Err("contest failed".into());
        }
        Ok(data)
    }

    /// Resolve a contested envelope.
    pub fn resolve(
        &self,
        id: FixedBytes<32>,
        outcome: u8,
        resolution: &[u8],
    ) -> Result<(), String> {
        let key = [
            b"erc8312:contestable:resolve:",
            id.as_slice(),
            &[outcome],
            resolution,
        ]
        .concat();
        let data = self.provider.fetch(&key);
        if data.is_empty() {
            return Err("resolve failed".into());
        }
        Ok(())
    }
}
