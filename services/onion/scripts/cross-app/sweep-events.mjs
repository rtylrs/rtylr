// Inventory the asynchronous backbone: every wsbus publish site with its
// topic, every registered job, and the webhook subsystem's emission status.
// Output: backend_events.csv
import { readFileSync } from 'node:fs';
import { join } from 'node:path';
import { walk, paths, writeCsv, relRepo } from './lib.mjs';

const rows = [];

// ---- wsbus publishes ----
const pubRe = /wsbus\.Publish\(wsbus\.Event\{([\s\S]*?)\}\)/g;
for (const f of walk(paths.modules, ['.go'])) {
  const mod = f.slice(paths.modules.length + 1).split('/')[0];
  const t = readFileSync(f, 'utf8');
  let m;
  while ((m = pubRe.exec(t))) {
    const body = m[1];
    const topic = /Topic:\s*"([^"]+)"/.exec(body) || /Topic:\s*([\w.()]+)/.exec(body);
    const type = /Type:\s*"([^"]+)"/.exec(body);
    rows.push({
      mechanism: 'wsbus-publish',
      module: mod,
      name: topic ? topic[1] : 'topic-not-literal',
      detail: type ? `type=${type[1]}` : '',
      consumers: '',
      file: relRepo(f),
      line: t.slice(0, m.index).split('\n').length,
    });
  }
}

// wsbus subscriptions inside modules.
for (const f of walk(paths.modules, ['.go'])) {
  const mod = f.slice(paths.modules.length + 1).split('/')[0];
  const t = readFileSync(f, 'utf8');
  const re = /\.(?:Subscribe|Handle)\(\s*"([^"]+)"/g;
  let m;
  while ((m = re.exec(t))) {
    if (!/wsbus|bus|hub/i.test(t.slice(Math.max(0, m.index - 400), m.index))) continue;
    rows.push({
      mechanism: 'wsbus-subscribe',
      module: mod,
      name: m[1],
      detail: '',
      consumers: '',
      file: relRepo(f),
      line: t.slice(0, m.index).split('\n').length,
    });
  }
}

// ---- jobs ----
for (const f of walk(paths.modules, ['.go'])) {
  const mod = f.slice(paths.modules.length + 1).split('/')[0];
  const t = readFileSync(f, 'utf8');
  const re = /registry\.Register\((?:platformjobs|jobs)\.Job\{([\s\S]*?)\}\)/g;
  let m;
  while ((m = re.exec(t))) {
    const body = m[1];
    const name = /Name:\s*"([^"]+)"/.exec(body);
    const sched = /Schedule:\s*"([^"]+)"|ScheduleFromEnv\("([^"]+)",\s*"([^"]+)"\)/.exec(body);
    rows.push({
      mechanism: 'job',
      module: mod,
      name: name ? name[1] : 'name-not-literal',
      detail: sched ? `schedule=${sched[1] || sched[3] || 'env:' + (sched[2] || '')}` : '',
      consumers: 'worker scheduler',
      file: relRepo(f),
      line: t.slice(0, m.index).split('\n').length,
    });
  }
}

// Scheduled jobs via chrono or RegisterScheduledJobs without registry.Job.
// Recorded as a coverage note rather than guessed.

// ---- webhook emission check ----
// The webhooks module stores webhook registrations. Search the whole repo for
// any code that POSTs to a stored webhook URL (an emission engine).
let emissionSites = [];
for (const f of walk(join(paths.platform, '..'), ['.go'])) {
  const t = readFileSync(f, 'utf8');
  if (/webhook/i.test(t) && /(http\.Post|http\.NewRequest|client\.Do)/.test(t)) {
    emissionSites.push(relRepo(f));
  }
}
for (const s of emissionSites) {
  rows.push({
    mechanism: 'webhook-emission',
    module: 'search',
    name: 'POST-to-webhook-URL',
    detail: '',
    consumers: '',
    file: s,
    line: 0,
  });
}

rows.sort((a, b) => a.mechanism.localeCompare(b.mechanism) || a.module.localeCompare(b.module));
writeCsv(
  'backend_events.csv',
  ['mechanism', 'module', 'name', 'detail', 'consumers', 'file', 'line'],
  rows.map((r) => [r.mechanism, r.module, r.name, r.detail, r.consumers, r.file, r.line]),
);
