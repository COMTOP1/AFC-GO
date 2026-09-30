export interface NavItem {
  label: string;
  /** The legacy page; used until the page is ported. */
  legacyHref: string;
  /** The SPA route (relative to /app); every public page is ported (4a). */
  to?: string;
}

export const navItems: NavItem[] = [
  { label: 'Home', legacyHref: '/', to: '/' },
  { label: 'Teams', legacyHref: '/teams', to: '/teams' },
  { label: 'News', legacyHref: '/news', to: '/news' },
  { label: "What's On", legacyHref: '/whatson', to: '/whatson' },
  { label: 'Gallery', legacyHref: '/gallery', to: '/gallery' },
  { label: 'Documents', legacyHref: '/documents', to: '/documents' },
  { label: 'Programmes', legacyHref: '/programmes', to: '/programmes' },
  { label: 'Sponsors', legacyHref: '/sponsors', to: '/sponsors' },
  { label: 'Info', legacyHref: '/info', to: '/info' },
  { label: 'Contact', legacyHref: '/contact', to: '/contact' },
];
