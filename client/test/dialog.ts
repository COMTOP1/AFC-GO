const focusable =
  'button:not([disabled]), input:not([disabled]), select:not([disabled]), textarea:not([disabled]), a[href], [tabindex]:not([tabindex="-1"])';

/**
 * The browser's dialog focusing steps: showModal() focuses the first [autofocus]
 * descendant, else the first focusable one. jsdom won't focus inside a closed
 * dialog, so React's autoFocus (which runs before showModal) relies on this.
 */
function focusDialog(dialog: HTMLDialogElement): void {
  const target =
    dialog.querySelector<HTMLElement>('[autofocus]') ??
    dialog.querySelector<HTMLElement>(focusable);
  target?.focus();
}

/** jsdom has no showModal/close; emulate the parts our Modal relies on. */
export function installDialogPolyfill(): void {
  const proto = HTMLDialogElement.prototype;
  if (typeof proto.showModal !== 'function') {
    proto.showModal = function showModal(this: HTMLDialogElement) {
      this.setAttribute('open', '');
      focusDialog(this);
    };
  }
  if (typeof proto.show !== 'function') {
    proto.show = function show(this: HTMLDialogElement) {
      this.setAttribute('open', '');
    };
  }
  if (typeof proto.close !== 'function') {
    proto.close = function close(this: HTMLDialogElement) {
      if (!this.hasAttribute('open')) {
        return;
      }
      this.removeAttribute('open');
      this.dispatchEvent(new Event('close'));
    };
  }
}
