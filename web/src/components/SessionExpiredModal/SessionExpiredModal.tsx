import { useEffect, useRef, useState } from "react";
import { Lock } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useAuth } from "@/contexts/AuthContext";

interface AuthResponse {
  token: string;
  user: {
    id: string;
    username: string;
    email: string;
    role: string;
    createdAt: string;
    lastLogin: string;
  };
}

/**
 * Shown over the page when a request is rejected for an expired session.
 *
 * Deliberately not a redirect to the login screen: a redirect discards scroll
 * position, open dialogs, filters and half-typed input, and leaves the user to
 * remember what they had been doing. Logging in here puts them back exactly
 * where they were, and the request that triggered this is replayed, so the
 * action they originally asked for still happens.
 */
export function SessionExpiredModal() {
  const { sessionExpired, user, resumeSession, logout } = useAuth();
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const passwordRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (sessionExpired) {
      setPassword("");
      setError("");
      passwordRef.current?.focus();
    }
  }, [sessionExpired]);

  if (!sessionExpired) return null;

  const email = user?.email ?? "";

  const submit = async (event: React.FormEvent) => {
    event.preventDefault();

    setSubmitting(true);
    setError("");

    try {
      // Plain fetch: this request is what fixes the session, so it must not be
      // parked by the wrapper that handles 401s.
      const response = await fetch("/api/auth/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email, password }),
      });

      if (!response.ok) {
        setError(
          response.status === 401
            ? "That password was not accepted."
            : (await response.text()) || "Could not sign in.",
        );
        return;
      }

      const auth: AuthResponse = await response.json();
      resumeSession(auth.user, auth.token);
    } catch {
      setError("Could not reach the server.");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="fixed inset-0 z-[100] flex items-center justify-center bg-black/50 p-4">
      <div className="w-full max-w-sm rounded-xl bg-white p-6 shadow-xl dark:bg-gray-800">
        <div className="flex items-start gap-3">
          <div className="rounded-lg bg-amber-100 p-2 dark:bg-amber-900/40">
            <Lock className="h-5 w-5 text-amber-600 dark:text-amber-400" />
          </div>
          <div className="min-w-0">
            <h2 className="text-base font-semibold text-gray-900 dark:text-gray-100">
              Your session expired
            </h2>
            <p className="mt-1 text-sm text-gray-600 dark:text-gray-400">
              Sign in to continue. Nothing is lost - whatever you were doing will
              carry on.
            </p>
          </div>
        </div>

        <form onSubmit={submit} className="mt-4 space-y-3">
          {email && (
            <div className="rounded-lg bg-gray-50 px-3 py-2 text-sm text-gray-700 dark:bg-gray-900 dark:text-gray-300">
              {email}
            </div>
          )}

          <input
            ref={passwordRef}
            type="password"
            value={password}
            onChange={(event) => setPassword(event.target.value)}
            placeholder="Password"
            autoComplete="current-password"
            className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm text-gray-900 focus:border-blue-500 focus:outline-none dark:border-gray-600 dark:bg-gray-900 dark:text-gray-100"
          />

          {error && (
            <p className="text-sm text-red-600 dark:text-red-400">{error}</p>
          )}

          <div className="flex items-center gap-2">
            <Button
              type="submit"
              disabled={submitting || password.length === 0}
              className="flex-1"
            >
              {submitting ? "Signing in..." : "Sign in and continue"}
            </Button>
            <Button type="button" variant="ghost" onClick={logout}>
              Sign out
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
}

export default SessionExpiredModal;
