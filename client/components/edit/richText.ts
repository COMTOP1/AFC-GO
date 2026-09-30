export interface RichTextEditorProps {
  label: string;
  /** Initial HTML (read once when the editor mounts). */
  value: string;
  /** Receives the editor's HTML ('' for an empty document). */
  onChange: (html: string) => void;
  error?: string;
}
