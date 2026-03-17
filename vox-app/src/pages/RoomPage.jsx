import { useState, useRef, useEffect } from "react";
import Logo from "./../components/Logo";

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

export default function RoomPage({ roomId, navigate }) {
  const [lang, setLang] = useState("en");
  const [connected, setConnected] = useState(false);
  const [connecting, setConnecting] = useState(true);
  const [active, setActive] = useState(false);
  const waveRef = useRef(null);

  useEffect(() => {
    // STUB: WebSocket/SSE connection to room
    // const ws = new WebSocket(`wss://your-api.example.com/rooms/${roomId}/listen`);
    // ws.onopen = () => setConnected(true);
    const t = setTimeout(() => {
      setConnecting(false);
      setConnected(true);
    }, 1200);

    // Simulate occasional audio activity
    const sim = setInterval(() => setActive((a) => !a), 2800);

    return () => {
      clearTimeout(t);
      clearInterval(sim);
    };
  }, [roomId]);

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
            <p style={{ color: "#444", fontSize: 13 }}>Connecting to room…</p>
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
              <div
                style={{
                  width: 100,
                  height: 100,
                  borderRadius: "50%",
                  background: active
                    ? "rgba(232,255,94,0.08)"
                    : "rgba(232,255,94,0.03)",
                  border: `1.5px solid ${active ? "rgba(232,255,94,0.4)" : "rgba(232,255,94,0.12)"}`,
                  display: "flex",
                  alignItems: "center",
                  justifyContent: "center",
                  transition: "all 0.6s ease",
                }}
              >
                {active ? (
                  <div
                    style={{
                      display: "flex",
                      gap: 4,
                      alignItems: "flex-end",
                      height: 28,
                    }}
                  >
                    {[1, 2, 3, 4, 5].map((i) => (
                      <div
                        key={i}
                        className="wave-bar"
                        style={{
                          height: `${8 + i * 4}px`,
                          animationDelay: `${i * 0.12}s`,
                        }}
                      />
                    ))}
                  </div>
                ) : (
                  <svg width="24" height="24" viewBox="0 0 24 24" fill="none">
                    <path
                      d="M12 2a4 4 0 014 4v6a4 4 0 01-8 0V6a4 4 0 014-4z"
                      stroke="#333"
                      strokeWidth="1.5"
                    />
                    <path
                      d="M4 12c0 4.4 3.6 8 8 8s8-3.6 8-8"
                      stroke="#333"
                      strokeWidth="1.5"
                      strokeLinecap="round"
                    />
                  </svg>
                )}
              </div>
            </div>

            {/* STATUS TEXT */}
            <div style={{ textAlign: "center" }}>
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
              <div style={{ position: "relative" }}>
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
}
