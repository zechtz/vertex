import { useEffect, useState } from "react";
import { Clock } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useAuth } from "@/contexts/AuthContext";
import { SystemApi } from "@/services/systemApi";
import { tokenExpiresAt } from "@/services/sessionExpiry";

function formatRemaining(ms: number): string {
  if (ms <= 0) return "less than a minute";

  const minutes = Math.ceil(ms / 60000);
  return minutes === 1 ? "1 minute" : `${minutes} minutes`;
}

/**
 * Warns shortly before the session ends and offers to extend it.
 *
 * This is the half that keeps expiry from happening at all. Without it the
 * first sign of trouble is an action that fails, which the expired-session
 * prompt then recovers from - correct, but it still costs the user their
 * password. One click here costs nothing.
 */
export function SessionExpiringDialog() {
  const { sessionExpiring, token, renewSession, dismissExpiryWarning, logout } =
    useAuth();
  const [renewing, setRenewing] = useState(false);
  const [error, setError] = useState("");
  const [remaining, setRemaining] = useState("");

  // Count down while the prompt is open, so "a few minutes" does not sit there
  // claiming the same thing after four of them have passed.
  useEffect(() => {
    if (!sessionExpiring || !token) return;

    const expiry = tokenExpiresAt(token);
    if (!expiry) return;

    const tick = () => setRemaining(formatRemaining(expiry.getTime() - Date.now()));
    tick();

    const timer = setInterval(tick, 30000);
    return () => clearInterval(timer);
  }, [sessionExpiring, token]);

  if (!sessionExpiring) return null;

  const renew = async () => {
    setRenewing(true);
    setError("");

    try {
      const auth = await SystemApi.refreshSession();
      renewSession(auth.token);
    } catch {
      setError("Could not extend the session. You may need to sign in again.");
    } finally {
      setRenewing(false);
    }
  };

  return (
    <div className="fixed bottom-4 right-4 z-[90] w-full max-w-sm rounded-xl border border-amber-200 bg-white p-4 shadow-lg dark:border-amber-900 dark:bg-gray-800">
      <div className="flex items-start gap-3">
        <div className="rounded-lg bg-amber-100 p-2 dark:bg-amber-900/40">
          <Clock className="h-4 w-4 text-amber-600 dark:text-amber-400" />
        </div>

        <div className="min-w-0 flex-1">
          <p className="text-sm font-medium text-gray-900 dark:text-gray-100">
            Your session ends in {remaining}
          </p>
          <p className="mt-0.5 text-xs text-gray-600 dark:text-gray-400">
            Continue to stay signed in. No password needed.
          </p>

          {error && (
            <p className="mt-2 text-xs text-red-600 dark:text-red-400">{error}</p>
          )}

          <div className="mt-3 flex items-center gap-2">
            <Button onClick={renew} disabled={renewing} size="sm">
              {renewing ? "Extending..." : "Continue"}
            </Button>
            <Button
              onClick={dismissExpiryWarning}
              variant="ghost"
              size="sm"
              className="text-xs"
            >
              Dismiss
            </Button>
            <Button
              onClick={logout}
              variant="ghost"
              size="sm"
              className="ml-auto text-xs"
            >
              Sign out
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}

export default SessionExpiringDialog;
