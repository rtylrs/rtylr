import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { RING_CAP_MS, createAlarm } from '../../../../shared/alarm/alarm';
import { createFakeAlarmEnvironment } from '../../../../shared/alarm/testing/fakeAlarmEnvironment';
import { ticketAlarmState, type AlarmTicket } from './ticketAlarm';

// The ticket page's whole alarm path, minus React: each poll answer goes
// through ticketAlarmState into the real alarm, with only the audio faked.
function setup() {
  const fake = createFakeAlarmEnvironment();
  const alarm = createAlarm({ bellSrc: 'bell.wav', env: fake.env });
  alarm.unlock();
  fake.bell.play.mockClear();
  const poll = (ticket: AlarmTicket): void => alarm.sync(ticketAlarmState(ticket));
  return { ...fake, alarm, poll };
}

const READY_AT = '2026-11-19T10:05:00.000Z';
const waiting: AlarmTicket = { status: 'waiting' };
const ready: AlarmTicket = { status: 'ready', readyAt: READY_AT };

beforeEach(() => {
  vi.useFakeTimers();
});

afterEach(() => {
  vi.useRealTimers();
});

describe('ticket alarm', () => {
  it('stays quiet while waiting and rings when the order becomes ready', () => {
    const { alarm, bell, poll } = setup();

    poll(waiting);
    poll(waiting);
    expect(alarm.getSnapshot().ringing).toBe(false);
    expect(bell.play).not.toHaveBeenCalled();

    poll(ready);
    expect(alarm.getSnapshot().ringing).toBe(true);
    expect(bell.play).toHaveBeenCalledTimes(1);
  });

  it('does not start over on every poll of the same ready order', () => {
    const { bell, poll } = setup();

    poll(ready);
    poll(ready);
    poll(ready);

    expect(bell.play).toHaveBeenCalledTimes(1);
  });

  it('stops when the ticket is acknowledged, on whichever phone that happened', () => {
    const { alarm, bell, poll } = setup();
    poll(ready);

    poll({ ...ready, acknowledgedAt: '2026-11-19T10:05:20.000Z' });

    expect(alarm.getSnapshot().ringing).toBe(false);
    expect(bell.pause).toHaveBeenCalled();
  });

  it('rings again when the staff call again', () => {
    const { alarm, bell, poll } = setup();
    poll(ready);
    poll({ ...ready, acknowledgedAt: '2026-11-19T10:05:20.000Z' });
    bell.play.mockClear();

    // A reminder clears the acknowledgement and stamps reminded_at.
    const reminded = { ...ready, remindedAt: '2026-11-19T10:07:00.000Z' };
    poll(reminded);
    expect(alarm.getSnapshot().ringing).toBe(true);
    expect(bell.play).toHaveBeenCalledTimes(1);

    poll(reminded);
    expect(bell.play).toHaveBeenCalledTimes(1);

    poll({ ...ready, remindedAt: '2026-11-19T10:09:00.000Z' });
    expect(bell.play).toHaveBeenCalledTimes(2);
  });

  it('stays silent after "stop sound" until the next reminder', () => {
    const { alarm, bell, poll } = setup();
    poll(ready);
    alarm.silence();
    bell.play.mockClear();

    poll(ready);
    expect(alarm.getSnapshot().ringing).toBe(false);

    poll({ ...ready, remindedAt: '2026-11-19T10:07:00.000Z' });
    expect(alarm.getSnapshot().ringing).toBe(true);
    expect(bell.play).toHaveBeenCalledTimes(1);
  });

  it('gives up after two minutes without an acknowledgement and does not restart', async () => {
    const { alarm, bell, poll } = setup();
    poll(ready);

    await vi.advanceTimersByTimeAsync(RING_CAP_MS);
    expect(alarm.getSnapshot().ringing).toBe(false);

    bell.play.mockClear();
    poll(ready);
    expect(bell.play).not.toHaveBeenCalled();
  });

  it('rings again when an order is put back to waiting and made ready a second time', () => {
    const { bell, poll } = setup();
    poll(ready);
    poll(waiting);

    poll({ status: 'ready', readyAt: '2026-11-19T10:12:00.000Z' });

    expect(bell.play).toHaveBeenCalledTimes(2);
  });

  it.each(['collected', 'expired'])('stops and lets the screen sleep once %s', async (status) => {
    const { alarm, poll, wakeLocks } = setup();
    poll(ready);
    await vi.advanceTimersByTimeAsync(0);

    poll({ status, readyAt: READY_AT });

    expect(alarm.getSnapshot().ringing).toBe(false);
    expect(wakeLocks[0].release).toHaveBeenCalledTimes(1);
  });

  it('keeps the screen awake while waiting', async () => {
    const { navigator, poll } = setup();

    poll(waiting);
    await vi.advanceTimersByTimeAsync(0);

    expect(navigator.wakeLock.request).toHaveBeenCalledWith('screen');
  });
});
