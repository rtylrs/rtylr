import { describe, it } from 'node:test';
import assert from 'node:assert/strict';
import {
  parseHtml,
  htmlToMarkdown,
  extractPageMeta,
  buildPagesIndex,
  buildLlmsTxt,
  versionFromReleaseTag,
} from '../../scripts/generate-agent-artifacts.mjs';

const PAGE = (main) => `<!DOCTYPE html><html><head><title>T · rtylr</title></head><body>${main}</body></html>`;

describe('htmlToMarkdown', () => {
  it('scopes conversion to <main> and drops chrome like nav/footer/script', () => {
    const md = htmlToMarkdown(
      PAGE(`<nav><a href="/menu/">Menu</a></nav>
            <main><h1>Retail OS</h1><p>Sell everywhere.</p></main>
            <footer>Contact us</footer>
            <script>window.x=1;</script>`)
    );
    assert.ok(md.includes('# Retail OS'));
    assert.ok(md.includes('Sell everywhere.'));
    assert.ok(!md.includes('Menu'), 'nav must be stripped');
    assert.ok(!md.includes('Contact us'), 'footer must be stripped');
    assert.ok(!md.includes('window.x'), 'script must be stripped');
    assert.ok(!md.includes('<'), 'no raw HTML tags may leak');
  });

  it('renders heading levels', () => {
    const md = htmlToMarkdown(PAGE('<main><h1>A</h1><h2>B</h2><h3>C</h3></main>'));
    assert.ok(md.includes('# A'));
    assert.ok(md.includes('## B'));
    assert.ok(md.includes('### C'));
  });

  it('absolutizes relative links and keeps text', () => {
    const md = htmlToMarkdown(
      PAGE('<main><p>See <a href="/pricing/">pricing</a> and <a href="https://ext.example/x">external</a>.</p></main>')
    );
    assert.ok(md.includes('[pricing](https://rtylr.com/pricing/)'));
    assert.ok(md.includes('[external](https://ext.example/x)'));
  });

  it('drops pure-decoration links (# and empty)', () => {
    const md = htmlToMarkdown(PAGE('<main><a href="#"><span>CTA</span></a></main>'));
    assert.ok(md.includes('CTA'));
    assert.ok(!md.includes('](#'));
  });

  it('renders emphasis, code and buttons-as-text', () => {
    const md = htmlToMarkdown(
      PAGE('<main><p><strong>fast</strong>, <em>simple</em>, <code>npm i</code></p><button>Book a demo</button></main>')
    );
    assert.ok(md.includes('**fast**'));
    assert.ok(md.includes('*simple*'));
    assert.ok(md.includes('`npm i`'));
    assert.ok(md.includes('Book a demo'));
  });

  it('renders lists (ul/ol, nested)', () => {
    const md = htmlToMarkdown(
      PAGE(`<main><ul><li>POS</li><li>Retail<ul><li>Scanner</li></ul></li></ul><ol><li>First</li><li>Second</li></ol></main>`)
    );
    assert.ok(md.includes('- POS'));
    assert.ok(md.match(/- Retail\n\s+- Scanner/));
    assert.ok(md.includes('1. First'));
    assert.ok(md.includes('2. Second'));
  });

  it('renders GFM tables', () => {
    const md = htmlToMarkdown(
      PAGE(`<main><table><thead><tr><th>Plan</th><th>Price</th></tr></thead><tbody><tr><td>Core</td><td>$49</td></tr></tbody></table></main>`)
    );
    assert.ok(md.includes('| Plan | Price |'));
    assert.ok(md.includes('| --- | --- |'));
    assert.ok(md.includes('| Core | $49 |'));
  });

  it('escapes pipes inside table cells', () => {
    const md = htmlToMarkdown(
      PAGE('<main><table><thead><tr><th>A|B</th></tr></thead><tbody><tr><td>c</td></tr></tbody></table></main>')
    );
    assert.ok(md.includes('| A\\|B |'));
  });

  it('fences code blocks preserving newlines and language hints', () => {
    const md = htmlToMarkdown(
      PAGE('<main><pre><code class="language-js">a()\nb()</code></pre></main>')
    );
    assert.ok(md.includes('```js\na()\nb()\n```'));
  });

  it('renders images with alt text and absolute src', () => {
    const md = htmlToMarkdown(PAGE('<main><figure><img src="/images/pos.png" alt="POS terminal"></figure></main>'));
    assert.ok(md.includes('![POS terminal](https://rtylr.com/images/pos.png)'));
  });

  it('renders blockquotes and horizontal rules', () => {
    const md = htmlToMarkdown(PAGE('<main><blockquote>Trusted by retailers.</blockquote><hr></main>'));
    assert.ok(md.includes('> Trusted by retailers.'));
    assert.ok(md.includes('\n---\n') || md.trimEnd().endsWith('---'));
  });

  it('decodes HTML entities in text and attributes', () => {
    const md = htmlToMarkdown(
      PAGE('<main><p>Tom &amp; Jerry &#39;own&#39; retail&nbsp;ops &lt;today&gt;</p><p><a href="/search?a=1&amp;b=2">q</a></p></main>')
    );
    assert.ok(md.includes("Tom & Jerry 'own' retail ops <today>"));
    assert.ok(md.includes('(https://rtylr.com/search?a=1&b=2)'));
  });

  it('falls back to body when <main> is absent', () => {
    const md = htmlToMarkdown('<html><body><main is="fake"></main><article><h2>Fallback</h2></article></body></html>'.replace('<main is="fake"></main>', ''));
    assert.ok(md.includes('## Fallback'));
  });

  it('produces deterministic output for identical input', () => {
    const input = PAGE('<main><h1>X</h1><p>Same.</p></main>');
    assert.equal(htmlToMarkdown(input), htmlToMarkdown(input));
  });
});

