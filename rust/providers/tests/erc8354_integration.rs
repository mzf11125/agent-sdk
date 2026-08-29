//! ERC-8354 integration tests, testkit workflow.
//!
//! Deploy order: verifier, registry, guard (3 addresses).
//! Use `ERC8354_ADDRESSES` with whitespace-separated addresses.
//!
//! These tests exercise the SDK's own client (`agent_sdk_providers::erc8354`),
//! using raw alloy only for setup (registry admin) and the verifier toggle.

use std::sync::OnceLock;
use std::time::{SystemTime, UNIX_EPOCH};

use alloy::network::Ethereum;
use alloy::primitives::{keccak256, Address, FixedBytes, U256};
use alloy::providers::{Provider, ProviderBuilder};
use alloy::signers::local::PrivateKeySigner;
use alloy::signers::Signer;

use agent_sdk_providers::erc8354::{compute_verdict_digest, Erc8354Client, Verdict};

const ANVIL_RPC: &str = "http://127.0.0.1:8545";

fn signer_at(index: usize) -> PrivateKeySigner {
    let path = std::path::Path::new(env!("CARGO_MANIFEST_DIR")).join("../../testkit/.anvil-accounts.json");
    let data = std::fs::read_to_string(&path).expect("cannot read anvil accounts file");
    let parsed: serde_json::Value = serde_json::from_str(&data).expect("invalid JSON");
    parsed["accounts"][index]["privateKey"]
        .as_str()
        .expect("no account at index")
        .parse()
        .expect("invalid private key")
}

fn provider(signer: PrivateKeySigner) -> impl Provider<Ethereum> + Clone {
    ProviderBuilder::new()
        .wallet(signer)
        .connect_http(ANVIL_RPC.parse().expect("invalid url"))
}

fn addresses() -> Vec<Address> {
    std::env::var("ERC8354_ADDRESSES")
        .unwrap_or_default()
        .split_whitespace()
        .filter(|s| !s.is_empty())
        .map(|s| s.parse().expect("invalid address"))
        .collect()
}

async fn register_domain<P: Provider<Ethereum> + Clone>(
    provider: P,
    registry: Address,
    verifier: Address,
    domain_id: FixedBytes<32>,
    root: FixedBytes<32>,
) {
    use alloy::sol;

    sol! {
        #[allow(missing_docs)]
        #[sol(rpc)]
        interface IRegistryAdmin {
            function registerDomain(bytes32 domainId, address registrar, address verifier, bytes32 programKey, uint64 maxRootAge) external;
            function updateRoot(bytes32 domainId, bytes32 newRoot) external;
        }
    }

    let program_key: FixedBytes<32> = keccak256(b"interpreter-vkey");
    let registrar = Address::from([0xa1u8; 20]);

    let registry_admin = IRegistryAdmin::new(registry, provider.clone());
    registry_admin
        .registerDomain(domain_id, registrar, verifier, program_key, 3600)
        .send()
        .await
        .expect("registerDomain")
        .get_receipt()
        .await
        .expect("registerDomain receipt");

    registry_admin
        .updateRoot(domain_id, root)
        .send()
        .await
        .expect("updateRoot")
        .get_receipt()
        .await
        .expect("updateRoot receipt");
}

async fn set_verifier_result<P: Provider<Ethereum> + Clone>(
    provider: P,
    verifier: Address,
    result: bool,
) {
    use alloy::sol;

    sol! {
        #[allow(missing_docs)]
        #[sol(rpc)]
        interface IMockVerifier {
            function setResult(bool r) external;
        }
    }

    IMockVerifier::new(verifier, provider)
        .setResult(result)
        .send()
        .await
        .expect("setResult")
        .get_receipt()
        .await
        .expect("setResult receipt");
}

fn build_verdict(
    executor: Address,
    nullifier: &[u8],
    domain_id: FixedBytes<32>,
    root: FixedBytes<32>,
) -> Verdict {
    Verdict {
        agentId: U256::from(1u64),
        domainId: domain_id,
        policyRoot: root,
        actionCommitment: keccak256(b"action"),
        executor,
        expiry: 4_000_000_000,
        nullifier: keccak256(nullifier),
        decision: 1,
        policyKind: 0,
    }
}

// run_salt makes each test binary invocation use a distinct set of domain
// ids, so rerunning the suite against an already-deployed registry (without
// a fresh anvil deploy) does not collide with domains a prior run left
// registered.
fn run_salt() -> &'static [u8] {
    static SALT: OnceLock<Vec<u8>> = OnceLock::new();
    SALT.get_or_init(|| {
        let nanos = SystemTime::now()
            .duration_since(UNIX_EPOCH)
            .expect("system clock before unix epoch")
            .as_nanos();
        nanos.to_le_bytes().to_vec()
    })
}

