import { useRef, useState, useCallback } from 'react';

interface UseAudioRecorderReturn {
  isRecording: boolean;
  analyserNode: AnalyserNode | null;
  startRecording: (onChunk: (chunk: Blob) => void) => Promise<void>;
  stopRecording: () => void;
  error: string | null;
}

const CHUNK_INTERVAL_MS = 250; // send audio chunk every 250ms

export function useAudioRecorder(): UseAudioRecorderReturn {
  const [isRecording, setIsRecording] = useState(false);
  const [analyserNode, setAnalyserNode] = useState<AnalyserNode | null>(null);
  const [error, setError] = useState<string | null>(null);

  const mediaRecorderRef = useRef<MediaRecorder | null>(null);
  const streamRef = useRef<MediaStream | null>(null);
  const audioCtxRef = useRef<AudioContext | null>(null);

  const startRecording = useCallback(async (onChunk: (chunk: Blob) => void) => {
    setError(null);
    try {
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
      streamRef.current = stream;

      // Set up Web Audio API for visualization
      const audioCtx = new AudioContext();
      audioCtxRef.current = audioCtx;
      const source = audioCtx.createMediaStreamSource(stream);
      const analyser = audioCtx.createAnalyser();
      analyser.fftSize = 256;
      source.connect(analyser);
      setAnalyserNode(analyser);

      // Set up MediaRecorder for streaming
      const mediaRecorder = new MediaRecorder(stream, {
        mimeType: 'audio/webm;codecs=opus',
      });
      mediaRecorderRef.current = mediaRecorder;

      mediaRecorder.ondataavailable = (e) => {
        if (e.data && e.data.size > 0) {
          onChunk(e.data);
        }
      };

      mediaRecorder.start(CHUNK_INTERVAL_MS);
      setIsRecording(true);
    } catch (err) {
      const message =
        err instanceof Error ? err.message : 'Microphone access denied';
      setError(message);
    }
  }, []);

  const stopRecording = useCallback(() => {
    mediaRecorderRef.current?.stop();
    streamRef.current?.getTracks().forEach((t) => t.stop());
    audioCtxRef.current?.close();
    setAnalyserNode(null);
    setIsRecording(false);
  }, []);

  return { isRecording, analyserNode, startRecording, stopRecording, error };
}
