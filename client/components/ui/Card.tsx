import { clsx } from 'clsx';
import { useState, type ComponentProps } from 'react';

export function Card({ className, ...props }: ComponentProps<'div'>) {
  return (
    <div
      className={clsx('overflow-hidden rounded-lg border border-line bg-bg', className)}
      {...props}
    />
  );
}

export function CardBody({ className, ...props }: ComponentProps<'div'>) {
  return <div className={clsx('p-4', className)} {...props} />;
}

export interface CardMediaProps {
  /** The uploaded image; missing or empty shows the club gradient instead. */
  src?: string;
  alt: string;
  /** CSS aspect-ratio, default 16 / 9. */
  aspect?: string;
  className?: string;
}

export function CardMedia({ src, alt, aspect = '16 / 9', className }: CardMediaProps) {
  const [failedSrc, setFailedSrc] = useState<string | null>(null);
  const style = { aspectRatio: aspect };

  if (!src || failedSrc === src) {
    return (
      <div
        aria-hidden="true"
        data-fallback=""
        className={clsx('w-full bg-linear-135 from-blue to-red', className)}
        style={style}
      />
    );
  }
  return (
    <img
      src={src}
      alt={alt}
      loading="lazy"
      onError={() => setFailedSrc(src)}
      className={clsx('w-full object-cover', className)}
      style={style}
    />
  );
}
