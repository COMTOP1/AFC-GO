// Copies build/client into the Go embed target, builds build/afc, then resets
// the embed target to just .keep so the working tree stays clean.
import { execSync } from 'node:child_process';
import { cpSync, existsSync, mkdirSync, rmSync, writeFileSync } from 'node:fs';
import { join } from 'node:path';

const root = process.cwd();
const client = join(root, 'build', 'client');
const ui = join(root, 'server', 'cmd', 'afc', 'ui');

function resetUI() {
  rmSync(ui, { recursive: true, force: true });
  mkdirSync(ui, { recursive: true });
  writeFileSync(join(ui, '.keep'), '');
}

if (!existsSync(join(client, 'index.html'))) {
  console.warn(
    'build/client not found: building the server without the web client (/app will answer 503).',
  );
}

resetUI();
try {
  if (existsSync(client)) {
    cpSync(client, ui, { recursive: true });
  }
  execSync('go build -o build/afc ./server/cmd/afc', { stdio: 'inherit' });
  console.log('Built build/afc');
} finally {
  resetUI();
}
