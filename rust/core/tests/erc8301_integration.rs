#![cfg(feature = "std")]
/// ERC-8301 integration tests — testkit workflow.
///
/// ```bash
/// testkit/scripts/start-anvil.sh
/// testkit/scripts/deploy.sh execution/ERC8301 DeployERC8301
/// cargo test --manifest-path rust/core/Cargo.toml --test erc8301_integration -- --nocapture
/// testkit/scripts/stop-anvil.sh
/// ```
use alloy::primitives::{Address, Bytes, FixedBytes, U256};
use alloy::providers::ProviderBuilder;
use alloy::sol;

sol! {
    #[allow(missing_docs)]
    #[sol(rpc)]
    interface IAgentWorkflow {
        function run(bytes32 inputHash, bytes input, uint256 expiresAt) external returns (bytes32);
        function result(bytes32 workflowRunId) external view returns (uint8, bytes32, uint256);
        function getAgentTask(bytes32 taskHash) external view returns (
            uint8 stage, uint256 taskSeq, bytes32 inputHash, bytes input,
            uint256 timestamp, uint256 expiresAt, bytes32[] prevReplyHashes, bytes32 workflowRunId, bool proven
        );
    }
}

const ANVIL_RPC: &str = "http://127.0.0.1:8545";

/// Read anvil account #0 key from the testkit accounts file at runtime.
fn anvil_key() -> String {
    if let Ok(key) = std::env::var("ANVIL_KEY") {
        return key;
    }
    let path =
        std::path::Path::new(env!("CARGO_MANIFEST_DIR")).join("../../testkit/.anvil-accounts.json");
    if let Ok(data) = std::fs::read_to_string(&path) {
        if let Ok(parsed) = serde_json::from_str::<serde_json::Value>(&data) {
            if let Some(key) = parsed["accounts"][0]["privateKey"].as_str() {
                return key.to_string();
            }
        }
    }
    panic!("ANVIL_KEY env var not set and testkit/.anvil-accounts.json not found");
}

fn contract_address() -> Address {
    let addr = std::env::var("ERC8301_ADDRESS").unwrap_or_default();
    if addr.is_empty() {
        return Address::ZERO;
    }
    let addr = if !addr.starts_with("0x") {
        format!("0x{addr}")
    } else {
        addr
    };
    addr.parse().expect("invalid address")
}

#[tokio::test]
async fn run_and_read_result() {
    let addr = contract_address();
    assert!(
        !addr.is_zero(),
        "ERC8301_ADDRESS not set — deploy first via testkit/scripts/deploy.sh execution/ERC8301 DeployERC8301"
    );

    let key = anvil_key();
    let signer: alloy::signers::local::PrivateKeySigner = key.parse().expect("invalid key");
    let provider = ProviderBuilder::new()
        .wallet(signer)
        .connect_http(ANVIL_RPC.parse().unwrap());

    let contract = IAgentWorkflow::new(addr, provider);

    let expires_at = U256::from(2000000000u64);
    let tx = contract.run(FixedBytes::ZERO, Bytes::default(), expires_at);
    let receipt = tx
        .send()
        .await
        .expect("run")
        .get_receipt()
        .await
        .expect("receipt");
    eprintln!("run tx: {:?}", receipt.transaction_hash);

    // run() returns bytes32 — decode from the tx return data via a helper:
    // for now, just verify the transaction succeeded and result() works with a zero runId
    let r = contract
        .result(FixedBytes::ZERO)
        .call()
        .await
        .expect("result");
    eprintln!("result(zero) → status = {:?}", r._0);
}
