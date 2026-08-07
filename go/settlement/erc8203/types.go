package erc8203

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// JobStatus is the lifecycle state of an escrowed job (ERC-8203
// IConsultEscrow.Status).
type JobStatus uint8

const (
	// JobStatusNone is the default state of a never-opened job.
	JobStatusNone JobStatus = 0
	// JobStatusOpen is an escrowed job awaiting a signed release or refund.
	JobStatusOpen JobStatus = 1
	// JobStatusReleased is a job whose escrow was paid out to the provider.
	JobStatusReleased JobStatus = 2
	// JobStatusRefunded is a job whose escrow was returned to the consumer
	// after the deadline lapsed.
	JobStatusRefunded JobStatus = 3
)

// Job is the escrowed consultation job (ERC-8203 IConsultEscrow.jobs).
//
// A never-opened jobId reads back as the all-zero default: zero addresses,
// zero amount/deadline, and Status == JobStatusNone.
type Job struct {
	Consumer common.Address // Payer; refund recipient after a lapsed deadline.
	Provider common.Address // Paid on a valid signed release (the agent's wallet).
	Attestor common.Address // Key whose signature over the commitment authorizes release.
	Amount   *big.Int       // Locked ETH, in wei.
	Deadline *big.Int       // Unix timestamp after which the consumer may refund.
	Status   JobStatus      // Current lifecycle state.
}
