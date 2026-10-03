import { useState, type ComponentProps, type ReactNode } from 'react';

export interface ImageWithFallbackProps extends Omit<ComponentProps<'img'>, 'src' | 'onError'> {
  src?: string;
  /** An image to show instead (e.g. the crest) when src is missing or broken. */
  fallbackSrc?: string;
  /** Content to show instead when there's no fallbackSrc (e.g. a name tile). */
  fallback?: ReactNode;
}

function safeImageSrc(value?: string): string | undefined {
  if (!value) {
    return undefined;
  }
  try {
    if (value.startsWith('/')) {
      // Disallow protocol-relative forms like //example.com/path.
      if (value.startsWith('//')) {
        return undefined;
      }
      return value;
    }

    const parsed = new URL(value, window.location.origin);
    if (
      parsed.protocol === 'blob:' ||
      parsed.protocol === 'http:' ||
      parsed.protocol === 'https:'
    ) {
      return value;
    }
  } catch {
    return undefined;
  }
  return undefined;
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
  const safeSrc = safeImageSrc(src);
  const safeFallbackSrc = safeImageSrc(fallbackSrc);

  if (!safeSrc || failedSrc === safeSrc) {
    return safeFallbackSrc ? <img src={safeFallbackSrc} alt={alt} {...props} /> : <>{fallback}</>;
  }
  return <img src={safeSrc} alt={alt} onError={() => setFailedSrc(safeSrc)} {...props} />;
}
