#!/usr/bin/env node
// Writes src/shared/alarm/bell.wav: a three-note bell, synthesised here so the
// sound is ours. Deterministic; rerun after changing a constant below.
//
//   node scripts/generate-alarm-bell.mjs

import { writeFileSync } from 'node:fs';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

const SAMPLE_RATE = 24_000;
const DURATION_S = 1.25;
// C6, E6, G6: inside the band a phone speaker reproduces loudly.
const NOTES = [
  { at: 0, hz: 1046.5 },
  { at: 0.16, hz: 1318.5 },
  { at: 0.32, hz: 1568 },
];
// Slightly inharmonic upper partials are what make a sine sound struck.
const PARTIALS = [
  { ratio: 1, gain: 1, decay: 5 },
  { ratio: 2.0, gain: 0.45, decay: 7 },
  { ratio: 3.01, gain: 0.22, decay: 10 },
  { ratio: 4.17, gain: 0.1, decay: 14 },
];
const ATTACK_S = 0.004;

const total = Math.round(SAMPLE_RATE * DURATION_S);
const samples = new Float64Array(total);
for (const note of NOTES) {
  const start = Math.round(note.at * SAMPLE_RATE);
  for (let i = start; i < total; i += 1) {
    const t = (i - start) / SAMPLE_RATE;
    const attack = Math.min(1, t / ATTACK_S);
    let value = 0;
    for (const partial of PARTIALS) {
      value +=
        partial.gain *
        Math.exp(-partial.decay * t) *
        Math.sin(2 * Math.PI * note.hz * partial.ratio * t);
    }
    samples[i] += attack * value;
  }
}

// Fade the tail to zero so the loop point does not click, then soft-clip and
// normalise: the soft clip raises the average level without harsh distortion.
const fadeFrom = total - Math.round(0.05 * SAMPLE_RATE);
let peak = 0;
for (let i = 0; i < total; i += 1) {
  if (i > fadeFrom) samples[i] *= (total - i) / (total - fadeFrom);
  samples[i] = Math.tanh(1.4 * samples[i]);
  peak = Math.max(peak, Math.abs(samples[i]));
}

const data = Buffer.alloc(total * 2);
for (let i = 0; i < total; i += 1) {
  data.writeInt16LE(Math.round((samples[i] / peak) * 0.97 * 32767), i * 2);
}

const header = Buffer.alloc(44);
header.write('RIFF', 0);
header.writeUInt32LE(36 + data.length, 4);
header.write('WAVEfmt ', 8);
header.writeUInt32LE(16, 16);
header.writeUInt16LE(1, 20); // PCM
header.writeUInt16LE(1, 22); // mono
header.writeUInt32LE(SAMPLE_RATE, 24);
header.writeUInt32LE(SAMPLE_RATE * 2, 28);
header.writeUInt16LE(2, 32);
header.writeUInt16LE(16, 34);
header.write('data', 36);
header.writeUInt32LE(data.length, 40);

const out = resolve(dirname(fileURLToPath(import.meta.url)), '../src/shared/alarm/bell.wav');
writeFileSync(out, Buffer.concat([header, data]));
console.warn(`wrote ${out} (${44 + data.length} bytes)`);
