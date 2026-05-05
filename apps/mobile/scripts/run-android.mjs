// expo run:android --device with NODE_OPTIONS set for Node 22 compatibility.
// Also runs `adb reverse tcp:18080 tcp:80` so that the app can reach the local
// gateway at http://localhost:18080 regardless of the host machine's LAN IP.
import { spawnSync } from 'child_process';

const nodeMajor = parseInt(process.version.slice(1), 10);
const env = { ...process.env };
if (nodeMajor >= 22) {
  const opts = (env.NODE_OPTIONS ?? '').split(' ').filter(Boolean);
  if (!opts.includes('--no-strip-types')) opts.push('--no-strip-types');
  env.NODE_OPTIONS = opts.join(' ');
}

// Forward device's localhost:18080 → host's port 80 (nginx gateway).
// Metro bundler: use port 19000 to avoid conflict with Colima's SSH mux on port 8081.
// Forward device's localhost:8081 → host's port 19000 so the embedded Metro URL still works.
for (const [devicePort, hostPort] of [['18080', '80'], ['8081', '19000']]) {
  const adb = spawnSync('adb', ['reverse', `tcp:${devicePort}`, `tcp:${hostPort}`], { stdio: 'inherit' });
  if (adb.status !== 0) {
    console.warn(`[run-android] adb reverse tcp:${devicePort} tcp:${hostPort} failed`);
  }
}

const npx = process.platform === 'win32' ? 'npx.cmd' : 'npx';
const result = spawnSync(npx, ['expo', 'run:android', '--device', '--port', '19000'], {
  stdio: 'inherit',
  env,
});

process.exit(result.status ?? 1);
