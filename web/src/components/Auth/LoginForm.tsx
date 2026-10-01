import { useState } from "react";
import { Lock, Mail, Eye, EyeOff, LogIn, Activity } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { CopyButton } from "@/components/ui/copy-button";

interface LoginFormProps {
  onLogin: (email: string, password: string) => Promise<void>;
  onSwitchToRegister: () => void;
  isLoading: boolean;
  error: string | null;
}

export function LoginForm({
  onLogin,
  onSwitchToRegister,
  isLoading,
  error,
}: LoginFormProps) {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [showRecovery, setShowRecovery] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (email && password) {
      await onLogin(email, password);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-blue-50 to-purple-50 p-4">
      <Card className="w-full max-w-md shadow-xl">
        <CardHeader className="text-center pb-4">
          <div className="flex justify-center mb-4">
            <div className="p-3 bg-gradient-to-br from-blue-500 to-purple-600 rounded-full">
              <Activity className="h-8 w-8 text-white" />
            </div>
          </div>
          <CardTitle className="text-2xl font-bold bg-gradient-to-r from-blue-600 to-purple-600 bg-clip-text text-transparent">
            Welcome Back
          </CardTitle>
          <p className="text-gray-600 text-sm">Sign in to Vertex</p>
        </CardHeader>

        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            {error && (
              <div className="p-3 bg-red-50 border border-red-200 rounded-md">
                <p className="text-red-700 text-sm">{error}</p>
              </div>
            )}

            <div className="space-y-2">
              <Label
                htmlFor="email"
                className="text-sm font-medium text-gray-700"
              >
                Email Address
              </Label>
              <div className="relative">
                <Mail className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 h-4 w-4" />
                <Input
                  id="email"
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  placeholder="Enter your email"
                  className="pl-10"
                  required
                  disabled={isLoading}
                />
              </div>
            </div>

            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <Label
                  htmlFor="password"
                  className="text-sm font-medium text-gray-700"
                >
                  Password
                </Label>
                <button
                  type="button"
                  onClick={() => setShowRecovery(!showRecovery)}
                  className="text-xs text-blue-600 hover:text-blue-700 hover:underline"
                >
                  Forgot password?
                </button>
              </div>
              <div className="relative">
                <Lock className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 h-4 w-4" />
                <Input
                  id="password"
                  type={showPassword ? "text" : "password"}
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  placeholder="Enter your password"
                  className="pl-10 pr-10"
                  required
                  disabled={isLoading}
                />
                <button
                  type="button"
                  onClick={() => setShowPassword(!showPassword)}
                  className="absolute right-3 top-1/2 transform -translate-y-1/2 text-gray-400 hover:text-gray-600"
                  disabled={isLoading}
                >
                  {showPassword ? (
                    <EyeOff className="h-4 w-4" />
                  ) : (
                    <Eye className="h-4 w-4" />
                  )}
                </button>
              </div>
            </div>

            <Button
              type="submit"
              disabled={isLoading || !email || !password}
              className="w-full bg-gradient-to-r from-blue-500 to-purple-600 hover:from-blue-600 hover:to-purple-700"
            >
              {isLoading ? (
                <div className="flex items-center gap-2">
                  <div className="h-4 w-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
                  Signing in...
                </div>
              ) : (
                <div className="flex items-center gap-2">
                  <LogIn className="h-4 w-4" />
                  Sign In
                </div>
              )}
            </Button>
          </form>

          {showRecovery && <PasswordRecoveryHelp email={email} />}

          <div className="mt-6 text-center">
            <p className="text-sm text-gray-600">
              Don't have an account?{" "}
              <button
                onClick={onSwitchToRegister}
                disabled={isLoading}
                className="text-blue-600 hover:text-blue-700 font-medium hover:underline"
              >
                Create one
              </button>
            </p>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

/**
 * Vertex sends no email, so a forgotten password is reset from the machine it
 * runs on - access to that machine is the proof of identity. This shows the
 * command, with the email already typed into the form filled in.
 */
function PasswordRecoveryHelp({ email }: { email: string }) {
  const account = email || "you@example.com";
  const commands = [
    { label: "On the machine running Vertex", command: `vertex reset-password ${account}` },
    { label: "If Vertex runs in Docker", command: `docker exec -it vertex ./vertex reset-password ${account}` },
  ];

  return (
    <div className="mt-4 p-4 bg-blue-50 border border-blue-200 rounded-md space-y-3">
      <p className="text-sm text-gray-700">
        Reset your password by running this where Vertex is installed. It asks
        for a new password.
      </p>
      {commands.map(({ label, command }) => (
        <div key={label} className="space-y-1">
          <p className="text-xs font-medium text-gray-600">{label}</p>
          <div className="flex items-center gap-2">
            <code className="flex-1 min-w-0 px-2 py-1.5 bg-white border border-gray-200 rounded text-xs font-mono text-gray-800 break-words">
              {command}
            </code>
            <CopyButton text={command} label="Copy" />
          </div>
        </div>
      ))}
    </div>
  );
}
