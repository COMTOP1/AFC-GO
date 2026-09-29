// Compiles server/internal/emails/mjml/<name>.mjml to server/internal/emails/<name>.tmpl.
import { readdirSync, readFileSync, writeFileSync } from 'node:fs';
import { basename, join } from 'node:path';
import mjml2html from 'mjml';

const dir = join(process.cwd(), 'server', 'internal', 'emails');
const src = join(dir, 'mjml');

let failed = false;
for (const file of readdirSync(src).filter((f) => f.endsWith('.mjml'))) {
  const { html, errors } = mjml2html(readFileSync(join(src, file), 'utf8'), {
    validationLevel: 'soft',
  });
  for (const e of errors) {
    console.error(`${file}: line ${e.line}: ${e.message}`);
  }
  if (errors.length > 0) {
    failed = true;
    continue;
  }
  const out = join(dir, `${basename(file, '.mjml')}.tmpl`);
  writeFileSync(out, html);
  console.log(`wrote ${out}`);
}
process.exit(failed ? 1 : 0);
