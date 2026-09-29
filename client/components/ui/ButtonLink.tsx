import type { ComponentProps } from 'react';
import { Link } from 'react-router';

import { buttonClasses, type ButtonSize, type ButtonVariant } from './buttonStyles';

export interface ButtonLinkProps extends Omit<ComponentProps<'a'>, 'href'> {
  /** An SPA route (relative to /app). */
  to?: string;
  /** A full-page URL, e.g. a legacy page. */
  href?: string;
  variant?: ButtonVariant;
  size?: ButtonSize;
}

export function ButtonLink({ to, href, variant, size, className, ...props }: ButtonLinkProps) {
  const classes = buttonClasses(variant, size, className);
  if (to !== undefined) {
    return <Link to={to} className={classes} {...props} />;
  }
  return <a href={href} className={classes} {...props} />;
}
