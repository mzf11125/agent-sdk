use agent_sdk_core::DataProvider;

/// Guest-side data provider: reads from a zkVM preimage oracle cache.
/// No network — all data was committed to the cache before the guest ran.
pub struct GuestProvider;

impl DataProvider for GuestProvider {
    fn fetch(&self, _key: &[u8]) -> Vec<u8> {
        // TODO: wire to zkVM precompile (e.g. SP1 preimage oracle)
        unimplemented!("GuestProvider not yet wired")
    }
}
