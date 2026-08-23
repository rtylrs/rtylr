import { describe, it, afterEach } from 'node:test';
import assert from 'node:assert/strict';
import vm from 'node:vm';
import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';

import {
  CONSENT_COOKIE,
  CONSENT_EVENT,
  CONSENT_MODE_BOOTSTRAP,
  GA_MEASUREMENT_ID,
  syncConsentMode,
  watchConsentMode,
} from '../../apps/landing/lib/consentMode.ts';

// ---------------------------------------------------------------------------
// Why this file exists.
//
// GA4 shipped to production firing on every page load while the site's
// own consent banner sat next to it, working, ignored. Both halves of
// the bridge that now connects them are easy to break silently:
//
//   - the cookie name is duplicated from @voxire/consent, because the
//     bootstrap runs in the document head before any module loads
//   - the bootstrap is a string, so nothing type-checks the order of
//     the gtag commands inside it, and `consent default` after `config`
//     is ignored by Google rather than rejected
//
// Both are asserted below against behaviour, not presence.
// ---------------------------------------------------------------------------

/**
 * Execute the head bootstrap the way a browser would and return the
 * dataLayer it produced, as plain arrays.
 *
 * `window` is made self-referential because that is what it is in a
 * browser: the snippet assigns `window.dataLayer` and then reads the
 * bare identifier `dataLayer`, which only resolves if the two are the
 * same object.
 *
 * The result is round-tripped through JSON because objects built inside
 * a vm context carry that context's Object prototype, and deepEqual
 * compares prototypes.
 */
function runBootstrap(cookie) {
  const context = { document: { cookie }, Date };
  context.window = context;
  vm.createContext(context);
  vm.runInContext(CONSENT_MODE_BOOTSTRAP, context);
  return JSON.parse(JSON.stringify(context.dataLayer.map((args) => Array.from(args))));
}

const indexOfCommand = (calls, kind, subtype) =>
  calls.findIndex((c) => c[0] === kind && (subtype === undefined || c[1] === subtype));

describe('consent cookie contract with @voxire/consent', () => {
  const consentSource = readFileSync(
    fileURLToPath(new URL('../../packages/consent/src/consent.ts', import.meta.url)),
    'utf8'
  );

  it('reads the same cookie name that @voxire/consent writes', () => {
    const match = consentSource.match(/const COOKIE_NAME = '([^']+)'/);
    assert.ok(match, 'could not find COOKIE_NAME in @voxire/consent');
    assert.equal(
      match[1],
      CONSENT_COOKIE,
      'consentMode.ts is reading a cookie @voxire/consent does not write, so GA4 would never see consent'
    );
  });

  it('listens for the same event name that @voxire/consent dispatches', () => {
    assert.ok(
      consentSource.includes(`'${CONSENT_EVENT}'`),
      `@voxire/consent no longer dispatches '${CONSENT_EVENT}'`
    );
  });
});

describe('CONSENT_MODE_BOOTSTRAP', () => {
  it('denies every storage type before configuring GA4', () => {
    const calls = runBootstrap('');

    const defaultIdx = indexOfCommand(calls, 'consent', 'default');
    const configIdx = indexOfCommand(calls, 'config');

    assert.notEqual(defaultIdx, -1, 'no consent default command was queued');
    assert.notEqual(configIdx, -1, 'GA4 was never configured');
    assert.ok(
      defaultIdx < configIdx,
      'consent default must be queued before config or Google ignores it'
    );

    assert.deepEqual(calls[defaultIdx][2], {
      ad_storage: 'denied',
      ad_user_data: 'denied',
      ad_personalization: 'denied',
      analytics_storage: 'denied',
      wait_for_update: 500,
    });
  });

  it('configures the rtylr.com measurement ID', () => {
    const calls = runBootstrap('');
    assert.deepEqual(calls[indexOfCommand(calls, 'config')], ['config', GA_MEASUREMENT_ID]);
  });

  it('grants analytics immediately for a visitor who already accepted', () => {
    const calls = runBootstrap(`${CONSENT_COOKIE}=granted`);

    const updateIdx = indexOfCommand(calls, 'consent', 'update');
    assert.notEqual(updateIdx, -1, 'an accepted visitor should not be measured as denied');
    assert.deepEqual(calls[updateIdx][2], { analytics_storage: 'granted' });
    assert.ok(
      updateIdx < indexOfCommand(calls, 'config'),
      'the grant must land before config or the first pageview is sent cookieless'
    );
  });

  it('finds the cookie when other cookies precede it', () => {
    const calls = runBootstrap(`_ga=GA1.1.x; ${CONSENT_COOKIE}=granted; rtylr_variant=a`);
    assert.notEqual(indexOfCommand(calls, 'consent', 'update'), -1);
  });

  it('stays denied when the visitor declined', () => {
    const calls = runBootstrap(`${CONSENT_COOKIE}=denied`);
    assert.equal(indexOfCommand(calls, 'consent', 'update'), -1);
  });

  it('stays denied when no decision has been made', () => {
    const calls = runBootstrap('rtylr_variant=a');
    assert.equal(indexOfCommand(calls, 'consent', 'update'), -1);
  });

  it('is not fooled by a cookie whose name merely ends with the consent name', () => {
    const calls = runBootstrap(`not-${CONSENT_COOKIE}=granted`);
    assert.equal(
      indexOfCommand(calls, 'consent', 'update'),
      -1,
      'a different cookie granted analytics storage'
    );
  });

  it('never grants an advertising consent the banner does not ask for', () => {
    const calls = runBootstrap(`${CONSENT_COOKIE}=granted`);
    for (const call of calls.filter((c) => c[0] === 'consent' && c[1] === 'update')) {
      for (const key of ['ad_storage', 'ad_user_data', 'ad_personalization']) {
        assert.notEqual(call[2][key], 'granted', `${key} was granted without being asked`);
      }
    }
  });
});

