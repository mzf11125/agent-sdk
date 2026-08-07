// Package erc8299 implements the ERC-8299 WYRIWE — Input Provenance for AI
// Inference (Go) SDK.
//
// ERC-8299 binds what a user submitted (`rawInputHash`) to what the model
// actually received (`inputHash`) via a public, replayable sanitization
// pipeline (`sanitizationPipelineHash`). Both hashes are pure recomputes:
// any verifier holding the raw input and the pinned sanitization spec CID
// can re-derive them without trusting a third party. The on-chain layer
// (IWyriweAttestation / IJudgmentExecutionAttestation) authenticates
// attestations via an EIP-712-style signature checked by the contract's
// verify(attestation, signature) view.
//
// The L4 layer (judgment validator chain-of-custody) is sha256-based
// rather than keccak256-based: it is designed to anchor off-chain
// (Nostr-relay-published) verdicts as well as on-chain ones, matching
// invinoveritas's reference implementation instead of the EVM-native hash
// the L1-L3 layer uses for its on-chain contract calls.
package erc8299

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// ComputeRawInputHash computes rawInputHash = keccak256(raw_user_input)
// (ERC-8299 §45) — the first step of the WYRIWE triple-hash construction.
//
// The raw user input bytes are hashed to produce the hash the attestation
// pins. Keccak-256 never fails, so the function returns a single value
// rather than (result, error), and never panics.
//
// Golden vector: ComputeRawInputHash([]byte("hello")) =
// 0x1c8aff950685c2ed4bc3174f3472287b56d9517b9c948127319a09a7a36deac8.
func ComputeRawInputHash(rawInput []byte) common.Hash {
	return crypto.Keccak256Hash(rawInput)
}

// ComputeSanitizationPipelineHash computes
// sanitizationPipelineHash = keccak256(utf8(cid) || rawInputHash)
// (ERC-8299 §46) — the second step of the WYRIWE triple-hash construction.
//
// The spec CID is converted to UTF-8 bytes, concatenated with the
// rawInputHash bytes (the plan's `append(rawHash[:], innerHash...)`
// pattern), and hashed. Keccak-256 never fails, so the function returns a
// single value rather than (result, error), and never panics.
//
// Golden vector:
// ComputeSanitizationPipelineHash("ipfs://QmccvoM6aRVgZ2dtFWvT6Wm3DmTvoAUHHotK7uQufnStVR",
// hexToHash("0x1c8aff950685c2ed4bc3174f3472287b56d9517b9c948127319a09a7a36deac8")) =
// 0x5798efed4aa92f96a0622fc30268042b067294bdb5fd06f599bf8d84fd5d734b.
func ComputeSanitizationPipelineHash(cid string, rawInputHash common.Hash) common.Hash {
	buf := make([]byte, 0, len(cid)+common.HashLength)
	buf = append(buf, cid...)
	buf = append(buf, rawInputHash[:]...)
	return crypto.Keccak256Hash(buf)
}

// ComputeRawProposalHash computes rawProposalHash = sha256(artifact)
// (ERC-8299, L4 judgment validator binding) — sha256 over the UTF-8 bytes
// of the proposed-action artifact string (a diff, command, plan, etc.),
// returned as 0x-prefixed lowercase hex.
//
// NOTE: this is sha256, NOT keccak256. The L4 layer anchors off-chain
// (Nostr-relay-published) verdicts as well as on-chain ones, so it uses
// sha256 over the raw artifact string directly — matching invinoveritas's
// reference implementation (services/proof_signing.py:
// `artifact_hash = sha256(artifact)`), not the EVM-native keccak256 the
// L1-L3 layer uses for its on-chain contract calls. Sha-256 never fails,
// so the function returns a single value rather than (result, error), and
// never panics.
//
// Golden vector: ComputeRawProposalHash(
// "test artifact content for cross-language verification") =
// 0xb8f70a237da212a272ecd09370acedbce6ca1d7df90745beafcac77e39697a88
// (testkit/vectors/erc8299-l4.vectors.json, "8299-l4/raw-proposal-hash").
func ComputeRawProposalHash(artifact string) string {
	h := sha256.Sum256([]byte(artifact))
	return "0x" + hex.EncodeToString(h[:])
}

// ComputeVerdictHash computes verdictHash = sha256(JCS({preimage fields}))
// (ERC-8299, L4 judgment validator binding), returned as "sha256:<hex>" —
// matching invinoveritas's decision_ref format.
//
// The verdict metadata binds to rawProposalHash so a verdict cannot be
// replayed against a different proposal than the one it judged
// ("verdict-shopping"). The canonical preimage is the JCS (RFC 8785)
// serialization of the preimageFields subset of fields:
//
//   - keys sorted by code point — Go's sort.Strings compares UTF-8 byte
//     order, which is identical to code-point order (unlike JS UTF-16
//     code-unit order);
//   - each key and present value JSON-encoded with no extraneous
//     whitespace and literal UTF-8 (never \uXXXX-escaped), so the preimage
//     bytes recompute identically across implementations/languages;
//   - a key from preimageFields that is NOT present in fields serializes
//     as JSON null (never as "").
//
// The caller passes the producer's own decision_ref_preimage_fields
// (published on each proof) so a recompute matches the policy version that
// was actually in force when the verdict was issued.
//
// Golden vectors (testkit/vectors/erc8299-l4.vectors.json,
// "8299-l4/verdict-hash"): real /ledger entry #236 ->
// sha256:5bca0bf044c8e1c8e16a01bf3ee44b12c305ce6a50dd9789ff73cbd13482b9b9;
// null-valued field -> sha256:2970854c035d5aedb673b8523128665712895f62dd525c91fc8e858ad588ce58;
// non-ASCII key sort -> sha256:36e2e43ff6d7062ebb64c209604b7ce028b4eb88d4db2892e872194d16f36bca.
func ComputeVerdictHash(fields map[string]string, preimageFields []string) string {
	keys := append([]string(nil), preimageFields...)
	sort.Strings(keys)

	var b strings.Builder
	b.Grow(64 + 32*len(keys))
	b.WriteByte('{')
	for i, k := range keys {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(jcsString(k))
		b.WriteByte(':')
		if v, ok := fields[k]; ok {
			b.WriteString(jcsString(v))
		} else {
			b.WriteString("null")
		}
	}
	b.WriteByte('}')

	h := sha256.Sum256([]byte(b.String()))
	return "sha256:" + hex.EncodeToString(h[:])
}

// jcsString returns the JSON encoding of v in the exact byte form JCS
// (RFC 8785) requires: json.Marshal's string encoding with HTML escaping
// disabled, so '<', '>', '&' are emitted literally instead of as
// backslash-u003c-style escapes. The other SDK ports (TypeScript
// JSON.stringify, Python json.dumps with ensure_ascii=False) never escape
// those characters, so Go's default json.Marshal would break cross-lane
// byte compatibility on any verdict field containing them. Encoding a
// string never fails.
func jcsString(v string) string {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(v)
	return strings.TrimSuffix(buf.String(), "\n")
}
