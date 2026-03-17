import { useState, useRef } from "react";
import Logo from "./../components/Logo";
import { authState } from "./../App";
import GoogleIcon from "./../components/Google";
import GithubIcon from "./../components/GitHub";

const MOCK_ADMIN = {
  email: "admin@vox.io",
  password: "vox2024",
  name: "Alex Morgan",
};

export default function LoginPage({ navigate, onLogin }) {
  const [email, setEmail] = useState("");
  const [pass, setPass] = useState("");
  const [err, setErr] = useState("");
  const [loading, setLoading] = useState(false);

  const submit = async () => {
    setLoading(true);
    setErr("");
    await new Promise((r) => setTimeout(r, 600));
    if (email === MOCK_ADMIN.email && pass === MOCK_ADMIN.password) {
      authState.user = MOCK_ADMIN;
      onLogin(MOCK_ADMIN);
      navigate("#/admin");
    } else {
      setErr("Invalid credentials.");
    }
    setLoading(false);
  };

  const handleGoogle = () => {
    // TODO: подключить реальный Google OAuth
    // import { GoogleAuthProvider, signInWithPopup } from "firebase/auth"
    // или использовать @react-oauth/google
    alert("Google OAuth — подключи SDK");
  };

  const handleGithub = () => {
    // TODO: подключить GitHub OAuth
    alert("GitHub OAuth — подключи SDK");
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
          <span style={{ color: "#555", fontSize: 13 }}>admin login</span>
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
              marginBottom: 8,
            }}
          >
            Sign in
          </h2>

          {/* OAuth кнопки */}
          <button
            onClick={handleGoogle}
            style={{
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
              gap: 10,
              background: "#111",
              border: "1px solid #1e1e1e",
              borderRadius: 12,
              padding: "12px 20px",
              color: "#f0ede8",
              fontSize: 14,
              cursor: "pointer",
              transition: "border-color 0.2s",
              fontFamily: "inherit",
              width: "100%",
            }}
            onMouseEnter={(e) => (e.currentTarget.style.borderColor = "#333")}
            onMouseLeave={(e) =>
              (e.currentTarget.style.borderColor = "#1e1e1e")
            }
          >
            <GoogleIcon /> Continue with Google
          </button>

          <button
            onClick={handleGithub}
            style={{
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
              gap: 10,
              background: "#111",
              border: "1px solid #1e1e1e",
              borderRadius: 12,
              padding: "12px 20px",
              color: "#f0ede8",
              fontSize: 14,
              cursor: "pointer",
              transition: "border-color 0.2s",
              fontFamily: "inherit",
              width: "100%",
            }}
            onMouseEnter={(e) => (e.currentTarget.style.borderColor = "#333")}
            onMouseLeave={(e) =>
              (e.currentTarget.style.borderColor = "#1e1e1e")
            }
          >
            <GithubIcon /> Continue with GitHub
          </button>

          {/* Разделитель */}
          <div style={{ display: "flex", alignItems: "center", gap: 12 }}>
            <div style={{ flex: 1, height: 1, background: "#1a1a1a" }} />
            <span style={{ color: "#333", fontSize: 12 }}>or</span>
            <div style={{ flex: 1, height: 1, background: "#1a1a1a" }} />
          </div>

          {/* Email/password */}
          <input
            className="input-field"
            placeholder="Email"
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            onKeyDown={(e) => e.key === "Enter" && submit()}
          />
          <input
            className="input-field"
            placeholder="Password"
            type="password"
            value={pass}
            onChange={(e) => setPass(e.target.value)}
            onKeyDown={(e) => e.key === "Enter" && submit()}
          />
          {err && <p style={{ color: "#ff5e5e", fontSize: 13 }}>{err}</p>}
          <button
            className="btn-primary"
            onClick={submit}
            disabled={loading}
            style={{ marginTop: 8, opacity: loading ? 0.7 : 1 }}
          >
            {loading ? "Signing in…" : "Continue →"}
          </button>
        </div>

        <p
          style={{
            color: "#333",
            fontSize: 12,
            textAlign: "center",
            marginTop: 20,
          }}
        >
          Hint: admin@vox.io / vox2024
        </p>
      </div>
    </div>
  );
}
