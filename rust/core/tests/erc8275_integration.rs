/// ERC-8275 integration tests — testkit workflow.
///
/// ```bash
/// testkit/scripts/start-anvil.sh
/// testkit/scripts/deploy.sh reputation/ERC8275 DeployERC8275
/// cargo test --manifest-path rust/core/Cargo.toml --test erc8275_integration -- --nocapture
/// testkit/scripts/stop-anvil.sh
/// ```

use alloy::primitives::{Address, FixedBytes};
use alloy::providers::ProviderBuilder;
use alloy::sol;

sol! {
    #[allow(missing_docs)]
    #[sol(rpc)]
    interface IAgentReputation {
        function getReputation(bytes32 agentId) external view returns (
            uint64 completedOrders,
            uint64 disputedOrders,
            uint64 totalVolume,
            uint64 lastActiveAt,
            uint16 score
        );
        function getDecayWeight(bytes32 agentId) external view returns (uint16 weight);
        function verifyOutcome(bytes32 orderId, bytes proof) external view returns (bool valid);
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
    std::env::var("ERC8275_ADDRESS")
        .unwrap_or_default()
        .parse()
        .unwrap_or(Address::ZERO)
}

#[tokio::test]
async fn get_reputation_reads_default() {
    let addr = contract_address();
    if addr.is_zero() {
        eprintln!("SKIP: ERC8275_ADDRESS not set");
        return;
    }

    let key = anvil_key();
    let signer: alloy::signers::local::PrivateKeySigner = key.parse().expect("invalid key");
    let provider = ProviderBuilder::new()
        .wallet(signer)
        .connect_http(ANVIL_RPC.parse().unwrap());

    let contract = IAgentReputation::new(addr, provider);
    let agent_id = FixedBytes::ZERO;
    let rep = contract.getReputation(agent_id).call().await.expect("getReputation");
    eprintln!("getReputation(zero) = completed={}, disputed={}, volume={}, lastActive={}, score={}",
        rep.completedOrders, rep.disputedOrders, rep.totalVolume, rep.lastActiveAt, rep.score);

    let weight = contract.getDecayWeight(agent_id).call().await.expect("getDecayWeight");
    eprintln!("getDecayWeight(zero) = {weight}");
}
