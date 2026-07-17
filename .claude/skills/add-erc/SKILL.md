---
name: add-erc
description: Generate TypeScript, Python, and Rust SDK clients for an ERC defined in agent-ercs that doesn't have SDK support yet, including a recompute-to-verify classification, pure recompute functions (Layer 2), and tests run against golden conformance vectors (offline) then a local anvil deployment.
---

# Add ERC

Generate off-chain client support for one ERC that `agent-sdk` doesn't yet implement.

The output has two layers:
- **Layer 1** — contract wrappers: `client.ts` / `client.py` that call the on-chain contract via an RPC node (read/send/verify).
- **Layer 2** — pure recompute: `recompute.ts` / `recompute.py` with stateless functions that reproduce the ERC's deterministic computations from public inputs, verified against golden conformance vectors without requiring any blockchain node.

## Process

1. **Determine which ERC.** If not specified, ask which ERC (by number, e.g. "ERC-8301") to add. If the `agent-ercs` submodule should be read from something other than its currently checked-out `main`, ask for the branch, tag, commit, or local path to use, and check that out in `agent-ercs/` before continuing (default: whatever `main` currently points to).

2. **Read the spec.** Read the ERC's interface file(s) and `README.md` under `agent-ercs/contracts/<category>/<ERCXXXX>/`.

