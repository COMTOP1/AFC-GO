import { EditorContent, useEditor, type Editor } from '@tiptap/react';
import StarterKit from '@tiptap/starter-kit';
import { clsx } from 'clsx';
import { useEffect, useId, useState, type FormEvent } from 'react';

import { normalizeLink } from '../../lib/links';
import { Button } from '../ui/Button';
import { buttonClasses } from '../ui/buttonStyles';
import { Input } from '../ui/controls';
import { Field } from '../ui/Field';
import { Modal } from '../ui/Modal';
import type { RichTextEditorProps } from './richText';

interface Tool {
  label: string;
  text: string;
  isActive?: (e: Editor) => boolean;
  run: (e: Editor) => void;
}

const tools: Tool[] = [
  {
    label: 'Bold',
    text: 'B',
    isActive: (e) => e.isActive('bold'),
    run: (e) => e.chain().focus().toggleBold().run(),
  },
  {
    label: 'Italic',
    text: 'I',
    isActive: (e) => e.isActive('italic'),
    run: (e) => e.chain().focus().toggleItalic().run(),
  },
  {
    label: 'Underline',
    text: 'U',
    isActive: (e) => e.isActive('underline'),
    run: (e) => e.chain().focus().toggleUnderline().run(),
  },
  {
    label: 'Strikethrough',
    text: 'S',
    isActive: (e) => e.isActive('strike'),
    run: (e) => e.chain().focus().toggleStrike().run(),
  },
  {
    label: 'Heading 2',
    text: 'H2',
    isActive: (e) => e.isActive('heading', { level: 2 }),
    run: (e) => e.chain().focus().toggleHeading({ level: 2 }).run(),
  },
  {
    label: 'Heading 3',
    text: 'H3',
    isActive: (e) => e.isActive('heading', { level: 3 }),
    run: (e) => e.chain().focus().toggleHeading({ level: 3 }).run(),
  },
  {
    label: 'Bullet list',
    text: '•',
    isActive: (e) => e.isActive('bulletList'),
    run: (e) => e.chain().focus().toggleBulletList().run(),
  },
  {
    label: 'Numbered list',
    text: '1.',
    isActive: (e) => e.isActive('orderedList'),
    run: (e) => e.chain().focus().toggleOrderedList().run(),
  },
  {
    label: 'Quote',
    text: '❝',
    isActive: (e) => e.isActive('blockquote'),
    run: (e) => e.chain().focus().toggleBlockquote().run(),
  },
  { label: 'Horizontal rule', text: '―', run: (e) => e.chain().focus().setHorizontalRule().run() },
  { label: 'Undo', text: '↶', run: (e) => e.chain().focus().undo().run() },
  { label: 'Redo', text: '↷', run: (e) => e.chain().focus().redo().run() },
];

const toolClass = (active: boolean) =>
  clsx(buttonClasses('ghost', 'sm', 'min-w-8 px-2 text-ink'), active && 'bg-surface text-red');

function attributes(label: string, error: string | undefined, errorId: string) {
  return {
    role: 'textbox',
    'aria-multiline': 'true',
    'aria-label': label,
    ...(error ? { 'aria-describedby': errorId, 'aria-invalid': 'true' } : {}),
    class:
      'prose min-h-48 rounded-b-md border border-line bg-field px-3 py-2 text-ink focus:outline-none focus-visible:border-blue',
  };
}

export default function RichTextEditorImpl({ label, value, onChange, error }: RichTextEditorProps) {
  const errorId = useId();
  const [linkOpen, setLinkOpen] = useState(false);
  const [linkValue, setLinkValue] = useState('');
  const [linkError, setLinkError] = useState<string | undefined>();

  const editor = useEditor({
    extensions: [
      StarterKit.configure({
        heading: { levels: [2, 3] },
        code: false,
        codeBlock: false,
        link: { openOnClick: false, autolink: true, protocols: ['http', 'https', 'mailto'] },
      }),
    ],
    content: value,
    immediatelyRender: true,
    // Re-render on every change so the toolbar's pressed states stay current.
    shouldRerenderOnTransaction: true,
    editorProps: { attributes: attributes(label, error, errorId) },
    onUpdate: ({ editor: e }) => onChange(e.isEmpty ? '' : e.getHTML()),
  });

  // Keep the label and error wired to the editable area for screen readers.
  useEffect(() => {
    editor?.setOptions({ editorProps: { attributes: attributes(label, error, errorId) } });
  }, [editor, label, error, errorId]);

  if (!editor) {
    return null;
  }

  function openLink() {
    setLinkValue(editor?.getAttributes('link').href ?? '');
    setLinkError(undefined);
    setLinkOpen(true);
  }

  function applyLink(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const href = normalizeLink(linkValue);
    if (!href) {
      setLinkError('Enter a web address (https://…) or an email link (mailto:…).');
      return;
    }
    editor?.chain().focus().extendMarkRange('link').setLink({ href }).run();
    setLinkOpen(false);
  }

  return (
    <div className="flex flex-col gap-1.5">
      <span className="text-sm font-semibold">{label}</span>
      <div
        role="toolbar"
        aria-label={`${label} formatting`}
        className="flex flex-wrap gap-1 rounded-t-md border border-b-0 border-line bg-surface p-1"
      >
        {tools.map((tool) => {
          const active = tool.isActive?.(editor) ?? false;
          return (
            <button
              key={tool.label}
              type="button"
              aria-label={tool.label}
              aria-pressed={tool.isActive ? active : undefined}
              title={tool.label}
              onClick={() => tool.run(editor)}
              className={toolClass(active)}
            >
              {tool.text}
            </button>
          );
        })}
        <button
          type="button"
          aria-label="Link"
          aria-pressed={editor.isActive('link')}
          title="Link"
          onClick={openLink}
          className={toolClass(editor.isActive('link'))}
        >
          🔗
        </button>
        <button
          type="button"
          aria-label="Remove link"
          title="Remove link"
          disabled={!editor.isActive('link')}
          onClick={() => editor.chain().focus().unsetLink().run()}
          className={toolClass(false)}
        >
          ⛓
        </button>
      </div>
      <EditorContent editor={editor} />
      {error && (
        <p id={errorId} className="text-sm text-red">
          {error}
        </p>
      )}
      <Modal open={linkOpen} onClose={() => setLinkOpen(false)} title="Add a link">
        <form onSubmit={applyLink} className="flex flex-col gap-4" noValidate>
          <Field
            label="Link address"
            error={linkError}
            help="For example https://thefa.com or mailto:someone@example.com"
          >
            <Input value={linkValue} onChange={(e) => setLinkValue(e.target.value)} autoFocus />
          </Field>
          <div className="flex justify-end gap-2">
            <Button variant="secondary" onClick={() => setLinkOpen(false)}>
              Cancel
            </Button>
            <Button type="submit">Add link</Button>
          </div>
        </form>
      </Modal>
    </div>
  );
}
