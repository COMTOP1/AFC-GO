import axe from 'axe-core';

/**
 * Runs axe over node and returns "rule: selectors" strings, so a failing
 * expect(...).toEqual([]) says exactly what broke. Colour contrast is off
 * because jsdom does not compute styles; the palette was checked by hand.
 */
export async function axeViolations(node: Element): Promise<string[]> {
  const results = await axe.run(node, {
    // The Contact map is a cross-origin frame axe can't audit in jsdom.
    iframes: false,
    rules: { 'color-contrast': { enabled: false } },
  });
  return results.violations.map(
    (v) => `${v.id}: ${v.nodes.map((n) => n.target.join(' ')).join(', ')}`,
  );
}
