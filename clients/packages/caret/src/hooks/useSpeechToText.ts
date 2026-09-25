import { useCallback, useEffect, useRef, useState } from 'react';

// Minimal Web Speech API shape. We don't import lib.dom types here because
// `webkitSpeechRecognition` isn't in standard TS lib definitions.
interface SpeechRecognitionAlternative {
  transcript: string;
}
interface SpeechRecognitionResult {
  isFinal: boolean;
  0: SpeechRecognitionAlternative;
}
interface SpeechRecognitionEventLike {
  results: ArrayLike<SpeechRecognitionResult>;
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
  interimTranscript: string;
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
  const [interimTranscript, setInterimTranscript] = useState('');
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
      let finalOut = '';
      let interimOut = '';
      for (let i = 0; i < e.results.length; i++) {
        const result = e.results[i];
        if (!result) continue;
        const text = result[0]?.transcript ?? '';
        if (result.isFinal) finalOut += text;
        else interimOut += text;
      }
      // The browser keeps prior results in the list and revisits interim
      // results as they change. Rebuilding from final results prevents the
      // same words from being appended once per interim event.
      setTranscript(finalOut);
      setInterimTranscript(interimOut);
    };
    rec.onend = () => {
      setListening(false);
      setInterimTranscript('');
    };
    rec.onerror = (ev) => {
      setError(ev.error);
      setListening(false);
      setInterimTranscript('');
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
    setTranscript('');
    setInterimTranscript('');
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

  const reset = useCallback((): void => {
    setTranscript('');
    setInterimTranscript('');
  }, []);

  return {
    supported,
    listening,
    transcript,
    interimTranscript,
    error,
    start,
    stop,
    reset,
  };
}
