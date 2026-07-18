def check_stateful_bound(reserved: int, confirmed: int, cap: int) -> bool:
    """
    Check that the sum of reserved and confirmed amounts does not exceed the cap.

    ERC-8312 (StatefulBound variant): (reserved + confirmed) <= cap.

    Args:
        reserved: The reserved amount (non-negative integer).
        confirmed: The confirmed amount (non-negative integer).
        cap: The cap (total capacity).

    Returns:
        True if the sum is within the cap, False otherwise.
    """
    return reserved + confirmed <= cap


def check_cursor_headroom(aggregate: int, cap: int) -> bool:
    """
    Check that an aggregate cursor value does not exceed the cap.

    ERC-8312 (Orbmis/headroom variant): aggregate <= cap.

    Args:
        aggregate: The aggregate cursor value (non-negative integer).
        cap: The cap (total capacity).

    Returns:
        True if the aggregate is within the cap, False otherwise.
    """
    return aggregate <= cap


def compute_remaining_headroom(cap: int, spent: int) -> int:
    """
    Compute the remaining headroom from cap and cumulative spent.

    ERC-8312 §IBudgetSubstrate: remaining = cap - spent.
    Returns 0 if spent exceeds cap (exhausted or inactive envelope).

    Args:
        cap: The maximum capacity (non-negative integer).
        spent: The cumulative amount consumed (non-negative integer).

    Returns:
        The remaining headroom (non-negative integer).
    """
    return max(0, cap - spent)


def verify_remaining(cap: int, spent: int, reported_remaining: int) -> bool:
    """
    Verify that reported remaining matches cap - spent.

    ERC-8312 §IBudgetSubstrate: remaining(id) is recomputed, never trusted.

    Args:
        cap: The maximum capacity.
        spent: The cumulative amount consumed.
        reported_remaining: The reported remaining headroom.

    Returns:
        True if the reported value matches the recomputed headroom.
    """
    return spent <= cap and compute_remaining_headroom(cap, spent) == reported_remaining
