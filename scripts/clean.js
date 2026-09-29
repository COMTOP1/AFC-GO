// Removes build output and resets the Go embed target to just .keep.
import { mkdirSync, rmSync, writeFileSync } from 'node:fs';
import { join } from 'node:path';

const root = process.cwd();
rmSync(join(root, 'build'), { recursive: true, force: true });

const ui = join(root, 'server', 'cmd', 'afc', 'ui');
rmSync(ui, { recursive: true, force: true });
mkdirSync(ui, { recursive: true });
writeFileSync(join(ui, '.keep'), '');
