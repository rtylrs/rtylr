import { describe, expect, it } from 'vitest';

import { sanitizeHtml } from './html';

/**
 * header_html is written by a business member and rendered to anonymous
 * visitors through dangerouslySetInnerHTML. These tests pin the two halves of
 * the contract: hostile markup is neutralised, ordinary formatting survives.
 *
 * Absolute URLs point at the discard port because happy-dom resolves parsed
 * documents and would otherwise start real requests for them.
 */
const HOST = 'http://127.0.0.1:9';
const IMG = `${HOST}/logo.png`;

const FORBIDDEN_TAGS = new Set([
  'script',
  'style',
  'iframe',
  'object',
  'embed',
  'svg',
  'math',
  'form',
  'input',
  'noscript',
  'template',
  'base',
  'link',
  'meta',
]);

/**
 * Re-parses sanitizer output exactly the way innerHTML does, then asserts the
 * tree holds nothing executable. Attribute escaping only counts as correct if
 * the browser cannot rebuild a handler from it.
 */
function audit(html: string): void {
  const doc = new DOMParser().parseFromString(`<body>${html}</body>`, 'text/html');

  const walk = (node: Node): void => {
    if (node.nodeType === 1) {
      const el = node as Element;
      const tag = el.localName.toLowerCase();
      expect(FORBIDDEN_TAGS.has(tag), `<${tag}> survived sanitisation`).toBe(false);
      for (const name of el.getAttributeNames()) {
        const attr = name.toLowerCase();
        expect(attr.startsWith('on'), `handler ${attr} survived sanitisation`).toBe(false);
        if (attr === 'href' || attr === 'src') {
          const value = (el.getAttribute(name) ?? '').replace(/\s+/g, '');
          expect(
            /^(javascript|data|vbscript|blob|file):/i.test(value),
            `unsafe url ${value} survived`
          ).toBe(false);
        }
      }
    }
    for (const child of Array.from(node.childNodes)) walk(child);
  };

  for (const child of Array.from(doc.body.childNodes)) walk(child);
}

describe('sanitizeHtml - hostile input', () => {
  it('drops a script element and its body', () => {
    expect(sanitizeHtml('<script>alert(document.cookie)</script>')).toBe('');
  });

  it('drops event handler attributes', () => {
    expect(sanitizeHtml(`<img src="${IMG}" onerror="alert(1)">`)).toBe(`<img src="${IMG}">`);
  });

  it('drops javascript: hrefs', () => {
    expect(sanitizeHtml('<a href="javascript:alert(1)">click</a>')).toBe('<a>click</a>');
  });

  it('drops entity-encoded javascript: hrefs', () => {
    expect(sanitizeHtml('<a href="java&#115;cript:alert(1)">click</a>')).toBe('<a>click</a>');
  });

  it('drops mixed-case javascript: hrefs, which browsers resolve case-insensitively', () => {
    for (const href of ['JaVaScRiPt:alert(1)', 'JAVASCRIPT:alert(1)', 'jAvAsCrIpT:alert(1)']) {
      expect(sanitizeHtml(`<a href="${href}">click</a>`)).toBe('<a>click</a>');
    }
  });

  it('drops mixed-case schemes for the other dangerous protocols too', () => {
    expect(sanitizeHtml('<a href="DaTa:text/html,<b>x</b>">x</a>')).toBe('<a>x</a>');
    expect(sanitizeHtml('<a href="VBScript:msgbox(1)">x</a>')).toBe('<a>x</a>');
  });

  it('drops the whole svg subtree, including nested script', () => {
    expect(sanitizeHtml('<svg><script>alert(1)</script></svg>')).toBe('');
  });

  it('drops iframes and their sources', () => {
    expect(sanitizeHtml(`<iframe src="${HOST}/frame"></iframe>`)).toBe('');
  });

  it('drops forms and inputs', () => {
    expect(sanitizeHtml(`<form action="${HOST}/steal"><input name="pw"></form>`)).toBe('');
  });

  it('drops style elements and style attributes', () => {
    expect(sanitizeHtml('<style>body{display:none}</style>')).toBe('');
    expect(sanitizeHtml('<p style="background:url(javascript:alert(1))">hi</p>')).toBe('<p>hi</p>');
  });

  it('drops comments, which can hide conditional-comment payloads', () => {
    expect(sanitizeHtml('<!--[if IE]><script>alert(1)</script><![endif]-->')).toBe('');
  });

  it('drops noscript, which the parser fills as markup when scripting is off', () => {
    expect(sanitizeHtml('<noscript><img src="x" onerror="alert(1)"></noscript>')).toBe('');
  });

  it('escapes attribute values so a handler cannot be rebuilt from them', () => {
    const out = sanitizeHtml('<p title="x&quot; onmouseover=&quot;alert(1)">hi</p>');
    audit(out);
    expect(out).toContain('&quot;');
  });

  it('leaves nothing executable for a battery of payloads', () => {
    const payloads = [
      '<img src=x onerror=alert(1)>',
      '<a href="javascript:alert(1)">x</a>',
      '<a href="JaVaScRiPt:alert(1)">x</a>',
      '<a href="jAvAsCrIpT:alert(1)">x</a>',
      '<a href="jav&#x09;ascript:alert(1)">x</a>',
      '<a href=" javascript:alert(1)">x</a>',
      '<a href="data:text/html;base64,PHNjcmlwdD5hbGVydCgxKTwvc2NyaXB0Pg==">x</a>',
      '<div><span onmouseover="alert(1)">x</span></div>',
      '<math><mtext><script>alert(1)</script></mtext></math>',
      '<template><script>alert(1)</script></template>',
      `<base href="${HOST}/">`,
      `<meta http-equiv="refresh" content="0;url=${HOST}">`,
      `<object data="${HOST}/x.swf"></object>`,
      '<a href="//evil.test">x</a>',
      '<p title="</p><img src=x onerror=alert(1)>">x</p>',
    ];
    for (const payload of payloads) {
      audit(sanitizeHtml(payload));
    }
  });
});

