/// ERC-8323 integration tests — testkit workflow.

use alloy::primitives::Address;
use alloy::providers::ProviderBuilder;
use alloy::sol;

sol! {
    #[allow(missing_docs)]
    #[sol(rpc)]
    interface IAgentSourceBinding {
        function boundCollection() external view returns (address);
        function getSourceNFT(uint256 agentId) external view returns (address, uint256);
        function hasSourceNFT(uint256 agentId) external view returns (bool);
        function isSourceNFTOwnershipValid(uint256 agentId) external view returns (bool);
        function register(uint256 sourceTokenId) external returns (uint256 agentId);
    }
}

const ANVIL_RPC: &str = "http://127.0.0.1:8545";

fn anvil_key() -> String {
    if let Ok(key) = std::env::var("ANVIL_KEY") {
        return key;
    }
    let path = std::path::Path::new(env!("CARGO_MANIFEST_DIR"))
        .join("../../testkit/.anvil-accounts.json");
    let data = std::fs::read_to_string(&path).expect("cannot read anvil accounts file");
    let parsed: serde_json::Value = serde_json::from_str(&data).expect("invalid JSON");
    parsed["accounts"][0]["privateKey"].as_str().expect("no account").to_string()
}

fn contract_address() -> Address {
    std::env::var("ERC8323_ADDRESS")
        .unwrap_or_default()
        .parse()
        .unwrap_or(Address::ZERO)
}

#[tokio::test]
async fn bound_collection_reads() {
    let addr = contract_address();
    if addr.is_zero() {
        eprintln!("SKIP: ERC8323_ADDRESS not set");
        return;
    }

    let key = anvil_key();
    let signer: alloy::signers::local::PrivateKeySigner = key.parse().expect("invalid key");
    let provider = ProviderBuilder::new()
        .wallet(signer)
        .connect_http(ANVIL_RPC.parse().unwrap());

    let contract = IAgentSourceBinding::new(addr, provider);
    let collection = contract.boundCollection().call().await.expect("boundCollection");
    eprintln!("boundCollection = {collection:?}");
}
