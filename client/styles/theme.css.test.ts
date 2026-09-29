/// <reference types="node" />
import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';

import { describe, expect, it } from 'vitest';

// Read from disk: the Vitest config (css: false) blanks CSS imports, even with ?raw.
const css = readFileSync(resolve(process.cwd(), 'client/styles/app.css'), 'utf8');

describe('app.css', () => {
  it('pins native widgets to light when the light theme is forced', () => {
    expect(css).toMatch(/\[data-theme='light'\]\s*\{\s*color-scheme:\s*light;/);
  });
});
