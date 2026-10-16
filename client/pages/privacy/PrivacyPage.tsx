import { useState } from 'react';
import { Link } from 'react-router';

import { useSite } from '../../api/queries';
import { usePageTitle } from '../../components/page/usePageTitle';
import { Checkbox } from '../../components/ui/Checkbox';
import { PageHeader } from '../../components/ui/PageHeader';
import { Table, TBody, Td, Th, THead, Tr } from '../../components/ui/Table';
import { isVisitCountOptedOut, setVisitCountOptOut } from '../../lib/visitCount';

const LAST_UPDATED = '16 October 2026';

const cookies = [
  {
    name: 'Session cookie',
    purpose: 'Keeps club officials signed in. Only set when you sign in.',
    lasts: '24 hours, or 31 days if you tick "Remember me"',
    kind: 'Strictly necessary',
  },
  {
    name: '_csrf',
    purpose: 'Protects forms on this site from being submitted by other websites.',
    lasts: '24 hours',
    kind: 'Strictly necessary',
  },
  {
    name: 'afc_aldermaston_visited',
    purpose: 'Counts unique visitors to the site, so we only count you once a day.',
    lasts: '24 hours',
    kind: 'Statistics (you can opt out below)',
  },
  {
    name: 'afc_no_visit_count',
    purpose: 'Remembers that you opted out of the visitor count.',
    lasts: '1 year',
    kind: 'Strictly necessary',
  },
  {
    name: 'afc-theme (stored in your browser, not a cookie)',
    purpose: 'Remembers whether you chose the light or dark theme.',
    lasts: 'Until you clear it',
    kind: 'Strictly necessary',
  },
];

function VisitCountOptOut() {
  const [optedOut, setOptedOut] = useState(isVisitCountOptedOut);
  return (
    <Checkbox
      label="Don't count my visits"
      checked={optedOut}
      onChange={(e) => {
        setVisitCountOptOut(e.target.checked);
        setOptedOut(e.target.checked);
      }}
    />
  );
}

function ContactLink() {
  const email = useSite().data?.displayEmail;
  if (email) {
    return <a href={`mailto:${email}`}>{email}</a>;
  }
  return (
    <>
      the Club Secretary through our <Link to="/contact">contact page</Link>
    </>
  );
}

