import React, { useState, useRef, useEffect, useCallback } from "react";
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

// ─── Individual hub row ───────────────────────────────────────────────────────

interface HubRowProps {
  hubId: string;
  userId: string;
  navigate: (to: string) => void;
  onDeleted: (hubId: string) => void;
}

const HubRow: React.FC<HubRowProps> = ({
  hubId,
  userId,
  navigate,
  onDeleted,
}) => {
  const [lang, setLang] = useState("en");
  const [broadcasting, setBroadcasting] = useState(false);
  const [active, setActive] = useState(false);
  const [deleting, setDeleting] = useState(false);
  const [confirmDelete, setConfirmDelete] = useState(false);
  const [err, setErr] = useState("");
  const [copied, setCopied] = useState(false);

  const mediaRecorderRef = useRef<MediaRecorder | null>(null);
  const chunksRef = useRef<Blob[]>([]);
  const intervalRef = useRef<number | null>(null);

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
    setBroadcasting(false);
    setActive(false);
  };

  const handleDelete = async () => {
    if (!confirmDelete) {
      setConfirmDelete(true);
      setTimeout(() => setConfirmDelete(false), 3000);
      return;
    }
    setDeleting(true);
    try {
      await hubApi.delete(hubId, userId);
      onDeleted(hubId);
    } catch {
      setErr("Failed to delete.");
      setDeleting(false);
      setConfirmDelete(false);
    }
  };

  const handleReconnect = () => {
    hubApi.reconnect(hubId);
  };

  const handleCopy = () => {
    navigator.clipboard.writeText(`${window.location.origin}/#/room/${hubId}`);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  useEffect(() => () => stopBroadcast(), []);

  return (
    <div
      style={{
        background: "#0c0c0c",
        border: `1px solid ${broadcasting ? "#e8ff5e22" : "#191919"}`,
        borderRadius: 12,
        padding: "12px 14px",
        display: "flex",
        flexDirection: "column",
        gap: 10,
        transition: "border-color 0.3s",
      }}
    >
      {/* Top row: indicator + id + actions */}
      <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
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
            color: "#ccc",
            flex: 1,
            overflow: "hidden",
            textOverflow: "ellipsis",
            whiteSpace: "nowrap",
          }}
        >
          {hubId}
        </span>
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
              padding: "1px 6px",
              flexShrink: 0,
            }}
          >
            LIVE
          </span>
        )}

        {/* Icon buttons */}
        <button
          title="Open listener room"
          onClick={() => navigate(`#/room/${hubId}`)}
          style={iconBtnStyle}
        >
          <IconShare />
        </button>
        <button
          title="Reconnect stream"
          onClick={handleReconnect}
          style={iconBtnStyle}
        >
          <IconReconnect />
        </button>
        <button
          title={confirmDelete ? "Click again to confirm" : "Delete hub"}
          onClick={handleDelete}
          disabled={deleting}
          style={{
            ...iconBtnStyle,
            color: confirmDelete ? "#ff5e5e" : "#555",
            background: confirmDelete ? "#ff5e5e0a" : "none",
            border: `1px solid ${confirmDelete ? "#ff5e5e33" : "transparent"}`,
          }}
        >
          {deleting ? <MiniSpinner /> : <IconTrash />}
        </button>
      </div>

      {/* Error / confirm notice */}
      {confirmDelete && !err && (
        <p style={{ fontSize: 11, color: "#ff5e5e", margin: 0 }}>
          Click delete again to confirm.
        </p>
      )}
      {err && (
        <p style={{ fontSize: 11, color: "#ff5e5e", margin: 0 }}>{err}</p>
      )}

      {/* Bottom row: lang + broadcast + copy */}
      <div
        style={{
          display: "flex",
          alignItems: "center",
          gap: 8,
          flexWrap: "wrap",
        }}
      >
        <select
          value={lang}
          onChange={(e) => setLang(e.target.value)}
          className="lang-select"
          style={{ fontSize: 12, padding: "4px 8px" }}
        >
          {LANGUAGES.map((l) => (
            <option key={l.code} value={l.code}>
              {l.label}
            </option>
          ))}
        </select>

        {!broadcasting ? (
          <button
            className="btn-primary"
            onClick={startBroadcast}
            style={{ fontSize: 12, padding: "5px 14px" }}
          >
            ▶ Broadcast
          </button>
        ) : (
          <button
            className="btn-ghost"
            onClick={stopBroadcast}
            style={{
              fontSize: 12,
              padding: "5px 14px",
              borderColor: "#ff5e5e",
              color: "#ff5e5e",
            }}
          >
            ■ Stop
          </button>
        )}

        <button
          onClick={handleCopy}
          style={{
            marginLeft: "auto",
            background: "none",
            border: "1px solid #1e1e1e",
            borderRadius: 6,
            fontSize: 12,
            fontFamily: "inherit",
            padding: "5px 12px",
            cursor: "pointer",
            color: copied ? "#6fff6f" : "#555",
            transition: "color 0.2s",
          }}
        >
          {copied ? "✓ Copied" : "Copy link"}
        </button>
      </div>

      {/* Waveform when live */}
      {broadcasting && (
        <div
          style={{
            display: "flex",
            alignItems: "center",
            gap: 10,
            paddingTop: 4,
            borderTop: "1px solid #141414",
          }}
        >
          <WaveVisualizer active={active} size={36} />
          <span style={{ fontSize: 11, color: "#444" }}>Streaming…</span>
        </div>
      )}
    </div>
  );
};

// ─── Main page ────────────────────────────────────────────────────────────────

