# ERC-8203 — ConsultEscrow Settlement (TypeScript)

**Recompute-to-verify: YES.**

## What IS recompute-to-verify: the settlement commitment

The `release()` function recomputes `commitmentHash = keccak256(abi.encode(jobId, resultHash))` on-chain and checks the recovered signer matches `job.attestor`. Both `jobId` and `resultHash` are public — `jobId` is a parameter, `resultHash` is posted as part of the `release()` call and visible in the `Released` event. Anyone can independently recompute `commitmentHash` from `jobId` and `resultText` and verify it matches, or call `jobs(jobId)` to get the stored attestor and check signatures against it.

This is why:
- `ConsultEscrowClient.verify()` is a **pure recompute** (Layer 2) — no contract call or gas needed — that recomputes the verdict hash and compares it to a claimed value.
- The `computeVerdictHash()` function (Layer 2) is the pure mathematical function at the heart of the release check, testable offline against golden conformance vectors.

## API

### ConsultEscrowClient

Wraps `IConsultEscrow` for a single deployed escrow contract:

| Method | Description | State |
|--------|-------------|-------|
| `open(jobId, provider, attestor, deadline, value)` | Lock ETH for a consultation | broadcast |
| `release(jobId, resultHash, signature)` | Release funds on attestor's signed commitment | broadcast |
| `refund(jobId)` | Refund consumer after deadline | broadcast |
| `getJob(jobId)` | Read the escrowed job struct | read-only |
| `verify(commitmentHash, jobId, resultText)` | Pure recompute-to-verify check | offline (no RPC) |

### Event parsing helpers

- `parseOpened(receipt)` — parse `Opened` events
- `parseReleased(receipt)` — parse `Released` events
- `parseRefunded(receipt)` — parse `Refunded` events

## Layer 2 — Pure recompute

**Pure recompute: YES (one function).**

`computeVerdictHash(jobId, resultText)` — computes the release commitment:
```
verdictHash = keccak256(abi.encode(
    bytes32 jobId,
    keccak256(utf8(resultText))
))
```

Golden vector (step `8203/settlement-proof`):
- `jobId: "0xbc01b40fe7a3509f35470053d4bc1844d50c9782546cf0fc11154adcb90caa56"`
- `resultText: "No intermediaries required, cryptographic verification only."`
- expected: `"0xdc568bd1cbacdd1ead8231e9d3d6f4e475f5168f3cc9f72b31935d46cfdd48f7"`

This lives in `recompute.ts` and is tested against golden vectors from `recompute-kit/conformance/agent-flow.vectors.json`. The recompute tests are pure function calls with no RPC, no anvil, and no deployed contract.
