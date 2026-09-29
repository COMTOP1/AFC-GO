import { clsx } from 'clsx';

export type ButtonVariant = 'primary' | 'secondary' | 'ghost' | 'danger';
export type ButtonSize = 'sm' | 'md';

const base =
  'inline-flex items-center justify-center gap-2 rounded-md border font-semibold transition-colors disabled:cursor-not-allowed disabled:opacity-50';

const variants: Record<ButtonVariant, string> = {
  primary: 'border-transparent bg-red text-on-red hover:bg-red-hover',
  secondary: 'border-line bg-bg text-ink hover:bg-surface',
  ghost: 'border-transparent bg-transparent text-red hover:bg-surface',
  danger: 'border-red bg-transparent text-red hover:bg-red hover:text-on-red',
};

const sizes: Record<ButtonSize, string> = {
  sm: 'px-2.5 py-1 text-[13px]',
  md: 'px-3.5 py-2 text-sm',
};

export function buttonClasses(
  variant: ButtonVariant = 'primary',
  size: ButtonSize = 'md',
  className?: string,
): string {
  return clsx(base, variants[variant], sizes[size], className);
}
