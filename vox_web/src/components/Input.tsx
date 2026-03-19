import React, { type InputHTMLAttributes, type ReactNode } from 'react';

interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  label?: string;
  error?: string;
  icon?: ReactNode;
}

export function Input({ label, error, icon, style, ...rest }: InputProps) {
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '6px' }}>
      {label && (
        <label
          style={{
            fontSize: '13px',
            color: 'var(--text-secondary)',
            fontWeight: 500,
          }}
        >
          {label}
        </label>
      )}
      <div style={{ position: 'relative' }}>
        {icon && (
          <span
            style={{
              position: 'absolute',
              left: '12px',
              top: '50%',
              transform: 'translateY(-50%)',
              color: 'var(--text-muted)',
              display: 'flex',
              alignItems: 'center',
            }}
          >
            {icon}
          </span>
        )}
        <input
          style={{
            width: '100%',
            background: 'var(--bg-elevated)',
            border: `1px solid ${error ? 'var(--error)' : 'var(--border)'}`,
            borderRadius: '10px',
            padding: icon ? '10px 14px 10px 38px' : '10px 14px',
            fontSize: '14px',
            color: 'var(--text-primary)',
            outline: 'none',
            transition: 'border-color 0.2s ease',
            ...style,
          }}
          onFocus={(e) => {
            (e.target as HTMLInputElement).style.borderColor =
              'var(--border-accent)';
          }}
          onBlur={(e) => {
            (e.target as HTMLInputElement).style.borderColor = error
              ? 'var(--error)'
              : 'var(--border)';
          }}
          {...rest}
        />
      </div>
      {error && (
        <span style={{ fontSize: '12px', color: 'var(--error)' }}>{error}</span>
      )}
    </div>
  );
}