const AdminPage: React.FC<AdminPageProps> = ({
  navigate,
  currentUser,
  onLogout,
}) => {
  const [tab, setTab] = useState<Tab>("broadcast");

  // Broadcast / hubs state
  const [hubs, setHubs] = useState<string[]>([]);
  const [hubsLoading, setHubsLoading] = useState(false);
  const [hubsErr, setHubsErr] = useState("");
  const [creating, setCreating] = useState(false);

  // Voice state
  const [textRef, setTextRef] = useState("");
  const [voiceRecording, setVoiceRecording] = useState(false);
  const [voiceStatus, setVoiceStatus] = useState<
    "idle" | "recording" | "uploading" | "done" | "error"
  >("idle");
  const voiceRecorderRef = useRef<MediaRecorder | null>(null);
  const voiceChunksRef = useRef<Blob[]>([]);

  const userId = currentUser?.id ?? "";

  // Fetch hub list on mount / tab switch
  const fetchHubs = useCallback(async () => {
    if (!userId) return;
    setHubsLoading(true);
    setHubsErr("");
    try {
      const data = await hubApi.listMine();
      setHubs(data.hub_ids ?? []);
    } catch {
      setHubsErr("Failed to load hubs.");
    } finally {
      setHubsLoading(false);
    }
  }, [userId]);

  useEffect(() => {
    if (tab === "broadcast") fetchHubs();
  }, [tab, fetchHubs]);

  const createHub = async () => {
    setCreating(true);
    setHubsErr("");
    try {
      const data = await hubApi.createAuto();
      setHubs((prev) => [data.hub_id, ...prev]);
    } catch (e) {
      if (e instanceof ApiError) {
        setHubsErr(e.message);
      } else {
        setHubsErr("Failed to create hub.");
      }
    } finally {
      setCreating(false);
    }
  };

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
          justifyContent: "flex-start",
          padding: 32,
          gap: 24,
          maxWidth: 500,
          margin: "0 auto",
          width: "100%",
        }}
      >
        {/* ─── BROADCAST TAB ─── */}
        {tab === "broadcast" && (
          <div
            className="slide-up"
            style={{
              width: "100%",
              display: "flex",
              flexDirection: "column",
              gap: 16,
            }}
          >
            {/* Header */}
            <div
              style={{
                display: "flex",
                alignItems: "flex-start",
                justifyContent: "space-between",
                gap: 12,
              }}
            >
              <div>
                <h2
                  style={{
                    fontSize: 20,
                    fontWeight: 400,
                    letterSpacing: "-0.03em",
                    margin: 0,
                  }}
                >
                  Hubs
                </h2>
                <p
                  style={{
                    color: "#555",
                    fontSize: 13,
                    lineHeight: 1.6,
                    margin: "4px 0 0",
                  }}
                >
                  Your broadcast rooms.
                </p>
              </div>
              <button
                className="btn-primary"
                onClick={createHub}
                disabled={creating || !userId}
                style={{
                  opacity: creating ? 0.7 : 1,
                  whiteSpace: "nowrap",
                  flexShrink: 0,
                }}
              >
                {creating ? "Creating…" : "+ New hub"}
              </button>
            </div>

            {hubsErr && (
              <p style={{ color: "#ff5e5e", fontSize: 13, margin: 0 }}>
                {hubsErr}
              </p>
            )}

            {/* List */}
            {hubsLoading ? (
              <div
                style={{
                  display: "flex",
                  alignItems: "center",
                  gap: 10,
                  padding: "16px 0",
                }}
              >
                <div
                  style={{
                    width: 18,
                    height: 18,
                    border: "2px solid #222",
                    borderTopColor: "#e8ff5e",
                    borderRadius: "50%",
                    animation: "spin 0.8s linear infinite",
                  }}
                />
                <span style={{ color: "#444", fontSize: 13 }}>Loading…</span>
              </div>
            ) : hubs.length === 0 ? (
              <div
                style={{
                  border: "1px dashed #1e1e1e",
                  borderRadius: 12,
                  padding: "32px 24px",
                  textAlign: "center",
                }}
              >
                <p style={{ color: "#333", fontSize: 13, margin: 0 }}>
                  No hubs yet.
                </p>
                <p style={{ color: "#2a2a2a", fontSize: 12, marginTop: 6 }}>
                  Create one to start broadcasting.
                </p>
              </div>
            ) : (
              <div style={{ display: "flex", flexDirection: "column", gap: 8 }}>
                {hubs.map((id) => (
                  <HubRow
                    key={id}
                    hubId={id}
                    userId={userId}
                    navigate={navigate}
                    onDeleted={(id) =>
                      setHubs((prev) => prev.filter((h) => h !== id))
                    }
                  />
                ))}
              </div>
            )}
          </div>
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
                <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
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
            <div style={{ display: "flex", flexDirection: "column", gap: 8 }}>
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

// ─── Shared styles / icons ────────────────────────────────────────────────────

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

const IconShare = () => (
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
    <path d="M4 12v8a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2v-8" />
    <polyline points="16 6 12 2 8 6" />
    <line x1="12" y1="2" x2="12" y2="15" />
  </svg>
);

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

const IconTrash = () => (
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
    <polyline points="3 6 5 6 21 6" />
    <path d="M19 6l-1 14a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2L5 6" />
    <path d="M10 11v6M14 11v6M9 6V4h6v2" />
  </svg>
);

const MiniSpinner = () => (
  <div
    style={{
      width: 11,
      height: 11,
      border: "1.5px solid #333",
      borderTopColor: "currentColor",
      borderRadius: "50%",
      animation: "spin 0.8s linear infinite",
    }}
  />
);

export default AdminPage;
