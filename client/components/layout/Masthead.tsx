import { Link } from 'react-router';

import crest from '../../assets/crest.png';
import faLogo from '../../assets/fa-logo.jpeg';

export function Masthead() {
  return (
    <div className="grid grid-cols-[auto_1fr_auto] items-center gap-3 px-3.5 py-3 md:gap-4 md:px-7 md:py-4.5">
      <Link to="/" className="rounded-full">
        <img
          src={crest}
          alt="AFC Aldermaston home"
          className="size-13 rounded-full bg-white md:size-24"
        />
      </Link>
      <div className="text-center">
        <p className="font-display text-xl leading-none font-extrabold tracking-wider uppercase md:text-[40px]">
          AFC Aldermaston
        </p>
        <p className="mt-1.5 text-[10px] tracking-[0.08em] text-muted uppercase md:text-[13px] md:tracking-[0.14em]">
          Facta Non Verba · <span className="md:hidden">1952</span>
          <span className="hidden md:inline">Founded 1952</span>
        </p>
      </div>
      <img
        src={faLogo}
        alt="The FA Charter Standard"
        className="h-11.5 w-auto rounded-md bg-white p-1 md:h-21"
      />
    </div>
  );
}
