# ERC-8323 — Source-Token Agent Binding

Client bindings for `IAgentSourceBinding` (ERC-165 id **`0x27eba962`**). Exposes
both the *read* side (which source NFT an agent is bound to, and whether that
binding is currently valid) and the *write* side (registering a new agent from
a source token in the bound collection).

## API

### `SourceBindingClient`

Requires a `LocalAccount` for transaction signing. View methods work over the
public client; write methods (`register`) broadcast via the signing middleware.

| Method | Description | State |
| --- | --- | --- |
| `bound_collection()` | The source ERC-721 collection this registry is bound to | read |
| `get_source_nft(agent_id)` | The `(source_contract, source_token_id)` the agent is bound to | read |
| `has_source_nft(agent_id)` | Whether the agent claims a source NFT | read |
| `is_source_nft_ownership_valid(agent_id)` | Whether the bound NFT is still owned by the agent's controller | read |
| `register(source_token_id, value = 0)` | Register an agent from `source_token_id` in the bound collection. Pass the registry's mint price (wei) if it charges one -- the call reverts on insufficient value | write, payable |
| `supports_source_binding()` | ERC-165 check for `0x27eba962` | read |

`SOURCE_BINDING_INTERFACE_ID` (`0x27eba962`) is exported for direct ERC-165
checks.

## No Layer-2 recompute

Unlike the hash-based ERCs (8299, 8301, 8203, ...), source binding is an **on-chain
fact** — the registry stores `agentId -> source NFT` and validates current ownership
on read. Verification is therefore a direct `get_source_nft` / `is_source_nft_ownership_valid`
call, not an off-chain hash re-derivation, so there is no `recompute` module here.
