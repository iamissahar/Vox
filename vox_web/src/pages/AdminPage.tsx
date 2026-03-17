import React, { useState, useRef, useEffect } from "react";
import Logo from "../components/Logo";
import WaveVisualizer from "../components/WaveVisualizer";
import { hub as hubApi, user as userApi, ApiError } from "../api/client";
import type { UserInfo } from "../types";

interface AdminPageProps {
  navigate: (to: string) => void;
  currentUser: UserInfo | null;
  onLogout: () => void;
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

type Tab = "broadcast" | "voice" | "profile";

const AdminPage: React.FC<AdminPageProps> = ({
  navigate,
  currentUser,
  onLogout,
}) => {
  const [tab, setTab] = useState<Tab>("broadcast");

  // Broadcast state
  const [hubId, setHubId] = useState("");
  const [lang, setLang] = useState("en");
  const [hubCreated, setHubCreated] = useState(false);
  const [broadcasting, setBroadcasting] = useState(false);
  const [active, setActive] = useState(false);
  const [hubErr, setHubErr] = useState("");
  const [hubLoading, setHubLoading] = useState(false);

  // Voice state
  const [textRef, setTextRef] = useState("");
  const [voiceRecording, setVoiceRecording] = useState(false);
  const [voiceStatus, setVoiceStatus] = useState<"idle" | "recording" | "uploading" | "done" | "error">("idle");
  const voiceRecorderRef = useRef<MediaRecorder | null>(null);
  const voiceChunksRef = useRef<Blob[]>([]);

  // Broadcast recording
  const mediaRecorderRef = useRef<MediaRecorder | null>(null);
  const chunksRef = useRef<Blob[]>([]);
  const broadcastIntervalRef = useRef<number | null>(null);

  const createHub = async () => {
    if (!hubId.trim()) {
      setHubErr("Please enter a hub ID.");
      return;
    }
    setHubLoading(true);
    setHubErr("");
    try {
      await hubApi.create(hubId.trim());
      setHubCreated(true);
    } catch (e) {
      if (e instanceof ApiError) {
        setHubErr(
          e.code === 409
            ? "A hub with this ID already exists."
            : e.message
        );
      } else {
        setHubErr("Failed to create hub.");
      }
    } finally {
      setHubLoading(false);
    }
  };

  const startBroadcast = async () => {
    try {
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
      const recorder = new MediaRecorder(stream, { mimeType: "audio/webm" });
      mediaRecorderRef.current = recorder;
      chunksRef.current = [];

      recorder.ondataavailable = (e) => {
        if (e.data.size > 0) chunksRef.current.push(e.data);
      };

      // Send chunks every 2 seconds
      recorder.start(2000);

      broadcastIntervalRef.current = window.setInterval(async () => {
        if (chunksRef.current.length === 0) return;
        const blob = new Blob(chunksRef.current, { type: "audio/webm" });
        chunksRef.current = [];
        try {
          await hubApi.publish(hubId.trim(), blob, lang);
          setActive(true);
        } catch {
          setActive(false);
        }
      }, 2500);

      setBroadcasting(true);
      setActive(true);
    } catch {
      setHubErr("Microphone access denied.");
    }
  };

  const stopBroadcast = () => {
    mediaRecorderRef.current?.stop();
    mediaRecorderRef.current?.stream
      .getTracks()
      .forEach((t) => t.stop());
    if (broadcastIntervalRef.current)
      clearInterval(broadcastIntervalRef.current);
    setBroadcasting(false);
    setActive(false);
  };

  useEffect(() => {
    return () => {
      stopBroadcast();
    };
  }, []);

  // Voice recording
  const startVoiceRecording = async () => {
    try {
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
      const recorder = new MediaRecorder(stream, { mimeType: "audio/webm" });
      voiceRecorderRef.current = recorder;
      voiceChunksRef.current = [];
      recorder.ondataavailable = (e) => {
        if (e.data.size > 0) voiceChunksRef.current.push(e.data);
      };
      recorder.start();
      setVoiceRecording(true);
      setVoiceStatus("recording");
    } catch {
      setVoiceStatus("error");
    }
  };

  const stopVoiceRecording = async () => {
    if (!voiceRecorderRef.current) return;
    voiceRecorderRef.current.stop();
    voiceRecorderRef.current.stream.getTracks().forEach((t) => t.stop());
    setVoiceRecording(false);
    setVoiceStatus("uploading");

    await new Promise<void>((res) => {
      voiceRecorderRef.current!.onstop = async () => {
        const blob = new Blob(voiceChunksRef.current, { type: "audio/webm" });
        try {
          await userApi.uploadVoice(blob, textRef);
          setVoiceStatus("done");
        } catch {
          setVoiceStatus("error");
        }
        res();
      };
    });
  };

  const tabStyle = (t: Tab): React.CSSProperties => ({
    background: "none",
    border: "none",
    fontFamily: "inherit",
    fontSize: 13,
    cursor: "pointer",
    padding: "8px 0",
    color: tab === t ? "#f0ede8" : "#444",
    borderBottom: `1px solid ${tab === t ? "#e8ff5e" : "transparent"}`,
    transition: "all 0.2s",
    letterSpacing: "0.02em",
  });

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
        <div style={{ display: "flex", alignItems: "center", gap: 16 }}>
          {currentUser && (
            <span style={{ color: "#444", fontSize: 13 }}>
              {currentUser.name || currentUser.email}
            </span>
          )}
          <button
            className="btn-ghost"
            style={{ fontSize: 12, padding: "6px 14px" }}
            onClick={onLogout}
          >
            Log out
          </button>
        </div>
      </nav>

      {/* TABS */}
      <div
        style={{
          padding: "0 28px",
          borderBottom: "1px solid #111",
          display: "flex",
          gap: 24,
        }}
      >
        {(["broadcast", "voice", "profile"] as Tab[]).map((t) => (
          <button key={t} style={tabStyle(t)} onClick={() => setTab(t)}>
            {t.charAt(0).toUpperCase() + t.slice(1)}
          </button>
        ))}
      </div>

      <main
        style={{
          flex: 1,
          display: "flex",
          flexDirection: "column",
          alignItems: "center",
          justifyContent: "center",
          padding: 32,
          gap: 32,
          maxWidth: 500,
          margin: "0 auto",
          width: "100%",
        }}
      >
        {/* ─── BROADCAST TAB ─── */}
        {tab === "broadcast" && (
          <>
            {!hubCreated ? (
              <div
                className="card slide-up"
                style={{
                  width: "100%",
                  display: "flex",
                  flexDirection: "column",
                  gap: 16,
                }}
              >
                <h2
                  style={{
                    fontSize: 20,
                    fontWeight: 400,
                    letterSpacing: "-0.03em",
                  }}
                >
                  Create a hub
                </h2>
                <p style={{ color: "#555", fontSize: 13, lineHeight: 1.6 }}>
                  Choose a unique ID for your broadcast room. Share it with
                  listeners.
                </p>
                <input
                  className="input-field"
                  placeholder="Hub ID (e.g. my-event-2025)"
                  value={hubId}
                  onChange={(e) => setHubId(e.target.value)}
                  onKeyDown={(e) => e.key === "Enter" && createHub()}
                />
                <div style={{ display: "flex", alignItems: "center", gap: 12 }}>
                  <label style={{ fontSize: 12, color: "#555" }}>
                    Broadcast language
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
                {hubErr && (
                  <p style={{ color: "#ff5e5e", fontSize: 13 }}>{hubErr}</p>
                )}
                <button
                  className="btn-primary"
                  onClick={createHub}
                  disabled={hubLoading}
                  style={{ opacity: hubLoading ? 0.7 : 1 }}
                >
                  {hubLoading ? "Creating…" : "Create hub →"}
                </button>
              </div>
            ) : (
              <div
                className="slide-up"
                style={{
                  display: "flex",
                  flexDirection: "column",
                  alignItems: "center",
                  gap: 32,
                  width: "100%",
                }}
              >
                {/* Live tag */}
                <div style={{ display: "flex", alignItems: "center", gap: 16 }}>
                  <div className="tag">
                    <div
                      className="dot-live"
                      style={{
                        background: broadcasting ? "#e8ff5e" : "#333",
                      }}
                    />
                    <span
                      style={{
                        fontFamily: "DM Mono",
                        letterSpacing: "0.06em",
                      }}
                    >
                      {hubId}
                    </span>
                  </div>
                  <span
                    style={{
                      fontSize: 12,
                      color: "#333",
                    }}
                  >
                    {broadcasting ? "live" : "ready"}
                  </span>
                </div>

                {/* Wave visualizer */}
                <div
                  style={{
                    position: "relative",
                    width: 200,
                    height: 200,
                    display: "flex",
                    alignItems: "center",
                    justifyContent: "center",
                  }}
                >
                  {broadcasting && (
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
                  <WaveVisualizer active={active} size={120} />
                </div>

                {/* Controls */}
                <div
                  style={{
                    display: "flex",
                    gap: 12,
                    width: "100%",
                    flexWrap: "wrap",
                    justifyContent: "center",
                  }}
                >
                  {!broadcasting ? (
                    <button
                      className="btn-primary"
                      onClick={startBroadcast}
                      style={{ flex: 1 }}
                    >
                      Start broadcasting
                    </button>
                  ) : (
                    <button
                      className="btn-ghost"
                      onClick={stopBroadcast}
                      style={{
                        flex: 1,
                        borderColor: "#ff5e5e",
                        color: "#ff5e5e",
                      }}
                    >
                      Stop
                    </button>
                  )}
                  <button
                    className="btn-ghost"
                    onClick={() => {
                      navigate(`#/room/${hubId}`);
                    }}
                    style={{ flex: 1 }}
                  >
                    Share link →
                  </button>
                </div>

                {/* Share URL */}
                <div
                  style={{
                    background: "#0a0a0a",
                    border: "1px solid #1a1a1a",
                    borderRadius: 10,
                    padding: "10px 16px",
                    width: "100%",
                    display: "flex",
                    alignItems: "center",
                    justifyContent: "space-between",
                    gap: 12,
                  }}
                >
                  <span
                    style={{
                      fontFamily: "DM Mono",
                      fontSize: 12,
                      color: "#555",
                      overflow: "hidden",
                      textOverflow: "ellipsis",
                      whiteSpace: "nowrap",
                    }}
                  >
                    {window.location.origin}/#/room/{hubId}
                  </span>
                  <button
                    style={{
                      background: "none",
                      border: "none",
                      color: "#e8ff5e",
                      fontSize: 12,
                      cursor: "pointer",
                      fontFamily: "inherit",
                      whiteSpace: "nowrap",
                    }}
                    onClick={() =>
                      navigator.clipboard.writeText(
                        `${window.location.origin}/#/room/${hubId}`
                      )
                    }
                  >
                    Copy
                  </button>
                </div>
              </div>
            )}
          </>
        )}

        {/* ─── VOICE TAB ─── */}
        {tab === "voice" && (
          <div
            className="card slide-up"
            style={{
              width: "100%",
              display: "flex",
              flexDirection: "column",
              gap: 20,
            }}
          >
            <div>
              <h2
                style={{
                  fontSize: 20,
                  fontWeight: 400,
                  letterSpacing: "-0.03em",
                  marginBottom: 8,
                }}
              >
                Voice reference
              </h2>
              <p style={{ color: "#555", fontSize: 13, lineHeight: 1.6 }}>
                Record a voice sample for better synthesis quality. Read the
                reference text aloud while recording.
              </p>
            </div>

            <textarea
              className="input-field"
              placeholder="Reference text (read this aloud during recording)"
              value={textRef}
              onChange={(e) => setTextRef(e.target.value)}
              rows={4}
              style={{ resize: "none", lineHeight: 1.6 }}
            />

            <div
              style={{
                display: "flex",
                flexDirection: "column",
                alignItems: "center",
                gap: 16,
              }}
            >
              {voiceStatus === "recording" && (
                <div
                  style={{ display: "flex", alignItems: "center", gap: 8 }}
                >
                  <div className="dot-live" />
                  <span style={{ fontSize: 13, color: "#e8ff5e" }}>
                    Recording…
                  </span>
                </div>
              )}
              {voiceStatus === "uploading" && (
                <div
                  style={{
                    width: 20,
                    height: 20,
                    border: "2px solid #222",
                    borderTopColor: "#e8ff5e",
                    borderRadius: "50%",
                    animation: "spin 0.8s linear infinite",
                  }}
                />
              )}
              {voiceStatus === "done" && (
                <p style={{ color: "#6fff6f", fontSize: 13 }}>
                  ✓ Voice reference uploaded
                </p>
              )}
              {voiceStatus === "error" && (
                <p style={{ color: "#ff5e5e", fontSize: 13 }}>
                  Failed to upload. Try again.
                </p>
              )}

              {!voiceRecording ? (
                <button
                  className="btn-primary"
                  onClick={startVoiceRecording}
                  disabled={!textRef.trim()}
                  style={{ opacity: !textRef.trim() ? 0.4 : 1, width: "100%" }}
                >
                  Start recording
                </button>
              ) : (
                <button
                  className="btn-ghost"
                  onClick={stopVoiceRecording}
                  style={{
                    width: "100%",
                    borderColor: "#ff5e5e",
                    color: "#ff5e5e",
                  }}
                >
                  Stop & upload
                </button>
              )}
            </div>
          </div>
        )}

        {/* ─── PROFILE TAB ─── */}
        {tab === "profile" && currentUser && (
          <div
            className="card slide-up"
            style={{
              width: "100%",
              display: "flex",
              flexDirection: "column",
              gap: 20,
            }}
          >
            <h2
              style={{
                fontSize: 20,
                fontWeight: 400,
                letterSpacing: "-0.03em",
              }}
            >
              Profile
            </h2>

            {currentUser.picture && (
              <img
                src={currentUser.picture}
                alt="avatar"
                style={{
                  width: 64,
                  height: 64,
                  borderRadius: "50%",
                  border: "1px solid #222",
                }}
              />
            )}

            <div
              style={{
                display: "flex",
                flexDirection: "column",
                gap: 8,
              }}
            >
              {[
                { label: "Name", value: currentUser.name },
                { label: "Email", value: currentUser.email },
                { label: "ID", value: currentUser.id },
              ].map(({ label, value }) => (
                <div
                  key={label}
                  style={{
                    display: "flex",
                    justifyContent: "space-between",
                    padding: "10px 0",
                    borderBottom: "1px solid #141414",
                  }}
                >
                  <span style={{ color: "#555", fontSize: 13 }}>{label}</span>
                  <span
                    style={{
                      fontSize: 13,
                      fontFamily: label === "ID" ? "DM Mono" : "inherit",
                      color: "#888",
                    }}
                  >
                    {value || "—"}
                  </span>
                </div>
              ))}
            </div>

            <button className="btn-ghost" onClick={onLogout}>
              Log out
            </button>
          </div>
        )}
      </main>
    </div>
  );
};

export default AdminPage;
