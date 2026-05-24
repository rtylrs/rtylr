import React, { useState, useRef, useEffect, useCallback } from 'react';
import { useDesktopStore } from '../../stores/useDesktopStore';

interface HistoryEntry {
  type: 'input' | 'output';
  text: string;
}

const NEOFETCH_LOGO = `  _ __ _        _
 |_)| |_|\\|__ | |_
 |  | | |/|   |_|`;

function buildNeofetch(): string {
  const vw = typeof window !== 'undefined' ? window.innerWidth : 0;
  const vh = typeof window !== 'undefined' ? window.innerHeight : 0;
  let mem = '';
  if (
    typeof window !== 'undefined' &&
    (performance as unknown as { memory?: { usedJSHeapSize: number } }).memory
  ) {
    const mb = Math.round(
      (performance as unknown as { memory: { usedJSHeapSize: number } }).memory.usedJSHeapSize /
        1048576
    );
    mem = `${mb} MB`;
  } else {
    mem = 'N/A';
  }
  return [
    NEOFETCH_LOGO,
    '',
    `OS:         rtylr Desktop 1.0`,
    `Shell:      rtylr-sh 1.0`,
    `Resolution: ${vw}x${vh}`,
    `Memory:     ${mem}`,
  ].join('\n');
}

function runCommand(
  cmd: string,
  addWindow: (type: Parameters<ReturnType<typeof useDesktopStore>['addWindow']>[0]) => void
): string {
  const trimmed = cmd.trim();
  const lower = trimmed.toLowerCase();

  if (lower === 'help') {
    return [
      'Available commands:',
      '  help          show this help',
      '  ls            list apps',
      '  ls apps/      list apps',
      '  cat pricing.txt',
      '  cat about.txt',
      '  ping rtylr.com',
      '  neofetch      system info',
      '  whoami        current user',
      '  date          current date/time',
      '  clear         clear terminal',
      '  open <app>    open a window (pos, erp, crm, hr, finance, flow, pricing, docs, settings, about, terminal, snake, doom, status)',
    ].join('\n');
  }

  if (lower === 'ls' || lower === 'ls apps/' || lower === 'ls apps') {
    return 'pos  erp  crm  hr  finance  flow';
  }

  if (lower === 'cat pricing.txt') {
    return 'rtylr starts at $49/mo — everything your business needs, one flat price. Visit /pricing for details.';
  }

  if (lower === 'cat about.txt') {
    return 'rtylr is an all-in-one business OS — POS, ERP, CRM, HR, Finance & Flow — built for modern retail and hospitality.';
  }

  if (lower === 'ping rtylr.com') {
    const packets = Array.from({ length: 4 }, (_, i) => {
      const ms = (Math.random() * 18 + 4).toFixed(3);
      return `64 bytes from rtylr.com: icmp_seq=${i + 1} ttl=57 time=${ms} ms`;
    });
    return [
      'PING rtylr.com (104.21.0.1): 56 data bytes',
      ...packets,
      '',
      '--- rtylr.com ping statistics ---',
      `4 packets transmitted, 4 received, 0% packet loss`,
    ].join('\n');
  }

  if (lower === 'neofetch') {
    return buildNeofetch();
  }

  if (lower === 'whoami') {
    return 'guest@rtylr';
  }

  if (lower === 'date') {
    return new Date().toString();
  }

  if (lower === 'clear') {
    return '__CLEAR__';
  }

  if (lower.startsWith('open ')) {
    const app = lower.slice(5).trim();
    const validApps: Record<
      string,
      Parameters<ReturnType<typeof useDesktopStore>['addWindow']>[0]
    > = {
      pos: 'pos',
      erp: 'erp',
      crm: 'crm',
      hr: 'hr',
      finance: 'finance',
      flow: 'flow',
      pricing: 'pricing',
      docs: 'docs',
      settings: 'settings',
      about: 'about',
      terminal: 'terminal',
      snake: 'snake',
      doom: 'doom',
      status: 'status',
    };
    if (validApps[app]) {
      addWindow(validApps[app]);
      return `Opening ${app}...`;
    }
    return `open: no such app: ${app}`;
  }

  if (trimmed === '') return '';

  return `command not found: ${trimmed}`;
}