describe('syncConsentMode', () => {
  afterEach(() => {
    delete globalThis.window;
  });

  const withGtag = () => {
    const calls = [];
    globalThis.window = { gtag: (...args) => calls.push(args) };
    return calls;
  };

  it('grants analytics storage on accept', () => {
    const calls = withGtag();
    syncConsentMode('granted');
    assert.deepEqual(calls, [['consent', 'update', { analytics_storage: 'granted' }]]);
  });

  it('denies analytics storage on decline', () => {
    const calls = withGtag();
    syncConsentMode('denied');
    assert.deepEqual(calls, [['consent', 'update', { analytics_storage: 'denied' }]]);
  });

  it('denies on undecided, which is what reopening the banner produces', () => {
    const calls = withGtag();
    syncConsentMode('undecided');
    assert.deepEqual(calls, [['consent', 'update', { analytics_storage: 'denied' }]]);
  });

  it('does nothing when gtag was blocked rather than throwing', () => {
    globalThis.window = {};
    assert.doesNotThrow(() => syncConsentMode('granted'));
  });
});

describe('watchConsentMode', () => {
  afterEach(() => {
    delete globalThis.window;
    delete globalThis.document;
  });

  /**
   * Minimal window with an event target and a gtag spy. The real
   * listener plumbing matters here: the bug this covers was in what
   * watchConsentMode does *before* any event arrives.
   */
  function mount(cookie) {
    const calls = [];
    const listeners = {};
    globalThis.document = { cookie };
    globalThis.window = {
      gtag: (...args) => calls.push(args),
      addEventListener: (name, fn) => {
        (listeners[name] ??= []).push(fn);
      },
      removeEventListener: (name, fn) => {
        listeners[name] = (listeners[name] ?? []).filter((f) => f !== fn);
      },
    };
    const off = watchConsentMode();
    const emit = (detail) => {
      for (const fn of listeners[CONSENT_EVENT] ?? []) fn({ detail });
    };
    return { calls, emit, off, listeners };
  }

  it('pushes nothing for an undecided visitor', () => {
    // An update, even a denial, ends the bootstrap's wait_for_update
    // grace window, so a first-time visitor who accepts a moment later
    // has already been counted as denied.
    const { calls } = mount('');
    assert.deepEqual(calls, []);
  });

  it('re-grants on mount for a visitor who already accepted', () => {
    const { calls } = mount(`${CONSENT_COOKIE}=granted`);
    assert.deepEqual(calls, [['consent', 'update', { analytics_storage: 'granted' }]]);
  });

  it('re-denies on mount for a visitor who already declined', () => {
    const { calls } = mount(`${CONSENT_COOKIE}=denied`);
    assert.deepEqual(calls, [['consent', 'update', { analytics_storage: 'denied' }]]);
  });

  it('grants when the banner dispatches an accept', () => {
    const { calls, emit } = mount('');
    emit('granted');
    assert.deepEqual(calls, [['consent', 'update', { analytics_storage: 'granted' }]]);
  });

  it('denies when the banner dispatches a decline', () => {
    const { calls, emit } = mount('');
    emit('denied');
    assert.deepEqual(calls, [['consent', 'update', { analytics_storage: 'denied' }]]);
  });

  it('revokes when consent is cleared and the banner reopens', () => {
    const { calls, emit } = mount(`${CONSENT_COOKIE}=granted`);
    emit('undecided');
    assert.deepEqual(calls.at(-1), ['consent', 'update', { analytics_storage: 'denied' }]);
  });

  it('accepts the { state } detail shape as well as the bare string', () => {
    const { calls, emit } = mount('');
    emit({ state: 'granted' });
    assert.deepEqual(calls, [['consent', 'update', { analytics_storage: 'granted' }]]);
  });

  it('stops listening after the returned cleanup runs', () => {
    const { calls, emit, off } = mount('');
    off();
    emit('granted');
    assert.deepEqual(calls, [], 'the effect cleanup left a listener attached');
  });
});
