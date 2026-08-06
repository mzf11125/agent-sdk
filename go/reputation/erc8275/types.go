package erc8275

// ReputationData is the current reputation snapshot for an agent (ERC-8275
// IAgentReputation.getReputation), recomputable from public Escrow events.
type ReputationData struct {
	CompletedOrders uint64 // Count of orders settled without dispute.
	DisputedOrders  uint64 // Count of orders that entered dispute/challenge.
	TotalVolume     uint64 // Cumulative settled volume across all orders (implementation-defined unit).
	LastActiveAt    uint64 // Unix timestamp of the most recent settlement event.
	Score           uint16 // Derived score per f(attestationCount, counterpartyDiversity, winRate, volumeCap).
}
