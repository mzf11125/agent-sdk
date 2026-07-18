/// ERC-8312 (StatefulBound variant): (reserved + confirmed) <= cap.
pub fn check_stateful_bound(reserved: u64, confirmed: u64, cap: u64) -> bool {
    reserved.saturating_add(confirmed) <= cap
}

/// ERC-8312 (Orbmis/headroom variant): aggregate <= cap.
pub fn check_cursor_headroom(aggregate: u64, cap: u64) -> bool {
    aggregate <= cap
}

#[cfg(test)]
mod tests {
    use super::*;

    /// Golden vector from recompute-kit "8312/cap-conservation": holds
    #[test]
    fn golden_holds() {
        assert!(check_stateful_bound(100, 0, 150));
    }

    /// Golden vector: headroom
    #[test]
    fn golden_headroom() {
        assert!(check_cursor_headroom(0, 8000));
    }

    /// Golden vector: breach
    #[test]
    fn golden_breach() {
        assert!(!check_stateful_bound(100, 60, 150));
    }

    #[test]
    fn exact_cap_holds() {
        assert!(check_stateful_bound(100, 50, 150));
    }

    #[test]
    fn zero_aggregate_always_holds() {
        assert!(check_cursor_headroom(0, 0));
        assert!(check_cursor_headroom(0, 1));
    }
}
