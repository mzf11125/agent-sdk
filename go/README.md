# Go SDK

Go client library for the ERCs defined in [trustless-ai/agent-ercs](https://github.com/trustless-ai/agent-ercs).

## Install

```bash
go get github.com/trustless-ai/agent-sdk/go
```

## Usage

```go
import "github.com/trustless-ai/agent-sdk/go/<category>/<erc_lowercase>"
```

Each ERC's own `README.md` documents the full method list, types, and whether `verify()` is available.

## Supported ERCs

| ERC | Category | Contract Calls | Recompute |
|-----|----------|----------------|-----------|
| ERC-8004 | identity | IdentityRegistryClient | ComputeAgentId |
| ERC-8203 | settlement | ConsultEscrowClient | ComputeVerdictHash |
| ERC-8263 | anchor | OnChainProofClient | — |
| ERC-8274 | verify | ProofVerifierClient, AgentVerifierClient | — |
| ERC-8275 | reputation | AgentReputationClient | ComputeWinRate |
| ERC-8281 | verify | ObservationCommitmentClient | ComputeObservationDigest |
| ERC-8299 | verify | WyriweAttestationClient, JudgmentExecutionClient | ComputeRawInputHash, ComputeSanitizationPipelineHash |
| ERC-8301 | execution | AgentWorkflowClient | ComputeTaskHash, ComputeReplyHash |
| ERC-8312 | metering | BoundedAgentActionClient, BudgetSubstrateClient, ContestableEnvelopeClient | CheckStatefulBound, CheckCursorHeadroom |
| ERC-8323 | identity | SourceBindingClient | — |

## License

Apache 2.0
