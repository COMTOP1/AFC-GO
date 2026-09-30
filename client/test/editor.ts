import { act, screen } from '@testing-library/react';
import type { Editor } from '@tiptap/react';

/** The Tiptap editor behind a labelled editable area (Tiptap puts it on the DOM node). */
export async function editorFor(label: string): Promise<Editor> {
  const el = await screen.findByLabelText(label, undefined, { timeout: 3000 });
  return (el as unknown as { editor: Editor }).editor;
}

/** Replaces the editor's content, as if typed, so onUpdate fires. */
export async function typeInEditor(label: string, html: string): Promise<void> {
  const editor = await editorFor(label);
  await act(async () => {
    editor.commands.selectAll();
    editor.commands.insertContent(html);
  });
}
