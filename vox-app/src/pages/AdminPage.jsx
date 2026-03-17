import { useState, useRef } from "react";
import { authState } from "../App";
import Logo from "./../components/Logo";
import WaveVisualizer from "./../components/WaveVisualizer";
import QRPlaceholder from "../components/QRPlaceholder";

const ROOM_ID = "vox-" + Math.random().toString(36).slice(2, 8).toUpperCase();

export default function AdminPage({ navigate, user, onLogout }) {
  const [isRecording, setIsRecording] = useState(false);
  const [showProfile, setShowProfile] = useState(false);
  const [showQR, setShowQR] = useState(false);
  const [copied, setCopied] = useState(false);
  const [streamBytes, setStreamBytes] = useState(0);
  const [duration, setSec] = useState(0);
  const [audioLevel, setAudioLevel] = useState(0);
  const [freqs, setFreqs] = useState(new Array(8).fill(0));
  const mediaRef = useRef(null);
  const streamRef = useRef(null);
  const timerRef = useRef(null);
  const bytesRef = useRef(null);
  const analyserRef = useRef(null);
  const animFrameRef = useRef(null);

  if (!user) {
    navigate("#/login");
    return null;
  }

  const startRecording = async () => {
    try {
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
      streamRef.current = stream;
      const recorder = new MediaRecorder(stream, {
        mimeType: "audio/webm;codecs=opus",
      });
      mediaRef.current = recorder;
      recorder.ondataavailable = (e) => {
        if (e.data.size > 0) {
          // STUB: send to API
          // fetch("https://your-api.example.com/stream/" + ROOM_ID, {
          //   method: "POST", body: e.data,
          //   headers: { "Content-Type": "audio/webm" }
          // });
          setStreamBytes((b) => b + e.data.size);
        }
      };
      recorder.start(250);
      setIsRecording(true);
      setSec(0);
      timerRef.current = setInterval(() => setSec((s) => s + 1), 1000);

      // Web Audio API — анализатор
      const audioCtx = new AudioContext();
      const source = audioCtx.createMediaStreamSource(stream);
      const analyser = audioCtx.createAnalyser();
      analyser.fftSize = 64;
      source.connect(analyser);
      analyserRef.current = analyser;

      const dataArray = new Uint8Array(analyser.frequencyBinCount);
      const tick = () => {
        analyser.getByteFrequencyData(dataArray);
        const avg = Math.min(
          (dataArray.reduce((a, b) => a + b, 0) / dataArray.length / 255) * 4,
          1,
        );
        setAudioLevel(avg);
        const step = Math.floor(dataArray.length / 8);
        setFreqs(
          Array.from({ length: 8 }, (_, i) =>
            Math.min((dataArray[i * step] / 255) * 3, 1),
          ),
        );
        animFrameRef.current = requestAnimationFrame(tick);
      };
      tick();
    } catch (_) {
      alert("Microphone access denied.");
    }
  };

  const stopRecording = () => {
    mediaRef.current?.stop();
    streamRef.current?.getTracks().forEach((t) => t.stop());
    clearInterval(timerRef.current);
    cancelAnimationFrame(animFrameRef.current); // ← добавлено
    setIsRecording(false);
    setAudioLevel(0); // ← добавлено
    setFreqs(new Array(8).fill(0)); // ← добавлено
  };

  const fmt = (s) =>
    `${String(Math.floor(s / 60)).padStart(2, "0")}:${String(s % 60).padStart(2, "0")}`;
  const fmtBytes = (b) =>
    b < 1024
      ? b + " B"
      : b < 1048576
        ? (b / 1024).toFixed(1) + " KB"
        : (b / 1048576).toFixed(2) + " MB";

  const copyId = () => {
    navigator.clipboard.writeText(ROOM_ID);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div
      style={{ minHeight: "100vh", display: "flex", flexDirection: "column" }}
    >
      {/* NAV */}
      <nav
        style={{
          padding: "20px 28px",
          display: "flex",
          alignItems: "center",
          justifyContent: "space-between",
          borderBottom: "1px solid #111",
        }}
      >
        <Logo size={22} />
        <div style={{ position: "relative" }}>
          <button
            onClick={() => setShowProfile((p) => !p)}
            style={{
              background: "#151515",
              border: "1px solid #1e1e1e",
              borderRadius: "100px",
              padding: "8px 16px",
              color: "#f0ede8",
              cursor: "pointer",
              display: "flex",
              alignItems: "center",
              gap: 8,
              fontSize: 13,
            }}
          >
            <div
              style={{
                width: 24,
                height: 24,
                borderRadius: "50%",
                background: "#e8ff5e",
                display: "flex",
                alignItems: "center",
                justifyContent: "center",
                color: "#080808",
                fontSize: 11,
                fontWeight: 600,
              }}
            >
              {user.name[0]}
            </div>
            <span style={{ color: "#888" }}>▾</span>
          </button>
          {showProfile && (
            <div
              className="slide-up"
              style={{
                position: "absolute",
                right: 0,
                top: "calc(100% + 8px)",
                background: "#0f0f0f",
                border: "1px solid #1a1a1a",
                borderRadius: 16,
                padding: 20,
                width: 220,
                zIndex: 100,
              }}
            >
              <div
                style={{
                  display: "flex",
                  alignItems: "center",
                  gap: 12,
                  marginBottom: 16,
                }}
              >
                <div
                  style={{
                    width: 36,
                    height: 36,
                    borderRadius: "50%",
                    background: "#e8ff5e",
                    display: "flex",
                    alignItems: "center",
                    justifyContent: "center",
                    color: "#080808",
                    fontWeight: 600,
                    fontSize: 15,
                  }}
                >
                  {user.name[0]}
                </div>
                <div>
                  <p style={{ fontSize: 14, fontWeight: 500 }}>{user.name}</p>
                  <p style={{ fontSize: 12, color: "#555" }}>{user.email}</p>
                </div>
              </div>
              <div
                style={{
                  background: "#141414",
                  borderRadius: 10,
                  padding: "8px 12px",
                  marginBottom: 12,
                }}
              >
                <p
                  style={{
                    fontSize: 11,
                    color: "#444",
                    textTransform: "uppercase",
                    letterSpacing: "0.06em",
                    marginBottom: 4,
                  }}
                >
                  Role
                </p>
                <p style={{ fontSize: 13, color: "#e8ff5e" }}>Administrator</p>
              </div>
              <button
                className="btn-ghost"
                style={{ width: "100%", fontSize: 13 }}
                onClick={() => {
                  authState.user = null;
                  onLogout();
                  navigate("#/");
                }}
              >
                Sign out
              </button>
            </div>
          )}
        </div>
      </nav>

      {/* MAIN */}
      <main
        style={{
          flex: 1,
          display: "flex",
          flexDirection: "column",
          alignItems: "center",
          justifyContent: "center",
          padding: 24,
          gap: 40,
        }}
        className="fade-in"
      >
        {/* STATUS */}
        {isRecording && <WaveVisualizer bars={8} freqs={freqs} />}

        {/* BIG RECORD BUTTON */}
        <div
          style={{
            position: "relative",
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
          }}
        >
          {isRecording && (
            <>
              <div
                style={{
                  position: "absolute",
                  width: 200,
                  height: 200,
                  borderRadius: "50%",
                  border: "1.5px solid #e8ff5e",
                  transform: `scale(${1 + audioLevel * 0.4})`,
                  opacity: 0.2 + audioLevel * 0.6,
                  transition: "transform 0.08s ease, opacity 0.08s ease",
                }}
              />
              <div
                style={{
                  position: "absolute",
                  width: 240,
                  height: 240,
                  borderRadius: "50%",
                  border: "1px solid #e8ff5e",
                  transform: `scale(${1 + audioLevel * 0.6})`,
                  opacity: 0.1 + audioLevel * 0.3,
                  transition: "transform 0.08s ease, opacity 0.08s ease",
                }}
              />
            </>
          )}
          <button
            onClick={isRecording ? stopRecording : startRecording}
            style={{
              width: 160,
              height: 160,
              borderRadius: "50%",
              background: isRecording ? "#ff3b3b" : "#e8ff5e",
              border: "none",
              cursor: "pointer",
              display: "flex",
              flexDirection: "column",
              alignItems: "center",
              justifyContent: "center",
              gap: 8,
              transition: "all 0.3s cubic-bezier(.34,1.56,.64,1)",
              transform: isRecording ? "scale(0.95)" : "scale(1)",
              boxShadow: isRecording
                ? "0 0 60px rgba(255,59,59,0.25)"
                : "0 0 40px rgba(232,255,94,0.15)",
            }}
          >
            <svg width="32" height="32" viewBox="0 0 32 32" fill="none">
              {isRecording ? (
                <rect
                  x="9"
                  y="9"
                  width="14"
                  height="14"
                  rx="3"
                  fill="#080808"
                />
              ) : (
                <>
                  <rect
                    x="12"
                    y="4"
                    width="8"
                    height="16"
                    rx="4"
                    fill="#080808"
                  />
                  <path
                    d="M6 16c0 5.5 4 9 10 9s10-3.5 10-9"
                    stroke="#080808"
                    strokeWidth="2"
                    strokeLinecap="round"
                  />
                  <line
                    x1="16"
                    y1="25"
                    x2="16"
                    y2="30"
                    stroke="#080808"
                    strokeWidth="2"
                    strokeLinecap="round"
                  />
                </>
              )}
            </svg>
            <span
              style={{
                color: "#080808",
                fontSize: 12,
                fontWeight: 600,
                letterSpacing: "0.04em",
                textTransform: "uppercase",
              }}
            >
              {isRecording ? "Stop" : "Record"}
            </span>
          </button>
        </div>

        {/* STATS */}
        {isRecording && (
          <div className="slide-up" style={{ display: "flex", gap: 16 }}>
            <div style={{ display: "flex", gap: 8, alignItems: "center" }}>
              <div className="dot-live" />
              <span
                style={{
                  fontFamily: "DM Mono",
                  fontSize: 13,
                  color: "#e8ff5e",
                }}
              >
                {fmt(duration)}
              </span>
            </div>
            <span style={{ color: "#222" }}>·</span>
            <span
              style={{ fontFamily: "DM Mono", fontSize: 13, color: "#555" }}
            >
              {fmtBytes(streamBytes)}
            </span>
            <span style={{ color: "#222" }}>·</span>
            <span
              style={{
                fontSize: 12,
                color: "#333",
                textTransform: "uppercase",
                letterSpacing: "0.04em",
              }}
            >
              streaming to {ROOM_ID}
            </span>
          </div>
        )}

        {!isRecording && (
          <p style={{ color: "#333", fontSize: 13, letterSpacing: "-0.01em" }}>
            Press to start broadcasting to room{" "}
            <span style={{ color: "#555", fontFamily: "DM Mono" }}>
              {ROOM_ID}
            </span>
          </p>
        )}

        {/* ROOM ACTIONS */}
        <div
          style={{
            display: "flex",
            gap: 10,
            flexWrap: "wrap",
            justifyContent: "center",
          }}
        >
          <button
            className="btn-ghost"
            onClick={copyId}
            style={{ fontFamily: "DM Mono", fontSize: 13 }}
          >
            {copied ? "✓ Copied" : `⊞  ${ROOM_ID}`}
          </button>
          <button
            className="btn-ghost"
            onClick={() => setShowQR((p) => !p)}
            style={{ fontSize: 13 }}
          >
            {showQR ? "Hide QR" : "Show QR"}
          </button>
        </div>

        {showQR && (
          <div
            className="slide-up card"
            style={{
              display: "flex",
              flexDirection: "column",
              alignItems: "center",
              gap: 16,
              padding: 28,
            }}
          >
            <QRPlaceholder value={ROOM_ID} size={160} />
            <p
              style={{
                fontSize: 12,
                color: "#444",
                letterSpacing: "0.04em",
                textTransform: "uppercase",
              }}
            >
              Scan to join room
            </p>
          </div>
        )}
      </main>
    </div>
  );
}
