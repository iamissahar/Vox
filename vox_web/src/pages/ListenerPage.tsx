import React, { useState, useRef, useEffect, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { hubApi } from '../api';
import { VoxLogo } from '../components/VoxLogo';
import { Button } from '../components/Button';
import { Input } from '../components/Input';
import { Card } from '../components/Card';
import { Waveform } from '../components/Waveform';

type Phase = 'idle' | 'listening' | 'error';

// ─── Listener audio analyser from <audio> element ─────────────────────────────

function useAudioAnalyser(audioRef: React.RefObject<HTMLAudioElement | null>) {
  const [analyser, setAnalyser] = useState<AnalyserNode | null>(null);
  const audioCtxRef = useRef<AudioContext | null>(null);
  const sourceRef = useRef<MediaElementAudioSourceNode | null>(null);

  const connect = useCallback(() => {
    if (!audioRef.current || audioCtxRef.current) return;
    const audioCtx = new AudioContext();
    audioCtxRef.current = audioCtx;
    const source = audioCtx.createMediaElementSource(audioRef.current);
    sourceRef.current = source;
    const analyserNode = audioCtx.createAnalyser();
    analyserNode.fftSize = 256;
    source.connect(analyserNode);
    analyserNode.connect(audioCtx.destination);
    setAnalyser(analyserNode);
  }, [audioRef]);

  const disconnect = useCallback(() => {
    audioCtxRef.current?.close();
    audioCtxRef.current = null;
    setAnalyser(null);
  }, []);

  return { analyser, connect, disconnect };
}

// ─── Main ─────────────────────────────────────────────────────────────────────

export function ListenerPage() {
  const navigate = useNavigate();

  const [hubId, setHubId] = useState('');
  const [phase, setPhase] = useState<Phase>('idle');
  const [errorMsg, setErrorMsg] = useState('');

  const audioRef = useRef<HTMLAudioElement>(null);
  const { analyser, connect, disconnect } = useAudioAnalyser(audioRef);

  const handleListen = useCallback(async () => {
    if (!hubId.trim()) return;
    setErrorMsg('');
    setPhase('listening');

    const streamUrl = hubApi.getListenUrl(hubId.trim());

    if (!audioRef.current) return;
    audioRef.current.src = streamUrl;
    audioRef.current.load();

    try {
      await audioRef.current.play();
      connect();
    } catch (err) {
      setPhase('error');
      setErrorMsg('Could not connect to the hub. Please check the ID and try again.');
    }
  }, [hubId, connect]);

  const handleStop = useCallback(() => {
    if (audioRef.current) {
      audioRef.current.pause();
      audioRef.current.src = '';
    }
    disconnect();
    setPhase('idle');
  }, [disconnect]);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      handleStop();
    };
  }, [handleStop]);

  const isListening = phase === 'listening';

  return (
    <div style={{ minHeight: '100vh', display: 'flex', flexDirection: 'column' }}>
      {/* Nav */}
      <nav
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          padding: '16px 40px',
          borderBottom: '1px solid var(--border)',
        }}
      >
        <VoxLogo size={26} />
        <Button variant="ghost" size="sm" onClick={() => navigate('/')}>
          ← Home
        </Button>
      </nav>

      {/* Hidden audio element */}
      <audio ref={audioRef} style={{ display: 'none' }} />

      {/* Main */}
      <main
        style={{
          flex: 1,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          padding: '48px 24px',
        }}
      >
        <div style={{ width: '100%', maxWidth: '560px', display: 'flex', flexDirection: 'column', gap: '24px' }}>
          {/* Header */}
          <div>
            <div style={{ display: 'flex', alignItems: 'center', gap: '10px', marginBottom: '8px' }}>
              <div
                style={{
                  width: '10px',
                  height: '10px',
                  borderRadius: '50%',
                  background: isListening ? 'var(--success)' : 'var(--text-muted)',
                  boxShadow: isListening ? '0 0 8px var(--success)' : 'none',
                  animation: isListening ? 'pulse 2s ease infinite' : 'none',
                }}
              />
              <h1
                style={{
                  fontFamily: "'Syne', sans-serif",
                  fontSize: '26px',
                  fontWeight: 700,
                }}
              >
                {isListening ? 'Listening…' : 'Join a Hub'}
              </h1>
            </div>
            <p style={{ color: 'var(--text-secondary)', fontSize: '14px' }}>
              {isListening
                ? 'You are receiving the live translated audio stream.'
                : 'Enter the hub ID shared by your host to start listening.'}
            </p>
          </div>

          {/* Input card */}
          {!isListening && (
            <Card>
              <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
                <Input
                  label="Hub ID"
                  placeholder="Paste the hub ID here…"
                  value={hubId}
                  onChange={(e) => setHubId(e.target.value)}
                  onKeyDown={(e) => e.key === 'Enter' && handleListen()}
                  error={phase === 'error' ? errorMsg : undefined}
                />
                <Button onClick={handleListen} size="lg" fullWidth disabled={!hubId.trim()}>
                  🎧 Start Listening
                </Button>
              </div>
            </Card>
          )}

          {/* Waveform card (active state) */}
          {isListening && (
            <Card glowing>
              <div style={{ marginBottom: '12px', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <span style={{ fontSize: '12px', color: 'var(--text-muted)', fontWeight: 500, textTransform: 'uppercase', letterSpacing: '0.08em' }}>
                  Translated audio
                </span>
                <span style={{ fontSize: '12px', color: 'var(--success)' }}>
                  ● Live
                </span>
              </div>

              <Waveform analyser={analyser} isActive={isListening} variant="listener" height={90} />

              <div style={{ marginTop: '20px', display: 'flex', gap: '12px', alignItems: 'center' }}>
                <Button
                  onClick={handleStop}
                  variant="secondary"
                  fullWidth
                  size="md"
                >
                  ⏹ Stop
                </Button>
              </div>

              {/* Hub ID reminder */}
              <div
                style={{
                  marginTop: '16px',
                  padding: '10px 14px',
                  background: 'var(--bg-elevated)',
                  borderRadius: '8px',
                  fontSize: '12px',
                  color: 'var(--text-muted)',
                  display: 'flex',
                  gap: '8px',
                  alignItems: 'center',
                }}
              >
                <span>Connected to:</span>
                <code style={{ color: 'var(--text-secondary)', wordBreak: 'break-all' }}>
                  {hubId}
                </code>
              </div>
            </Card>
          )}

          {/* Info note */}
          {!isListening && (
            <p
              style={{
                textAlign: 'center',
                fontSize: '13px',
                color: 'var(--text-muted)',
                lineHeight: 1.6,
              }}
            >
              No account needed to listen. Just paste the ID and press play.
            </p>
          )}
        </div>
      </main>

      <style>{`
        @keyframes pulse {
          0%, 100% { opacity: 1; }
          50% { opacity: 0.4; }
        }
      `}</style>
    </div>
  );
}
