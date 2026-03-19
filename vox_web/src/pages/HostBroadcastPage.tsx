import React, { useState, useRef, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useAudioRecorder } from '../hooks/useAudioRecorder';
import { hubApi } from '../api';
import { VoxLogo } from '../components/VoxLogo';
import { Button } from '../components/Button';
import { Waveform } from '../components/Waveform';
import { Card } from '../components/Card';

// ─── Main ─────────────────────────────────────────────────────────────────────

export function HostBroadcastPage() {
  const { hubId } = useParams<{ hubId: string }>();
  const navigate = useNavigate();

  const { isRecording, analyserNode, startRecording, stopRecording, error } =
    useAudioRecorder();

  const [copied, setCopied] = useState(false);

  // XHR ref — each audio chunk is sent as a separate POST
  const xhrRef = useRef<XMLHttpRequest | null>(null);

  // ── Send each audio chunk to the publish endpoint ─────────────────────────
  const publishChunk = useCallback(
    (chunk: Blob) => {
      if (!hubId) return;
      const url = hubApi.getPublishUrl(hubId, 'ru');
      const xhr = new XMLHttpRequest();
      xhrRef.current = xhr;
      xhr.open('POST', url, true);
      xhr.withCredentials = true;
      xhr.setRequestHeader('Content-Type', 'application/octet-stream');
      xhr.send(chunk);
    },
    [hubId]
  );

  const handleStart = async () => {
    await startRecording(publishChunk);
  };

  const handleStop = () => {
    stopRecording();
    xhrRef.current?.abort();
  };

  const handleCopyId = () => {
    if (hubId) {
      navigator.clipboard.writeText(hubId);
      setCopied(true);
      setTimeout(() => setCopied(false), 1800);
    }
  };

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
        <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
          <span style={{ fontSize: '13px', color: 'var(--text-muted)' }}>Hub:</span>
          <code
            style={{
              background: 'var(--bg-elevated)',
              border: '1px solid var(--border)',
              borderRadius: '6px',
              padding: '4px 10px',
              fontSize: '12px',
              color: 'var(--text-secondary)',
              maxWidth: '200px',
              overflow: 'hidden',
              textOverflow: 'ellipsis',
              whiteSpace: 'nowrap',
            }}
          >
            {hubId}
          </code>
          <Button size="sm" variant="secondary" onClick={handleCopyId}>
            {copied ? '✓' : 'Copy'}
          </Button>
          <Button size="sm" variant="ghost" onClick={() => navigate('/host')}>
            ← Back
          </Button>
        </div>
      </nav>

      {/* Main */}
      <main
        style={{
          flex: 1,
          maxWidth: '760px',
          margin: '0 auto',
          padding: '64px 24px',
          width: '100%',
          display: 'flex',
          flexDirection: 'column',
          gap: '28px',
        }}
      >
        {/* Status */}
        <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
          <div
            style={{
              width: '10px',
              height: '10px',
              borderRadius: '50%',
              background: isRecording ? 'var(--error)' : 'var(--text-muted)',
              boxShadow: isRecording ? '0 0 10px var(--error)' : 'none',
              animation: isRecording ? 'pulse 1.5s ease infinite' : 'none',
            }}
          />
          <span
            style={{
              fontFamily: "'Syne', sans-serif",
              fontSize: '24px',
              fontWeight: 700,
            }}
          >
            {isRecording ? 'Broadcasting Live' : 'Ready to Broadcast'}
          </span>
        </div>

        {/* Waveform card — the centrepiece */}
        <Card glowing={isRecording} style={{ padding: '32px' }}>
          <div
            style={{
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
              marginBottom: '24px',
            }}
          >
            <span
              style={{
                fontSize: '12px',
                color: 'var(--text-muted)',
                fontWeight: 500,
                textTransform: 'uppercase',
                letterSpacing: '0.08em',
              }}
            >
              Microphone input
            </span>
            <span
              style={{
                fontSize: '12px',
                color: isRecording ? 'var(--error)' : 'var(--text-muted)',
                fontWeight: 500,
              }}
            >
              {isRecording ? '● REC · Russian → English' : '○ Idle'}
            </span>
          </div>

          {/* Large waveform */}
          <Waveform
            analyser={analyserNode}
            isActive={isRecording}
            variant="recorder"
            height={120}
          />

          {error && (
            <p
              style={{
                marginTop: '16px',
                color: 'var(--error)',
                fontSize: '13px',
                background: 'rgba(248,113,113,0.08)',
                border: '1px solid rgba(248,113,113,0.15)',
                borderRadius: '8px',
                padding: '10px 14px',
              }}
            >
              ⚠ {error}
            </p>
          )}

          <div style={{ marginTop: '28px' }}>
            {!isRecording ? (
              <Button onClick={handleStart} size="lg" fullWidth>
                🎙 Start Broadcasting
              </Button>
            ) : (
              <Button
                onClick={handleStop}
                size="lg"
                fullWidth
                style={{ background: 'var(--error)' }}
              >
                ⏹ Stop Broadcasting
              </Button>
            )}
          </div>
        </Card>

        {/* Share hub ID */}
        <Card>
          <p
            style={{
              fontSize: '14px',
              color: 'var(--text-secondary)',
              marginBottom: '14px',
            }}
          >
            Share this Hub ID with your listeners — they don't need an account:
          </p>
          <div style={{ display: 'flex', gap: '10px', alignItems: 'center' }}>
            <code
              style={{
                flex: 1,
                background: 'var(--bg-elevated)',
                border: '1px solid var(--border)',
                borderRadius: '8px',
                padding: '12px 16px',
                fontSize: '13px',
                color: 'var(--text-primary)',
                wordBreak: 'break-all',
              }}
            >
              {hubId}
            </code>
            <Button variant="secondary" onClick={handleCopyId}>
              {copied ? '✓ Copied' : 'Copy'}
            </Button>
          </div>
        </Card>

        {/* Tip */}
        <p
          style={{
            fontSize: '13px',
            color: 'var(--text-muted)',
            textAlign: 'center',
            lineHeight: 1.6,
          }}
        >
          Want your voice to sound like you in translation?{' '}
          <button
            onClick={() => navigate('/profile')}
            style={{
              color: 'var(--accent)',
              background: 'none',
              border: 'none',
              cursor: 'pointer',
              fontSize: '13px',
            }}
          >
            Upload a voice sample →
          </button>
        </p>
      </main>

      <style>{`
        @keyframes pulse {
          0%, 100% { opacity: 1; }
          50% { opacity: 0.3; }
        }
      `}</style>
    </div>
  );
}
