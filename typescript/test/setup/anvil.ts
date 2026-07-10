import { execFileSync } from 'node:child_process'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const testkitDir = path.resolve(fileURLToPath(new URL('.', import.meta.url)), '../../../testkit')

export default async function setup() {
  execFileSync(path.join(testkitDir, 'scripts', 'start-anvil.sh'))

  return async function teardown() {
    execFileSync(path.join(testkitDir, 'scripts', 'stop-anvil.sh'))
  }
}