export default function TerminalWindow() {
  const addWindow = useDesktopStore((s) => s.addWindow);
  const [history, setHistory] = useState<HistoryEntry[]>([
    { type: 'output', text: 'rtylr Terminal v1.0 — type "help" for commands' },
  ]);
  const [input, setInput] = useState('');
  const [cmdHistory, setCmdHistory] = useState<string[]>([]);
  const [histIdx, setHistIdx] = useState(-1);
  const bottomRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [history]);

  useEffect(() => {
    inputRef.current?.focus();
  }, []);

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent<HTMLInputElement>) => {
      if (e.key === 'Enter') {
        e.preventDefault();
        const cmd = input;
        const output = runCommand(cmd, addWindow);
        if (output === '__CLEAR__') {
          setHistory([]);
        } else {
          setHistory((h) => [
            ...h,
            { type: 'input', text: cmd },
            ...(output ? [{ type: 'output' as const, text: output }] : []),
          ]);
        }
        if (cmd.trim()) {
          setCmdHistory((h) => [cmd, ...h]);
        }
        setHistIdx(-1);
        setInput('');
        return;
      }
      if (e.key === 'ArrowUp') {
        e.preventDefault();
        setCmdHistory((h) => {
          const next = Math.min(histIdx + 1, h.length - 1);
          setHistIdx(next);
          if (h[next] !== undefined) setInput(h[next]);
          return h;
        });
        return;
      }
      if (e.key === 'ArrowDown') {
        e.preventDefault();
        const next = histIdx - 1;
        if (next < 0) {
          setHistIdx(-1);
          setInput('');
        } else {
          setHistIdx(next);
          setCmdHistory((h) => {
            if (h[next] !== undefined) setInput(h[next]);
            return h;
          });
        }
      }
    },
    [input, histIdx, addWindow]
  );

  return (
    <div
      style={{
        background: '#0a0a0a',
        height: '100%',
        display: 'flex',
        flexDirection: 'column',
        fontFamily: "'Menlo', 'Monaco', 'Courier New', monospace",
        fontSize: 13,
        color: '#22c55e',
        padding: '12px 16px',
        boxSizing: 'border-box',
        overflowY: 'auto',
        cursor: 'text',
      }}
      onClick={() => inputRef.current?.focus()}
    >
      <div style={{ flex: 1, overflowY: 'auto' }}>
        {history.map((entry, i) => (
          <div key={i} style={{ marginBottom: 2, whiteSpace: 'pre-wrap', wordBreak: 'break-all' }}>
            {entry.type === 'input' ? (
              <span>
                <span style={{ color: 'rgba(255,255,255,0.5)' }}>guest@rtylr</span>
                <span style={{ color: 'rgba(255,255,255,0.3)' }}> % </span>
                <span style={{ color: '#ffffff' }}>{entry.text}</span>
              </span>
            ) : (
              <span style={{ color: '#22c55e' }}>{entry.text}</span>
            )}
          </div>
        ))}
        <div ref={bottomRef} />
      </div>

      <div style={{ display: 'flex', alignItems: 'center', marginTop: 4 }}>
        <span style={{ color: 'rgba(255,255,255,0.5)' }}>guest@rtylr</span>
        <span style={{ color: 'rgba(255,255,255,0.3)' }}>&nbsp;%&nbsp;</span>
        <input
          ref={inputRef}
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={handleKeyDown}
          style={{
            flex: 1,
            background: 'transparent',
            border: 'none',
            outline: 'none',
            color: '#ffffff',
            fontFamily: 'inherit',
            fontSize: 'inherit',
            caretColor: '#22c55e',
          }}
          autoComplete="off"
          spellCheck={false}
          autoCorrect="off"
          autoCapitalize="off"
        />
      </div>
    </div>
  );
}
