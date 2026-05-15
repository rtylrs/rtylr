import React, { useEffect } from 'react';
import { Mic, MicOff } from 'lucide-react';
import IconButton from '../../ui/IconButton';
import Tooltip from '../../ui/Tooltip';
import { useSpeechToText } from '../../hooks/useSpeechToText';

export interface MicButtonProps {
  onTranscript: (text: string) => void;
}

export const MicButton: React.FC<MicButtonProps> = ({ onTranscript }) => {
  const { supported, listening, transcript, start, stop, reset } = useSpeechToText();

  // Stream interim transcripts into the composer as they arrive.
  useEffect(() => {
    if (!listening || !transcript) return;
    onTranscript(transcript);
  }, [transcript, listening, onTranscript]);

  // When the user stops, clear the buffer so the next session starts fresh.
  useEffect(() => {
    if (!listening) reset();
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
