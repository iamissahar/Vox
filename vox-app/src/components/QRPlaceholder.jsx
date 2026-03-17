export default function QRPlaceholder({ value, size = 160 }) {
  // Simple visual placeholder — in prod use qrcode.react or similar
  const cells = 21;
  const cell = size / cells;
  // Deterministic "random" pattern based on value
  const hash = (str) => {
    let h = 0;
    for (let i = 0; i < str.length; i++)
      h = (Math.imul(31, h) + str.charCodeAt(i)) | 0;
    return h;
  };
  const grid = Array.from({ length: cells }, (_, r) =>
    Array.from({ length: cells }, (_, c) => {
      if (r < 7 && c < 7)
        return (
          r === 0 ||
          r === 6 ||
          c === 0 ||
          c === 6 ||
          (r >= 2 && r <= 4 && c >= 2 && c <= 4)
        );
      if (r < 7 && c >= cells - 7)
        return (
          r === 0 ||
          r === 6 ||
          c === cells - 7 ||
          c === cells - 1 ||
          (r >= 2 && r <= 4 && c >= cells - 5 && c <= cells - 3)
        );
      if (r >= cells - 7 && c < 7)
        return (
          r === cells - 7 ||
          r === cells - 1 ||
          c === 0 ||
          c === 6 ||
          (r >= cells - 5 && r <= cells - 3 && c >= 2 && c <= 4)
        );
      return Math.abs(hash(value + r * 23 + c * 17)) % 2 === 0;
    }),
  );
  return (
    <svg
      width={size}
      height={size}
      viewBox={`0 0 ${size} ${size}`}
      style={{ borderRadius: 8 }}
    >
      <rect width={size} height={size} fill="white" />
      {grid.map((row, r) =>
        row.map(
          (on, c) =>
            on && (
              <rect
                key={`${r}-${c}`}
                x={c * cell}
                y={r * cell}
                width={cell}
                height={cell}
                fill="#0a0a0a"
              />
            ),
        ),
      )}
    </svg>
  );
}
