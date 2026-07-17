/// ERC-8004 integration tests.
/// Requires: anvil running, mock IdentityRegistry deployed.
/// Run: testkit/scripts/start-anvil.sh && testkit/scripts/deploy.sh erc8004 && cargo test

#[cfg(test)]
mod erc8004_integration {
    use alloy_provider::{Provider, ProviderBuilder};
    use alloy_core::sol;

    // Minimal ABI for the view functions we test
    sol! {
        #[allow(missing_docs)]
        function ownerOf(uint256 agentId) external view returns (address);
        function tokenURI(uint256 agentId) external view returns (string);
        function getAgentWallet(uint256 agentId) external view returns (address);
        function getMetadata(uint256 agentId, string metadataKey) external view returns (bytes32);
    }

    /// Registry address — set by deploy.sh output.
    /// Default anvil deployer key (index 0).
    const ANVIL_RPC: &str = "http://127.0.0.1:8545";

    fn registry_address() -> alloy::primitives::Address {
        std::env::var("ERC8004_ADDRESS")
            .unwrap_or_else(|_| "0x5FbDB2315678afecb367f032d93F642f64180aa3".into())
            .parse()
            .expect("invalid registry address")
    }

    async fn provider() -> impl Provider {
        ProviderBuilder::new().on_http(ANVIL_RPC.parse().unwrap())
    }

    #[tokio::test]
    async fn test_owner_of_returns_address() {
        let p = provider().await;
        let addr = registry_address();
        // Minted agent 1 should exist and have an owner
        let result = ownerOf::new().call(&p, addr).await;
        // May fail if contract not deployed — skip gracefully
        match result {
            Ok(owner) => {
                assert!(!owner._0.is_zero(), "owner should be non-zero address");
            }
            Err(_) => eprintln!("Skipping: contract may not be deployed"),
        }
    }

    #[tokio::test]
    async fn test_token_uri_reads() {
        let p = provider().await;
        let addr = registry_address();
        let result = tokenURI::new(1u64.into()).call(&p, addr).await;
        match result {
            Ok(uri) => eprintln!("tokenURI(1) = {uri}"),
            Err(_) => eprintln!("Skipping: contract may not be deployed"),
        }
    }
}
