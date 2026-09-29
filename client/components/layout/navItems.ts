export interface NavItem {
  label: string;
  /** The legacy page; used until the page is ported. */
  legacyHref: string;
  /** The SPA route (relative to /app) once sub-project 4 ports the page. */
  to?: string;
}

export const navItems: NavItem[] = [
  { label: 'Home', legacyHref: '/', to: '/' },
  { label: 'Teams', legacyHref: '/teams' },
  { label: 'News', legacyHref: '/news' },
  { label: "What's On", legacyHref: '/whatson' },
  { label: 'Gallery', legacyHref: '/gallery' },
  { label: 'Documents', legacyHref: '/documents' },
  { label: 'Programmes', legacyHref: '/programmes' },
  { label: 'Sponsors', legacyHref: '/sponsors' },
  { label: 'Info', legacyHref: '/info' },
  { label: 'Contact', legacyHref: '/contact' },
];
