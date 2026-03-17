export default function Logo({ size = 28, showLabel = true }) {
  return (
    <div style={{ display: "flex", alignItems: "center", gap: 10 }}>
      <svg width={size} height={size} viewBox="0 0 28 28" fill="none">
        <circle cx="14" cy="14" r="13" stroke="#e8ff5e" strokeWidth="1.5" />
        <path
          d="M9 10v8M13 7v14M17 10v8M21 12v4"
          stroke="#e8ff5e"
          strokeWidth="1.8"
          strokeLinecap="round"
        />
      </svg>
      {showLabel && (
        <span
          style={{
            fontFamily: "DM Sans",
            fontWeight: 500,
            fontSize: size * 0.85,
            letterSpacing: "-0.04em",
            color: "#f0ede8",
          }}
        >
          vox
        </span>
      )}
    </div>
  );
}