3. **Classify recompute-to-verify capability.** An ERC can make more than one distinct claim — classify each central claim separately rather than forcing one verdict for the whole ERC. For each claim, determine: can a caller independently obtain the same authoritative answer without trusting whoever originally submitted it — either by recomputing off-chain from public data, or by calling a deterministic, callable-by-anyone, immutable on-chain check themselves — or does the guarantee terminate at trusting a specific deployment's unfixed convention (a signing scheme, a derived value) with no way to even know what to check? Write out the reasoning — this is the part of the job that isn't mechanical.
   - See `typescript/src/identity/ERC8004/README.md` for a clean NOT-verifiable case: the interface leaves a signing convention completely unfixed, so there's nothing a generic SDK can check at all.
   - See `typescript/src/verify/ERC8274/README.md` for a split verdict on one ERC: the core validity check *is* recompute-to-verify (anyone can call the deployed, immutable verifier contract themselves and get an authoritative answer — that's a real instance of recompute-to-verify, not "just asking the same contract again"), while a separate derived value (an audit-trail digest) is NOT, because one of its inputs isn't exposed anywhere in the interface. Don't let one claim's verdict force the other's.

4. **Identify pure recompute functions (Layer 2).** For each claim the ERC makes, identify whether it involves a deterministic computation that can be reproduced off-chain from public inputs as a pure mathematical function. This is SEPARATE from the recompute-to-verify classification in step 3:
   - **recompute-to-verify** (step 3): can a caller *independently verify* a claim by calling a contract or recomputing it? The verdict can be YES, NO, or SPLIT.
   - **pure recompute** (this step): is there a *mathematical function* that anyone can compute from public inputs to derive the expected output? The answer is a list of functions, regardless of whether the overall verdict is verifiable.
   
   An ERC can have pure recompute functions even if its overall verdict is NOT verifiable. For example, ERC-8004 is NOT recompute-to-verify (its signing convention is unfixed), but `agentId = bytes32(uint256(registryId))` is a pure recompute function — anyone can compute it from the registry ID.

   Examples of pure recompute from existing conformance vectors:
   - ERC-8004: `agentId = bytes32(uint256(registryId))` — left-padded, no hash (`step: "8004/agent-id"`)
   - ERC-8299: `raw_input_hash = keccak256(raw_user_input)` (`step: "wyriwe/raw"`)
   - ERC-8299: `sanitization_pipeline_hash = keccak256(utf8(cid) \|\| raw_input_hash)` (`step: "wyriwe/pipeline"`)
   - ERC-8301: `taskHash = keccak256(abi.encode(stage, taskSeq, inputHash, timestamp, expiresAt, innerHash, workflowRunId))` (`step: "8301/task-hash"`)
   - Scope-contestation: `scopeRoot = keccak256(abi.encode(merkleRoot, count))` (`step: "scope/binding"`)
   - ENS: `namehash(name)` per EIP-137 (`step: "ens/namehash"`)
   - ERC-8275: `winRate = gated_wins / (gated_wins + gated_losses)` (`step: "8275/reputation"`)
   - ERC-8203: `verdictHash = keccak256(abi.encode(jobId, keccak256(utf8(resultText))))` (`step: "8203/settlement-proof"`)
   - Scope-contestation bond: "standing" computed from public bond state (`step: "scope/bond-standing"`)
   - Scope-contestation: resolution root from sorted votes (`step: "scope/value-fidelity"`)
   - Scope-contestation: materiality contest check (`step: "scope/contest-verify"`)
   - ERC-8312: cap conservation invariant (`step: "8312/cap-conservation"`)

   Read the conformance vectors at `/Users/shakku/code/recompute-kit/conformance/agent-flow.vectors.json` — they are the source of truth for what pure recompute functions exist. Each vector has a `step` field identifying the computation and `inputs`/`expected` for testing. Cross-reference steps with the ERC's claims.

   If pure recompute functions exist, add them to the proposed API in step 5. Propose one function per distinct computation, each documented with the ERC section it comes from. Don't bundle unrelated computations into one function.

5. **Propose the client API (both layers).** Based on the interface, the classification (step 3), and the identified pure recompute functions (step 4), propose the method list for both languages — including whether a `verify()` method is warranted for Layer 1 and which recompute functions are exposed for Layer 2 — and get the user's approval before writing any code.

6. **Write the per-ERC READMEs.** Under both `typescript/src/<category>/<ERCXXXX>/README.md` and `python/src/agent_sdk/<category>/<ercxxxx>/README.md` (lowercase ERC segment for Python), record the verdict, its rationale, the API summary (both layers), and whether pure recompute functions exist.

7. **Implement Layer 2 (pure recompute) — offline, no contract needed.** For each language, generate:

   **a) `recompute.ts` / `recompute.py`:**
   - Pure stateless functions, each documented with the ERC section it comes from.
   - TypeScript: import each utility as a named import from `'viem'` (top-level package), never from deep paths like `'viem/utils'`. Use these and similar pure utilities: `keccak256`, `encodeAbiParameters`, `concat`, `stringToHex`, `toHex`, `namehash`, `bytesToHex`, `hexToBytes`. NO dependency on contract clients, ABIs, viem chains, or deployed addresses.
   - Python: import only from `eth_utils` or `web3.py` equivalents (e.g. `Web3.keccak`, `eth_utils.keccak`, `eth_utils.to_bytes`, `eth_utils.to_hex`). Use `eth_utils.keccak` for hash computations — it's the standard pattern in this project. Avoid `eth_hash` unless it is a declared dependency in `pyproject.toml`. NO dependency on `web3.eth.contract` or RPC. If a computation is trivial (e.g. integer-to-bytes32 padding), pure Python without dependencies is acceptable.
   - Each function signature: clear input types (native JS/Python types, not ABI-encoded blobs) -> deterministic output. Accept `Hex` (0x-prefixed strings) for hash inputs, `number`/`bigint` for integers, `boolean` for booleans, strings for text.
   - Naming convention: TypeScript uses camelCase parameter names (matching Solidity event/function parameter style); Python uses snake_case. Export all functions as named exports. Do not wrap them in a class — they are pure module-level functions.
   - Include a JSDoc/docstring comment on each function describing what it computes and referencing the ERC section (e.g. `// ERC-8299 §45: raw_input_hash = keccak256(raw_user_input)`).

   **b) `recompute.test.ts` / `test_recompute.py`:**
   - Read golden vectors from `recompute-kit/conformance/agent-flow.vectors.json`. Use a relative path from the test file — either resolve via a symlink or compute relative to the repo root. The path to look for: `../../../../../recompute-kit/conformance/agent-flow.vectors.json` from the TypeScript test directory, or an equivalent relative path from Python.
   - For each vector whose `step` matches one of the recompute functions in this ERC, assert `recomputeFn(inputs) === expected`.
   - Test each conformance-file vector individually, not all vectors inside one test function. TypeScript: one `it(...)` per vector. Python: one parametrized test method per vector using `@pytest.mark.parametrize`.
   - Also write a self-contained inline golden vector test (duplicate the vector's inputs and expected values directly in the test file) so tests work even when recompute-kit is not present on disk or the vectors file is unreachable. This inline test is the primary assertion; the file-based reader is a secondary conformance check that vectors haven't drifted.
   - When recompute-kit vectors are not found on disk, use a guard clause / early return pattern: check existence at the top and return/skip immediately, rather than wrapping the bulk of the test logic inside an `if (vectors.length)` branch.
   - Generate an empty `__init__.py` in any new Python test directory (e.g. `python/tests/<category>/<ercxxxx>/__init__.py`) to match sibling ERC test directory patterns.
   - Tests must run without any deployed contract, anvil node, or RPC connection — pure function calls only.
   - TypeScript: use `vitest`. Python: use `pytest`.
   - Cover both happy path (inputs produce expected output) and edge cases where applicable. Edge-case checklist by operation type:

     | Operation | Edge cases to test |
     |---|---|
     | `toHex` / zero-padding | zero, one, max value within bytes32 (2^248-1) |
     | `keccak256` | empty input (`0x`), known short input (1 byte), longer input (32+ bytes), verify against a known golden hash |
     | `abi.encode` / `encodeAbiParameters` | single-argument, multi-argument, each type combination (uint, bytes32, address, tuple/bool — note Solidity encodes booleans as uint8) |
     | `concat` | empty first segment, empty second segment, both empty, multi-segment |
     | `namehash` | single-label (TLD), two-label (`name.eth`), three-label (`sub.name.eth`) |
     | Boolean / invariant checks | true/truthy and false/falsy cases |

   **c) Post-generation cleanup:** After writing all recompute files, scan each file for unused imports and remove them before finalizing. Generated code should be clean on arrival — no dangling imports from copy-paste or removed computations.

   **d) Rust `recompute.rs` + inline tests:**
   - Generate `rust/core/src/<erc_lowercase>/recompute.rs` (lowercase ERC segment, e.g. `erc8004`).
   - **Imports (module-level):** Only what the public functions need. Use `alloy_primitives::{keccak256, FixedBytes}` for hashes and bytes32 types (use `FixedBytes<32>` explicitly, not the `B256` alias). For ABI encoding, use `alloy_core::sol_types::SolValue` — this requires `sol-types` feature enabled on `alloy-core` in `Cargo.toml`. For integer-to-bytes32 padding, use `u64::to_be_bytes()` into a `[0u8; 32]` buffer.
   - **Test imports inside `#[cfg(test)]`:** Put `use alloy_primitives::hex;` inside the test module (not module-level, or it triggers "unused import" in non-test builds). Use `hex!("...")` macro for inline golden vector hex literals.
   - Pure `no_std` functions — no networking, no alloc unless needed. Each function: doc comment with ERC section, clean input types (`u64`, `FixedBytes<32>`, `&str`), `FixedBytes<32>` output (not `B256`).
   - Write inline `#[cfg(test)] mod tests { ... }` directly in `recompute.rs`. Inline golden vectors (duplicate expected values from recompute-kit) are the primary test. Each golden vector gets its own `#[test]` function. Edge cases: zero, empty input, different-inputs-different-hash, max values.
   - Tests run with `cargo test -p agent-sdk-core` — no anvil or network needed.
   - **Concrete patterns by operation type:**

     | Operation | Rust pattern |
     |-----------|-------------|
     | `bytes32(uint256(x))` left-pad | `u64_to_bytes32`: `let mut buf = [0u8; 32]; buf[24..].copy_from_slice(&x.to_be_bytes()); FixedBytes::new(buf)` |
     | `keccak256(bytes)` | `keccak256(&bytes)` — returns `FixedBytes<32>` |
     | `keccak256(utf8(str))` | `keccak256(s.as_bytes())` |
     | `concat(a, b)` then keccak | `let mut v = Vec::new(); v.extend_from_slice(a); v.extend_from_slice(b); keccak256(&v)` |
     | `abi.encode(type1, type2)` | `(val1, val2).abi_encode()` — requires `SolValue` in scope |