export default function PrivacyPage() {
  usePageTitle('Privacy notice');
  return (
    <>
      <PageHeader title="Privacy notice" subtitle={`Last updated ${LAST_UPDATED}`} />
      <div className="prose max-w-3xl">
        <h2>Who we are</h2>
        <p>
          AFC Aldermaston (&ldquo;the club&rdquo;, &ldquo;we&rdquo;) runs this website and is the
          controller of the personal information it holds. To ask anything about your information,
          contact <ContactLink />.
        </p>

        <h2>What we hold and why</h2>
        <h3>Club officials and volunteers with a website account</h3>
        <p>
          Your name, email address, phone number (if given), role, team (for managers), photo (if
          given) and a securely hashed password. We use these to let you manage the website, to show
          club officials on the contact page, and to show each team&apos;s managers on its team
          page. Our lawful basis is our legitimate interest in running the club. We use your photo
          only with your consent; you can remove it at any time from your account page.
        </p>
        <h3>Players</h3>
        <p>
          Your name, date of birth, team, position, whether you are captain, and a photo (if given).
          We use these to organise teams and to show senior squads on the website. Our lawful basis
          is our legitimate interest in running the club and its teams. We use photos only with
          consent, given by a parent or guardian for anyone under 18.
        </p>
        <h3>Team coaches and physios</h3>
        <p>
          The name of each team&apos;s coach and physio, shown on its team page. Our lawful basis is
          our legitimate interest in letting players and families know who runs each team.
        </p>
        <h3>Everyone who visits the site</h3>
        <p>
          We count unique visitors. To do this we hold your IP address in memory for no more than
          two hours; it is never written to our database. Our servers also keep technical logs
          (which can include your IP address, browser type and, for a failed sign-in, the email
          address used) to keep the site secure and working. To stop password guessing, failed
          sign-ins are also counted against the email address and IP address for 15 minutes. Our
          lawful basis for all of these is our legitimate interest in running a secure website.
        </p>

        <h2>Children</h2>
        <ul>
          <li>Youth team squads are never shown on public pages.</li>
          <li>Senior teams are for adults only: nobody under 18 can be added to one.</li>
          <li>
            Player photos of youth team players, and of anyone under 18, are never shown on the
            site.
          </li>
          <li>
            Dates of birth are seen only by club officials who manage player records, and by a
            team&apos;s own manager.
          </li>
          <li>
            Photos of children in the gallery or news are only used with a parent or guardian&apos;s
            consent, checked by our Safeguarding Officer before upload.
          </li>
          <li>A parent or guardian can make any of the requests below on a child&apos;s behalf.</li>
        </ul>

        <h2>Who can see it</h2>
        <ul>
          <li>
            <strong>Anyone:</strong> the names, roles, photos and contact emails of the club&apos;s
            main officials (Chairperson, Club Secretary, Safeguarding Officer, Treasurer, League
            Secretary and Programme Editor) on the contact page; each team&apos;s managers&apos;
            names and email addresses, coach and physio on its team page; senior players&apos;
            names, positions, captaincy and photos on team pages.
          </li>
          <li>
            <strong>Signed-in club officials:</strong> the players list, with dates of birth limited
            as described above. Programme editors see only players&apos; names, positions, teams,
            captaincy and photos of adults.
          </li>
          <li>
            <strong>Officials who manage accounts:</strong> account details of other officials.
          </li>
        </ul>
        <p>
          The website is provided and run for us by BSWDI, which hosts it, its database, file
          storage and email on servers it manages, and monitors it for errors. BSWDI handles your
          information only on our instructions. We never sell your information or use it for
          advertising.
        </p>

        <h2>How long we keep it</h2>
        <ul>
          <li>
            Accounts are kept until the official leaves their role and the account is removed.
          </li>
          <li>Player records are kept while you play for the club and removed when you leave.</li>
          <li>Removing an account or player also deletes their photo.</li>
          <li>Emails we send are deleted from our mail server after 30 days.</li>
          <li>
            Backups are kept for 30 days, so anything we delete is gone from our backups within 30
            days too.
          </li>
          <li>
            Links to set or reset a password expire within 48 hours, or as soon as they are used.
          </li>
          <li>Technical logs and error monitoring data are deleted after 60 days.</li>
        </ul>

        <h2>Cookies</h2>
        <p>
          We don&apos;t use advertising or tracking cookies. These are the cookies and browser
          storage this site uses:
        </p>
      </div>
      <div className="my-4 max-w-3xl">
        <Table>
          <THead>
            <Tr>
              <Th>Name</Th>
              <Th>What it does</Th>
              <Th>How long</Th>
              <Th>Type</Th>
            </Tr>
          </THead>
          <TBody>
            {cookies.map((c) => (
              <Tr key={c.name}>
                <Td className="font-mono text-xs">{c.name}</Td>
                <Td>{c.purpose}</Td>
                <Td>{c.lasts}</Td>
                <Td>{c.kind}</Td>
              </Tr>
            ))}
          </TBody>
        </Table>
      </div>
      <div className="prose max-w-3xl">
        <p>
          Our <Link to="/contact">contact page</Link> can show a Google map. It only loads when you
          press &ldquo;Show map&rdquo;, and Google may then set its own cookies; see{' '}
          <a href="https://policies.google.com/privacy" target="_blank" rel="noopener noreferrer">
            Google&apos;s privacy policy
          </a>
          .
        </p>
        <h3>Opting out of the visitor count</h3>
        <p>
          Tick this to stop your visits being counted on this browser. We also don&apos;t count you
          if your browser sends a Global Privacy Control signal.
        </p>
        <VisitCountOptOut />

        <h2>Your rights</h2>
        <p>Under UK data protection law you can ask us to:</p>
        <ul>
          <li>give you a copy of the information we hold about you;</li>
          <li>correct anything that is wrong;</li>
          <li>delete your information;</li>
          <li>restrict or stop using it, or object to how we use it;</li>
          <li>give your information to you in a format you can take elsewhere.</li>
        </ul>
        <p>
          Where we rely on your consent, such as for photos, you can withdraw it at any time.
          Contact <ContactLink /> and we will reply within one month.
        </p>
        <p>
          If you are unhappy with how we have handled your information, you can complain to the
          Information Commissioner&apos;s Office at{' '}
          <a href="https://ico.org.uk/make-a-complaint/" target="_blank" rel="noopener noreferrer">
            ico.org.uk
          </a>{' '}
          or on 0303 123 1113.
        </p>

        <h2>Changes to this notice</h2>
        <p>
          We will update this page if how we use your information changes. The date at the top shows
          when it last changed.
        </p>
      </div>
    </>
  );
}
