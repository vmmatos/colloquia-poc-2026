// Wrapper around `expo prebuild` that:
//  - Skips if android/ already exists (idempotent in CI and locally)
//  - Passes --no-strip-types to the child Node process on Node 22+ so that
//    expo-modules-core's .ts sources don't trigger ERR_UNSUPPORTED_NODE_MODULES_TYPE_STRIPPING
import { spawnSync } from 'child_process';
import { existsSync } from 'fs';

if (existsSync('android')) {
  console.log('android/ already exists — skipping expo prebuild.');
  process.exit(0);
}

const nodeMajor = parseInt(process.version.slice(1), 10);
const env = { ...process.env };

if (nodeMajor >= 22) {
  const opts = (env.NODE_OPTIONS ?? '').split(' ').filter(Boolean);
  // Node 22+ promoted TypeScript type-stripping; it blocks .ts files in node_modules.
  // Pass the flag only to the child process (not this script) to avoid "unknown flag" errors
  // on older Node versions if this script is invoked with a shared NODE_OPTIONS.
  opts.push('--no-strip-types');
  env.NODE_OPTIONS = opts.join(' ');
}

const npx = process.platform === 'win32' ? 'npx.cmd' : 'npx';
const result = spawnSync(
  npx,
  ['expo', 'prebuild', '--platform', 'android', '--no-install'],
  { stdio: 'inherit', env },
);

process.exit(result.status ?? 1);
