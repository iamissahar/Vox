import React, { useState, useEffect, useRef } from "react";
import Logo from "../components/Logo";
import WaveVisualizer from "../components/WaveVisualizer";
import { hub as hubApi } from "../api/client";

interface RoomPageProps {
  roomId: string;
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

const RoomPage: React.FC<RoomPageProps> = ({ roomId, navigate }) => {
  const [lang, setLang] = useState("en");
  const [connected, setConnected] = useState(false);
  const [connecting, setConnecting] = useState(true);
  const [active, setActive] = useState(false);
  const [error, setError] = useState("");

  const audioRef = useRef<HTMLAudioElement | null>(null);
  const abortRef = useRef<AbortController | null>(null);

  useEffect(() => {
    if (!roomId) return;

    const connect = async () => {
      setConnecting(true);
      setError("");

      try {
        // Use fetch with streaming to receive audio/mpeg chunks
        const controller = new AbortController();
        abortRef.current = controller;

        const streamUrl = hubApi.listenUrl(roomId);

        // Create an audio element that plays the stream
        const audio = new Audio(streamUrl);
        audioRef.current = audio;

        audio.oncanplay = () => {
          setConnecting(false);
          setConnected(true);
        };

        audio.onplaying = () => setActive(true);
        audio.onpause = () => setActive(false);
        audio.onwaiting = () => setActive(false);
        audio.onerror = () => {
          setConnected(false);
          setConnecting(false);
          setError("Stream unavailable. The host may not have started yet.");
          setActive(false);
        };

        await audio.play().catch(() => {
          // Autoplay blocked — user needs to interact
          setConnecting(false);
          setConnected(true);
        });
      } catch {
        setConnecting(false);
        setError("Could not connect to hub.");
      }
    };

    connect();

    return () => {
      abortRef.current?.abort();
      if (audioRef.current) {
        audioRef.current.pause();
        audioRef.current.src = "";
      }
    };
  }, [roomId]);

  const handlePlay = () => {
    if (audioRef.current) {
      audioRef.current.play().catch(() => {});
    }
  };

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
        <div className="tag">
          <div
            className="dot-live"
            style={{ background: connected ? "#e8ff5e" : "#333" }}
          />
          <span style={{ fontFamily: "DM Mono", letterSpacing: "0.06em" }}>
            {roomId}
          </span>
        </div>
      </nav>

      <main
        style={{
          flex: 1,
          display: "flex",
          flexDirection: "column",
          alignItems: "center",
          justifyContent: "center",
          padding: 24,
          gap: 48,
        }}
      >
        {connecting ? (
          <div
            style={{
              display: "flex",
              flexDirection: "column",
              alignItems: "center",
              gap: 16,
            }}
          >
            <div
              style={{
                width: 28,
                height: 28,
                border: "2px solid #222",
                borderTopColor: "#e8ff5e",
                borderRadius: "50%",
                animation: "spin 0.8s linear infinite",
              }}
            />
            <p style={{ color: "#444", fontSize: 13 }}>Connecting to hub…</p>
          </div>
        ) : (
          <>
            {/* VISUAL */}
            <div
              style={{
                position: "relative",
                display: "flex",
                alignItems: "center",
                justifyContent: "center",
                width: 200,
                height: 200,
              }}
            >
              {active && (
                <>
                  <div
                    style={{
                      position: "absolute",
                      width: 160,
                      height: 160,
                      borderRadius: "50%",
                      border: "1px solid rgba(232,255,94,0.15)",
                      animation: "pulse-ring 2s ease infinite",
                    }}
                  />
                  <div
                    style={{
                      position: "absolute",
                      width: 190,
                      height: 190,
                      borderRadius: "50%",
                      border: "1px solid rgba(232,255,94,0.07)",
                      animation: "pulse-ring2 2s ease infinite",
                    }}
                  />
                </>
              )}
              <WaveVisualizer active={active} size={100} />
            </div>

            {/* STATUS */}
            <div style={{ textAlign: "center" }}>
              {error ? (
                <>
                  <p
                    style={{
                      fontSize: 16,
                      color: "#555",
                      marginBottom: 8,
                    }}
                  >
                    {error}
                  </p>
                  <button
                    className="btn-ghost"
                    style={{ fontSize: 12, padding: "8px 16px" }}
                    onClick={() => window.location.reload()}
                  >
                    Retry
                  </button>
                </>
              ) : (
                <>
                  <p
                    style={{
                      fontSize: 18,
                      fontWeight: 400,
                      letterSpacing: "-0.02em",
                      marginBottom: 8,
                    }}
                  >
                    {active ? "Receiving audio…" : "Waiting for broadcast"}
                  </p>
                  <p style={{ color: "#444", fontSize: 13 }}>
                    {active
                      ? `Translating to ${LANGUAGES.find((l) => l.code === lang)?.label}`
                      : "The host hasn't started yet"}
                  </p>
                  {connected && !active && (
                    <button
                      className="btn-ghost"
                      style={{
                        marginTop: 16,
                        fontSize: 12,
                        padding: "8px 16px",
                      }}
                      onClick={handlePlay}
                    >
                      ▶ Tap to play
                    </button>
                  )}
                </>
              )}
            </div>

            {/* LANGUAGE SELECTOR */}
            <div
              style={{
                display: "flex",
                flexDirection: "column",
                alignItems: "center",
                gap: 12,
              }}
            >
              <label
                style={{
                  fontSize: 12,
                  color: "#444",
                  textTransform: "uppercase",
                  letterSpacing: "0.06em",
                }}
              >
                Output language
              </label>
              <select
                className="lang-select"
                value={lang}
                onChange={(e) => setLang(e.target.value)}
              >
                {LANGUAGES.map((l) => (
                  <option key={l.code} value={l.code}>
                    {l.label}
                  </option>
                ))}
              </select>
            </div>
          </>
        )}
      </main>

      {/* FOOTER */}
      <div
        style={{
          padding: "20px 28px",
          borderTop: "1px solid #0e0e0e",
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
        }}
      >
        <button
          className="btn-ghost"
          style={{ fontSize: 12, padding: "8px 16px" }}
          onClick={() => navigate("#/room")}
        >
          Leave
        </button>
        <span style={{ fontSize: 12, color: "#2a2a2a" }}>
          vox · live translation
        </span>
      </div>
    </div>
  );
};

export default RoomPage;
