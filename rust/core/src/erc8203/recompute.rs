use alloy_primitives::{hex, keccak256, B256};
use alloy_core::sol_types::SolValue;

/// ERC-8203 settlement: compute the verdict hash from jobId and resultText.
///
/// verdictHash = keccak256(abi.encode(jobId, resultHash))
///   where resultHash = keccak256(utf8(resultText))
///
/// This is the commitment ConsultEscrow.release() recomputes on-chain
/// before checking the attestor's signature.
pub fn compute_verdict_hash(job_id: B256, result_text: &str) -> B256 {
    let result_hash = keccak256(result_text.as_bytes());
    // abi.encode(bytes32, bytes32)
    let encoded = (job_id, result_hash).abi_encode();
    keccak256(&encoded)
}

#[cfg(test)]
mod tests {
    use super::*;

    /// Golden vector from recompute-kit "8203/settlement-proof":
    /// jobId: 0xbc01b40fe7a3509f35470053d4bc1844d50c9782546cf0fc11154adcb90caa56
    /// resultText: "No intermediaries required, cryptographic verification only."
    /// expected: 0xdc568bd1cbacdd1ead8231e9d3d6f4e475f5168f3cc9f72b31935d46cfdd48f7
    #[test]
    fn golden_verdict_hash() {
        let job_id = B256::from(hex!("bc01b40fe7a3509f35470053d4bc1844d50c9782546cf0fc11154adcb90caa56"));
        let text = "No intermediaries required, cryptographic verification only.";
        let hash = compute_verdict_hash(job_id, text);
        let expected = B256::from(hex!("dc568bd1cbacdd1ead8231e9d3d6f4e475f5168f3cc9f72b31935d46cfdd48f7"));
        assert_eq!(hash, expected);
    }

    #[test]
    fn empty_text_produces_keccak_of_empty_string() {
        let job_id = B256::ZERO;
        let hash = compute_verdict_hash(job_id, "");
        // resultHash = keccak256("") = 0xc5d2...a470
        let result_hash = keccak256(b"");
        let expected = keccak256(&(job_id, result_hash).abi_encode());
        assert_eq!(hash, expected);
    }

    #[test]
    fn different_job_different_hash() {
        let text = "same text";
        let a = compute_verdict_hash(B256::ZERO, text);
        let b = compute_verdict_hash(B256::from(hex!("0000000000000000000000000000000000000000000000000000000000000001")), text);
        assert_ne!(a, b);
    }

    #[test]
    fn different_text_different_hash() {
        let job_id = B256::ZERO;
        let a = compute_verdict_hash(job_id, "option A");
        let b = compute_verdict_hash(job_id, "option B");
        assert_ne!(a, b);
    }
}
