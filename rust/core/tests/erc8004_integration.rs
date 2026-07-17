/// ERC-8004 integration tests — testkit workflow.
/// Prerequisites:
/// ```bash
/// testkit/scripts/start-anvil.sh
/// testkit/scripts/deploy.sh identity/ERC8004 DeployERC8004
/// cargo test --manifest-path rust/core/Cargo.toml --test erc8004_integration -- --nocapture
/// testkit/scripts/stop-anvil.sh
/// ```

use alloy::primitives::{Address, U256};
use alloy::providers::ProviderBuilder;
use alloy::sol;

sol! {
    #[allow(missing_docs)]
    #[sol(rpc)]
    interface IIdentityRegistry {
        function register(string agentURI, (string, bytes)[] metadata) external returns (uint256);
        function ownerOf(uint256 agentId) external view returns (address);
        function tokenURI(uint256 agentId) external view returns (string);
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

fn registry_address() -> Address {
    std::env::var("ERC8004_ADDRESS")
        .unwrap_or_else(|_| String::from("0x0000000000000000000000000000000000000000"))
        .parse()
        .unwrap_or(Address::ZERO)
}

#[tokio::test]
async fn register_and_read_owner() {
    let addr = registry_address();
    if addr.is_zero() {
        eprintln!("SKIP: ERC8004_ADDRESS not set");
        return;
    }
    let key = anvil_key();
    let signer: alloy::signers::local::PrivateKeySigner = key.parse().expect("invalid key");
    let provider = ProviderBuilder::new()
        .wallet(signer)
        .connect_http(ANVIL_RPC.parse().unwrap());

    let contract = IIdentityRegistry::new(addr, provider);
    let tx = contract.register("test-agent".to_string(), vec![]);
    let receipt = tx.send().await.expect("register").get_receipt().await.expect("receipt");
    eprintln!("register tx: {:?}", receipt.transaction_hash);

    let owner = contract.ownerOf(U256::from(1)).call().await.expect("ownerOf");
    eprintln!("ownerOf(1) = {owner:?}");
    assert!(!owner.0.is_zero());
}
