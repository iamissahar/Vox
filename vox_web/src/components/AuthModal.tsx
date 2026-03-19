import React, { useState } from 'react';
import { authApi } from '../api';
import { useAuth } from '../hooks/useAuth';
import { Button } from './Button';
import { Input } from './Input';

interface AuthModalProps {
  onClose: () => void;
  /** Where to redirect after successful login */
  onSuccess?: () => void;
}

type Mode = 'login' | 'signup';

// ─── OAuth Icon ───────────────────────────────────────────────────────────────

function GoogleIcon() {
  return (
    <svg width="18" height="18" viewBox="0 0 18 18" fill="none">
      <path d="M17.64 9.2c0-.637-.057-1.251-.164-1.84H9v3.481h4.844a4.14 4.14 0 01-1.796 2.716v2.259h2.908c1.702-1.567 2.684-3.875 2.684-6.615z" fill="#4285F4"/>
      <path d="M9 18c2.43 0 4.467-.806 5.956-2.18l-2.908-2.259c-.806.54-1.837.86-3.048.86-2.344 0-4.328-1.584-5.036-3.711H.957v2.332A8.997 8.997 0 009 18z" fill="#34A853"/>
      <path d="M3.964 10.71A5.41 5.41 0 013.682 9c0-.593.102-1.17.282-1.71V4.958H.957A8.996 8.996 0 000 9c0 1.452.348 2.827.957 4.042l3.007-2.332z" fill="#FBBC05"/>
      <path d="M9 3.58c1.321 0 2.508.454 3.44 1.345l2.582-2.58C13.463.891 11.426 0 9 0A8.997 8.997 0 00.957 4.958L3.964 7.29C4.672 5.163 6.656 3.58 9 3.58z" fill="#EA4335"/>
    </svg>
  );
}

function GithubIcon() {
  return (
    <svg width="18" height="18" viewBox="0 0 18 18" fill="currentColor">
      <path d="M9 .5A8.5 8.5 0 00.5 9a8.5 8.5 0 005.81 8.07c.42.08.58-.18.58-.4v-1.4c-2.36.51-2.86-1.14-2.86-1.14-.39-1-.95-1.26-.95-1.26-.77-.53.06-.52.06-.52.85.06 1.3.87 1.3.87.76 1.3 1.99.92 2.47.7.08-.55.3-.92.54-1.13-1.88-.21-3.86-.94-3.86-4.19 0-.93.33-1.69.87-2.28-.09-.21-.38-1.08.08-2.25 0 0 .71-.23 2.33.87a8.1 8.1 0 014.24 0c1.62-1.1 2.33-.87 2.33-.87.46 1.17.17 2.04.08 2.25.54.59.87 1.35.87 2.28 0 3.26-1.99 3.98-3.88 4.19.3.26.58.78.58 1.57v2.33c0 .22.15.49.58.4A8.5 8.5 0 009 .5z"/>
    </svg>
  );
}

// ─── Component ────────────────────────────────────────────────────────────────

