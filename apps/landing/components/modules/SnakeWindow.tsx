import React, { useEffect, useRef, useState, useCallback } from 'react';

const GRID = 20;
const CELL = 20;
const CANVAS = GRID * CELL;
const TICK = 150;

type Dir = 'UP' | 'DOWN' | 'LEFT' | 'RIGHT';
type Pt = { x: number; y: number };

function rand(): Pt {
  return { x: Math.floor(Math.random() * GRID), y: Math.floor(Math.random() * GRID) };
}

function newFood(snake: Pt[]): Pt {
  let f: Pt;
  do {
    f = rand();
  } while (snake.some((s) => s.x === f.x && s.y === f.y));
  return f;
}

function initState() {
  const snake = [
    { x: 10, y: 10 },
    { x: 9, y: 10 },
    { x: 8, y: 10 },
  ];
  return {
    snake,
    food: newFood(snake),
    dir: 'RIGHT' as Dir,
    nextDir: 'RIGHT' as Dir,
    score: 0,
    dead: false,
  };
}

type GameState = ReturnType<typeof initState>;

export default function SnakeWindow() {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const stateRef = useRef<GameState>(initState());
  const [score, setScore] = useState(0);
  const [dead, setDead] = useState(false);
  const pausedRef = useRef(false);
  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const draw = useCallback(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const ctx = canvas.getContext('2d');
    if (!ctx) return;
    const gs = stateRef.current;

    ctx.fillStyle = '#0f0f0f';
    ctx.fillRect(0, 0, CANVAS, CANVAS);

    ctx.fillStyle = '#22c55e';
    gs.snake.forEach((s, i) => {
      ctx.globalAlpha = i === 0 ? 1 : 0.75 - (i / gs.snake.length) * 0.3;
      ctx.fillRect(s.x * CELL + 1, s.y * CELL + 1, CELL - 2, CELL - 2);
    });
    ctx.globalAlpha = 1;

    ctx.fillStyle = '#ef4444';
    ctx.fillRect(gs.food.x * CELL + 2, gs.food.y * CELL + 2, CELL - 4, CELL - 4);
  }, []);

  const tick = useCallback(() => {
    if (pausedRef.current) return;
    const gs = stateRef.current;
    if (gs.dead) return;

    gs.dir = gs.nextDir;
    const head = gs.snake[0];
    const newHead: Pt = { x: head.x, y: head.y };
    if (gs.dir === 'UP') newHead.y -= 1;
    if (gs.dir === 'DOWN') newHead.y += 1;
    if (gs.dir === 'LEFT') newHead.x -= 1;
    if (gs.dir === 'RIGHT') newHead.x += 1;

    if (newHead.x < 0 || newHead.x >= GRID || newHead.y < 0 || newHead.y >= GRID) {
      gs.dead = true;
      setDead(true);
      draw();
      return;
    }
    if (gs.snake.some((s) => s.x === newHead.x && s.y === newHead.y)) {
      gs.dead = true;
      setDead(true);
      draw();
      return;
    }

    const ate = newHead.x === gs.food.x && newHead.y === gs.food.y;
    gs.snake = [newHead, ...gs.snake];
    if (!ate) gs.snake.pop();
    else {
      gs.food = newFood(gs.snake);
      gs.score += 1;
      setScore(gs.score);
    }
    draw();
  }, [draw]);

  const restart = useCallback(() => {
    stateRef.current = initState();
    setScore(0);
    setDead(false);
    draw();
  }, [draw]);

  useEffect(() => {
    draw();
    timerRef.current = setInterval(tick, TICK);
    return () => {
      if (timerRef.current) clearInterval(timerRef.current);
    };
  }, [tick, draw]);

  useEffect(() => {
    const handleKey = (e: KeyboardEvent) => {
      const gs = stateRef.current;
      if (e.key === 'ArrowUp' || e.key === 'w') {
        e.preventDefault();
        if (gs.dir !== 'DOWN') gs.nextDir = 'UP';
      }
      if (e.key === 'ArrowDown' || e.key === 's') {
        e.preventDefault();
        if (gs.dir !== 'UP') gs.nextDir = 'DOWN';
      }
      if (e.key === 'ArrowLeft' || e.key === 'a') {
        e.preventDefault();
        if (gs.dir !== 'RIGHT') gs.nextDir = 'LEFT';
      }
      if (e.key === 'ArrowRight' || e.key === 'd') {
        e.preventDefault();
        if (gs.dir !== 'LEFT') gs.nextDir = 'RIGHT';
      }
      if ((e.key === ' ' || e.key === 'Enter') && gs.dead) restart();
    };
    window.addEventListener('keydown', handleKey);
    return () => window.removeEventListener('keydown', handleKey);
  }, [restart]);

  useEffect(() => {
    const onVis = () => {
      pausedRef.current = document.hidden;
    };
    document.addEventListener('visibilitychange', onVis);
    return () => document.removeEventListener('visibilitychange', onVis);
  }, []);

  return (
    <div
      style={{
        background: '#151515',
        height: '100%',
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        gap: 12,
      }}
    >
      <div style={{ fontSize: 13, color: 'rgba(255,255,255,0.6)', fontFamily: 'monospace' }}>
        Score: <span style={{ color: '#22c55e', fontWeight: 700 }}>{score}</span>
      </div>
      <div style={{ position: 'relative' }}>
        <canvas
          ref={canvasRef}
          width={CANVAS}
          height={CANVAS}
          style={{ display: 'block', borderRadius: 4, border: '1px solid rgba(255,255,255,0.08)' }}
        />
        {dead && (
          <div
            style={{
              position: 'absolute',
              inset: 0,
              background: 'rgba(0,0,0,0.75)',
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'center',
              justifyContent: 'center',
              gap: 8,
              borderRadius: 4,
            }}
          >
            <div style={{ fontSize: 20, fontWeight: 700, color: '#ef4444' }}>Game Over</div>
            <div style={{ fontSize: 13, color: 'rgba(255,255,255,0.6)' }}>Score: {score}</div>
            <div style={{ fontSize: 11, color: 'rgba(255,255,255,0.4)', marginTop: 4 }}>
              Press Space / Enter to restart
            </div>
          </div>
        )}
      </div>
      <div style={{ fontSize: 11, color: 'rgba(255,255,255,0.3)' }}>Arrow keys or WASD to move</div>
    </div>
  );
}