// Each test registers its own domain so the tests are independent of each
// other and of run order. A shared fixed domain id collides: the second
// registerDomain reverts with the registry's "already registered" error.
fn domain_for(seed: &[u8]) -> (FixedBytes<32>, FixedBytes<32>) {
    let salt = run_salt();
    let mut domain_seed = Vec::with_capacity(seed.len() + salt.len());
    domain_seed.extend_from_slice(seed);
    domain_seed.extend_from_slice(salt);
    let domain_id = keccak256(&domain_seed);

    let mut root_seed = Vec::with_capacity(seed.len() + salt.len() + 7);
    root_seed.extend_from_slice(b"root-v1");
    root_seed.extend_from_slice(seed);
    root_seed.extend_from_slice(salt);
    (domain_id, keccak256(&root_seed))
}

#[tokio::test]
async fn verify_domain_and_direct_consume() {
    let addrs = addresses();
    assert!(addrs.len() >= 3, "ERC8354_ADDRESSES must hold 3 addresses");
    let verifier = addrs[0];
    let registry = addrs[1];
    let guard = addrs[2];
    let (domain_id, root) = domain_for(b"test-domain-direct");

    let admin_provider = provider(signer_at(0));
    register_domain(admin_provider.clone(), registry, verifier, domain_id, root).await;

    let executor_signer = signer_at(1);
    let executor = executor_signer.address();
    let client = Erc8354Client::new(guard, registry, provider(executor_signer));

    let verdict = build_verdict(executor, b"nf-verify", domain_id, root);

    assert!(client.verify(&verdict, b"proof").await.expect("verify"));
    assert!(client
        .is_root_acceptable(domain_id, root)
        .await
        .expect("isRootAcceptable"));

    let domain = client.domain(domain_id).await.expect("domain");
    assert!(domain.active);
    // No identity registry was declared, so the field decodes as the zero
    // address and the guard's agent-existence check is a no-op.
    assert_eq!(domain.identity_registry, Address::ZERO);

    let current = client.current_root(domain_id).await.expect("currentRoot");
    assert_eq!(current.root, root);

    // Direct consume: msg.sender == executor.
    client.consume(&verdict, b"proof").await.expect("consume");
    assert!(client
        .is_consumed(domain_id, verdict.nullifier)
        .await
        .expect("isConsumed"));
}

#[tokio::test]
async fn consume_relayed_happy_path() {
    let addrs = addresses();
    assert!(addrs.len() >= 3, "ERC8354_ADDRESSES must hold 3 addresses");
    let verifier = addrs[0];
    let registry = addrs[1];
    let guard = addrs[2];
    let (domain_id, root) = domain_for(b"test-domain-relayed");

    register_domain(provider(signer_at(0)), registry, verifier, domain_id, root).await;

    let executor_signer = signer_at(1);
    let executor = executor_signer.address();
    let verdict = build_verdict(executor, b"nf-relayed", domain_id, root);

    let digest = compute_verdict_digest(&verdict, U256::from(31337u64), guard);
    let sig = executor_signer
        .sign_hash(&digest.into())
        .await
        .expect("sign verdict digest");
    let executor_auth = sig.as_bytes(); // 65 bytes, v in {27, 28}

    // The relayer (account 0) submits on the executor's behalf.
    let relay_client = Erc8354Client::new(guard, registry, provider(signer_at(0)));
    relay_client
        .consume_relayed(&verdict, b"proof", &executor_auth)
        .await
        .expect("consumeRelayed");
    assert!(relay_client
        .is_consumed(domain_id, verdict.nullifier)
        .await
        .expect("isConsumed"));
}

#[tokio::test]
async fn rejects_bad_executor_signature() {
    let addrs = addresses();
    assert!(addrs.len() >= 3, "ERC8354_ADDRESSES must hold 3 addresses");
    let verifier = addrs[0];
    let registry = addrs[1];
    let guard = addrs[2];
    let (domain_id, root) = domain_for(b"test-domain-bad-sig");

    register_domain(provider(signer_at(0)), registry, verifier, domain_id, root).await;

    let executor = signer_at(1).address();
    let verdict = build_verdict(executor, b"nf-bad-sig", domain_id, root);

    let relay_client = Erc8354Client::new(guard, registry, provider(signer_at(0)));
    let bad_auth = [0u8; 65];
    assert!(relay_client
        .consume_relayed(&verdict, b"proof", &bad_auth)
        .await
        .is_err());
}

#[tokio::test]
async fn rejects_invalid_proof() {
    let addrs = addresses();
    assert!(addrs.len() >= 3, "ERC8354_ADDRESSES must hold 3 addresses");
    let verifier = addrs[0];
    let registry = addrs[1];
    let guard = addrs[2];
    let (domain_id, root) = domain_for(b"test-domain-invalid-proof");

    let admin_provider = provider(signer_at(0));
    register_domain(admin_provider.clone(), registry, verifier, domain_id, root).await;
    set_verifier_result(admin_provider.clone(), verifier, false).await;

    let executor_signer = signer_at(1);
    let executor = executor_signer.address();
    let verdict = build_verdict(executor, b"nf-invalid-proof", domain_id, root);

    let client = Erc8354Client::new(guard, registry, provider(executor_signer));
    assert!(client.consume(&verdict, b"proof").await.is_err());

    // Restore the shared verifier so later tests are not affected.
    set_verifier_result(admin_provider, verifier, true).await;
}