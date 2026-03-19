import React, { useState } from "react";
import Logo from "../components/Logo";
import GoogleIcon from "../components/Google";
import GithubIcon from "../components/GitHub";
import { auth, user as userApi, ApiError } from "../api/client";
import type { UserInfo } from "../types";

interface LoginPageProps {
  navigate: (to: string) => void;
  onLogin: (user: UserInfo) => void;
}

type Mode = "login" | "signup";

const LoginPage: React.FC<LoginPageProps> = ({ navigate, onLogin }) => {
  const [mode, setMode] = useState<Mode>("login");
  const [login, setLogin] = useState("");
  const [email, setEmail] = useState("");
  const [name, setName] = useState("");
  const [pass, setPass] = useState("");
  const [err, setErr] = useState("");
  const [loading, setLoading] = useState(false);

  const handleSubmit = async () => {
    setLoading(true);
    setErr("");

    try {
      if (mode === "login") {
        await auth.login({ login, password: pass });
      } else {
        await auth.signUp({ email, login, name, password: pass });
      }
      // After login/signup, fetch user info
      const info = await userApi.getInfo();
      onLogin(info);
      navigate("#/admin");
    } catch (e) {
      if (e instanceof ApiError) {
        setErr(e.message);
      } else {
        setErr("Something went wrong. Please try again.");
      }
    } finally {
      setLoading(false);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter") handleSubmit();
  };

  const oauthButtonStyle: React.CSSProperties = {
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
        {/* Header */}
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
          <span style={{ color: "#555", fontSize: 13 }}>
            {mode === "login" ? "sign in" : "sign up"}
          </span>
        </div>

        <div
          className="card"
          style={{ display: "flex", flexDirection: "column", gap: 16 }}
        >
          <div
            style={{
              display: "flex",
              alignItems: "center",
              justifyContent: "space-between",
              marginBottom: 8,
            }}
          >
            <h2
              style={{
                fontSize: 22,
                fontWeight: 400,
                letterSpacing: "-0.03em",
              }}
            >
              {mode === "login" ? "Sign in" : "Create account"}
            </h2>
            <button
              onClick={() => {
                setMode(mode === "login" ? "signup" : "login");
                setErr("");
              }}
              style={{
                background: "none",
                border: "none",
                color: "#555",
                fontSize: 13,
                cursor: "pointer",
                fontFamily: "inherit",
                textDecoration: "underline",
                textUnderlineOffset: 3,
              }}
            >
              {mode === "login" ? "Sign up instead" : "Sign in instead"}
            </button>
          </div>

          {/* OAuth buttons */}
          <button
            style={oauthButtonStyle}
            onClick={() => auth.oauthLogin("google")}
            onMouseEnter={(e) =>
              (e.currentTarget.style.borderColor = "#333")
            }
            onMouseLeave={(e) =>
              (e.currentTarget.style.borderColor = "#1e1e1e")
            }
          >
            <GoogleIcon /> Continue with Google
          </button>

          <button
            style={oauthButtonStyle}
            onClick={() => auth.oauthLogin("github")}
            onMouseEnter={(e) =>
              (e.currentTarget.style.borderColor = "#333")
            }
            onMouseLeave={(e) =>
              (e.currentTarget.style.borderColor = "#1e1e1e")
            }
          >
            <GithubIcon /> Continue with GitHub
          </button>

          {/* Divider */}
          <div style={{ display: "flex", alignItems: "center", gap: 12 }}>
            <div style={{ flex: 1, height: 1, background: "#1a1a1a" }} />
            <span style={{ color: "#333", fontSize: 12 }}>or</span>
            <div style={{ flex: 1, height: 1, background: "#1a1a1a" }} />
          </div>

          {/* Fields */}
          {mode === "signup" && (
            <>
              <input
                className="input-field"
                placeholder="Name"
                type="text"
                value={name}
                onChange={(e) => setName(e.target.value)}
                onKeyDown={handleKeyDown}
              />
              <input
                className="input-field"
                placeholder="Email"
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                onKeyDown={handleKeyDown}
              />
            </>
          )}

          <input
            className="input-field"
            placeholder="Login"
            type="text"
            value={login}
            onChange={(e) => setLogin(e.target.value)}
            onKeyDown={handleKeyDown}
          />
          <input
            className="input-field"
            placeholder="Password"
            type="password"
            value={pass}
            onChange={(e) => setPass(e.target.value)}
            onKeyDown={handleKeyDown}
          />

          {err && (
            <p style={{ color: "#ff5e5e", fontSize: 13, marginTop: -4 }}>
              {err}
            </p>
          )}

          <button
            className="btn-primary"
            onClick={handleSubmit}
            disabled={loading}
            style={{ marginTop: 8, opacity: loading ? 0.7 : 1 }}
          >
            {loading
              ? mode === "login"
                ? "Signing in…"
                : "Creating account…"
              : "Continue →"}
          </button>
        </div>
      </div>
    </div>
  );
};

export default LoginPage;
