/** The shape of a password-reset token (the server issues UUIDs). */
export const RESET_TOKEN = /^[A-Za-z0-9-]{1,100}$/;

/**
 * The server's reset URL ("/reset/<token>") as an in-app route, or null when
 * it isn't one. Reset links use the same path as the app's reset page.
 */
export function appResetPath(url: string): string | null {
  const match = /^\/reset\/([^/]+)$/.exec(url);
  return match && RESET_TOKEN.test(match[1]) ? `/reset/${match[1]}` : null;
}