describe('extractPageMeta', () => {
  it('reads title, description and canonical path', () => {
    const meta = extractPageMeta(
      `<html><head>
        <title>Pricing · rtylr</title>
        <meta name="description" content="Plans and per-seat pricing.">
        <link rel="canonical" href="https://rtylr.com/pricing/">
      </head><body><main><h1>Pricing</h1></main></body></html>`,
      '/fallback/'
    );
    assert.equal(meta.title, 'Pricing');
    assert.equal(meta.description, 'Plans and per-seat pricing.');
    assert.equal(meta.path, '/pricing/');
    assert.equal(meta.heading, 'Pricing');
  });

  it('uses the fallback path when no canonical exists', () => {
    const meta = extractPageMeta(PAGE('<main>x</main>'), '/about/');
    assert.equal(meta.path, '/about/');
  });

  it('strips only the site suffix from titles', () => {
    const meta = extractPageMeta(
      '<html><head><title>rtylr — retail OS · rtylr</title></head><body></body></html>',
      '/'
    );
    assert.equal(meta.title, 'rtylr — retail OS');
  });
});

describe('artifact builders', () => {
  const pages = [
    { path: '/', title: 'Home', heading: 'Retail OS', description: 'One platform.' },
    { path: '/pricing/', title: 'Pricing', heading: '', description: '' },
  ];

  it('pages index sorts by path, counts and absolutizes URLs', () => {
    const index = buildPagesIndex([...pages].reverse(), '2026-08-22T00:00:00.000Z');
    assert.equal(index.count, 2);
    assert.deepEqual(index.pages.map((page) => page.url), [
      'https://rtylr.com/',
      'https://rtylr.com/pricing/',
    ]);
  });

  it('llms.txt follows the llmstxt.org shape and advertises machine surfaces', () => {
    const txt = buildLlmsTxt(pages, '2026-08-22T00:00:00.000Z');
    assert.ok(txt.startsWith('# rtylr\n'));
    assert.ok(txt.includes('> One platform.'));
    assert.ok(txt.includes('Accept: text/markdown'));
    assert.ok(txt.includes('/openapi.json'));
    assert.ok(txt.includes('[Pricing](https://rtylr.com/pricing/)'));
  });
});

describe('release versions', () => {
  it('extracts the version from a landing release tag', () => {
    assert.equal(versionFromReleaseTag('rtylr-landing-v0.2.7'), '0.2.7');
  });

  it('rejects unrelated tags', () => {
    assert.equal(versionFromReleaseTag('main'), null);
  });
});

describe('parseHtml tolerance', () => {
  it('handles void tags, comments and self-closing syntax', () => {
    const tree = parseHtml('<div><!-- c --><br><img src="x.png"/><span>s</span></div>');
    const div = tree.children[0];
    assert.equal(div.tag, 'div');
    const tags = div.children.map((child) => child.tag);
    assert.deepEqual(tags.filter((tag) => tag !== '#text'), ['br', 'img', 'span']);
  });
});
