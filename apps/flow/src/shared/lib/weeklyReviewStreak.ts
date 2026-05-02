import { mondayStartMs } from './lifeRhythm';

const KEY = 'flow.weeklyReviewStreak.v1';

interface Stored {
  streak: number;
  /** Monday 00:00 local ms for the last week you logged a review */
  lastWeekMonday: number | null;
}

function load(): Stored {
  try {
    const raw = localStorage.getItem(KEY);
    if (!raw) return { streak: 0, lastWeekMonday: null };
    const p = JSON.parse(raw) as Stored;
    return {
      streak: typeof p.streak === 'number' ? p.streak : 0,
      lastWeekMonday: typeof p.lastWeekMonday === 'number' ? p.lastWeekMonday : null,
    };
  } catch {
    return { streak: 0, lastWeekMonday: null };
  }
}

function save(s: Stored): void {
  try {
    localStorage.setItem(KEY, JSON.stringify(s));
  } catch {
    /* ignore */
  }
}

export function getWeeklyReviewStreak(): { streak: number; lastWeekMonday: number | null } {
  const s = load();
  return { streak: s.streak, lastWeekMonday: s.lastWeekMonday };
}

/** Mark “weekly review” done for the current week; extends streak across consecutive weeks. */
export function recordWeeklyReviewDone(): { streak: number; alreadyLogged: boolean } {
  const thisWeek = mondayStartMs(new Date());
  const data = load();
  if (data.lastWeekMonday === thisWeek) {
    return { streak: data.streak, alreadyLogged: true };
  }
  const prevWeek = thisWeek - 7 * 86400000;
  let streak = 1;
  if (data.lastWeekMonday === prevWeek) {
    streak = data.streak + 1;
  }
  const next = { streak, lastWeekMonday: thisWeek };
  save(next);
  return { streak, alreadyLogged: false };
}
