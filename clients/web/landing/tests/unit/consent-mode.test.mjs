import { describe, it, afterEach } from 'node:test';
import assert from 'node:assert/strict';
import vm from 'node:vm';
import {
  CONSENT_MODE_BOOTSTRAP,
  CONSENT_EVENT,
  GA_MEASUREMENT_ID,
  readConsentCookie,
  watchConsentMode,
} from '../../lib/consentMode.ts';

function boot(cookie = '') {
  const scripts = [],
    writes = [];
  const document = {
    get cookie() {
      return cookie;
    },
    set cookie(value) {
      writes.push(value);
    },
    createElement: () => ({}),
    head: { appendChild: (script) => scripts.push(script) },
  };
  const context = { document, Date, location: { hostname: 'rtylr.com' } };
  context.window = context;
  vm.runInNewContext(CONSENT_MODE_BOOTSTRAP, context);
  return { context, scripts, writes };
}
const commands = (ctx) => JSON.parse(JSON.stringify(ctx.dataLayer.map((args) => Array.from(args))));

describe('basic analytics consent', () => {
  for (const cookie of [
    '',
    'rtylr-consent=denied',
    'rtylr-consent=%ZZ',
    'not-rtylr-consent=granted',
  ]) {
    it(`does not load or configure analytics for ${cookie || 'no decision'}`, () => {
      const { context, scripts } = boot(cookie);
      assert.equal(scripts.length, 0);
      assert.equal(
        commands(context).some((c) => c[0] === 'config'),
        false
      );
      assert.equal(context[`ga-disable-${GA_MEASUREMENT_ID}`], true);
    });
  }
  it('loads once after a grant, with consent before configuration', () => {
    const { context, scripts } = boot();
    context.rtylrSyncAnalytics('granted');
    context.rtylrSyncAnalytics('granted');
    assert.equal(scripts.length, 1);
    assert.equal(
      scripts[0].src,
      `https://www.googletagmanager.com/gtag/js?id=${GA_MEASUREMENT_ID}`
    );
    const calls = commands(context);
    assert.ok(
      calls.findIndex((c) => c[0] === 'consent' && c[1] === 'update') <
        calls.findIndex((c) => c[0] === 'config')
    );
    assert.equal(calls.filter((c) => c[0] === 'config').length, 1);
    assert.equal(context[`ga-disable-${GA_MEASUREMENT_ID}`], false);
  });
  it('recognizes returning consent, disables on withdrawal and clears only GA identifiers', () => {
    const { context, scripts, writes } = boot(
      'session=keep; rtylr-consent=granted; _ga=abc; _ga_TEST=def'
    );
    assert.equal(scripts.length, 1);
    context.rtylrSyncAnalytics('undecided');
    assert.equal(context[`ga-disable-${GA_MEASUREMENT_ID}`], true);
    assert.ok(writes.includes('_ga=; Path=/; Max-Age=0; SameSite=Lax; Domain=.rtylr.com'));
    assert.ok(writes.some((s) => s.startsWith('_ga_TEST=')));
    assert.equal(
      writes.some((s) => s.startsWith('session=')),
      false
    );
    context.rtylrSyncAnalytics('granted');
    assert.equal(scripts.length, 1);
    assert.equal(context[`ga-disable-${GA_MEASUREMENT_ID}`], false);
  });
  it('never grants advertising consent', () => {
    const { context } = boot('rtylr-consent=granted');
    for (const c of commands(context).filter((c) => c[0] === 'consent')) {
      for (const key of ['ad_storage', 'ad_user_data', 'ad_personalization'])
        assert.notEqual(c[2][key], 'granted');
    }
  });
});

describe('consent lifecycle bridge', () => {
  afterEach(() => {
    delete globalThis.window;
    delete globalThis.document;
  });
  it('fails closed on malformed cookies', () => {
    globalThis.document = { cookie: 'rtylr-consent=%ZZ' };
    assert.equal(readConsentCookie(), 'undecided');
  });
  it('reads the persisted decision on events and focus, and cleans up', () => {
    const calls = [],
      target = new EventTarget();
    globalThis.document = { cookie: '' };
    globalThis.window = Object.assign(target, { rtylrSyncAnalytics: (state) => calls.push(state) });
    const off = watchConsentMode();
    globalThis.document.cookie = 'rtylr-consent=granted';
    target.dispatchEvent(new Event(CONSENT_EVENT));
    globalThis.document.cookie = 'rtylr-consent=denied';
    target.dispatchEvent(new Event('focus'));
    assert.deepEqual(calls, ['undecided', 'granted', 'denied']);
    off();
    target.dispatchEvent(new Event(CONSENT_EVENT));
    assert.equal(calls.length, 3);
  });
});
