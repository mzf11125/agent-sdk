# ERC-8323 — Source-Token Agent Binding (view side)

Client bindings for `IAgentSourceBindingView` (ERC-165 id **`0x8b3597c9`**) — the
query-only subset of source-binding that a self-sourced ("genesis") agent honestly
implements. It exposes the *read* side (which source NFT an agent is bound to, and
whether that binding is currently valid) without the bridge methods
(`boundCollection` / `registerWithSource`) of the full `IAgentSourceBinding`
(`0x27eba962`).

## API

### `SourceBindingViewClient`

| Method | Description | State |
| --- | --- | --- |
| `getSourceNFT(agentId)` | The `(sourceContract, sourceTokenId)` the agent is bound to | read |
| `hasSourceNFT(agentId)` | Whether the agent claims a source NFT | read |
| `isSourceNFTOwnershipValid(agentId)` | Whether the bound NFT is still owned by the agent's controller | read |
| `supportsSourceBindingView()` | ERC-165 check for `0x8b3597c9` | read |

`SOURCE_BINDING_VIEW_INTERFACE_ID` (`0x8b3597c9`) is exported for direct ERC-165
checks.

## No Layer-2 recompute

Unlike the hash-based ERCs (8299, 8301, 8203, …), source binding is an **on-chain
fact** — the registry stores `agentId → source NFT` and validates current ownership
on read. Verification is therefore a direct `getSourceNFT` / `isSourceNFTOwnershipValid`
call, not an off-chain hash re-derivation, so there is no `recompute` module here.

## Notes

- The reference implementation is `GenesisAgentRegistry` in `ens-dynamic-kit`
  (self-sourced agents) and `AgentIdentityRegistry` (source-bound agents), both of
  which use a **uint256 `agentId`** (ERC-721 tokenId) per-collection registry model.
- Integration tests need a deployed registry in the testkit; a `GenesisAgentRegistry`
  fixture is the natural fit.
