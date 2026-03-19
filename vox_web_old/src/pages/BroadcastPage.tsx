import React, { useState, useRef, useEffect } from "react";
import Logo from "../components/Logo";
import WaveVisualizer from "../components/WaveVisualizer";
import { hub as hubApi } from "../api/client";

interface BroadcastPageProps {
  hubId: string;
  navigate: (to: string) => void;
}

const LANGUAGES = [
  { code: "en", label: "English" },
  { code: "ru", label: "Русский" },
  { code: "es", label: "Español" },
  { code: "de", label: "Deutsch" },
  { code: "fr", label: "Français" },
  { code: "zh", label: "中文" },
  { code: "ar", label: "العربية" },
  { code: "ja", label: "日本語" },
  { code: "pt", label: "Português" },
  { code: "it", label: "Italiano" },
  { code: "ko", label: "한국어" },
  { code: "tr", label: "Türkçe" },
];

const BroadcastPage: React.FC<BroadcastPageProps> = ({ hubId, navigate }) => {
  const [lang, setLang] = useState("en");
  const [broadcasting, setBroadcasting] = useState(false);
  const [active, setActive] = useState(false);
  const [err, setErr] = useState("");
  const [copied, setCopied] = useState(false);
  const [elapsed, setElapsed] = useState(0);

  const mediaRecorderRef = useRef<MediaRecorder | null>(null);
  const chunksRef = useRef<Blob[]>([]);
  const intervalRef = useRef<number | null>(null);
  const timerRef = useRef<number | null>(null);
  const startTimeRef = useRef<number>(0);

  const startBroadcast = async () => {
    setErr("");
    try {
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
      const recorder = new MediaRecorder(stream, { mimeType: "audio/webm" });
      mediaRecorderRef.current = recorder;
      chunksRef.current = [];

      recorder.ondataavailable = (e) => {
        if (e.data.size > 0) chunksRef.current.push(e.data);
      };
      recorder.start(2000);

      intervalRef.current = window.setInterval(async () => {
        if (chunksRef.current.length === 0) return;
        const blob = new Blob(chunksRef.current, { type: "audio/webm" });
        chunksRef.current = [];
        try {
          await hubApi.publish(hubId, blob, lang);
          setActive(true);
        } catch {
          setActive(false);
        }
      }, 2500);

      startTimeRef.current = Date.now();
      timerRef.current = window.setInterval(() => {
        setElapsed(Math.floor((Date.now() - startTimeRef.current) / 1000));
      }, 1000);

      setBroadcasting(true);
      setActive(true);
    } catch {
      setErr("Microphone access denied.");
    }
  };

  const stopBroadcast = () => {
    mediaRecorderRef.current?.stop();
    mediaRecorderRef.current?.stream.getTracks().forEach((t) => t.stop());
    if (intervalRef.current) clearInterval(intervalRef.current);
    if (timerRef.current) clearInterval(timerRef.current);
    setBroadcasting(false);
    setActive(false);
    setElapsed(0);
  };

  const handleReconnect = () => {
    hubApi.reconnect(hubId);
  };

  const handleCopy = () => {
    navigator.clipboard.writeText(`${window.location.origin}/#/room/${hubId}`);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const formatElapsed = (s: number) => {
    const m = Math.floor(s / 60)
      .toString()
      .padStart(2, "0");
    const sec = (s % 60).toString().padStart(2, "0");
    return `${m}:${sec}`;
  };

  useEffect(() => () => stopBroadcast(), []);

  const listenerUrl = `${window.location.origin}/#/room/${hubId}`;

  return (
    <div
      className="fade-in"
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
        <div style={{ display: "flex", alignItems: "center", gap: 12 }}>
          {broadcasting && (
            <span
              style={{
                fontSize: 10,
                fontFamily: "DM Mono",
                letterSpacing: "0.1em",
                color: "#e8ff5e",
                background: "#e8ff5e14",
                border: "1px solid #e8ff5e33",
                borderRadius: 4,
                padding: "2px 8px",
              }}
            >
              LIVE · {formatElapsed(elapsed)}
            </span>
          )}
          <button
            className="btn-ghost"
            style={{ fontSize: 12, padding: "6px 14px" }}
            onClick={() => {
              stopBroadcast();
              navigate("#/admin");
            }}
          >
            ← Back
          </button>
        </div>
      </nav>

      <main
        style={{
          flex: 1,
          display: "flex",
          flexDirection: "column",
          alignItems: "center",
          justifyContent: "center",
          padding: "32px 24px",
          gap: 40,
          maxWidth: 480,
          margin: "0 auto",
          width: "100%",
        }}
      >
        {/* HUB ID */}
        <div
          style={{
            width: "100%",
            background: "#0c0c0c",
            border: "1px solid #191919",
            borderRadius: 12,
            padding: "14px 16px",
            display: "flex",
            alignItems: "center",
            gap: 10,
          }}
        >
          <div
            style={{
              width: 7,
              height: 7,
              borderRadius: "50%",
              flexShrink: 0,
              background: broadcasting ? "#e8ff5e" : "#2a2a2a",
              boxShadow: broadcasting ? "0 0 7px #e8ff5e66" : "none",
              transition: "all 0.3s",
            }}
          />
          <span
            style={{
              fontFamily: "DM Mono",
              fontSize: 13,
              color: "#666",
              flex: 1,
              overflow: "hidden",
              textOverflow: "ellipsis",
              whiteSpace: "nowrap",
            }}
          >
            {hubId}
          </span>
          <button
            onClick={handleReconnect}
            title="Reconnect stream"
            style={iconBtnStyle}
          >
            <IconReconnect />
          </button>
          <button
            onClick={handleCopy}
            style={{
              background: "none",
              border: "1px solid #1e1e1e",
              borderRadius: 6,
              fontSize: 12,
              fontFamily: "inherit",
              padding: "4px 12px",
              cursor: "pointer",
              color: copied ? "#6fff6f" : "#555",
              transition: "color 0.2s",
              flexShrink: 0,
            }}
          >
            {copied ? "✓ Copied" : "Copy link"}
          </button>
        </div>

        {/* VISUALIZER */}
        <div
          style={{
            position: "relative",
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            width: 180,
            height: 180,
          }}
        >
          {broadcasting && (
            <>
              <div
                style={{
                  position: "absolute",
                  width: 140,
                  height: 140,
                  borderRadius: "50%",
                  border: "1px solid rgba(232,255,94,0.12)",
                  animation: "pulse-ring 2s ease infinite",
                }}
              />
              <div
                style={{
                  position: "absolute",
                  width: 170,
                  height: 170,
                  borderRadius: "50%",
                  border: "1px solid rgba(232,255,94,0.05)",
                  animation: "pulse-ring2 2s ease infinite",
                }}
              />
            </>
          )}
          <WaveVisualizer active={active} size={90} />
        </div>

        {/* STATUS TEXT */}
        <div style={{ textAlign: "center" }}>
          <p
            style={{
              fontSize: 18,
              fontWeight: 400,
              letterSpacing: "-0.02em",
              margin: "0 0 6px",
              color: broadcasting ? "#f0ede8" : "#444",
              transition: "color 0.3s",
            }}
          >
            {broadcasting ? "Broadcasting…" : "Ready to broadcast"}
          </p>
          <p style={{ color: "#333", fontSize: 13, margin: 0 }}>
            {broadcasting
              ? `Translating from ${
                  LANGUAGES.find((l) => l.code === lang)?.label
                }`
              : "Select a language and start"}
          </p>
        </div>

        {/* CONTROLS */}
        <div
          style={{
            width: "100%",
            display: "flex",
            flexDirection: "column",
            gap: 12,
          }}
        >
          {/* Language selector */}
          <div style={{ display: "flex", alignItems: "center", gap: 10 }}>
            <label
              style={{
                fontSize: 12,
                color: "#444",
                textTransform: "uppercase",
                letterSpacing: "0.06em",
                whiteSpace: "nowrap",
              }}
            >
              Source lang
            </label>
            <select
              value={lang}
              onChange={(e) => setLang(e.target.value)}
              disabled={broadcasting}
              className="lang-select"
              style={{
                flex: 1,
                fontSize: 13,
                padding: "6px 10px",
                opacity: broadcasting ? 0.4 : 1,
              }}
            >
              {LANGUAGES.map((l) => (
                <option key={l.code} value={l.code}>
                  {l.label}
                </option>
              ))}
            </select>
          </div>

          {/* Start / Stop */}
          {!broadcasting ? (
            <button
              className="btn-primary"
              onClick={startBroadcast}
              style={{ width: "100%", padding: "12px", fontSize: 14 }}
            >
              ▶ Start broadcast
            </button>
          ) : (
            <button
              className="btn-ghost"
              onClick={stopBroadcast}
              style={{
                width: "100%",
                padding: "12px",
                fontSize: 14,
                borderColor: "#ff5e5e",
                color: "#ff5e5e",
              }}
            >
              ■ Stop broadcast
            </button>
          )}
        </div>

        {/* Waveform strip */}
        {broadcasting && (
          <div
            style={{
              width: "100%",
              display: "flex",
              alignItems: "center",
              gap: 10,
              padding: "10px 14px",
              background: "#0c0c0c",
              border: "1px solid #141414",
              borderRadius: 10,
            }}
          >
            <WaveVisualizer active={active} size={28} />
            <span style={{ fontSize: 11, color: "#333" }}>
              Streaming audio…
            </span>
            <span
              style={{
                marginLeft: "auto",
                fontFamily: "DM Mono",
                fontSize: 11,
                color: "#e8ff5e66",
              }}
            >
              {formatElapsed(elapsed)}
            </span>
          </div>
        )}

        {/* Listener link */}
        <div
          style={{
            width: "100%",
            borderTop: "1px solid #111",
            paddingTop: 20,
            display: "flex",
            flexDirection: "column",
            gap: 8,
          }}
        >
          <span
            style={{
              fontSize: 11,
              color: "#333",
              textTransform: "uppercase",
              letterSpacing: "0.06em",
            }}
          >
            Listener link
          </span>
          <div
            style={{
              display: "flex",
              alignItems: "center",
              gap: 8,
              background: "#0a0a0a",
              border: "1px solid #181818",
              borderRadius: 8,
              padding: "8px 12px",
            }}
          >
            <span
              style={{
                fontFamily: "DM Mono",
                fontSize: 11,
                color: "#333",
                flex: 1,
                overflow: "hidden",
                textOverflow: "ellipsis",
                whiteSpace: "nowrap",
              }}
            >
              {listenerUrl}
            </span>
            <button
              onClick={handleCopy}
              style={{
                background: "none",
                border: "none",
                cursor: "pointer",
                color: copied ? "#6fff6f" : "#444",
                fontSize: 11,
                fontFamily: "inherit",
                padding: 0,
                flexShrink: 0,
                transition: "color 0.2s",
              }}
            >
              {copied ? "✓" : "Copy"}
            </button>
          </div>
        </div>

        {err && (
          <p style={{ color: "#ff5e5e", fontSize: 13, margin: 0 }}>{err}</p>
        )}
      </main>
    </div>
  );
};

const iconBtnStyle: React.CSSProperties = {
  background: "none",
  border: "1px solid transparent",
  borderRadius: 6,
  color: "#555",
  cursor: "pointer",
  width: 28,
  height: 28,
  display: "flex",
  alignItems: "center",
  justifyContent: "center",
  padding: 0,
  transition: "all 0.15s",
  flexShrink: 0,
};

const IconReconnect = () => (
  <svg
    width="13"
    height="13"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    strokeWidth="2"
    strokeLinecap="round"
    strokeLinejoin="round"
  >
    <polyline points="23 4 23 10 17 10" />
    <path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10" />
  </svg>
);

export default BroadcastPage;
