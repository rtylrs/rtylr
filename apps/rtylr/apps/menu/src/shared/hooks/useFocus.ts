import { useState, useCallback, useRef } from 'react';

interface UseFocusReturn<T extends HTMLElement> {
  ref: React.RefObject<T | null>;
  isFocused: boolean;
  focus: () => void;
  blur: () => void;
  onFocus: () => void;
  onBlur: () => void;
}

/**
 * Track focus state for an element and provide programmatic focus/blur.
 * Attach `ref`, `onFocus`, and `onBlur` to the target element.
 */
export function useFocus<T extends HTMLElement = HTMLElement>(): UseFocusReturn<T> {
  const ref = useRef<T>(null);
  const [isFocused, setIsFocused] = useState(false);

  const onFocus = useCallback(() => setIsFocused(true), []);
  const onBlur = useCallback(() => setIsFocused(false), []);

  const focus = useCallback(() => ref.current?.focus(), []);
  const blur = useCallback(() => ref.current?.blur(), []);

  return { ref, isFocused, focus, blur, onFocus, onBlur };
}