export function AuthModal({ onClose, onSuccess }: AuthModalProps) {
  const { refreshUser } = useAuth();
  const [mode, setMode] = useState<Mode>('login');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');

  // Form state
  const [login, setLogin] = useState('');
  const [password, setPassword] = useState('');
  const [email, setEmail] = useState('');
  const [name, setName] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setIsLoading(true);

    try {
      if (mode === 'login') {
        await authApi.login({ login, password });
      } else {
        await authApi.signUp({ login, password, email, name });
      }
      await refreshUser();
      onSuccess?.();
      onClose();
    } catch (err: any) {
      const msg =
        err?.response?.data?.error?.message ||
        'Something went wrong. Please try again.';
      setError(msg);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    // Backdrop
    <div
      onClick={onClose}
      style={{
        position: 'fixed',
        inset: 0,
        background: 'rgba(0,0,0,0.7)',
        backdropFilter: 'blur(8px)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        zIndex: 1000,
        padding: '16px',
      }}
    >
      {/* Modal */}
      <div
        onClick={(e) => e.stopPropagation()}
        style={{
          background: 'var(--bg-card)',
          border: '1px solid var(--border)',
          borderRadius: 'var(--radius-xl)',
          padding: '36px',
          width: '100%',
          maxWidth: '420px',
          animation: 'fadeUp 0.25s ease',
        }}
      >
        {/* Header */}
        <div style={{ marginBottom: '28px' }}>
          <h2 style={{ fontSize: '22px', marginBottom: '6px' }}>
            {mode === 'login' ? 'Welcome back' : 'Create account'}
          </h2>
          <p style={{ color: 'var(--text-secondary)', fontSize: '14px' }}>
            {mode === 'login'
              ? 'Sign in to manage your hubs and voice.'
              : 'Join Vox to start broadcasting.'}
          </p>
        </div>

        {/* OAuth buttons */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: '10px', marginBottom: '24px' }}>
          <Button
            variant="secondary"
            fullWidth
            onClick={() => authApi.oauthLogin('google')}
            style={{ justifyContent: 'center', gap: '10px' }}
          >
            <GoogleIcon />
            Continue with Google
          </Button>
          <Button
            variant="secondary"
            fullWidth
            onClick={() => authApi.oauthLogin('github')}
            style={{ justifyContent: 'center', gap: '10px' }}
          >
            <GithubIcon />
            Continue with GitHub
          </Button>
        </div>

        {/* Divider */}
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: '12px',
            marginBottom: '24px',
          }}
        >
          <div style={{ flex: 1, height: '1px', background: 'var(--border)' }} />
          <span style={{ color: 'var(--text-muted)', fontSize: '12px' }}>or</span>
          <div style={{ flex: 1, height: '1px', background: 'var(--border)' }} />
        </div>

        {/* Form */}
        <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: '14px' }}>
          {mode === 'signup' && (
            <Input
              label="Full name"
              placeholder="Ada Lovelace"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
            />
          )}
          {mode === 'signup' && (
            <Input
              label="Email"
              type="email"
              placeholder="ada@example.com"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
            />
          )}
          <Input
            label="Username"
            placeholder="your_login"
            value={login}
            onChange={(e) => setLogin(e.target.value)}
            required
          />
          <Input
            label="Password"
            type="password"
            placeholder="••••••••"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
          />

          {error && (
            <p
              style={{
                color: 'var(--error)',
                fontSize: '13px',
                background: 'rgba(248,113,113,0.08)',
                border: '1px solid rgba(248,113,113,0.2)',
                borderRadius: '8px',
                padding: '10px 12px',
              }}
            >
              {error}
            </p>
          )}

          <Button type="submit" fullWidth isLoading={isLoading} size="lg" style={{ marginTop: '4px' }}>
            {mode === 'login' ? 'Sign in' : 'Create account'}
          </Button>
        </form>

        {/* Toggle */}
        <p
          style={{
            textAlign: 'center',
            marginTop: '20px',
            fontSize: '13px',
            color: 'var(--text-secondary)',
          }}
        >
          {mode === 'login' ? "Don't have an account? " : 'Already have an account? '}
          <button
            onClick={() => { setMode(mode === 'login' ? 'signup' : 'login'); setError(''); }}
            style={{
              color: 'var(--accent)',
              background: 'none',
              border: 'none',
              cursor: 'pointer',
              fontSize: '13px',
              fontWeight: 500,
            }}
          >
            {mode === 'login' ? 'Sign up' : 'Sign in'}
          </button>
        </p>

        {/* Close */}
        <button
          onClick={onClose}
          style={{
            position: 'absolute',
            top: '16px',
            right: '16px',
            color: 'var(--text-muted)',
            fontSize: '18px',
            lineHeight: 1,
            background: 'none',
            border: 'none',
            cursor: 'pointer',
            display: 'none', // handled by backdrop click
          }}
        >
          ×
        </button>
      </div>
    </div>
  );
}
