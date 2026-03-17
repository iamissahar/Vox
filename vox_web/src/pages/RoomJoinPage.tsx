import React, { useState } from "react";
import Logo from "../components/Logo";

interface RoomJoinPageProps {
  navigate: (to: string) => void;
}

const RoomJoinPage: React.FC<RoomJoinPageProps> = ({ navigate }) => {
  const [roomId, setRoomId] = useState("");

  const join = () => {
    const id = roomId.trim();
    if (id) navigate(`#/room/${id}`);
  };

  return (
    <div
      className="fade-in"
      style={{
        minHeight: "100vh",
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        padding: 24,
      }}
    >
      <div style={{ width: "100%", maxWidth: 380 }}>
        <div
          style={{
            marginBottom: 40,
            display: "flex",
            alignItems: "center",
            gap: 12,
          }}
        >
          <Logo size={24} />
          <span style={{ color: "#333", fontSize: 13 }}>/</span>
          <span style={{ color: "#555", fontSize: 13 }}>join room</span>
        </div>

        <div
          className="card"
          style={{ display: "flex", flexDirection: "column", gap: 16 }}
        >
          <h2
            style={{
              fontSize: 22,
              fontWeight: 400,
              letterSpacing: "-0.03em",
              marginBottom: 4,
            }}
          >
            Join a room
          </h2>
          <p style={{ color: "#555", fontSize: 13, lineHeight: 1.6 }}>
            Enter the hub ID shared by the host.
          </p>

          <input
            className="input-field"
            placeholder="Hub ID"
            value={roomId}
            onChange={(e) => setRoomId(e.target.value)}
            onKeyDown={(e) => e.key === "Enter" && join()}
            autoFocus
          />

          <button
            className="btn-primary"
            onClick={join}
            disabled={!roomId.trim()}
            style={{ marginTop: 4, opacity: roomId.trim() ? 1 : 0.5 }}
          >
            Join →
          </button>

          <button
            className="btn-ghost"
            onClick={() => navigate("#/")}
            style={{ fontSize: 13 }}
          >
            ← Back
          </button>
        </div>
      </div>
    </div>
  );
};

export default RoomJoinPage;
