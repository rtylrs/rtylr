import bellSrc from './bell.wav';
import { createAlarm } from './alarm';

// One instance for the whole document: the tap that unlocks sound happens on
// one screen and the ring on the next, and the unlocked audio element has to
// be the same object on both.
export const readyAlarm = createAlarm({ bellSrc });
