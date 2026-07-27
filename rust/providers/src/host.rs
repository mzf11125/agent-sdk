use agent_sdk_core::DataProvider;

/// Host-side data provider: fetches from real external sources (RPC, filesystem, etc.).
pub struct HostProvider;

impl DataProvider for HostProvider {
    fn fetch(&self, _key: &[u8]) -> Vec<u8> {
        // TODO: implement alloy RPC-backed fetch
        unimplemented!("HostProvider not yet wired")
    }
}
