import { useEffect, useRef, type RefObject } from 'react';

/** While open, calls onDismiss on a mousedown outside ref or on Esc anywhere. */
export function useDismiss(
  ref: RefObject<HTMLElement | null>,
  open: boolean,
  onDismiss: () => void,
): void {
  const callback = useRef(onDismiss);
  useEffect(() => {
    callback.current = onDismiss;
  });

  useEffect(() => {
    if (!open) {
      return;
    }
    function onPointer(e: MouseEvent) {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        callback.current();
      }
    }
    function onKey(e: KeyboardEvent) {
      if (e.key === 'Escape') {
        callback.current();
      }
    }
    document.addEventListener('mousedown', onPointer);
    document.addEventListener('keydown', onKey);
    return () => {
      document.removeEventListener('mousedown', onPointer);
      document.removeEventListener('keydown', onKey);
    };
  }, [open, ref]);
}
