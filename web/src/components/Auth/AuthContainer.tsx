import { useState } from 'react';
import { LoginForm } from './LoginForm';
import { RegisterForm } from './RegisterForm';

interface User {
  id: string;
  username: string;
  email: string;
  role: string;
  createdAt: string;
  lastLogin: string;
}

interface AuthResponse {
  user: User;
  token: string;
}

interface AuthContainerProps {
  onAuthSuccess: (user: User, token: string) => void;
}

export function AuthContainer({ onAuthSuccess }: AuthContainerProps) {
  const [mode, setMode] = useState<'login' | 'register'>('login');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleLogin = async (email: string, password: string) => {
    setIsLoading(true);
    setError(null);

    try {
      const response = await fetch('/api/auth/login', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ email, password }),
      });

      if (!response.ok) {
        const errorData = await response.text();
        throw new Error(errorData || 'Login failed');
      }

      const authResponse: AuthResponse = await response.json();
      
      // Store token in localStorage
      localStorage.setItem('authToken', authResponse.token);
      localStorage.setItem('user', JSON.stringify(authResponse.user));
      
      onAuthSuccess(authResponse.user, authResponse.token);
    } catch (error) {
      console.error('Login error:', error);
      setError(error instanceof Error ? error.message : 'Login failed');
    } finally {
      setIsLoading(false);
    }
  };

  const handleRegister = async (username: string, email: string, password: string) => {
    setIsLoading(true);
    setError(null);

    try {
      const response = await fetch('/api/auth/register', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ username, email, password }),
      });

      if (!response.ok) {
        const errorData = await response.text();
        throw new Error(errorData || 'Registration failed');
      }

      const result = await response.json();
      
      // After successful registration, switch to login mode
      setMode('login');
      setError(null);
      
      // Show success message (you could use a toast here)
      console.log('Registration successful:', result.message);
      
    } catch (error) {
      console.error('Registration error:', error);
      setError(error instanceof Error ? error.message : 'Registration failed');
    } finally {
      setIsLoading(false);
    }
  };

  if (mode === 'login') {
    return (
      <LoginForm
        onLogin={handleLogin}
        onSwitchToRegister={() => {
          setMode('register');
          setError(null);
        }}
        isLoading={isLoading}
        error={error}
      />
    );
  }

  return (
    <RegisterForm
      onRegister={handleRegister}
      onSwitchToLogin={() => {
        setMode('login');
        setError(null);
      }}
      isLoading={isLoading}
      error={error}
    />
  );
}