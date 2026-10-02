import { test, expect } from '@playwright/test';
import * as fs from 'fs';
import * as path from 'path';

const FLOW_URL = 'http://localhost:5197/flow';
const EVIDENCE_DIR = '/Users/abedamouneh/Development/vault/rtylr/audit/flow-qa-2026-08-v2/evidence';

async function primeAuthCookie(context: import('@playwright/test').BrowserContext): Promise<void> {
  const fakeJwt = 'e2e-test-token';
  await context.addCookies([
    {
      name: 'token',
      value: fakeJwt,
      domain: 'localhost',
      path: '/',
      httpOnly: false,
      sameSite: 'Lax',
      secure: false,
    },
  ]);
}

test('QA Driver - Task Detail Live Interrogation', async ({ page, context }) => {
  await primeAuthCookie(context);

  if (!fs.existsSync(EVIDENCE_DIR)) {
    fs.mkdirSync(EVIDENCE_DIR, { recursive: true });
  }

  // Intercept network requests to record payloads
  const networkLogs: string[] = [];
  page.on('response', async (response) => {
    const url = response.url();
    if (url.includes('/flow/') || url.includes('/hub/')) {
      try {
        const body = await response.text();
        networkLogs.push(
          `[${response.status()}] ${response.request().method()} ${url}\nResponse: ${body.slice(0, 300)}`
        );
      } catch {
        networkLogs.push(`[${response.status()}] ${response.request().method()} ${url}`);
      }
    }
  });

  // 1. Load task detail for Task 101
  await page.goto(`${FLOW_URL}/task/101`);
  await page.waitForLoadState('networkidle');

  // Screenshot E-003: Task Detail Initial View
  await page.screenshot({ path: path.join(EVIDENCE_DIR, 'E-003.png'), fullPage: true });

  // 2. Perform interactive control checks
  // Edit title
  const titleInput = page.locator('input[value*="SSO"], input.font-bold').first();
  if (await titleInput.isVisible()) {
    await titleInput.fill('Implement SSO Auth Flow (Updated QA)');
    await titleInput.blur();
  }

  // Add subtask
  const subtaskInput = page.getByPlaceholder(/Add subtask/i).first();
  if (await subtaskInput.isVisible()) {
    await subtaskInput.fill('QA Subtask Test Item');
    await subtaskInput.press('Enter');
  }

  // Add comment
  const commentInput = page.getByPlaceholder(/Write a comment/i).first();
  if (await commentInput.isVisible()) {
    await commentInput.fill('QA Automated Test Comment Verification');
    const submitBtn = page.getByRole('button', { name: /Comment/i }).first();
    if (await submitBtn.isVisible()) {
      await submitBtn.click();
    }
  }

  await page.waitForTimeout(1000);

  // Screenshot E-004: Task Detail Post-Edits View
  await page.screenshot({ path: path.join(EVIDENCE_DIR, 'E-004.png'), fullPage: true });

  // Write captured network payloads
  fs.writeFileSync(path.join(EVIDENCE_DIR, 'E-005.txt'), networkLogs.join('\n\n'));
});
