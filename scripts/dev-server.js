// Runs the Go server with /app proxied to the Vite dev server (yarn dev:client).
import { spawn } from 'node:child_process';

const env = {
  ...process.env,
  AFC_UI_PROXY_URL: process.env.AFC_UI_PROXY_URL ?? 'http://localhost:5173',
};

const server = spawn('go', ['run', './server/cmd/afc'], { stdio: 'inherit', env });
server.on('exit', (code) => process.exit(code ?? 0));
for (const signal of ['SIGINT', 'SIGTERM']) {
  process.on(signal, () => server.kill(signal));
}
