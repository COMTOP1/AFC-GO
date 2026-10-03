import { act, fireEvent, screen, within } from '@testing-library/react';
import { useState } from 'react';
import { describe, expect, it } from 'vitest';

import { normalizeLink } from '../../lib/links';
import { editorFor, typeInEditor } from '../../test/editor';
import { publicRoutes } from '../../test/fixtures';
import { mockFetch } from '../../test/mockFetch';
import { renderWithProviders } from '../../test/render';
import { RichTextEditor } from './RichTextEditor';

function Harness({ initial = '' }: { initial?: string }) {
  const [html, setHtml] = useState(initial);
  return (
    <>
      <RichTextEditor label="Content" value={initial} onChange={setHtml} />
      <output data-testid="html">{html}</output>
    </>
  );
}

function renderEditor(initial?: string) {
  mockFetch(publicRoutes());
  renderWithProviders(<Harness initial={initial} />);
}

describe('normalizeLink', () => {
  it('accepts web and email links and adds https:// to bare domains', () => {
    expect(normalizeLink('https://league.example/table')).toBe('https://league.example/table');
    expect(normalizeLink('  league.example  ')).toBe('https://league.example');
    expect(normalizeLink('mailto:sec@example.test')).toBe('mailto:sec@example.test');
  });

  it('rejects other schemes and junk', () => {
    for (const bad of ['javascript:alert(1)', 'data:text/html,x', '', 'not a url']) {
      expect(normalizeLink(bad)).toBeNull();
    }
  });
});

describe('RichTextEditor', () => {
  it('loads existing HTML and emits HTML as you type', async () => {
    renderEditor('<p>Hello</p>');
    const editor = await editorFor('Content');
    expect(editor.getHTML()).toBe('<p>Hello</p>');
    await typeInEditor('Content', '<p>New <strong>text</strong></p>');
    expect(screen.getByTestId('html')).toHaveTextContent('<p>New <strong>text</strong></p>');
  });

  it('emits an empty string for an empty document', async () => {
    renderEditor('<p>Hello</p>');
    const editor = await editorFor('Content');
    await act(async () => {
      editor.commands.clearContent(true);
    });
    expect(screen.getByTestId('html')).toHaveTextContent(/^$/);
  });

  it('has a labelled toolbar whose buttons toggle formatting', async () => {
    renderEditor('<p>Hello</p>');
    const editor = await editorFor('Content');
    const toolbar = screen.getByRole('toolbar', { name: 'Content formatting' });
    await act(async () => {
      editor.commands.selectAll();
    });
    const bold = within(toolbar).getByRole('button', { name: 'Bold' });
    expect(bold).toHaveAttribute('aria-pressed', 'false');
    fireEvent.click(bold);
    expect(bold).toHaveAttribute('aria-pressed', 'true');
    expect(screen.getByTestId('html')).toHaveTextContent('<p><strong>Hello</strong></p>');
    fireEvent.click(within(toolbar).getByRole('button', { name: 'Bullet list' }));
    expect(screen.getByTestId('html')).toHaveTextContent(
      '<ul><li><p><strong>Hello</strong></p></li></ul>',
    );
  });

  it('aligns paragraphs left, centre and right', async () => {
    renderEditor('<p style="text-align: right">Hello</p>');
    const editor = await editorFor('Content');
    const toolbar = screen.getByRole('toolbar', { name: 'Content formatting' });
    await act(async () => {
      editor.commands.selectAll();
    });
    const left = within(toolbar).getByRole('button', { name: 'Align left' });
    const centre = within(toolbar).getByRole('button', { name: 'Align centre' });
    expect(within(toolbar).getByRole('button', { name: 'Align right' })).toHaveAttribute(
      'aria-pressed',
      'true',
    );
    fireEvent.click(centre);
    expect(centre).toHaveAttribute('aria-pressed', 'true');
    expect(screen.getByTestId('html')).toHaveTextContent(
      '<p style="text-align: center;">Hello</p>',
    );
    fireEvent.click(left);
    expect(left).toHaveAttribute('aria-pressed', 'true');
    expect(screen.getByTestId('html')).toHaveTextContent('<p style="text-align: left;">Hello</p>');
  });

  it('adds a link through the link dialog, fixing a bare domain and refusing javascript:', async () => {
    renderEditor('<p>Table</p>');
    const editor = await editorFor('Content');
    await act(async () => {
      editor.commands.selectAll();
    });
    fireEvent.click(screen.getByRole('button', { name: 'Link' }));
    const dialog = screen.getByRole('dialog', { name: 'Add a link' });
    const input = within(dialog).getByLabelText('Link address');
    fireEvent.change(input, { target: { value: 'javascript:alert(1)' } });
    fireEvent.click(within(dialog).getByRole('button', { name: 'Add link' }));
    expect(input).toHaveAccessibleDescription(
      'Enter a web address (https://…) or an email link (mailto:…).',
    );
    fireEvent.change(input, { target: { value: 'league.example' } });
    fireEvent.click(within(dialog).getByRole('button', { name: 'Add link' }));
    expect(screen.queryByRole('dialog', { name: 'Add a link' })).toBeNull();
    expect(screen.getByTestId('html').textContent).toContain('href="https://league.example"');
  });

  it('shows an error under the editor', async () => {
    mockFetch(publicRoutes());
    renderWithProviders(
      <RichTextEditor label="Content" value="" onChange={() => {}} error="content is too long" />,
    );
    const area = await screen.findByLabelText('Content', undefined, { timeout: 3000 });
    expect(area).toHaveAccessibleDescription('content is too long');
  });
});
