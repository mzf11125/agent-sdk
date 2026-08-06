# Go SDK Design

## Overview

Add Go as the 4th language to `agent-sdk`, alongside TypeScript, Python, and Rust.
The Go SDK provides the same dual-layer architecture: contract wrappers (Layer 1) and pure recompute functions (Layer 2), tested against golden conformance vectors and a local anvil node via testkit.

## Directory Structure

```
go/
├── go.mod                              # module github.com/trustless-ai/agent-sdk/go
├── go.sum
├── <category>/<erc_lowercase>/         # one package per ERC
│   ├── client.go                       # Layer 1: RPC contract wrapper
│   ├── recompute.go                    # Layer 2: pure stateless recompute + inline tests
│   ├── recompute_test.go               # separate test file when inline tests get too large
│   ├── abi.go                          # ABI fragments for the contract
│   ├── types.go                        # shared types
│   └── README.md                       # recompute-to-verify verdict + reasoning + API summary
├── test/                               # integration tests (require anvil)
│   └── <erc_lowercase>_integration_test.go
└── README.md                           # Go SDK overview
```

Aligns with TS (`typescript/src/<category>/<ERCXXXX>/`) and Python (`python/src/agent_sdk/<category>/<ercxxxx>/`) conventions: category grouping, lowercase ERC segment, flat per-ERC packages with `client.go` and `recompute.go` at the same level. Integration tests that need a running anvil live separately under `go/test/`, mirroring `rust/core/tests/`.

## Technology Stack

| Concern | Choice | Rationale |
|---------|--------|-----------|
| Go version | `go 1.21` | Compatible with go-ethereum v1.14.x; broad consumer reach |
| keccak256 | `go-ethereum/crypto.Keccak256Hash()` | Industry standard, same role as viem's `keccak256` in TS |
| bytes32 type | `go-ethereum/common.Hash` | Canonical `[32]byte` with hex helpers |
| RPC client | `go-ethereum/ethclient.Client` | Standard Ethereum Go client |
| ABI encoding | `go-ethereum/accounts/abi` | `abi.Arguments.Pack()` for `abi.encode` equivalents |
| Testing | stdlib `testing` | Zero-dependency, Go convention |
| Integer arithmetic | basis-points (10000 = 1.0) | Identical across all 4 languages by construction |

## Dual-Layer Architecture

### Layer 2 — Pure Recompute

Stateless functions that reproduce ERC-defined deterministic computations from public inputs. No RPC, no contract addresses, no network.

- Functions named with PascalCase Go convention (exported): `ComputeWinRate`, `ComputeAgentId`
- Each function documented with ERC section reference
- Return `(result, error)` — zero inputs produce an error, never a panic
- Inline golden-vector tests with `func TestGoldenXxx(t *testing.T)` alongside each function
- Edge-case checklist (same as TS/Python/Rust): zero, empty, max values, different-inputs-different-hash

Operation-to-Go mapping:

| Operation | Go pattern |
|-----------|-----------|
| `bytes32(uint256(x))` left-pad | `var b common.Hash; binary.BigEndian.PutUint64(b[24:], x)` |
| `keccak256(bytes)` | `crypto.Keccak256Hash(data)` |
| `keccak256(utf8(str))` | `crypto.Keccak256Hash([]byte(s))` |
| `concat(a, b)` then keccak | `append(a, b...)` then `crypto.Keccak256Hash(combined)` |
| `abi.encode(...)` | `abi.Arguments{{Type: t1}, {Type: t2}}.Pack(args...)` |
| integer arithmetic | standard Go `uint64`/`big.Int` operators, basis-points formula |

### Layer 1 — Contract Wrappers

- Concrete `struct` with `*ethclient.Client` and `common.Address` fields (no DataProvider generics — Go doesn't target zkVMs)
- Read methods return `(T, error)`, write/send methods return `(*types.Transaction, error)`
- Constructor follows TS/Python pattern: `NewXxxClient(rpc *ethclient.Client, addr common.Address) *XxxClient`
- ABI kept in `abi.go` as a string constant, parsed once at init or lazily

## Testing Strategy

### Recompute Tests (offline, Layer 2)

- Inline in `recompute.go` or adjacent `recompute_test.go`
- Primary: inline golden vectors (duplicated from recompute-kit)
- Secondary: read from `recompute-kit/conformance/agent-flow.vectors.json` when available
- Run: `go test ./go/<category>/<erc_lowercase>/...`
- No network, no anvil needed

### Integration Tests (anvil, Layer 1)

- Live under `go/test/<erc_lowercase>_integration_test.go`
- Build tag or env-var guard to skip when anvil isn't running
- Workflow: start anvil → forge deploy → run Go integration test → stop anvil
- Contract address from env var `ERCXXXX_ADDRESS`, signer key from `testkit/.anvil-accounts.json`
- Run: `ERCXXXX_ADDRESS=0x... go test -run Integration ./go/test/`

## Skill Updates

Both `/add-erc` and `/update-erc` skills need Go generation steps added:

1. **Step 7e** (add to existing step 7): Go `recompute.go` + inline tests — patterns mirroring the operation-type table above
2. **Step 9c** (add to existing step 9): Go `client.go` — concrete `*ethclient.Client` based struct
3. **Step 10** (wire up exports): Add Go module registration (new `go/<category>/<erc>/` directory)
4. **Step 12** (testkit verification): Add `go test` commands for both recompute and integration

## Iterative Refinement Process

Same as the Rust SDK development workflow:

1. Update `/add-erc` skill with Go generation steps
2. Add one ERC via the skill
3. Review the generated Go code for correctness and conventions
4. If issues found → fix the skill, delete the bad Go code, regenerate
5. Repeat until the ERC is clean
6. Move to the next ERC
7. All 10 ERCs done → skill is polished, Go SDK is complete

## What Gets Committed

- `go/go.mod`, `go/go.sum`
- Per-ERC: `client.go`, `recompute.go`, `recompute_test.go`, `abi.go`, `types.go`, `README.md`
- Integration tests under `go/test/`
- `go/README.md`
- Updated `.claude/skills/add-erc/SKILL.md` and `.claude/skills/update-erc/SKILL.md`
- Updated repo root `README.md` (add Go to language list)
