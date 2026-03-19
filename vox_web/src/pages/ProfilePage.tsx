import React, { useEffect, useState, useRef, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';
import { voiceApi } from '../api';
import { VoxLogo } from '../components/VoxLogo';
import { Button } from '../components/Button';
import { Card } from '../components/Card';
import { Waveform } from '../components/Waveform';
import type { VoiceReference } from '../types';

// ─── Voice sample text (Russian, to be read aloud by the host) ───────────────
// Keep it natural and at a comfortable reading pace (~25 words).

const VOICE_SAMPLE_TEXT =
  'Добро пожаловать. Меня зовут Александр. Сегодня я расскажу вам ' +
  'о нашем замечательном проекте по переводу речи в реальном времени. ' +
  'Мы создаём будущее общения между людьми по всему миру.';

const VOICE_SAMPLE_WORDS = VOICE_SAMPLE_TEXT.split(' ');

// Average milliseconds per word at a natural Russian speaking pace (~120 wpm)
const MS_PER_WORD = 500;

// ─── Highlighted text ─────────────────────────────────────────────────────────

function HighlightedText({
  currentWordIndex,
  isRecording,
}: {
  currentWordIndex: number;
  isRecording: boolean;
}) {
  return (
    <p
      style={{
        fontSize: '16px',
        lineHeight: 2,
        letterSpacing: '0.01em',
      }}
    >
      {VOICE_SAMPLE_WORDS.map((word, i) => {
        const isSpoken = i < currentWordIndex;
        const isCurrent = i === currentWordIndex && isRecording;

        return (
          <React.Fragment key={i}>
            <span
              style={{
                color: isSpoken
                  ? 'var(--text-muted)'
                  : isCurrent
                  ? 'var(--text-primary)'
                  : 'var(--text-secondary)',
                background: isCurrent ? 'var(--accent-dim)' : 'transparent',
                borderRadius: '4px',
                padding: isCurrent ? '2px 5px' : '0',
                transition: 'all 0.15s ease',
                fontWeight: isCurrent ? 500 : 400,
                outline: isCurrent ? '1px solid var(--border-accent)' : 'none',
              }}
            >
              {word}
            </span>
            {i < VOICE_SAMPLE_WORDS.length - 1 ? ' ' : ''}
          </React.Fragment>
        );
      })}
    </p>
  );
}

// ─── Voice reference card ─────────────────────────────────────────────────────

function VoiceRefCard({
  voiceRef,
  onDeleted,
}: {
  voiceRef: VoiceReference;
  onDeleted: () => void;
}) {
  const [isDeleting, setIsDeleting] = useState(false);

  const handleDelete = async () => {
    if (!window.confirm('Delete this voice reference?')) return;
    setIsDeleting(true);
    try {
      await voiceApi.delete(voiceRef.file_id);
      onDeleted();
    } catch {
      alert('Failed to delete voice reference.');
    } finally {
      setIsDeleting(false);
    }
  };

  return (
    <div
      style={{
        background: 'var(--bg-elevated)',
        border: '1px solid var(--border)',
        borderRadius: 'var(--radius-md)',
        padding: '16px',
        display: 'flex',
        alignItems: 'flex-start',
        justifyContent: 'space-between',
        gap: '16px',
      }}
    >
      <div style={{ flex: 1, minWidth: 0 }}>
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: '8px',
            marginBottom: '6px',
          }}
        >
          <span style={{ fontSize: '14px' }}>🎙</span>
          <span
            style={{
              fontSize: '11px',
              color: 'var(--text-muted)',
              fontWeight: 500,
              textTransform: 'uppercase',
              letterSpacing: '0.06em',
            }}
          >
            Voice sample
          </span>
        </div>
        <p
          style={{
            fontSize: '13px',
            color: 'var(--text-secondary)',
            lineHeight: 1.5,
            overflow: 'hidden',
            textOverflow: 'ellipsis',
            display: '-webkit-box',
            WebkitLineClamp: 2,
            WebkitBoxOrient: 'vertical',
          }}
        >
          {voiceRef.text || '(no reference text)'}
        </p>
        <p
          style={{ fontSize: '11px', color: 'var(--text-muted)', marginTop: '6px' }}
        >
          ID: {voiceRef.file_id.slice(0, 16)}…
        </p>
      </div>
      <button
        onClick={handleDelete}
        disabled={isDeleting}
        style={{
          background: 'none',
          border: 'none',
          cursor: 'pointer',
          color: 'var(--text-muted)',
          fontSize: '16px',
          flexShrink: 0,
          transition: 'color 0.2s',
        }}
        onMouseEnter={(e) =>
          ((e.target as HTMLElement).style.color = 'var(--error)')
        }
        onMouseLeave={(e) =>
          ((e.target as HTMLElement).style.color = 'var(--text-muted)')
        }
      >
        {isDeleting ? '…' : '×'}
      </button>
    </div>
  );
}

