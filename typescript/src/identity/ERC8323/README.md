# ERC-8323 — Source-Token Agent Binding

Client bindings for `IAgentSourceBinding` (ERC-165 id **`0x27eba962`**). Exposes
both the *read* side (which source NFT an agent is bound to, and whether that
binding is currently valid) and the *write* side (registering a new agent from
a source token in the bound collection).

## API

### `SourceBindingClient`

Requires an `Account` for transaction signing. View methods work over the
public client; write methods (`register`) broadcast via the wallet client.

| Method | Description | State |
| --- | --- | --- |
| `boundCollection()` | The source ERC-721 collection this registry is bound to | read |
| `getSourceNFT(agentId)` | The `(sourceContract, sourceTokenId)` the agent is bound to | read |
| `hasSourceNFT(agentId)` | Whether the agent claims a source NFT | read |
| `isSourceNFTOwnershipValid(agentId)` | Whether the bound NFT is still owned by the agent's controller | read |
| `register(sourceTokenId, value = 0n)` | Register an agent from `sourceTokenId` in the bound collection. Pass the registry's mint price (wei) if it charges one -- the call reverts on insufficient value | write, payable |
| `supportsSourceBinding()` | ERC-165 check for `0x27eba962` | read |

`SOURCE_BINDING_INTERFACE_ID` (`0x27eba962`) is exported for direct ERC-165
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
