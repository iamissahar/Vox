import React from "react";
import Logo from "../components/Logo";

interface HomePageProps {
  navigate: (to: string) => void;
  isAuthenticated: boolean;
}

const HomePage: React.FC<HomePageProps> = ({ navigate, isAuthenticated }) => {
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
        gap: 48,
      }}
    >
      <Logo size={40} />

      <div style={{ textAlign: "center", maxWidth: 480 }}>
        <h1
          style={{
            fontSize: "clamp(36px, 6vw, 64px)",
            fontWeight: 300,
            letterSpacing: "-0.04em",
            lineHeight: 1.05,
            marginBottom: 16,
          }}
        >
          Real-time voice,
          <br />
          <em style={{ fontStyle: "italic", color: "#e8ff5e" }}>
            translated live.
          </em>
        </h1>
        <p style={{ color: "#555", fontSize: 15, lineHeight: 1.7 }}>
          Stream audio to your audience. They choose their language. No delay,
          no friction.
        </p>
      </div>

      <div
        style={{
          display: "flex",
          gap: 12,
          flexWrap: "wrap",
          justifyContent: "center",
        }}
      >
        <button
          className="btn-primary"
          onClick={() =>
            navigate(isAuthenticated ? "#/admin" : "#/login")
          }
        >
          I'm a Host
        </button>
        <button className="btn-ghost" onClick={() => navigate("#/room")}>
          I'm a Guest
        </button>
      </div>

      <div
        style={{
          position: "fixed",
          bottom: 24,
          color: "#2a2a2a",
          fontSize: 12,
          letterSpacing: "0.06em",
          textTransform: "uppercase",
        }}
      >
        vox · live translation
      </div>
    </div>
  );
};

export default HomePage;
