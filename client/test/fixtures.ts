import type { ClubDocument } from '../api/documents';
import type { GalleryImage } from '../api/gallery';
import type { Affiliation, HomeData } from '../api/home';
import type { NewsArticle } from '../api/news';
import type { ContactData, InfoContent } from '../api/pages';
import type { Programme, Season } from '../api/programmes';
import type { Sponsor } from '../api/sponsors';
import type { TeamDetail } from '../api/teams';
import type { SiteInfo, TeamSummary } from '../api/types';
import type { WhatsOnEvent } from '../api/whatson';
import type { MockResponse, MockRoute } from './mockFetch';

export const anonymous: MockResponse = {
  status: 401,
  body: { error: { code: 401, message: 'login required' } },
};

function signedIn(name: string, role: string, canEdit: boolean, canManageGallery: boolean) {
  return {
    body: {
      id: 1,
      name,
      email: 'someone@example.test',
      role,
      permissions: { canEdit, canManageGallery, canManageUsers: false },
    },
  };
}

export const editor: MockResponse = signedIn('Ed Editor', 'Treasurer', true, true);
export const manager: MockResponse = signedIn('Mo Manager', 'Manager', false, false);

export const newsArticle: NewsArticle = {
  id: 1,
  title: 'First team win the cup',
  content: '<p>A late winner at <b>Aldermaston</b> in front of a record crowd.</p>',
  date: '2026-09-28T10:00:00Z',
  imageUrl: '/api/v1/files/news/1',
};
export const newsNoImage: NewsArticle = {
  id: 2,
  title: 'AGM announced',
  content: '<p>All members welcome.</p>',
  date: '2026-09-20T10:00:00Z',
};
export const news: NewsArticle[] = [newsArticle, newsNoImage];

export const event: WhatsOnEvent = {
  id: 3,
  title: 'Presentation evening',
  content: '<p>Clubhouse, all welcome.</p>',
  date: '2026-09-01T10:00:00Z',
  dateOfEvent: '2026-10-16T18:00:00Z',
};
export const events: WhatsOnEvent[] = [event];

export const sponsor: Sponsor = {
  id: 4,
  name: 'Acme Ltd',
  website: 'https://acme.example',
  purpose: 'Kit sponsor',
  team: 'First Team',
  imageUrl: '/api/v1/files/sponsor/4',
};
export const sponsorPlain: Sponsor = { id: 5, name: 'Corner Shop' };
export const sponsors: Sponsor[] = [sponsor, sponsorPlain];

export const affiliation: Affiliation = { id: 6, name: 'The FA', website: 'https://www.thefa.com' };

export const team: TeamSummary = {
  id: 7,
  name: 'First Team',
  league: 'Thames Valley Premier',
  division: 'Division 1',
  isActive: true,
  isYouth: false,
  ages: 99,
};
export const youthTeam: TeamSummary = {
  id: 8,
  name: 'Under 12s',
  isActive: true,
  isYouth: true,
  ages: 12,
};
export const teams: TeamSummary[] = [team, youthTeam];

export const teamDetail: TeamDetail = {
  team: {
    ...team,
    description: 'Our senior side.',
    coach: 'Sam Patel',
    physio: 'Alex Lee',
    leagueTableUrl: 'https://league.example/table',
    fixturesUrl: 'https://league.example/fixtures',
  },
  managers: [{ name: 'Jo Smith', email: 'jo@example.test' }],
  sponsors: [sponsor],
  players: [
    { id: 9, name: 'Chris Captain', position: 'Defender', isCaptain: true },
    { id: 10, name: 'Pat Player', position: 'Forward', isCaptain: false, imageUrl: '/p/10' },
  ],
};
export const youthTeamDetail: TeamDetail = {
  team: youthTeam,
  managers: [],
  sponsors: [],
  players: [],
};

export const galleryImages: GalleryImage[] = [
  { id: 11, caption: 'Cup final', imageUrl: '/api/v1/files/gallery/11' },
  { id: 12, imageUrl: '/api/v1/files/gallery/12' },
  { id: 13, caption: 'Presentation', imageUrl: '/api/v1/files/gallery/13' },
];

export const documents: ClubDocument[] = [
  { id: 14, name: 'Club constitution', fileUrl: '/api/v1/files/document/14' },
  { id: 15, name: 'Safeguarding policy', fileUrl: '/api/v1/files/document/15' },
];

export const seasons: Season[] = [
  { id: 1, name: '2025-26' },
  { id: 2, name: '2026-27' },
];
export const programmes: Programme[] = [
  {
    id: 16,
    name: 'vs Downton',
    date: '2026-09-05T13:00:00Z',
    fileUrl: '/api/v1/files/programme/16',
    season: seasons[1],
  },
  {
    id: 17,
    name: 'vs Marlow',
    date: '2026-03-01T15:00:00Z',
    fileUrl: '/api/v1/files/programme/17',
    season: seasons[0],
  },
  {
    id: 18,
    name: 'Pre-season friendly',
    date: '2025-07-20T13:00:00Z',
    fileUrl: '/api/v1/files/programme/18',
  },
];

export const info: InfoContent = { content: '<h2>Welcome</h2><p>All about the club.</p>' };

export const contact: ContactData = {
  people: [
    { id: 19, name: 'Sam Sec', email: 'sam@example.test', role: 'Club Secretary' },
    {
      id: 20,
      name: 'Cara Chair',
      email: 'cara@example.test',
      role: 'Chairperson',
      imageUrl: '/c/20',
    },
  ],
};

export const home: HomeData = {
  latestNews: newsArticle,
  nextEvent: event,
  sponsors,
  affiliations: [affiliation],
};

export const site: SiteInfo = { year: 2026, visitorCount: 42, version: 'test', teams };

export const resetToken = '5f0b6c1e-1a2b-4c3d-8e9f-000000000001';
export const expiredResetToken = '5f0b6c1e-1a2b-4c3d-8e9f-00000000dead';

/** Every public endpoint with fixture data; override entries per test. */
export function publicRoutes(overrides: Record<string, MockRoute> = {}): Record<string, MockRoute> {
  return {
    '/api/v1/site': { body: site },
    '/api/v1/auth/me': anonymous,
    '/api/v1/home': { body: home },
    '/api/v1/teams': { body: teams },
    [`/api/v1/teams/${team.id}`]: { body: teamDetail },
    [`/api/v1/teams/${youthTeam.id}`]: { body: youthTeamDetail },
    '/api/v1/news': { body: news },
    [`/api/v1/news/${newsArticle.id}`]: { body: newsArticle },
    '/api/v1/whatson?period=future': { body: events },
    '/api/v1/whatson?period=past': { body: [] },
    '/api/v1/whatson?period=all': { body: events },
    [`/api/v1/whatson/${event.id}`]: { body: event },
    '/api/v1/gallery': { body: galleryImages },
    '/api/v1/documents': { body: documents },
    '/api/v1/programmes': { body: programmes },
    '/api/v1/seasons': { body: seasons },
    '/api/v1/sponsors': { body: sponsors },
    '/api/v1/info': { body: info },
    '/api/v1/contact': { body: contact },
    [`/api/v1/auth/reset/${resetToken}`]: { status: 204 },
    [`/api/v1/auth/reset/${expiredResetToken}`]: {
      status: 404,
      body: { error: { code: 404, message: 'reset link is invalid or has expired' } },
    },
    ...overrides,
  };
}
