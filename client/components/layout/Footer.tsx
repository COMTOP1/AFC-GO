import { useSite } from '../../api/queries';
import { useAuth } from '../../auth/useAuth';

const socials = [
  {
    label: 'AFC Aldermaston on Facebook',
    href: 'https://www.facebook.com/AFC-Aldermaston-114651238068/',
    path: 'M14 8h3V4h-3c-2.8 0-5 2.2-5 5v2H7v4h2v9h4v-9h3l1-4h-4V9c0-.6.4-1 1-1z',
  },
  {
    label: 'AFC Aldermaston on X',
    href: 'https://x.com/afcaldermaston',
    path: 'M17.8 3h3.1l-6.8 7.8L22 21h-6.2l-4.9-6.4L5.3 21H2.2l7.3-8.3L2 3h6.4l4.4 5.8L17.8 3zm-1.1 16.2h1.7L7.4 4.7H5.6l11.1 14.5z',
  },
];

export function Footer() {
  const site = useSite();
  const { user } = useAuth();
  const year = site.data?.year ?? new Date().getFullYear();

  return (
    <footer className="mt-12 bg-footer text-white">
      <div className="mx-auto flex w-full max-w-6xl flex-wrap items-center justify-between gap-4 px-4 py-5 text-sm md:px-6">
        <div>
          <p>
            © 2020–{year} AFC Aldermaston · Website provided by{' '}
            <a
              href="https://bswdi.co.uk"
              target="_blank"
              rel="noopener noreferrer"
              className="font-bold underline"
            >
              BSWDI
            </a>
          </p>
          {user && site.data && (
            <p className="mt-1 opacity-75">
              Visitor count: {site.data.visitorCount.toLocaleString('en-GB')}
            </p>
          )}
        </div>
        <ul className="flex gap-2.5">
          {socials.map((s) => (
            <li key={s.href}>
              <a
                href={s.href}
                target="_blank"
                rel="noopener noreferrer"
                aria-label={s.label}
                className="grid size-8 place-items-center rounded-full bg-white/15 hover:bg-white/25"
              >
                <svg
                  width="16"
                  height="16"
                  viewBox="0 0 24 24"
                  fill="currentColor"
                  aria-hidden="true"
                >
                  <path d={s.path} />
                </svg>
              </a>
            </li>
          ))}
        </ul>
      </div>
    </footer>
  );
}
