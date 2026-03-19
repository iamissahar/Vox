import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';
import { AuthModal } from '../components/AuthModal';
import { VoxLogo } from '../components/VoxLogo';
import { Button } from '../components/Button';

// ─── Feature card ─────────────────────────────────────────────────────────────

function FeatureCard({
  icon,
  title,
  desc,
}: {
  icon: string;
  title: string;
  desc: string;
}) {
  return (
    <div
      style={{
        background: 'var(--bg-card)',
        border: '1px solid var(--border)',
        borderRadius: 'var(--radius-lg)',
        padding: '28px',
        display: 'flex',
        flexDirection: 'column',
        gap: '12px',
      }}
    >
      <span style={{ fontSize: '28px' }}>{icon}</span>
      <h3 style={{ fontSize: '16px', fontWeight: 600 }}>{title}</h3>
      <p style={{ fontSize: '14px', color: 'var(--text-secondary)', lineHeight: 1.6 }}>
        {desc}
      </p>
    </div>
  );
}

// ─── Main ─────────────────────────────────────────────────────────────────────

export function HomePage() {
  const navigate = useNavigate();
  const { isAuthenticated } = useAuth();
  const [showAuth, setShowAuth] = useState(false);

  const handleHostClick = () => {
    if (isAuthenticated) {
      navigate('/host');
    } else {
      setShowAuth(true);
    }
  };

  const handleGuestClick = () => {
    navigate('/listen');
  };

  return (
    <>
      <div
        style={{
          minHeight: '100vh',
          display: 'flex',
          flexDirection: 'column',
        }}
      >
        {/* Nav */}
        <nav
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            padding: '20px 40px',
            borderBottom: '1px solid var(--border)',
          }}
        >
          <VoxLogo size={32} />
          {isAuthenticated ? (
            <Button variant="ghost" size="sm" onClick={() => navigate('/host')}>
              Dashboard →
            </Button>
          ) : (
            <Button variant="secondary" size="sm" onClick={() => setShowAuth(true)}>
              Sign in
            </Button>
          )}
        </nav>

        {/* Hero */}
        <main
          style={{
            flex: 1,
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            justifyContent: 'center',
            padding: '80px 24px',
            textAlign: 'center',
            position: 'relative',
            overflow: 'hidden',
          }}
        >
          {/* Background glow */}
          <div
            style={{
              position: 'absolute',
              top: '20%',
              left: '50%',
              transform: 'translateX(-50%)',
              width: '600px',
              height: '300px',
              background: 'radial-gradient(ellipse, rgba(108,99,255,0.15) 0%, transparent 70%)',
              pointerEvents: 'none',
            }}
          />

          <div
            style={{
              display: 'inline-flex',
              alignItems: 'center',
              gap: '8px',
              background: 'var(--accent-dim)',
              border: '1px solid var(--border-accent)',
              borderRadius: '100px',
              padding: '6px 16px',
              fontSize: '12px',
              color: 'var(--accent)',
              fontWeight: 500,
              marginBottom: '32px',
              letterSpacing: '0.05em',
              textTransform: 'uppercase',
            }}
          >
            <span
              style={{
                width: '6px',
                height: '6px',
                borderRadius: '50%',
                background: 'var(--accent)',
                animation: 'pulse 2s ease infinite',
              }}
            />
            Real-time voice translation
          </div>

          <h1
            style={{
              fontFamily: "'Syne', sans-serif",
              fontSize: 'clamp(44px, 8vw, 80px)',
              fontWeight: 800,
              lineHeight: 1.05,
              letterSpacing: '-0.03em',
              marginBottom: '24px',
              maxWidth: '800px',
            }}
          >
            Speak once.
            <br />
            <span style={{ color: 'var(--accent)' }}>Everyone hears.</span>
          </h1>

          <p
            style={{
              fontSize: 'clamp(16px, 2vw, 18px)',
              color: 'var(--text-secondary)',
              maxWidth: '520px',
              lineHeight: 1.7,
              marginBottom: '52px',
            }}
          >
            Vox lets hosts broadcast live speech in Russian — translated to English
            in real-time — so any listener anywhere can follow along, without any
            special app.
          </p>

          {/* CTA row */}
          <div
            style={{
              display: 'flex',
              gap: '16px',
              flexWrap: 'wrap',
              justifyContent: 'center',
            }}
          >
            <button
              onClick={handleHostClick}
              style={{
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'flex-start',
                gap: '4px',
                background: 'var(--accent)',
                color: '#fff',
                border: 'none',
                borderRadius: 'var(--radius-lg)',
                padding: '20px 28px',
                cursor: 'pointer',
                transition: 'transform 0.2s ease, box-shadow 0.2s ease',
                minWidth: '180px',
                textAlign: 'left',
              }}
              onMouseEnter={(e) => {
                (e.currentTarget as HTMLElement).style.transform = 'translateY(-2px)';
                (e.currentTarget as HTMLElement).style.boxShadow = '0 12px 40px var(--accent-glow)';
              }}
              onMouseLeave={(e) => {
                (e.currentTarget as HTMLElement).style.transform = 'translateY(0)';
                (e.currentTarget as HTMLElement).style.boxShadow = 'none';
              }}
            >
              <span style={{ fontSize: '22px' }}>🎙</span>
              <span
                style={{
                  fontFamily: "'Syne', sans-serif",
                  fontSize: '18px',
                  fontWeight: 700,
                }}
              >
                I'm a Host
              </span>
              <span style={{ fontSize: '13px', opacity: 0.75 }}>
                Create a hub & broadcast
              </span>
            </button>

            <button
              onClick={handleGuestClick}
              style={{
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'flex-start',
                gap: '4px',
                background: 'var(--bg-card)',
                color: 'var(--text-primary)',
                border: '1px solid var(--border)',
                borderRadius: 'var(--radius-lg)',
                padding: '20px 28px',
                cursor: 'pointer',
                transition: 'transform 0.2s ease, border-color 0.2s ease',
                minWidth: '180px',
                textAlign: 'left',
              }}
              onMouseEnter={(e) => {
                (e.currentTarget as HTMLElement).style.transform = 'translateY(-2px)';
                (e.currentTarget as HTMLElement).style.borderColor = 'var(--border-accent)';
              }}
              onMouseLeave={(e) => {
                (e.currentTarget as HTMLElement).style.transform = 'translateY(0)';
                (e.currentTarget as HTMLElement).style.borderColor = 'var(--border)';
              }}
            >
              <span style={{ fontSize: '22px' }}>🎧</span>
              <span
                style={{
                  fontFamily: "'Syne', sans-serif",
                  fontSize: '18px',
                  fontWeight: 700,
                }}
              >
                I'm a Listener
              </span>
              <span style={{ fontSize: '13px', color: 'var(--text-secondary)' }}>
                Join a hub with its ID
              </span>
            </button>
          </div>
        </main>

        {/* Features */}
        <section
          style={{
            padding: '80px 40px',
            maxWidth: '1100px',
            margin: '0 auto',
            width: '100%',
          }}
        >
          <h2
            style={{
              textAlign: 'center',
              fontSize: '28px',
              marginBottom: '48px',
              color: 'var(--text-secondary)',
              fontWeight: 600,
            }}
          >
            How it works
          </h2>
          <div
            style={{
              display: 'grid',
              gridTemplateColumns: 'repeat(auto-fit, minmax(240px, 1fr))',
              gap: '20px',
            }}
          >
            <FeatureCard
              icon="🎙"
              title="Host creates a hub"
              desc="Sign in, create a hub in one click. You get a unique ID to share with your audience."
            />
            <FeatureCard
              icon="🔁"
              title="Live translation"
              desc="Vox transcribes your Russian speech via Deepgram, translates it via Groq, and synthesizes English audio with your voice."
            />
            <FeatureCard
              icon="🎧"
              title="Listeners join"
              desc="Anyone can join just by entering the hub ID — no account, no install, just a browser."
            />
            <FeatureCard
              icon="🗣"
              title="Your voice, any language"
              desc="Upload a short voice sample and Vox will use your voice's timbre when synthesizing translated speech."
            />
          </div>
        </section>

        {/* Footer */}
        <footer
          style={{
            borderTop: '1px solid var(--border)',
            padding: '24px 40px',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            flexWrap: 'wrap',
            gap: '12px',
          }}
        >
          <VoxLogo size={24} />
          <p style={{ fontSize: '13px', color: 'var(--text-muted)' }}>
            © 2025 Vox · Real-time voice translation
          </p>
        </footer>
      </div>

      {showAuth && (
        <AuthModal
          onClose={() => setShowAuth(false)}
          onSuccess={() => navigate('/host')}
        />
      )}

      <style>{`
        @keyframes pulse {
          0%, 100% { opacity: 1; }
          50% { opacity: 0.4; }
        }
        @keyframes fadeUp {
          from { opacity: 0; transform: translateY(16px); }
          to { opacity: 1; transform: translateY(0); }
        }
        @keyframes spin {
          to { transform: rotate(360deg); }
        }
      `}</style>
    </>
  );
}
