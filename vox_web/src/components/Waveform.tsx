import React, { useRef, useEffect } from 'react';

interface WaveformProps {
  analyser: AnalyserNode | null;
  isActive: boolean;
  /** 'recorder' = purple bars for host, 'listener' = cyan bars for listener */
  variant?: 'recorder' | 'listener';
  height?: number;
}

const COLORS = {
  recorder: {
    bar: 'rgba(108, 99, 255, 0.9)',
    glow: 'rgba(108, 99, 255, 0.3)',
    idle: 'rgba(108, 99, 255, 0.15)',
  },
  listener: {
    bar: 'rgba(56, 189, 248, 0.9)',
    glow: 'rgba(56, 189, 248, 0.3)',
    idle: 'rgba(56, 189, 248, 0.15)',
  },
};

export function Waveform({
  analyser,
  isActive,
  variant = 'recorder',
  height = 80,
}: WaveformProps) {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const animFrameRef = useRef<number>(0);
  const colors = COLORS[variant];

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    const draw = () => {
      const W = canvas.width;
      const H = canvas.height;
      ctx.clearRect(0, 0, W, H);

      if (!analyser || !isActive) {
        // Draw flat idle line with gentle pulse
        const centerY = H / 2;
        const time = Date.now() / 1000;
        const numBars = 40;
        const barW = 3;
        const gap = (W - numBars * barW) / (numBars + 1);

        for (let i = 0; i < numBars; i++) {
          const x = gap + i * (barW + gap);
          const pulse = Math.sin(time * 1.5 + i * 0.3) * 2 + 3;
          ctx.fillStyle = colors.idle;
          ctx.beginPath();
          ctx.roundRect(x, centerY - pulse / 2, barW, pulse, 2);
          ctx.fill();
        }

        animFrameRef.current = requestAnimationFrame(draw);
        return;
      }

      const bufferLen = analyser.frequencyBinCount;
      const dataArray = new Uint8Array(bufferLen);
      analyser.getByteFrequencyData(dataArray);

      const numBars = 48;
      const barW = Math.floor(W / (numBars * 1.6));
      const gap = (W - numBars * barW) / (numBars + 1);

      for (let i = 0; i < numBars; i++) {
        // Map bar index to frequency data with emphasis on lower freqs
        const dataIdx = Math.floor((i / numBars) * bufferLen * 0.6);
        const value = dataArray[dataIdx] / 255;

        const barH = Math.max(3, value * H * 0.85);
        const x = gap + i * (barW + gap);
        const y = (H - barH) / 2;

        // Gradient per bar
        const grad = ctx.createLinearGradient(0, y, 0, y + barH);
        grad.addColorStop(0, colors.glow);
        grad.addColorStop(0.5, colors.bar);
        grad.addColorStop(1, colors.glow);

        ctx.fillStyle = grad;
        ctx.beginPath();
        ctx.roundRect(x, y, barW, barH, barW / 2);
        ctx.fill();
      }

      animFrameRef.current = requestAnimationFrame(draw);
    };

    animFrameRef.current = requestAnimationFrame(draw);
    return () => cancelAnimationFrame(animFrameRef.current);
  }, [analyser, isActive, colors]);

  // Resize observer for responsive canvas
  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const observer = new ResizeObserver(() => {
      canvas.width = canvas.offsetWidth * window.devicePixelRatio;
      canvas.height = canvas.offsetHeight * window.devicePixelRatio;
      const ctx = canvas.getContext('2d');
      if (ctx) {
        ctx.scale(window.devicePixelRatio, window.devicePixelRatio);
      }
    });
    observer.observe(canvas);
    return () => observer.disconnect();
  }, []);

  return (
    <canvas
      ref={canvasRef}
      style={{
        width: '100%',
        height: `${height}px`,
        display: 'block',
      }}
    />
  );
}
