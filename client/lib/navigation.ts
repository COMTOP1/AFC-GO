/** Full-page navigation (e.g. to a legacy page). A module so tests can mock it. */
export function goTo(url: string): void {
  window.location.assign(url);
}
