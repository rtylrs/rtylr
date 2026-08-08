import React, { useEffect, useRef } from 'react';
import { Mic, MicOff } from 'lucide-react';
import IconButton from '../../ui/IconButton';
import Tooltip from '../../ui/Tooltip';
import { useSpeechToText } from '../../hooks/useSpeechToText';

export interface MicButtonProps {
  onTranscript: (text: string) => void;
}

export const MicButton: React.FC<MicButtonProps> = ({ onTranscript }) => {
  const { supported, listening, transcript, start, stop, reset } = useSpeechToText();
  const deliveredTranscript = useRef('');

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
    <Tooltip label={listening ? 'Stop dictating' : 'Dictate'}>
      <IconButton
        size="md"
        label={listening ? 'Stop dictation' : 'Start dictation'}
        active={listening}
        onClick={listening ? stop : start}
      >
        <Mic size={14} />
      </IconButton>
    </Tooltip>
  );
};

export default MicButton;