// ─── Record voice panel ───────────────────────────────────────────────────────

function RecordVoicePanel({ onUploaded }: { onUploaded: () => void }) {
  const [isRecording, setIsRecording] = useState(false);
  const [isUploading, setIsUploading] = useState(false);
  const [analyser, setAnalyser] = useState<AnalyserNode | null>(null);
  const [recordedBlob, setRecordedBlob] = useState<Blob | null>(null);
  const [currentWordIndex, setCurrentWordIndex] = useState(-1);
  const [error, setError] = useState('');

  const mediaRecorderRef = useRef<MediaRecorder | null>(null);
  const streamRef = useRef<MediaStream | null>(null);
  const audioCtxRef = useRef<AudioContext | null>(null);
  const wordTimerRef = useRef<ReturnType<typeof setInterval> | null>(null);

  // ── Word highlight timer: advances one word every MS_PER_WORD ms ──────────
  const startWordTracking = () => {
    setCurrentWordIndex(0);
    let idx = 0;
    wordTimerRef.current = setInterval(() => {
      idx++;
      if (idx < VOICE_SAMPLE_WORDS.length) {
        setCurrentWordIndex(idx);
      } else {
        if (wordTimerRef.current) clearInterval(wordTimerRef.current);
      }
    }, MS_PER_WORD);
  };

  const stopWordTracking = () => {
    if (wordTimerRef.current) clearInterval(wordTimerRef.current);
    setCurrentWordIndex(-1);
  };

  // ── Microphone + MediaRecorder ─────────────────────────────────────────────
  const startRecording = async () => {
    setError('');
    setRecordedBlob(null);
    try {
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
      streamRef.current = stream;

      // Web Audio for waveform
      const audioCtx = new AudioContext();
      audioCtxRef.current = audioCtx;
      const source = audioCtx.createMediaStreamSource(stream);
      const analyserNode = audioCtx.createAnalyser();
      analyserNode.fftSize = 256;
      source.connect(analyserNode);
      setAnalyser(analyserNode);

      // MediaRecorder to capture blob
      const mr = new MediaRecorder(stream, { mimeType: 'audio/webm;codecs=opus' });
      mediaRecorderRef.current = mr;
      const localChunks: Blob[] = [];

      mr.ondataavailable = (e) => {
        if (e.data.size > 0) localChunks.push(e.data);
      };
      mr.onstop = () => {
        setRecordedBlob(new Blob(localChunks, { type: 'audio/webm' }));
      };

      mr.start(250);
      setIsRecording(true);
      startWordTracking();
    } catch {
      setError('Microphone access denied.');
    }
  };

  const stopRecording = () => {
    mediaRecorderRef.current?.stop();
    streamRef.current?.getTracks().forEach((t) => t.stop());
    audioCtxRef.current?.close();
    setAnalyser(null);
    setIsRecording(false);
    stopWordTracking();
  };

  const handleUpload = async () => {
    if (!recordedBlob) return;
    setIsUploading(true);
    try {
      await voiceApi.upload(recordedBlob, VOICE_SAMPLE_TEXT);
      setRecordedBlob(null);
      onUploaded();
    } catch {
      setError('Upload failed. Please try again.');
    } finally {
      setIsUploading(false);
    }
  };

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      if (wordTimerRef.current) clearInterval(wordTimerRef.current);
    };
  }, []);

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
      {/* Instruction */}
      <p style={{ fontSize: '13px', color: 'var(--text-muted)' }}>
        Press <strong style={{ color: 'var(--text-secondary)' }}>Record</strong> and
        read the text below at a natural pace. The highlighted word follows your speech.
      </p>

      {/* Text with highlight */}
      <div
        style={{
          background: 'var(--bg-elevated)',
          border: `1px solid ${isRecording ? 'var(--border-accent)' : 'var(--border)'}`,
          borderRadius: 'var(--radius-md)',
          padding: '20px',
          transition: 'border-color 0.2s ease',
        }}
      >
        <HighlightedText
          currentWordIndex={currentWordIndex}
          isRecording={isRecording}
        />
      </div>

      {/* Waveform */}
      <div
        style={{
          background: 'var(--bg-elevated)',
          border: `1px solid ${isRecording ? 'var(--border-accent)' : 'var(--border)'}`,
          borderRadius: 'var(--radius-md)',
          padding: '16px',
          transition: 'border-color 0.2s ease',
        }}
      >
        <div
          style={{
            display: 'flex',
            justifyContent: 'space-between',
            marginBottom: '10px',
            fontSize: '12px',
            color: 'var(--text-muted)',
          }}
        >
          <span>Voice preview</span>
          {isRecording && (
            <span style={{ color: 'var(--error)' }}>● Recording</span>
          )}
          {recordedBlob && !isRecording && (
            <span style={{ color: 'var(--success)' }}>✓ Recorded</span>
          )}
        </div>
        <Waveform
          analyser={analyser}
          isActive={isRecording}
          variant="recorder"
          height={64}
        />
      </div>

      {error && (
        <p style={{ color: 'var(--error)', fontSize: '13px' }}>{error}</p>
      )}

      {/* Action buttons */}
      <div style={{ display: 'flex', gap: '10px' }}>
        {!isRecording ? (
          <Button onClick={startRecording} size="md" style={{ flex: 1 }}>
            {recordedBlob ? '🔁 Re-record' : '🎙 Record'}
          </Button>
        ) : (
          <Button
            onClick={stopRecording}
            size="md"
            style={{ flex: 1, background: 'var(--error)' }}
          >
            ⏹ Stop
          </Button>
        )}

        {recordedBlob && !isRecording && (
          <Button
            onClick={handleUpload}
            isLoading={isUploading}
            variant="secondary"
            size="md"
            style={{ flex: 1 }}
          >
            ↑ Upload sample
          </Button>
        )}
      </div>
    </div>
  );
}

