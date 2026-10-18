import { useState } from 'react';
import { Link } from 'react-router';

import { useContact } from '../../api/pages';
import crest from '../../assets/crest.png';
import { ImageWithFallback } from '../../components/page/ImageWithFallback';
import { QueryState } from '../../components/page/QueryState';
import { usePageTitle } from '../../components/page/usePageTitle';
import { Alert } from '../../components/ui/Alert';
import { Button } from '../../components/ui/Button';
import { PageHeader } from '../../components/ui/PageHeader';

// Copied from server/internal/legacy/templates/contact.tmpl.
const MAP_SRC =
  'https://www.google.com/maps/embed?pb=!1m18!1m12!1m3!1d8939.983502968527!2d-1.1655729266615455!3d51.361551652298395!2m3!1f0!2f0!3f0!3m2!1i1024!2i768!4f13.1!3m3!1m2!1s0x4876a0226ede6355%3A0xefc85c4cd2bbf09b!2sAldermaston%20Recreational%20Society!5e1!3m2!1sen!2suk!4v1587086773128!5m2!1sen!2suk';

const MAP_LINK =
  'https://www.google.com/maps/search/?api=1&query=Aldermaston+Recreational+Society+RG26+4QP';

/**
 * The map loads only when asked: Google sets its own cookies when it does,
 * and those need the visitor's consent.
 */
function ClubMap() {
  const [shown, setShown] = useState(false);
  if (shown) {
    return (
      <iframe
        src={MAP_SRC}
        title="Map to Aldermaston Recreational Society"
        referrerPolicy="no-referrer-when-downgrade"
        className="mt-4 h-[400px] w-full rounded-lg border border-line"
      />
    );
  }
  return (
    <div className="mt-4 flex h-[400px] w-full flex-col items-center justify-center gap-3 rounded-lg border border-line bg-surface p-6 text-center">
      <p className="max-w-md text-sm text-muted">
        The map is provided by Google, which may set cookies when it loads. See our{' '}
        <Link to="/privacy" className="text-red underline">
          privacy notice
        </Link>
        .
      </p>
      <Button onClick={() => setShown(true)}>Show map</Button>
      <a
        href={MAP_LINK}
        target="_blank"
        rel="noopener noreferrer"
        className="text-sm text-red underline"
      >
        Open in Google Maps instead
      </a>
    </div>
  );
}

export default function ContactPage() {
  usePageTitle('Contact');
  const contact = useContact();
  return (
    <>
      <PageHeader title="Contact" />
      <QueryState query={contact}>
        {({ displayEmail, people }) => (
          <ul className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {people.map((p) => {
              // Legacy rule: a site-wide contact address, when set, replaces personal ones.
              const email = displayEmail || p.email;
              return (
                <li key={p.id} className="rounded-lg border border-line p-4 text-center">
                  <ImageWithFallback
                    src={p.imageUrl}
                    fallbackSrc={crest}
                    alt=""
                    loading="lazy"
                    className="mx-auto mb-3 size-28 rounded-full border border-line bg-white object-cover"
                  />
                  <h2 className="font-display text-xl font-extrabold uppercase">{p.name}</h2>
                  <p className="text-sm text-muted">{p.role}</p>
                  <a
                    href={`mailto:${email}`}
                    className="mt-1 inline-block text-sm text-red underline"
                  >
                    {email}
                  </a>
                </li>
              );
            })}
          </ul>
        )}
      </QueryState>
      <Alert tone="info" className="mt-8">
        If you&apos;re using a satnav, use the postcode <strong>RG26 4QP</strong> — the postcode
        listed takes you some distance away.
      </Alert>
      <ClubMap />
    </>
  );
}
