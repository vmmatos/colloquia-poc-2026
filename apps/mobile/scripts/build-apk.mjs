// Builds the Android APK. debuggableVariants=[] in app/build.gradle means
// the JS bundle is always embedded — no Metro server needed at runtime.
// --release : release variant
// --all-abi : build for all 4 ABIs (default: arm64-v8a only for speed)
// --ci      : CI mode — all ABIs, daemon off
import { spawnSync } from 'child_process';

const args = process.argv.slice(2);
const ci      = args.includes('--ci');
const release = args.includes('--release');
const allAbi  = args.includes('--all-abi') || ci;

const variant = release ? 'assembleRelease' : 'assembleDebug';

// NODE_ENV must be set for expo export:embed to choose the right bundle mode.
const env = { ...process.env, NODE_ENV: release ? 'production' : 'development' };

// Strip any stale --no-strip-types from NODE_OPTIONS — that flag does not exist
// in Node 22 and causes Gradle's node sub-processes to fail silently.
if (env.NODE_OPTIONS) {
  env.NODE_OPTIONS = env.NODE_OPTIONS
    .split(' ')
    .filter((f) => f !== '--no-strip-types')
    .join(' ')
    .trim() || undefined;
}

const gradleArgs = [variant];

// Local builds: arm64-v8a only (all modern Android phones).
// CI / --all-abi: build all 4 ABIs for broader compatibility.
if (!allAbi) {
  gradleArgs.push('-PreactNativeArchitectures=arm64-v8a');
}

if (ci) gradleArgs.push('--no-daemon');

console.log(`▶ ./gradlew ${gradleArgs.join(' ')}`);
const result = spawnSync('./gradlew', gradleArgs, {
  cwd: new URL('../android', import.meta.url).pathname,
  stdio: 'inherit',
  env,
});

process.exit(result.status ?? 1);