// ─── Main page ────────────────────────────────────────────────────────────────

export function ProfilePage() {
  const navigate = useNavigate();
  const { user, logout } = useAuth();

  const [voiceRefs, setVoiceRefs] = useState<VoiceReference[]>([]);
  const [isLoadingVoice, setIsLoadingVoice] = useState(true);

  const loadVoiceRefs = useCallback(async () => {
    setIsLoadingVoice(true);
    try {
      const { data } = await voiceApi.getMeta();
      setVoiceRefs(data || []);
    } catch {
      setVoiceRefs([]);
    } finally {
      setIsLoadingVoice(false);
    }
  }, []);

  useEffect(() => {
    loadVoiceRefs();
  }, [loadVoiceRefs]);

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
        <div style={{ display: 'flex', gap: '12px' }}>
          <Button variant="ghost" size="sm" onClick={() => navigate('/host')}>
            ← Dashboard
          </Button>
          <Button variant="secondary" size="sm" onClick={logout}>
            Sign out
          </Button>
        </div>
      </nav>

      {/* Content */}
      <main
        style={{
          flex: 1,
          maxWidth: '720px',
          margin: '0 auto',
          padding: '48px 24px',
          width: '100%',
          display: 'flex',
          flexDirection: 'column',
          gap: '32px',
        }}
      >
        <h1 style={{ fontSize: '32px' }}>Profile</h1>

        {/* User info */}
        <Card>
          <div style={{ display: 'flex', alignItems: 'center', gap: '20px' }}>
            {user?.picture ? (
              <img
                src={user.picture}
                alt={user.name}
                style={{
                  width: '64px',
                  height: '64px',
                  borderRadius: '50%',
                  objectFit: 'cover',
                  border: '2px solid var(--border)',
                }}
              />
            ) : (
              <div
                style={{
                  width: '64px',
                  height: '64px',
                  borderRadius: '50%',
                  background: 'var(--accent-dim)',
                  border: '2px solid var(--border-accent)',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  fontSize: '26px',
                  fontWeight: 700,
                  color: 'var(--accent)',
                  fontFamily: "'Syne', sans-serif",
                }}
              >
                {user?.name?.[0]?.toUpperCase() ?? '?'}
              </div>
            )}
            <div>
              <p style={{ fontSize: '18px', fontWeight: 600, marginBottom: '4px' }}>
                {user?.name ?? '—'}
              </p>
              <p style={{ fontSize: '14px', color: 'var(--text-secondary)' }}>
                {user?.email ?? '—'}
              </p>
              <p
                style={{
                  fontSize: '12px',
                  color: 'var(--text-muted)',
                  marginTop: '4px',
                }}
              >
                ID: {user?.id?.slice(0, 16)}…
              </p>
            </div>
          </div>
        </Card>

        {/* Voice references */}
        <Card>
          <h2 style={{ fontSize: '18px', marginBottom: '6px' }}>Voice Sample</h2>
          <p
            style={{
              fontSize: '14px',
              color: 'var(--text-secondary)',
              marginBottom: '24px',
              lineHeight: 1.6,
            }}
          >
            Record yourself reading the text below. Vox uses this sample so
            translated speech sounds like your voice.
          </p>

          <RecordVoicePanel onUploaded={loadVoiceRefs} />

          {/* Saved samples */}
          <div style={{ marginTop: '32px' }}>
            <p
              style={{
                fontSize: '12px',
                color: 'var(--text-muted)',
                textTransform: 'uppercase',
                letterSpacing: '0.08em',
                fontWeight: 500,
                marginBottom: '12px',
              }}
            >
              Saved samples ({voiceRefs.length})
            </p>

            {isLoadingVoice ? (
              <p style={{ color: 'var(--text-muted)', fontSize: '14px' }}>
                Loading…
              </p>
            ) : voiceRefs.length === 0 ? (
              <p style={{ color: 'var(--text-muted)', fontSize: '14px' }}>
                No voice samples yet. Record one above.
              </p>
            ) : (
              <div style={{ display: 'flex', flexDirection: 'column', gap: '10px' }}>
                {voiceRefs.map((vr) => (
                  <VoiceRefCard
                    key={vr.file_id}
                    voiceRef={vr}
                    onDeleted={() =>
                      setVoiceRefs((prev) =>
                        prev.filter((v) => v.file_id !== vr.file_id)
                      )
                    }
                  />
                ))}
              </div>
            )}
          </div>
        </Card>

        {/* Account */}
        <Card>
          <h3
            style={{
              fontSize: '16px',
              marginBottom: '12px',
              color: 'var(--text-secondary)',
            }}
          >
            Account
          </h3>
          <Button variant="danger" onClick={logout}>
            Sign out
          </Button>
        </Card>
      </main>
    </div>
  );
}
