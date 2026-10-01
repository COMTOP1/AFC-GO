import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';
import { describe, expect, it } from 'vitest';

const root = resolve(__dirname, '..');
const read = (p: string) => readFileSync(resolve(root, p), 'utf8');

describe('served at the site root', () => {
  it('index.html, the router and Vite use root paths', () => {
    const html = read('client/index.html');
    expect(html).toContain('src="/theme-init.js"');
    expect(html).toContain('href="/favicon.png"');
    expect(html).toContain('href="/site.webmanifest"');
    expect(html).not.toContain('/app/');
    expect(read('client/main.tsx')).not.toContain('basename');
    expect(read('vite.config.ts')).toContain("base: '/'");
  });

  it('the web manifest and tile config point at root icons', () => {
    for (const file of ['client/public/site.webmanifest', 'client/public/browserconfig.xml']) {
      const text = read(file);
      expect(text).not.toContain('/public/');
      expect(text).not.toContain('/app/');
    }
  });
});
