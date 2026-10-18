/**
 * The visitor-count opt-out. Visits are counted under PECR's exemption for
 * statistics cookies, which requires a simple way to object; this cookie is
 * that objection. Mirrors OptOutCookie in server/internal/visitors.
 */

export const OPT_OUT_COOKIE = 'afc_no_visit_count';

const ONE_YEAR = 60 * 60 * 24 * 365;

export function isVisitCountOptedOut(): boolean {
  return document.cookie.split(';').some((c) => c.trim().startsWith(`${OPT_OUT_COOKIE}=`));
}

export function setVisitCountOptOut(optOut: boolean): void {
  const maxAge = optOut ? ONE_YEAR : 0;
  document.cookie = `${OPT_OUT_COOKIE}=1; Max-Age=${maxAge}; Path=/; SameSite=Lax; Secure`;
}
