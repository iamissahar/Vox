import React from 'react';

interface LogoProps {
  size?: number;
  showText?: boolean;
}

export function VoxLogo({ size = 36, showText = true }: LogoProps) {
  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
      {/* SVG mark: stylized soundwave inside a rounded square */}
      <svg
        width={size}
        height={size}
        viewBox="0 0 36 36"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
      >
        <rect width="36" height="36" rx="10" fill="var(--accent)" />
        {/* Soundwave bars – 5 bars of varying height */}
        <rect x="6"  y="14" width="3" height="8"  rx="1.5" fill="white" opacity="0.6" />
        <rect x="11" y="10" width="3" height="16" rx="1.5" fill="white" opacity="0.8" />
        <rect x="16" y="7"  width="4" height="22" rx="2"   fill="white" />
        <rect x="22" y="10" width="3" height="16" rx="1.5" fill="white" opacity="0.8" />
        <rect x="27" y="14" width="3" height="8"  rx="1.5" fill="white" opacity="0.6" />
      </svg>

      {showText && (
        <span
          style={{
            fontFamily: "'Syne', sans-serif",
            fontWeight: 800,
            fontSize: size * 0.67,
            color: 'var(--text-primary)',
            letterSpacing: '-0.02em',
          }}
        >
          vox
        </span>
      )}
    </div>
  );
}
