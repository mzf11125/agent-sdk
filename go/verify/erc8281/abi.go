package erc8281

import (
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

// observationCommitmentABIJSON is the ERC-8281 IObservationCommitment
// interface as JSON, hand-written to match the interface exactly (see
// agent-ercs/contracts/verify/ERC8281/IObservationCommitment.sol).
const observationCommitmentABIJSON = `[
  {
    "type": "function",
    "name": "record",
    "stateMutability": "nonpayable",
    "inputs": [{"name": "digest", "type": "bytes32"}],
    "outputs": []
  },
  {
    "type": "event",
    "name": "Recorded",
    "inputs": [
      {"name": "digest", "type": "bytes32", "indexed": true},
      {"name": "committer", "type": "address", "indexed": true}
    ]
  }
]`

var (
	ocABI     abi.ABI
	ocABIOnce sync.Once
	ocABIErr  error
)

// ObservationCommitmentABI returns the parsed ERC-8281
// IObservationCommitment ABI. Parsed once and cached; errors are returned,
// never panicked.
func ObservationCommitmentABI() (abi.ABI, error) {
	ocABIOnce.Do(func() {
		ocABI, ocABIErr = abi.JSON(strings.NewReader(observationCommitmentABIJSON))
	})
	return ocABI, ocABIErr
}
