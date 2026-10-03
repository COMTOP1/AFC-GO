import { describe, expect, it } from 'vitest';

import { cleanHtml, plainText } from './sanitise';

describe('cleanHtml', () => {
  it('removes scripts, handlers, javascript: links and iframes', () => {
    const out = cleanHtml(
      '<p onclick="steal()">Hi<script>bad()</script></p>' +
        '<a href="javascript:alert(1)">x</a><iframe src="https://evil.example"></iframe>' +
        '<img src="x" onerror="steal()">',
    );
    expect(out).not.toMatch(/script|onclick|onerror|javascript:|iframe/i);
    expect(out).toContain('<p>Hi</p>');
  });

  it('keeps the formatting the editor produces', () => {
    const html =
      '<h2>Title</h2><p><b>bold</b> <i>it</i> <u>u</u> <strong>s</strong> <em>e</em></p>' +
      '<ul><li>one</li></ul><ol><li>two</li></ol><blockquote>q</blockquote><a href="/news/1">in</a>';
    const out = cleanHtml(html);
    for (const tag of [
      '<h2>',
      '<b>',
      '<i>',
      '<u>',
      '<strong>',
      '<em>',
      '<ul>',
      '<ol>',
      '<li>',
      '<blockquote>',
    ]) {
      expect(out).toContain(tag);
    }
  });

  it('opens off-site links in a new tab, safely', () => {
    const out = cleanHtml('<a href="https://league.example/table">table</a>');
    expect(out).toContain('target="_blank"');
    expect(out).toContain('rel="noopener noreferrer"');
  });

  it('leaves email and phone links in the same tab', () => {
    const out = cleanHtml('<a href="mailto:sec@example.test">mail</a><a href="tel:0123">call</a>');
    expect(out).not.toContain('target=');
  });

  it('leaves same-site links alone', () => {
    const out = cleanHtml('<a href="/news/1">news</a>');
    expect(out).not.toContain('target=');
  });
});

describe('plainText', () => {
  it('strips tags and separates blocks with spaces', () => {
    expect(plainText('<p>A late <b>winner</b>.</p><p>Record crowd.</p>', 200)).toBe(
      'A late winner. Record crowd.',
    );
  });

  it('truncates at a word boundary with an ellipsis', () => {
    expect(plainText('<p>one two three four five</p>', 12)).toBe('one two…');
  });

  it('never runs scripts or keeps markup', () => {
    expect(plainText('<img src=x onerror="steal()"><p>safe</p>', 50)).toBe('safe');
  });
});
