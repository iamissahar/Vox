export default function WaveVisualizer({ bars = 8, freqs = [] }) {
  return (
    <div
      style={{ display: "flex", gap: 6, alignItems: "flex-end", height: 40 }}
    >
      {Array.from({ length: bars }).map((_, i) => (
        <div
          key={i}
          className="wave-bar"
          style={{
            height: `${6 + (freqs[i] || 0) * 34}px`,
            animationName: "none", // отключаем CSS анимацию
            transition: "height 0.08s ease",
            animationDelay: `${i * 0.1}s`,
          }}
        />
      ))}
    </div>
  );
}
