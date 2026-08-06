// Package erc8274 provides Go clients for ERC-8274 — AI Inference Proof
// Verification Interfaces.
//
// The ERC is three separate interfaces meant to be deployed as three
// separate contracts that reference each other by address:
//
//	Settlement Contract (IAgentVerifiable) — declares which IAgentVerifier it trusts
//	    └── IAgentVerifier — stateful: agent authorization + proof routing
//	            └── IProofVerifier — stateless: raw cryptographic proof check (zkML/opML/TEE)
//
// ProofVerifierClient wraps IProofVerifier: verify() is a read-only
// simulated call, not a broadcast transaction — a pure cryptographic check
// with no state to persist, so anyone can freely re-derive the answer
// without spending gas or holding a key with funds. AgentVerifierClient
// wraps IAgentVerifier: verify() is state-changing (it emits
// VerificationCompleted on-chain — that log is the point of calling it), so
// it is broadcast. GetTrustedVerifier reads IAgentVerifiable.agentVerifier()
// — the single getter of the declaration layer.
//
// Recompute-to-verify: NO. This package performs no off-chain recomputation
// (see README.md): the proof-validity result is delegated to the deployed,
// immutable IProofVerifier contract, and the verificationDigest cannot be
// reconstructed from the standard interface alone (agentProofProfile is not
// exposed by IAgentVerifier). The only digest computed here is
// AgentVerifierClient.GetDigest — the mock's expected-proof digest for the
// empty-metadata routing — which exists to make VerifyTask self-contained
// against the testkit reference implementation.
package erc8274

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// ErrNoSigner is returned by VerifyTask when the client was constructed
// without a private key.
var ErrNoSigner = errors.New("erc8274: VerifyTask requires a signer key (NewAgentVerifierClient with non-nil key)")

// ProofVerifierClient checks cryptographic proofs against a deployed
// IProofVerifier contract (ERC-8274 inner verification layer). All calls
// are read-only eth_call invocations — no gas or broadcast needed — which
// is this ERC's recompute-to-verify mechanism for its core claim: the
// deployed, immutable contract IS the public, callable-by-anyone checking
// procedure.
type ProofVerifierClient struct {
	rpc     *ethclient.Client
	address common.Address
}

// NewProofVerifierClient creates a client bound to a deployed ERC-8274
// IProofVerifier contract.
func NewProofVerifierClient(rpc *ethclient.Client, addr common.Address) *ProofVerifierClient {
	return &ProofVerifierClient{rpc: rpc, address: addr}
}

// ProofSystem reads the IProofVerifier.proofSystem view — the
// human-readable proof-system identifier (e.g. "zkML-Halo2", "opML-V1").
func (c *ProofVerifierClient) ProofSystem() (string, error) {
	out, err := callView(c.rpc, c.address, ProofVerifierABI, "proofSystem")
	if err != nil {
		return "", err
	}
	return unpackString(ProofVerifierABI, "proofSystem", out)
}

// ProofProfile reads the IProofVerifier.proofProfile view — the compact
// proof profile hash (e.g. keccak256 of proof system + version + circuit).
func (c *ProofVerifierClient) ProofProfile() (common.Hash, error) {
	out, err := callView(c.rpc, c.address, ProofVerifierABI, "proofProfile")
	if err != nil {
		return common.Hash{}, err
	}
	return unpackHash(ProofVerifierABI, "proofProfile", out)
}

// Verify checks whether proof is cryptographically valid for the given
// (inputHash, outputHash) pair via the IProofVerifier.verify view. The
// boolean is computed on-chain and never trusted off-chain. For the testkit
// mock, a proof is valid iff keccak256(proof) ==
// keccak256(abi.encodePacked(inputHash, outputHash, metadata)) — i.e.
// proof = keccak256(abi.encodePacked(inputHash, outputHash, metadata)) —
// not real zkML/opML/TEE cryptography.
func (c *ProofVerifierClient) Verify(inputHash, outputHash common.Hash, metadata []byte, proof common.Hash) (bool, error) {
	out, err := callView(c.rpc, c.address, ProofVerifierABI, "verify", inputHash, outputHash, metadata, proof[:])
	if err != nil {
		return false, err
	}
	return unpackBool(ProofVerifierABI, "verify", out)
}

// AgentVerifierClient verifies that an agent produced a given output from a
// given input, with a valid cryptographic proof, via a deployed
// IAgentVerifier contract (ERC-8274 outer verification layer). VerifyTask
// broadcasts — the contract emits VerificationCompleted on-chain, and that
// log is the point of calling it.
type AgentVerifierClient struct {
	rpc     *ethclient.Client
	address common.Address
	key     *ecdsa.PrivateKey // signs the verify transaction; nil for read-only clients
}

