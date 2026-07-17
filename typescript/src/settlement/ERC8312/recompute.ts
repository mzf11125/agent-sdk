/**
 * Check that the sum of reserved and confirmed amounts does not exceed the cap.
 *
 * ERC-8312 (StatefulBound variant): (reserved + confirmed) <= cap.
 *
 * @param reserved - The reserved amount (non-negative integer).
 * @param confirmed - The confirmed amount (non-negative integer).
 * @param cap - The cap (total capacity).
 * @returns true if the sum is within the cap, false otherwise.
 */
export function checkStatefulBound(
  reserved: number,
  confirmed: number,
  cap: number,
): boolean {
  return reserved + confirmed <= cap
}

/**
 * Check that an aggregate cursor value does not exceed the cap.
 *
 * ERC-8312 (Orbmis/headroom variant): aggregate <= cap.
 *
 * @param aggregate - The aggregate cursor value (non-negative integer).
 * @param cap - The cap (total capacity).
 * @returns true if the aggregate is within the cap, false otherwise.
 */
export function checkCursorHeadroom(
  aggregate: number,
  cap: number,
): boolean {
  return aggregate <= cap
}
