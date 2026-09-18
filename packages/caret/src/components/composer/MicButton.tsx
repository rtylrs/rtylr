import React, { useEffect, useRef } from 'react';
import { Mic, MicOff } from 'lucide-react';
import IconButton from '../../ui/IconButton';
import Tooltip from '../../ui/Tooltip';
import { useSpeechToText } from '../../hooks/useSpeechToText';
import { caretTokens } from '../../ui/tokens';

export interface VoiceInputStatus {
  supported: boolean;
  listening: boolean;
  interimTranscript: string;
  error: string | null;
}

export interface MicButtonProps {
  onTranscript: (text: string) => void;
  onStatusChange?: (status: VoiceInputStatus) => void;
}

export const MicButton: React.FC<MicButtonProps> = ({ onTranscript, onStatusChange }) => {
  const { supported, listening, transcript, interimTranscript, error, start, stop, reset } =
    useSpeechToText();
  const deliveredTranscript = useRef('');

  useEffect(() => {
    onStatusChange?.({ supported, listening, interimTranscript, error });
  }, [error, interimTranscript, listening, onStatusChange, supported]);

  // The speech hook returns the full set of final results. Only append the
  // newly recognized suffix so browser result updates never duplicate text.
  useEffect(() => {
    if (!transcript) return;
    const previous = deliveredTranscript.current;
    const next = transcript.startsWith(previous) ? transcript.slice(previous.length) : transcript;
    if (next) onTranscript(next);
    deliveredTranscript.current = transcript;
  }, [transcript, onTranscript]);

  // When the user stops, clear the buffer so the next session starts fresh.
  useEffect(() => {
    if (!listening) {
      deliveredTranscript.current = '';
      reset();
    }
  }, [listening, reset]);

  if (!supported) {
    return (
      <Tooltip label="Speech recognition unavailable">
        <IconButton size="md" label="Speech unavailable" disabled>
          <MicOff size={14} />
        </IconButton>
      </Tooltip>
    );
  }

  return (
    <Tooltip label={listening ? 'Stop dictation' : 'Dictate a prompt'}>
      <IconButton
        className="caret-composer-icon"
        size="md"
        label={listening ? 'Stop dictation' : 'Start dictation'}
        active={listening}
        aria-pressed={listening}
        onClick={listening ? stop : start}
        style={
          listening
            ? {
                background: caretTokens.dangerSoft,
                color: caretTokens.dangerText,
                boxShadow: `0 0 0 3px ${caretTokens.dangerSoft}`,
              }
            : undefined
        }
      >
        <Mic size={14} />
      </IconButton>
    </Tooltip>
  );
};

export default MicButton;
