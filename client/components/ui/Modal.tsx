import { clsx } from 'clsx';
import { useEffect, useId, useRef, type ReactNode } from 'react';

export interface ModalProps {
  open: boolean;
  onClose: () => void;
  title: string;
  children: ReactNode;
  actions?: ReactNode;
  className?: string;
  /** False: only the dialog's own buttons close it (no Esc, no backdrop click). */
  dismissible?: boolean;
}

/**
 * A native <dialog> shown with showModal(), so the browser traps focus and
 * handles Esc. The parent owns `open`; Esc and backdrop clicks ask it to close.
 */
export function Modal({
  open,
  onClose,
  title,
  children,
  actions,
  className,
  dismissible = true,
}: ModalProps) {
  const ref = useRef<HTMLDialogElement>(null);
  const titleId = useId();
  // True while this component wants the dialog open; stops our own close() from
  // echoing back as a second onClose.
  const wantOpen = useRef(false);

  useEffect(() => {
    const dialog = ref.current;
    if (!open || !dialog) {
      return;
    }
    const previous = document.activeElement instanceof HTMLElement ? document.activeElement : null;
    wantOpen.current = true;
    if (!dialog.open) {
      dialog.showModal();
    }
    return () => {
      wantOpen.current = false;
      if (dialog.open) {
        dialog.close();
      }
      previous?.focus();
    };
  }, [open]);

  return (
    <dialog
      ref={ref}
      aria-labelledby={titleId}
      // React passes a nested dialog's cancel/close up the component tree; only
      // react to this dialog's own.
      onCancel={(e) => {
        if (e.target !== e.currentTarget) {
          return;
        }
        e.preventDefault();
        if (dismissible) {
          onClose();
        }
      }}
      onClose={(e) => {
        if (e.target !== e.currentTarget) {
          return;
        }
        if (wantOpen.current) {
          wantOpen.current = false;
          onClose();
        }
      }}
      onClick={(e) => {
        if (dismissible && e.target === e.currentTarget) {
          onClose();
        }
      }}
      className={clsx(
        'm-auto w-[min(28rem,calc(100vw-2rem))] rounded-xl border border-line bg-bg p-0 text-ink shadow-2xl backdrop:bg-black/50',
        className,
      )}
    >
      {open && (
        <div>
          <div className="border-b border-line px-4 py-3">
            <h2
              id={titleId}
              className="font-display text-xl font-extrabold tracking-wide uppercase"
            >
              {title}
            </h2>
          </div>
          <div className="px-4 py-4">{children}</div>
          {actions && (
            <div className="flex justify-end gap-2 border-t border-line px-4 py-3">{actions}</div>
          )}
        </div>
      )}
    </dialog>
  );
}
