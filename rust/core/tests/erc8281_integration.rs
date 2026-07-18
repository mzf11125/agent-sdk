/// ERC-8281 integration tests — testkit workflow.
///
/// ```bash
/// testkit/scripts/start-anvil.sh
/// testkit/scripts/deploy.sh verify/ERC8281 DeployERC8281
/// cargo test --manifest-path rust/core/Cargo.toml --test erc8281_integration -- --nocapture
/// testkit/scripts/stop-anvil.sh
/// ```
use alloy::primitives::{Address, FixedBytes};
use alloy::providers::ProviderBuilder;
use alloy::sol;

sol! {
    #[allow(missing_docs)]
    #[sol(rpc)]
    interface IObservationCommitment {
        function record(bytes32 digest) external;
        event Recorded(bytes32 indexed digest, address indexed committer);
    }
}

const ANVIL_RPC: &str = "http://127.0.0.1:8545";

fn anvil_key() -> String {
    if let Ok(key) = std::env::var("ANVIL_KEY") {
        return key;
    }
    let path =
        std::path::Path::new(env!("CARGO_MANIFEST_DIR")).join("../../testkit/.anvil-accounts.json");
    let data = std::fs::read_to_string(&path).expect("cannot read anvil accounts file");
    let parsed: serde_json::Value = serde_json::from_str(&data).expect("invalid JSON");
    parsed["accounts"][0]["privateKey"]
        .as_str()
        .expect("no account")
        .to_string()
}

fn contract_address() -> Address {
    std::env::var("ERC8281_ADDRESS")
        .unwrap_or_default()
        .parse()
        .unwrap_or(Address::ZERO)
}

#[tokio::test]
async fn record_emits_event() {
    let addr = contract_address();
    if addr.is_zero() {
        eprintln!("SKIP: ERC8281_ADDRESS not set");
        return;
    }

    let key = anvil_key();
    let signer: alloy::signers::local::PrivateKeySigner = key.parse().expect("invalid key");
    let provider = ProviderBuilder::new()
        .wallet(signer)
        .connect_http(ANVIL_RPC.parse().unwrap());

    let contract = IObservationCommitment::new(addr, provider);
    let digest = FixedBytes::<32>::from([0xab; 32]);
    let tx = contract.record(digest);
    let receipt = tx
        .send()
        .await
        .expect("record")
        .get_receipt()
        .await
        .expect("receipt");
    eprintln!("record tx: {:?}", receipt.transaction_hash);

    // Check the Recorded event was emitted
    let logs = receipt.inner.logs();
    assert!(!logs.is_empty(), "should have Recorded event");
    let record_log = logs
        .iter()
        .find(|l| l.address() == addr)
        .expect("Recorded log");
    eprintln!(
        "Recorded event: topic1(digest)={:?}",
        record_log.topics().get(1)
    );
}