8. **Run the recompute tests separately first.** Before touching any contract infrastructure, run `npx vitest run <path-to-recompute.test.ts>` (TS), `pytest <path-to-test_recompute.py>` (Python), and `cargo test -p agent-sdk-core` (Rust). These must pass without any blockchain node. If they fail, debug the recompute implementation before proceeding to Layer 1.

9. **Implement Layer 1 (contract wrappers).**
   - Hand-write the ABI fragment for the functions/events the client uses, matching the interface exactly (no dynamic codegen from build artifacts).
   - **TypeScript and Python:**
     * Before writing the client, check existing ERC clients in the SAME category (identity, verify, etc.) for wallet and constructor patterns and match them. Specifically:
       * WalletClient: use the `createWalletClient({ chain: foundry, transport, account })` pattern (see ERC-8004, ERC-8274) — don't invent `{ account }` plain objects or other ad-hoc patterns.
       * Constructor: match the existing `(config, account)` signature pattern.
     * Implement the client, following the shape and conventions of `typescript/src/identity/ERC8004/client.ts` / `python/src/agent_sdk/identity/erc8004/client.py` for a single-contract ERC, or `typescript/src/verify/ERC8274/*Client.ts` / `python/src/agent_sdk/verify/erc8274/client.py` for an ERC that's really several interfaces meant to be deployed as separate, cross-referencing contracts — don't force a multi-contract ERC into one client class. For a claim classified as recompute-to-verify (a deterministic, callable-by-anyone check), expose it as a read-only simulated call/`.call()` rather than a broadcast transaction — nobody should need to spend gas or hold a funded key just to check something.
   - **Rust `client.rs`:**
     * Generate `rust/core/src/<erc_lowercase>/client.rs`. Define a generic struct `Client<D: DataProvider>` — no direct alloy transport dependency, no `tokio`. The `DataProvider` trait from `rust/core/src/trait.rs` supplies external data so the client compiles both in host (RPC-backed) and guest (preimage-backed) contexts.
     * Read-only contract calls: methods return `Result<T, ClientError>` where fetching happens via `self.provider.fetch(key)`. No `send`/broadcast in core — write methods belong in the `providers` crate or a separate host-only layer.
     * For ERCs with no contract interface (recompute-only), skip Rust Layer 1 entirely.
   - Generate `rust/core/src/<erc_lowercase>/mod.rs` that re-exports both `recompute` and `client` modules.

   - **Rust integration tests** (for ERCs with a contract interface):
     * Create `rust/core/tests/<erc_lowercase>_integration.rs`. This test uses the same testkit workflow as TS/Python: anvil running, contract deployed via Foundry deploy script, then calls the contract through the generated Rust client.
     * The integration test should use `alloy-provider` + `alloy-transport-http` (or the full `alloy` meta-crate) for a proper RPC client with signing, nonce management, and gas estimation. Raw `reqwest` + JSON-RPC works for `eth_call` (read-only) but NOT for `eth_sendTransaction` (writes) which need signing.
     * Dev-dependencies needed in `rust/core/Cargo.toml`: `alloy = { version = "0.11", features = ["full"] }` or compatible alloy version, `tokio` (with `rt` and `macros` features). If version conflicts arise with the main alloy-core/alloy-primitives deps, prefer a consistent alloy version across all deps.
     * Define ABI inline using `alloy_sol_types::sol!` macro (dev-dep). Create a provider with `ProviderBuilder::new().on_http("http://127.0.0.1:8545").wallet(signer)`. Call contract via `contract_instance.method(args).call().await`.
     * Tests match the same flow as TS/Python: deploy → register/setup → read → assert. Deployer account and RPC URL should match anvil defaults.
   - If the ERC needs a contract to deploy for testing and `agent-ercs` has no base implementation yet, write a minimal reference implementation under `testkit/contracts/mocks/<category>/<ERCXXXX>/` (one file per contract if the ERC needs more than one), clearly commented as local-testing-only (see `MockIdentityRegistry.sol` for a single-contract pattern, `MockProofVerifier.sol`/`MockAgentVerifier.sol`/`MockAgentVerifiable.sol` for a multi-contract one), plus a Foundry unit test for it/them under `testkit/test/<category>/<ERCXXXX>/`.
   - Write `testkit/script/<category>/<ERCXXXX>/Deploy<ERCXXXX>.s.sol` (file basename must match its contract name, e.g. `DeployERC8301.s.sol` containing `contract DeployERC8301` — Foundry keys broadcast artifacts by script basename only, so reusing a generic name like `Deploy.s.sol` across ERCs would collide). If the ERC needs several wired-together contracts, deploy all of them in one script (constructor-inject each into the next) — `testkit/scripts/deploy.sh` prints one address per line in the order each was deployed; use `deployContracts()`/`deploy_contracts()` (plural, returning the full list) instead of the single-address `deployContract()`/`deploy_contract()` to receive all of them (see `typescript/test/verify/ERC8274/erc.test.ts` / `python/tests/verify/erc8274/test_erc.py`).
   - Write tests for both languages that deploy via `testkit/scripts/deploy.sh` (see `typescript/test/identity/ERC8004/erc.test.ts` and `python/tests/identity/erc8004/test_erc.py` for the single-contract wiring pattern, or the ERC-8274 test files above for multi-contract) and call the client's methods. For any claim classified as recompute-to-verify, also test that the check rejects tampered/incorrect data (a bad proof, a bad signature) — some checks reject by returning a falsy result rather than reverting; assert whichever the contract actually does, don't assume a revert.
   - If double-checking a byte-encoding assumption against the actual Solidity (e.g. whether a hash was built with `abi.encode` vs `abi.encodePacked` — they differ for `bool` and other non-32-byte-aligned types), verify it against the real contract rather than assuming the two are interchangeable.