describe('sanitizeHtml - benign input', () => {
  it('keeps formatting tags untouched', () => {
    const html =
      '<nav><ul><li><strong>Open roles</strong> - <em>apply now</em></li></ul></nav>' +
      '<p>Come build with us.</p>';
    expect(sanitizeHtml(html)).toBe(html);
  });

  it('keeps links and images with safe urls', () => {
    expect(sanitizeHtml(`<a href="${HOST}/careers">Careers</a>`)).toBe(
      `<a href="${HOST}/careers">Careers</a>`
    );
    expect(sanitizeHtml('<a href="/about">About</a>')).toBe('<a href="/about">About</a>');
    expect(sanitizeHtml('<a href="mailto:jobs@acme.test">Mail</a>')).toBe(
      '<a href="mailto:jobs@acme.test">Mail</a>'
    );
    expect(sanitizeHtml(`<img src="${IMG}" alt="Acme">`)).toBe(`<img src="${IMG}" alt="Acme">`);
  });

  it('keeps table markup', () => {
    expect(sanitizeHtml('<table><tbody><tr><td>Remote</td></tr></tbody></table>')).toBe(
      '<table><tbody><tr><td>Remote</td></tr></tbody></table>'
    );
  });

  it('forces noopener on target=_blank links', () => {
    expect(sanitizeHtml(`<a href="${HOST}" target="_blank">Home</a>`)).toBe(
      `<a href="${HOST}" target="_blank" rel="noopener noreferrer">Home</a>`
    );
  });

  it('does not let an authored rel override the forced one', () => {
    expect(sanitizeHtml(`<a href="${HOST}" target="_blank" rel="opener">Home</a>`)).toBe(
      `<a href="${HOST}" target="_blank" rel="noopener noreferrer">Home</a>`
    );
  });

  it('unwraps unknown wrappers but keeps their text', () => {
    expect(sanitizeHtml('<font color="red">We hire</font>')).toBe('We hire');
  });

  it('returns an empty string for empty input', () => {
    expect(sanitizeHtml('')).toBe('');
    expect(sanitizeHtml(null)).toBe('');
    expect(sanitizeHtml(undefined)).toBe('');
  });
});
