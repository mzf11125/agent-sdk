use alloy_primitives::{keccak256, FixedBytes};

/// ERC-8299 §45: raw_input_hash = keccak256(raw_user_input)
pub fn compute_raw_input_hash(raw_user_input: &[u8]) -> FixedBytes<32> {
    keccak256(raw_user_input)
}

/// ERC-8299 §46: sanitization_pipeline_hash = keccak256(utf8(cid) || raw_input_hash)
pub fn compute_sanitization_pipeline_hash(cid: &str, raw_input_hash: FixedBytes<32>) -> FixedBytes<32> {
    let mut buf = Vec::with_capacity(cid.len() + 32);
    buf.extend_from_slice(cid.as_bytes());
    buf.extend_from_slice(raw_input_hash.as_slice());
    keccak256(&buf)
}

#[cfg(test)]
mod tests {
    use super::*;
    use alloy_primitives::hex;

    /// Golden vector: "wyriwe/raw" — raw_input_hex "0x68656c6c6f"
    /// expected: 0x1c8aff950685c2ed4bc3174f3472287b56d9517b9c948127319a09a7a36deac8
    #[test]
    fn golden_raw_input_hash() {
        let raw = hex!("68656c6c6f");
        let hash = compute_raw_input_hash(&raw);
        let expected = FixedBytes::<32>::from(hex!("1c8aff950685c2ed4bc3174f3472287b56d9517b9c948127319a09a7a36deac8"));
        assert_eq!(hash, expected);
    }

    /// Golden vector: "wyriwe/pipeline" — cid + raw_input_hash
    /// raw_input_hash: 0x1c8aff950685c2ed4bc3174f3472287b56d9517b9c948127319a09a7a36deac8
    /// cid: "ipfs://QmccvoM6aRVgZ2dtFWvT6Wm3DmTvoAUHHotK7uQufnStVR"
    /// expected: 0x5798efed4aa92f96a0622fc30268042b067294bdb5fd06f599bf8d84fd5d734b
    #[test]
    fn golden_sanitization_pipeline_hash() {
        let raw_hash = FixedBytes::<32>::from(hex!("1c8aff950685c2ed4bc3174f3472287b56d9517b9c948127319a09a7a36deac8"));
        let cid = "ipfs://QmccvoM6aRVgZ2dtFWvT6Wm3DmTvoAUHHotK7uQufnStVR";
        let hash = compute_sanitization_pipeline_hash(cid, raw_hash);
        let expected = FixedBytes::<32>::from(hex!("5798efed4aa92f96a0622fc30268042b067294bdb5fd06f599bf8d84fd5d734b"));
        assert_eq!(hash, expected);
    }

    #[test]
    fn empty_input_produces_keccak_of_empty() {
        let hash = compute_raw_input_hash(&[]);
        // keccak256("") = 0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470
        let expected = FixedBytes::<32>::from(hex!("c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470"));
        assert_eq!(hash, expected);
    }

    #[test]
    fn different_input_different_hash() {
        let a = compute_raw_input_hash(&[1u8]);
        let b = compute_raw_input_hash(&[2u8]);
        assert_ne!(a, b);
    }
}
