import { useEffect, useRef, useState } from 'react';

import {
  IMAGE_TYPE_MESSAGE,
  IMAGE_TYPES,
  isAcceptedImage,
  previewUrl,
  type ImageValue,
} from '../../lib/images';
import { FullImage } from '../page/FullImage';
import { Checkbox } from '../ui/Checkbox';
import { FileInput } from '../ui/controls';
import { Field } from '../ui/Field';

export interface ImageFieldProps {
  label: string;
  /** The image already saved, if any. */
  currentUrl?: string;
  required?: boolean;
  /** Offer "Remove image" when there's a current image (edit forms). */
  allowRemove?: boolean;
  value: ImageValue;
  onChange: (value: ImageValue) => void;
  error?: string;
}

export function ImageField({
  label,
  currentUrl,
  required,
  allowRemove,
  value,
  onChange,
  error,
}: ImageFieldProps) {
  const inputRef = useRef<HTMLInputElement>(null);
  const [preview, setPreview] = useState<string | null>(null);
  const [typeError, setTypeError] = useState<string | undefined>();

  useEffect(() => {
    if (!preview) {
      return;
    }
    return () => URL.revokeObjectURL(preview);
  }, [preview]);

  function pick(file: File | null) {
    setTypeError(undefined);
    if (file && !isAcceptedImage(file)) {
      setTypeError(IMAGE_TYPE_MESSAGE);
      setPreview(null);
      if (inputRef.current) {
        inputRef.current.value = '';
      }
      onChange({ file: null, remove: value.remove });
      return;
    }
    setPreview(file ? previewUrl(file) : null);
    onChange({ file, remove: file ? false : value.remove });
  }

  const shown = preview ?? (value.remove ? undefined : currentUrl);
  return (
    <div className="space-y-2">
      <FullImage src={shown} alt={preview ? 'Preview of the new image' : ''} className="max-w-xl" />
      <Field label={label} error={typeError ?? error}>
        <FileInput
          ref={inputRef}
          accept={IMAGE_TYPES}
          required={required}
          onChange={(e) => pick(e.target.files?.[0] ?? null)}
        />
      </Field>
      {allowRemove && currentUrl && !value.file && (
        <Checkbox
          label="Remove image"
          checked={value.remove}
          onChange={(e) => onChange({ file: null, remove: e.target.checked })}
        />
      )}
    </div>
  );
}
