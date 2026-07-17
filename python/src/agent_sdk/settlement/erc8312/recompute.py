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