10. **Wire up package exports.** After both layers are implemented, register the new ERC in the package's public API so consumers can import it.

    **TypeScript:**
    - Create a barrel `index.ts` in the ERC directory (`typescript/src/<category>/<ERCXXXX>/index.ts`) that re-exports the client class(es), recompute functions, and any public types. Use named re-exports. See `typescript/src/identity/ERC8004/index.ts` for the single-client pattern or `typescript/src/execution/ERC8301/index.ts` for a multi-client pattern.
    - Add subpath exports to `typescript/package.json` in the `"exports"` field, in the same alphabetical order as existing entries:
      * Full-entry: `"./<category>/<ERCXXXX>": { "types": "./dist/<category>/<ERCXXXX>/index.d.ts", "default": "./dist/<category>/<ERCXXXX>/index.js" }`
      * Recompute-only (if recompute functions exist): `"./<category>/<ERCXXXX>/recompute": { "types": "./dist/<category>/<ERCXXXX>/recompute.d.ts", "default": "./dist/<category>/<ERCXXXX>/recompute.js" }`

    **Python:**
    - Populate the ERC module's `__init__.py` (`python/src/agent_sdk/<category>/<ercxxxx>/__init__.py`) with proper named imports and `__all__`. See `python/src/agent_sdk/identity/erc8004/__init__.py` for a single-client pattern or `python/src/agent_sdk/execution/erc8301/__init__.py` for a multi-client pattern. Do not leave it empty — it must export all public classes and functions.
    - Update the category-level `__init__.py` (`python/src/agent_sdk/<category>/__init__.py`) with a docstring-only or import-based entry if it doesn't reference the new ERC yet.

    **Rust:**
    - Add `pub mod <erc_lowercase>;` to `rust/core/src/lib.rs` to register the new ERC module.
    - If the ERC introduces a new category that doesn't yet exist in `rust/core/src/`, create an empty category-level `mod.rs` and add the `pub mod` line for it from `lib.rs`.

