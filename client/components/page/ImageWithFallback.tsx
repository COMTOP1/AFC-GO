import { useState, type ComponentProps, type ReactNode } from 'react';

export interface ImageWithFallbackProps extends Omit<ComponentProps<'img'>, 'src' | 'onError'> {
  src?: string;
  /** An image to show instead (e.g. the crest) when src is missing or broken. */
  fallbackSrc?: string;
  /** Content to show instead when there's no fallbackSrc (e.g. a name tile). */
  fallback?: ReactNode;
}

/** An <img> that never shows the browser's broken-image state. */
export function ImageWithFallback({
  src,
  fallbackSrc,
  fallback = null,
  alt = '',
  ...props
}: ImageWithFallbackProps) {
  const [failedSrc, setFailedSrc] = useState<string | null>(null);
  if (!src || failedSrc === src) {
    return fallbackSrc ? <img src={fallbackSrc} alt={alt} {...props} /> : <>{fallback}</>;
  }
  return <img src={src} alt={alt} onError={() => setFailedSrc(src)} {...props} />;
}