// NewAgentVerifierClient creates a client bound to a deployed ERC-8274
// IAgentVerifier contract. key signs the verify() broadcast — pass nil for
// a read-only client (GetDigest only).
func NewAgentVerifierClient(rpc *ethclient.Client, addr common.Address, key *ecdsa.PrivateKey) *AgentVerifierClient {
	return &AgentVerifierClient{rpc: rpc, address: addr, key: key}
}

// VerifyTask broadcasts IAgentVerifier.verify(taskId, agentId, inputHash,
// outputHash, proof) and reports the valid flag parsed from the
// VerificationCompleted event in the mined receipt.
//
// The proof is derived by GetDigest: the ERC's agent verifier routes to an
// inner IProofVerifier with empty metadata (the ERC doesn't specify
// per-agent metadata, and the testkit mock follows that), so the proof that
// the bound verifier accepts is exactly the empty-metadata expected digest.
// Callers needing to exercise the valid=false path (a bad proof) must call
// the contract directly with their own proof bytes — this client always
// derives the correct one.
//
// Gas limit, base fee and nonce are resolved against the live node; the
// chain id is fetched from the RPC at call time. Returns ErrNoSigner if the
// client has no private key.
func (c *AgentVerifierClient) VerifyTask(taskId, agentId, inputHash, outputHash common.Hash) (bool, error) {
	if c.key == nil {
		return false, ErrNoSigner
	}
	proof, err := c.GetDigest(inputHash, outputHash)
	if err != nil {
		return false, err
	}
	receipt, err := c.transactAndWait("verify", taskId, agentId, inputHash, outputHash, proof[:])
	if err != nil {
		return false, err
	}
	return parseVerificationCompleted(receipt)
}

// GetDigest computes the expected proof digest for the given input/output
// pair: keccak256(abi.encodePacked(inputHash, outputHash)). This is the
// digest the bound agent verifier's empty-metadata routing accepts as a
// proof (the testkit MockProofVerifier's expectedDigest with empty
// metadata), and the proof VerifyTask derives for its broadcast. It is a
// pure off-chain computation — no RPC, no key — always returning a nil
// error.
func (c *AgentVerifierClient) GetDigest(inputHash, outputHash common.Hash) (common.Hash, error) {
	return crypto.Keccak256Hash(append(inputHash[:], outputHash[:]...)), nil
}

// GetTrustedVerifier reads the IAgentVerifiable.agentVerifier view — the
// IAgentVerifier address a settlement/execution contract declares it
// trusts. A standalone function, not a client struct: IAgentVerifiable is a
// single getter, so a struct would be one method wrapping a constructor for
// no benefit.
func GetTrustedVerifier(rpc *ethclient.Client, verifiableAddr common.Address) (common.Address, error) {
	out, err := callView(rpc, verifiableAddr, AgentVerifiableABI, "agentVerifier")
	if err != nil {
		return common.Address{}, err
	}
	return unpackAddress(AgentVerifiableABI, "agentVerifier", out)
}

// transactAndWait packs the inputs via a.Pack(name, args...) — which
// prepends the 4-byte method selector — signs, estimates gas (GasLimit 0),
// broadcasts through the same rpc client, and waits for the transaction to
// be mined. Reverting inputs surface as an error from gas estimation,
// before anything is broadcast.
func (c *AgentVerifierClient) transactAndWait(method string, args ...interface{}) (*types.Receipt, error) {
	a, err := AgentVerifierABI()
	if err != nil {
		return nil, fmt.Errorf("erc8274: parse ABI: %w", err)
	}
	ctx := context.Background()
	chainID, err := c.rpc.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("erc8274: fetch chain id: %w", err)
	}
	auth, err := bind.NewKeyedTransactorWithChainID(c.key, chainID)
	if err != nil {
		return nil, fmt.Errorf("erc8274: create transactor: %w", err)
	}
	bound := bind.NewBoundContract(c.address, a, c.rpc, c.rpc, c.rpc)
	tx, err := bound.Transact(auth, method, args...)
	if err != nil {
		return nil, fmt.Errorf("erc8274: %s: %w", method, err)
	}
	receipt, err := bind.WaitMined(ctx, c.rpc, tx)
	if err != nil {
		return nil, fmt.Errorf("erc8274: wait mined %s: %w", method, err)
	}
	return receipt, nil
}

