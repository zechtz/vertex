/**
 * Knowing when the session ends, before it does.
 *
 * The token already carries its own expiry, so the app can schedule a warning
 * locally rather than discovering expiry by having a request fail. That makes
 * the reactive recovery in apiFetch a fallback for the case where nobody was
 * watching the tab, instead of the normal path.
 */

interface JwtPayload {
  exp?: number;
}

/**
 * Milliseconds before expiry to warn. Long enough to notice and act, short
 * enough that the prompt is not a constant companion.
 */
export const EXPIRY_WARNING_LEAD_MS = 5 * 60 * 1000;

/**
 * Reads the expiry out of a JWT without verifying it - the server is the only
 * thing that can trust a token; this only needs to know when to ask.
 * Returns null if the token carries no usable expiry.
 */
export function tokenExpiresAt(token: string): Date | null {
  const payload = token.split(".")[1];
  if (!payload) return null;

  try {
    // JWT uses base64url, which atob does not accept directly.
    const normalised = payload.replace(/-/g, "+").replace(/_/g, "/");
    const claims: JwtPayload = JSON.parse(atob(normalised));

    if (typeof claims.exp !== "number") return null;

    return new Date(claims.exp * 1000);
  } catch {
    return null;
  }
}

/**
 * Milliseconds until the warning should appear, floored at zero so a token
 * already inside the warning window prompts immediately. Null when the token
 * has no expiry to schedule against.
 */
export function msUntilWarning(
  token: string,
  now: Date = new Date(),
): number | null {
  const expiry = tokenExpiresAt(token);
  if (!expiry) return null;

  return Math.max(0, expiry.getTime() - now.getTime() - EXPIRY_WARNING_LEAD_MS);
}

/**
 * Whether the token is already past its expiry.
 */
export function isTokenExpired(token: string, now: Date = new Date()): boolean {
  const expiry = tokenExpiresAt(token);
  if (!expiry) return false;

  return expiry.getTime() <= now.getTime();
}
