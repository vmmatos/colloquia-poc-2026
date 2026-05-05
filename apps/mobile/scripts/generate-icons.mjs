// Generates icon.png and adaptive-icon.png from assets/icon.svg using sharp.
// Run once, then commit the resulting PNGs to assets/.
//
// Usage:
//   npm install --save-dev sharp   (one time)
//   node scripts/generate-icons.mjs

import { createRequire } from 'module';
import { readFileSync, writeFileSync } from 'fs';
import { fileURLToPath } from 'url';
import { dirname, resolve } from 'path';

const __dir = dirname(fileURLToPath(import.meta.url));
const root = resolve(__dir, '..');

let sharp;
try {
  const req = createRequire(import.meta.url);
  sharp = req('sharp');
} catch {
  console.error('sharp not installed. Run: npm install --save-dev sharp');
  process.exit(1);
}

const svgPath = resolve(root, 'assets/icon.svg');
const svgSrc = readFileSync(svgPath);

// icon.png — 1024×1024, dark background baked in (SVG already has bg rect)
await sharp(svgSrc, { density: 300 })
  .resize(1024, 1024)
  .png()
  .toFile(resolve(root, 'assets/icon.png'));
console.log('✓ assets/icon.png');

// adaptive-icon.png — 1024×1024, transparent background (remove the bg rect via SVG edit)
const svgNoBg = svgSrc.toString()
  .replace(/<rect width="1024" height="1024" fill="#121212"\/>\n?\s*/, '');

await sharp(Buffer.from(svgNoBg), { density: 300 })
  .resize(1024, 1024)
  .png()
  .toFile(resolve(root, 'assets/adaptive-icon.png'));
console.log('✓ assets/adaptive-icon.png');

console.log('\nDone. Run `npm run prebuild` to apply icons to android/.');
