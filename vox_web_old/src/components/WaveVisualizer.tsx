import React from "react";

interface WaveVisualizerProps {
  active: boolean;
  size?: number;
}

const WaveVisualizer: React.FC<WaveVisualizerProps> = ({
  active,
  size = 100,
}) => {
  return (
    <div
      style={{
        width: size,
        height: size,
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
  );
};

export default WaveVisualizer;
