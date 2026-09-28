import assert from 'node:assert/strict'
import { readdirSync, readFileSync, statSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'
import test from 'node:test'

const sourceRoot = join(dirname(fileURLToPath(import.meta.url)), '..')

function sourceFiles(directory) {
  return readdirSync(directory).flatMap((name) => {
    const path = join(directory, name)
    if (name === 'locales') return []
    return statSync(path).isDirectory() ? sourceFiles(path) : /\.(?:ts|tsx|mjs)$/.test(name) ? [path] : []
  })
}

test('Web source has no Admin route, API, or module dependency', () => {
  for (const path of sourceFiles(sourceRoot)) {
    const source = readFileSync(path, 'utf8')
    assert.doesNotMatch(source, /\/api\/admin(?:\/|\b)/, path)
    assert.doesNotMatch(source, /(?:href|to)\s*=\s*[{\"']\/admin(?:\/|[\"'}])/, path)
    assert.doesNotMatch(source, /from\s+['\"][^'\"]*(?:\/|^)admin(?:\/|['\"])/, path)
  }
})
