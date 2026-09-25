import { mkdir, readdir, readFile, writeFile } from 'node:fs/promises';
import path from 'node:path';

const inputDir = process.argv[2] ?? 'desktop-downloads';
const outputDir = process.argv[3] ?? 'desktop-public';
const baseUrl = (process.env.DOWNLOADS_BASE_URL ?? 'https://pos-downloads.rtylr.com').replace(/\/$/, '');
const assetsBaseUrl = (process.env.DOWNLOADS_ASSETS_BASE_URL ?? baseUrl).replace(/\/$/, '');

const installerPattern = /\.(dmg|exe|msi|deb|rpm|appimage|app\.tar\.gz)$/i;

const fileNameFromUrl = (url) => {
  try {
    return decodeURIComponent(path.basename(new URL(url).pathname));
  } catch {
    return path.basename(url);
  }
};

const rewriteLatestJson = async () => {
  const latestPath = path.join(inputDir, 'latest.json');
  const latest = JSON.parse(await readFile(latestPath, 'utf8'));

  for (const platform of Object.values(latest.platforms ?? {})) {
    if (!platform || typeof platform !== 'object' || typeof platform.url !== 'string') continue;
    platform.url = `${assetsBaseUrl}/${fileNameFromUrl(platform.url)}`;
  }

  await writeFile(path.join(outputDir, 'latest.json'), `${JSON.stringify(latest, null, 2)}\n`);
};

const renderIndex = async (files) => {
  const installers = files.filter((file) => installerPattern.test(file));
  const rows = installers
    .map((file) => `<li><a href="${assetsBaseUrl}/${encodeURIComponent(file)}">${file}</a></li>`)
    .join('\n');

  await writeFile(
    path.join(outputDir, 'index.html'),
    `<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>rtylr POS Downloads</title>
    <style>
      :root { color-scheme: dark; font-family: Inter, ui-sans-serif, system-ui, sans-serif; }
      body { margin: 0; min-height: 100vh; display: grid; place-items: center; background: #101010; color: #f5f5f5; }
      main { width: min(720px, calc(100vw - 48px)); }
      h1 { font-size: 32px; margin: 0 0 12px; }
      p { color: #b8b8b8; margin: 0 0 28px; }
      ul { list-style: none; padding: 0; margin: 0; border: 1px solid #303030; border-radius: 12px; overflow: hidden; }
      li + li { border-top: 1px solid #303030; }
      a { display: block; padding: 16px 18px; color: #ffffff; text-decoration: none; background: #181818; }
      a:hover { background: #222222; }
    </style>
  </head>
  <body>
    <main>
      <h1>rtylr POS Downloads</h1>
      <p>Download the installer for your operating system.</p>
      <ul>
${rows}
      </ul>
    </main>
  </body>
</html>
`,
  );
};

await mkdir(outputDir, { recursive: true });

const files = await readdir(inputDir);

await rewriteLatestJson();
await renderIndex(files);
