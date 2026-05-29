import { useCallback, useEffect, useRef, useState } from 'react';

// Minimal Web Speech API shape. We don't import lib.dom types here because
// `webkitSpeechRecognition` isn't in standard TS lib definitions.
interface SpeechRecognitionResult {
  transcript: string;
  isFinal: boolean;
}
interface SpeechRecognitionEventLike {
  resultIndex: number;
  results: ArrayLike<ArrayLike<SpeechRecognitionResult>>;
}
interface SpeechRecognitionLike {
  continuous: boolean;
  interimResults: boolean;
  lang: string;
  onresult: ((e: SpeechRecognitionEventLike) => void) | null;
  onend: (() => void) | null;
  onerror: ((e: { error: string }) => void) | null;
  start: () => void;
  stop: () => void;
}

type SpeechCtor = new () => SpeechRecognitionLike;

function getRecognitionCtor(): SpeechCtor | null {
  if (typeof window === 'undefined') return null;
  const w = window as unknown as {
    SpeechRecognition?: SpeechCtor;
    webkitSpeechRecognition?: SpeechCtor;
  };
  return w.SpeechRecognition ?? w.webkitSpeechRecognition ?? null;
}

export interface UseSpeechToTextResult {
  supported: boolean;
  listening: boolean;
  transcript: string;
  error: string | null;
  start: () => void;
  stop: () => void;
  reset: () => void;
}

/** Browser speech recognition for the composer mic button. */
export function useSpeechToText(lang = 'en-US'): UseSpeechToTextResult {
  const recRef = useRef<SpeechRecognitionLike | null>(null);
  const [supported, setSupported] = useState(false);
  const [listening, setListening] = useState(false);
  const [transcript, setTranscript] = useState('');
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const Ctor = getRecognitionCtor();
    if (!Ctor) {
      setSupported(false);
      return;
    }
    const rec = new Ctor();
    rec.continuous = true;
    rec.interimResults = true;
    rec.lang = lang;
    rec.onresult = (e) => {
      let out = '';
      for (let i = e.resultIndex; i < e.results.length; i++) {
        out += e.results[i][0].transcript;
      }
      setTranscript((prev) => prev + out);
    };
    rec.onend = () => setListening(false);
    rec.onerror = (ev) => {
      setError(ev.error);
      setListening(false);
    };
    recRef.current = rec;
    setSupported(true);
    return (): void => {
      try {
        rec.stop();
      } catch {
        // ignore - already stopped
      }
    };
  }, [lang]);

  const start = useCallback((): void => {
    setError(null);
    const rec = recRef.current;
    if (!rec) return;
    try {
      rec.start();
      setListening(true);
    } catch (e) {
      setError((e as Error).message);
    }
  }, []);

  const stop = useCallback((): void => {
    const rec = recRef.current;
    if (!rec) return;
    try {
      rec.stop();
    } catch {
      // ignore
    }
  }, []);

  const reset = useCallback((): void => setTranscript(''), []);

  return { supported, listening, transcript, error, start, stop, reset };
}
