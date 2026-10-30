import { useEffect, useRef, useState } from 'react';
import { useCaretStore } from '@voxire/caret';
import type { CaretMessageType } from '@voxire/caret';
import type { BuddyMood } from './BuddySprite';

const LISTEN_MS = 1500;
const HAPPY_MS = 1500;

type Turn = { streaming: boolean; messages: CaretMessageType[] };

// moodOf reads what Buddy is doing from the open chat.
export function moodOf({ streaming, messages }: Turn): BuddyMood {
  const last = messages[messages.length - 1];
  const blocks = last?.role === 'assistant' ? last.blocks : [];
  const awaiting = blocks.some(
    (block) => block.type === 'tool_call' && block.status === 'pending' && block.requiresApproval
  );
  if (awaiting) return 'waiting';
  if (!streaming) return 'idle';
  if (blocks.some((block) => block.type === 'tool_call' && block.status === 'running'))
    return 'working';
  const said = blocks.some((block) => block.type === 'text' && block.content.trim() !== '');
  return said || last?.content?.trim() ? 'talking' : 'thinking';
}

/**
 * Buddy's live mood for the open chat: it listens while the person types,
 * thinks and talks through a reply, asks when it needs an approval, and
 * hops when a reply or a routine post lands.
 */
export function useBuddyMood(active: boolean): BuddyMood {
  const turn = useCaretStore((state) => {
    const tab = state.tabs.find((candidate) => candidate.id === state.activeTabId);
    return tab;
  });
  const streaming = Boolean(turn?.isStreaming);
  const count = turn?.messages.length ?? 0;
  const [typing, setTyping] = useState(false);
  const [happy, setHappy] = useState(false);
  const previous = useRef({ streaming, count, session: turn?.sessionId });

  useEffect(() => {
    if (!active) return;
    let timer: ReturnType<typeof setTimeout>;
    const onInput = (event: Event) => {
      const target = event.target as HTMLElement | null;
      if (!target?.closest?.('.caret-composer-wrap')) return;
      setTyping(true);
      clearTimeout(timer);
      timer = setTimeout(() => setTyping(false), LISTEN_MS);
    };
    document.addEventListener('input', onInput, true);
    return () => {
      document.removeEventListener('input', onInput, true);
      clearTimeout(timer);
    };
  }, [active]);

  useEffect(() => {
    const before = previous.current;
    previous.current = { streaming, count, session: turn?.sessionId };
    if (!active || before.session !== turn?.sessionId) return;
    // A reply just finished, or a routine posted while the chat was open.
    const finished = before.streaming && !streaming;
    const posted = !streaming && !before.streaming && count > before.count && before.count > 0;
    if (!finished && !posted) return;
    setHappy(true);
    const timer = setTimeout(() => setHappy(false), HAPPY_MS);
    return () => clearTimeout(timer);
  }, [active, count, streaming, turn?.sessionId]);

  if (!active) return 'idle';
  const mood = moodOf({ streaming, messages: turn?.messages ?? [] });
  if (mood !== 'idle') return mood;
  if (happy) return 'happy';
  if (typing) return 'listening';
  return 'idle';
}
