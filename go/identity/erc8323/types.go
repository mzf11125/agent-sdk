package erc8323

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// SourceNFT is the immutable source token an agent was derived from
// (IAgentSourceBinding.getSourceNFT, ERC-8323). Recorded once at registration
// and never changed — live ownership is exposed separately via
// IsSourceNFTOwnershipValid.
type SourceNFT struct {
	SourceContract common.Address // The bound ERC-721 collection.
	SourceTokenID  *big.Int       // Token id in that collection.
}

// SourceNFTLinkedEvent is the IAgentSourceBinding.SourceNFTLinked event
// (ERC-8323), emitted exactly once per agentId at registration. agentId is
// indexed (topic 1), sourceContract is indexed (topic 2), sourceTokenId is in
// the log data. Register parses this event to learn the minted agent id.
type SourceNFTLinkedEvent struct {
	AgentID        *big.Int       // The newly minted agent identity.
	SourceContract common.Address // The bound ERC-721 collection.
	SourceTokenID  *big.Int       // The source token the agent was derived from.
}
