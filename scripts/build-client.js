// Lints (unless BUILD_CLIENT_SKIP_LINT=true), type-checks and builds the client into build/client.
import { execSync } from 'node:child_process';

if (process.env.BUILD_CLIENT_SKIP_LINT !== 'true') {
  execSync('yarn lint:client', { stdio: 'inherit' });
}
execSync('tsc -b && vite build', { stdio: 'inherit' });
