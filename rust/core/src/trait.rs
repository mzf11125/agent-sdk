/// Data that a host or guest provider can resolve.
/// Op-succinct style: preimage key → data bytes.
pub trait DataProvider {
    /// Fetch raw bytes for a given preimage key.
    /// Host: goes to RPC / filesystem / network.
    /// Guest: reads from the zkVM preimage oracle cache.
    fn fetch(&self, key: &[u8]) -> Vec<u8>;
}
