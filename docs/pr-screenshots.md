# Screenshots in PR bodies

The UI standards addendum makes a screenshot part of the definition of done for
any change a user can see. This is how to produce one and get it into a PR body.

It is written down because the requirement was recorded as **"cannot be
delivered"** in the Flow workspace-removal tracker, on the belief that no image
file could be produced from an agent environment. That was wrong. A requirement
written off as impossible does not get revisited unless something in the repo
proves otherwise, so the helper, the tagged spec and this page exist together.

## Producing the images

```bash
pnpm test:screenshots
```

If that fails with `Executable doesn't exist`, the browser cache does not match
the Playwright version this repo pins. `npx playwright install chromium` is the
normal fix, but it can appear to work while delivering nothing: it left a 448 kB
stub directory and never exited. Point at a build that is actually on disk
instead:

```bash
PLAYWRIGHT_CHROMIUM_PATH="$HOME/Library/Caches/ms-playwright/chromium-1223/chrome-mac-arm64/Google Chrome for Testing.app/Contents/MacOS/Google Chrome for Testing" pnpm test:screenshots
```

`playwright.config.ts` reads that variable and is a no-op without it.

That runs only the specs tagged `@screenshots` and writes PNGs into
`e2e/__screenshots__/`, which is gitignored: they are review artefacts, not
fixtures, and nothing asserts against them.

To add a capture for your own screen, tag a `describe` with `@screenshots` and
call the helper:

```ts
import { captureResponsivePair } from './support/screenshot';

test.describe('my feature screenshots @screenshots', () => {
  test('captures the screen', async ({ page }) => {
    await installStubs(page);
    await page.goto(`${APP_URL}/my-route`);
    await expect(page.getByText('something real')).toBeVisible();

    const shots = await captureResponsivePair(page, 'my-feature');
    expect(shots.overflowsAt390).toBe(false);
  });
});
```

Reuse the stub installer the app's existing spec already has rather than writing
a new one. Four HR specs each carry their own copy; do not make it five.

## Before you put a capture on an authenticated screen

**Most authenticated specs currently do not reach their screen.** As of
2026-08-24, running the suite locally, the app redirects to the auth app at
`localhost:5176/?redirect=<app>` **before making a single API request**, so the
screen under test never renders:

| Spec                        | Result           |
| --------------------------- | ---------------- |
| `hr-staff.spec.ts`          | 7 of 8 failed    |
| `hr-payslip-search.spec.ts` | 5 of 5 failed    |
| `dash.spec.ts`              | 1 of 2 failed    |
| `flow-guards.spec.ts`       | 1 of 4 failed    |
| `menu-public.spec.ts`       | passed (no auth) |

That is why the reference capture lives on the **public** menu: a capture placed
on an authenticated screen would have shipped exercising nothing.

Two plausible causes were tested and **ruled out**, so nobody repeats them:

- **Not the setup gate.** `isHRSetupComplete` compares `Number(setupCompletedAt) > 0`
  and the fixtures send `'2026-01-01T00:00:00Z'`, which is `NaN`. Genuinely
  wrong, and fixing it to an epoch string changed nothing.
- **Not the role.** `AppGuard` fails closed on a role it cannot parse, even when
  the `apps` array would admit the user, and four HR fixtures use `'hr_admin'`,
  which is not a `PlatformRole`. Also genuinely wrong, and also changed nothing.

The decisive observation is that **zero requests reach the API**, so the guard
redirects on its cookie check before any of the above runs. The cookie is
present in the browser context in both `{ url }` and `{ domain, path }` form,
and `document.cookie` reads empty after the redirect. That is where the next
person should start.

## Why the helper rather than `page.screenshot()`

`captureForPR` finishes running animations and waits for fonts first. Both have
produced misleading images:

- **Stalled animations.** Entrance transitions do not advance when the browser
  surface is not visible, so a bare screenshot catches content mid-fade at
  partial opacity. The same stall made an axe contrast audit report 29 failures
  where 3 were real, because it measured colours mid-transition.
- **Unloaded fonts.** The first shot of a run can render in the fallback face,
  which makes a design review argue about the wrong type.

`captureResponsivePair` also shoots at 390px, the stated mobile floor, and
**measures** horizontal overflow rather than leaving it to be eyeballed off the
image. A screenshot cannot show you that a page scrolls sideways.

## Getting them into a PR body

GitHub's markdown will not render a local path, and this is a private repo, so a
`raw.githubusercontent.com` link is not publicly fetchable either. Commit the
images to the branch, link them by that commit's SHA, then remove them from the
tip:

```bash
mkdir -p .pr-shots && cp e2e/__screenshots__/*.png .pr-shots/
git add .pr-shots && git commit -m "Add the screenshots for the PR body"
git push
SHA=$(git rev-parse HEAD)          # the images live here, permanently
git rm -r .pr-shots && git commit -m "Remove the screenshot files from the branch tip"
git push
```

Then reference them at that SHA:

```html
<img
  src="https://github.com/tryrtylr/rtylr/blob/<SHA>/.pr-shots/01-desktop.png?raw=1"
  width="900"
  alt="what the image shows"
/>
```

The blob stays reachable at that commit even after the tip removes the files, so
the PR body keeps working and `main` never carries the PNGs.

Give every image a real `alt` describing what it shows. A reviewer on a slow
connection, and anyone reading the PR later through a screen reader, gets the
`alt` and nothing else.

## What a screenshot does not prove

Attach the measurements too. Today's HR reporting-lines PR is the pattern: the
images show the sidebar entry, and a table beside them records that the row
buttons are 44x77, the select is 44px at 16px font, and the nav row is 36px and
therefore **under** the touch-target floor. The last one is invisible in the
image and is the only real finding of the three.

Measure by accessible name, never `document.querySelector('button')` — that
returns the first button in DOM order, which is usually app chrome. Measuring
the wrong element produced a false "25px, under the floor" claim that had to be
retracted.
