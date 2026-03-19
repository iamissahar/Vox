import React, { type ReactNode, type CSSProperties } from 'react';

interface CardProps {
  children: ReactNode;
  style?: CSSProperties;
  className?: string;
  onClick?: () => void;
  /** Add a subtle glow border */
  glowing?: boolean;
}

export function Card({ children, style, onClick, glowing }: CardProps) {
  return (
    <div
      onClick={onClick}
      style={{
        background: 'var(--bg-card)',
        border: `1px solid ${glowing ? 'var(--border-accent)' : 'var(--border)'}`,
        borderRadius: 'var(--radius-lg)',
        padding: '24px',
        transition: 'border-color 0.2s ease, box-shadow 0.2s ease',
        boxShadow: glowing ? '0 0 24px var(--accent-glow)' : 'none',
        cursor: onClick ? 'pointer' : 'default',
        ...style,
      }}
    >
      {children}
    </div>
  );
}
