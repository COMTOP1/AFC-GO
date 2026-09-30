import crest from '../../assets/crest.png';
import { ImageWithFallback } from './ImageWithFallback';

/** A small round photo for table rows; the crest when there's none (or it's hidden). */
export function Thumb({ src }: { src?: string }) {
  return (
    <ImageWithFallback
      src={src}
      alt=""
      fallbackSrc={crest}
      loading="lazy"
      className="size-10 rounded-full border border-line bg-white object-cover"
    />
  );
}