// parseVerificationCompleted extracts the valid flag from the
// VerificationCompleted event log. taskId and agentId are indexed (topics
// 1-2); inputHash, outputHash, valid and verificationDigest are in the log
// data.
func parseVerificationCompleted(receipt *types.Receipt) (bool, error) {
	a, err := AgentVerifierABI()
	if err != nil {
		return false, fmt.Errorf("erc8274: parse ABI: %w", err)
	}
	evt, ok := a.Events["VerificationCompleted"]
	if !ok {
		return false, errors.New("erc8274: ABI has no VerificationCompleted event")
	}
	for _, log := range receipt.Logs {
		if log.Topics[0] != evt.ID {
			continue
		}
		if len(log.Topics) != 3 {
			return false, fmt.Errorf("erc8274: VerificationCompleted log has %d topics, want 3", len(log.Topics))
		}
		vals, err := evt.Inputs.NonIndexed().Unpack(log.Data)
		if err != nil {
			return false, fmt.Errorf("erc8274: unpack VerificationCompleted data: %w", err)
		}
		if len(vals) != 4 {
			return false, fmt.Errorf("erc8274: VerificationCompleted data has %d values, want 4", len(vals))
		}
		valid, ok := vals[2].(bool)
		if !ok {
			return false, fmt.Errorf("erc8274: VerificationCompleted valid is %T, want bool", vals[2])
		}
		return valid, nil
	}
	return false, fmt.Errorf("erc8274: VerificationCompleted event not found in receipt")
}

// callView packs the function inputs and performs a read-only eth_call
// against the given contract address.
//
// GOTCHA: the inputs must be packed via abi.ABI.Pack(name, args...), which
// prepends the 4-byte method selector — packing the arguments standalone
// with abi.Arguments.Pack produces calldata with no selector, which the
// node would reject.
func callView(rpc *ethclient.Client, address common.Address, abiGetter func() (abi.ABI, error), methodName string, args ...interface{}) ([]byte, error) {
	a, err := abiGetter()
	if err != nil {
		return nil, fmt.Errorf("erc8274: parse ABI: %w", err)
	}
	data, err := a.Pack(methodName, args...)
	if err != nil {
		return nil, fmt.Errorf("erc8274: pack %s inputs: %w", methodName, err)
	}
	msg := ethereum.CallMsg{To: &address, Data: data}
	return rpc.CallContract(context.Background(), msg, nil)
}

// unpackOutputs unpacks the raw call result with the method's declared
// output types.
func unpackOutputs(abiGetter func() (abi.ABI, error), methodName string, data []byte) ([]interface{}, error) {
	a, err := abiGetter()
	if err != nil {
		return nil, fmt.Errorf("erc8274: parse ABI: %w", err)
	}
	method, ok := a.Methods[methodName]
	if !ok {
		return nil, fmt.Errorf("erc8274: ABI has no method %q", methodName)
	}
	vals, err := method.Outputs.Unpack(data)
	if err != nil {
		return nil, fmt.Errorf("erc8274: unpack %s outputs: %w", methodName, err)
	}
	return vals, nil
}

// unpackBool decodes a single-boolean view result.
func unpackBool(abiGetter func() (abi.ABI, error), methodName string, data []byte) (bool, error) {
	vals, err := unpackOutputs(abiGetter, methodName, data)
	if err != nil {
		return false, err
	}
	if len(vals) != 1 {
		return false, fmt.Errorf("erc8274: %s returned %d outputs, want 1", methodName, len(vals))
	}
	v, ok := vals[0].(bool)
	if !ok {
		return false, fmt.Errorf("erc8274: %s output is %T, want bool", methodName, vals[0])
	}
	return v, nil
}

// unpackString decodes a single-string view result.
func unpackString(abiGetter func() (abi.ABI, error), methodName string, data []byte) (string, error) {
	vals, err := unpackOutputs(abiGetter, methodName, data)
	if err != nil {
		return "", err
	}
	if len(vals) != 1 {
		return "", fmt.Errorf("erc8274: %s returned %d outputs, want 1", methodName, len(vals))
	}
	s, ok := vals[0].(string)
	if !ok {
		return "", fmt.Errorf("erc8274: %s output is %T, want string", methodName, vals[0])
	}
	return s, nil
}

// unpackHash decodes a single-bytes32 view result.
func unpackHash(abiGetter func() (abi.ABI, error), methodName string, data []byte) (common.Hash, error) {
	vals, err := unpackOutputs(abiGetter, methodName, data)
	if err != nil {
		return common.Hash{}, err
	}
	if len(vals) != 1 {
		return common.Hash{}, fmt.Errorf("erc8274: %s returned %d outputs, want 1", methodName, len(vals))
	}
	b, ok := vals[0].([32]byte)
	if !ok {
		return common.Hash{}, fmt.Errorf("erc8274: %s output is %T, want [32]byte", methodName, vals[0])
	}
	return common.BytesToHash(b[:]), nil
}

// unpackAddress decodes a single-address view result.
func unpackAddress(abiGetter func() (abi.ABI, error), methodName string, data []byte) (common.Address, error) {
	vals, err := unpackOutputs(abiGetter, methodName, data)
	if err != nil {
		return common.Address{}, err
	}
	if len(vals) != 1 {
		return common.Address{}, fmt.Errorf("erc8274: %s returned %d outputs, want 1", methodName, len(vals))
	}
	addr, ok := vals[0].(common.Address)
	if !ok {
		return common.Address{}, fmt.Errorf("erc8274: %s output is %T, want common.Address", methodName, vals[0])
	}
	return addr, nil
}