11. **Update root README.** Append the new ERC to the "Supported ERCs" table in the repo root `README.md`. Match the existing row format: ERC name with link to agent-ercs, category, Contract Calls column (list client classes or `—`), Recompute column (list recompute functions or `—`). Insert in alphabetical order within its category.

12. **Run every new test to green** — first the recompute tests (offline, no anvil), then the full integration tests:
    - `npx vitest run <recompute test path>` (TS Layer 2 — offline, no anvil needed)
    - `pytest <recompute test path>` (Python Layer 2 — offline, no anvil needed)
    - `cargo test -p agent-sdk-core` (Rust Layer 2 — offline, no anvil needed)
    - Start anvil (`testkit/scripts/start-anvil.sh`), then:
    - `npx vitest run` (TS Layer 1 integration tests + any other TS tests)
    - `pytest` (Python Layer 1 integration tests + any other Python tests)
    - `cargo test -p agent-sdk-core` (Rust Layer 1 integration tests, if any)
    - If more than one ERC's tests now exist for a language, run that language's *full* suite, not just the new files in isolation — shared test infrastructure (one anvil instance and deployer account across all ERCs) can only reveal cross-file issues, such as a nonce race from parallel test execution, when everything runs together.
    - Stop anvil (`testkit/scripts/stop-anvil.sh`) when done.

## What gets committed

Only the final READMEs, recompute layer (recompute.ts/md, recompute.py, recompute.rs, recompute tests), client code (client.ts, client.py, client.rs, contract wrappers tests), barrel files (index.ts, __init__.py, mod.rs), package.json exports, and Rust module registrations. Discussion during the early steps is scratch and is not committed.
