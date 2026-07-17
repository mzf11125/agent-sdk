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
/// Anvil account #0 private key — pre-funded by anvil.
const ANVIL_KEY_0: &str = "0x48466688bd71f91cd657fe56ecb8af447c8302b70f7f71ac5fc48a66219c188f";

fn registry_address() -> Address {
    std::env::var("ERC8004_ADDRESS")
        .unwrap_or_else(|_| String::from("0x1dee0c0b73345dd324f79c6d170bd244d2075941"))
        .parse()
        .expect("invalid address")
}

#[tokio::test]
async fn register_and_read_owner() {
    let signer: alloy::signers::local::PrivateKeySigner = ANVIL_KEY_0.parse().expect("invalid key");
    let provider = ProviderBuilder::new()
        .wallet(signer)
        .connect_http(ANVIL_RPC.parse().unwrap());

    let contract = IIdentityRegistry::new(registry_address(), provider);
    let tx = contract.register("test-agent".to_string(), vec![]);
    let receipt = tx.send().await.expect("register").get_receipt().await.expect("receipt");
    eprintln!("register tx: {:?}", receipt.transaction_hash);

    let owner = contract.ownerOf(U256::from(1)).call().await.expect("ownerOf");
    eprintln!("ownerOf(1) = {owner:?}");
    assert!(!owner.0.is_zero());
}

