import { defineConfig } from 'vitest/config'

export default defineConfig({
  test: {
    globalSetup: ['./test/setup/anvil.ts'],
    testTimeout: 20000,
    // All ERC test files share one anvil instance and deployer account
    // (testkit/scripts/deploy.sh); running files in parallel races two
    // concurrent broadcasts against the same nonce. Serialize files —
    // tests within a file still run in the normal, fast in-process order.
    fileParallelism: false,
  },
})
