use alloy_primitives::{keccak256, FixedBytes};

/// ERC-8281 §1: compute `digest = keccak256(observation)`.
///
/// The core OCP commitment step: observation bytes are hashed to produce
/// the opaque digest anchored on-chain via `record(digest)`.
pub fn compute_observation_digest(observation: &[u8]) -> FixedBytes<32> {
    keccak256(observation)
}

#[cfg(test)]
mod tests {
    use super::*;
    use alloy_primitives::hex;

    /// keccak256("hello") inline golden vector for sanity.
    #[test]
    fn golden_hello() {
        let digest = compute_observation_digest(b"hello");
        // keccak256("hello")
        let expected = FixedBytes::<32>::from(hex!(
            "1c8aff950685c2ed4bc3174f3472287b56d9517b9c948127319a09a7a36deac8"
        ));
        assert_eq!(digest, expected);
    }

    #[test]
    fn empty_input_is_keccak_of_empty() {
        let digest = compute_observation_digest(&[]);
        let expected = FixedBytes::<32>::from(hex!(
            "c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470"
        ));
        assert_eq!(digest, expected);
    }

    #[test]
    fn different_input_different_digest() {
        let a = compute_observation_digest(b"a");
        let b = compute_observation_digest(b"b");
        assert_ne!(a, b);
    }
}
