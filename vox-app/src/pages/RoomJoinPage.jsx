import { useState } from "react";
import Logo from "./../components/Logo";

export default function RoomJoinPage({ navigate }) {
  const [id, setId] = useState("");
  const join = () => {
    if (id.trim()) navigate(`#/room/${id.trim().toUpperCase()}`);
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
        <div style={{ marginBottom: 40 }}>
          <Logo size={24} />
        </div>
        <div
          className="card"
          style={{ display: "flex", flexDirection: "column", gap: 16 }}
        >
          <div>
            <h2
              style={{
                fontSize: 22,
                fontWeight: 400,
                letterSpacing: "-0.03em",
                marginBottom: 6,
              }}
            >
              Join a room
            </h2>
            <p style={{ color: "#444", fontSize: 13 }}>
              Enter the room ID shared by the host
            </p>
          </div>
          <input
            className="input-field"
            placeholder="VOX-XXXXXX"
            value={id}
            onChange={(e) => setId(e.target.value)}
            onKeyDown={(e) => e.key === "Enter" && join()}
            style={{ textTransform: "uppercase", letterSpacing: "0.08em" }}
          />
          <button className="btn-primary" onClick={join} disabled={!id.trim()}>
            Enter →
          </button>
        </div>
        <button
          className="btn-ghost"
          onClick={() => navigate("#/")}
          style={{ width: "100%", marginTop: 12, fontSize: 13 }}
        >
          ← Back
        </button>
      </div>
    </div>
  );
}
