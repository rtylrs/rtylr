/**
 * Allowlist HTML sanitizer for business-supplied markup.
 *
 * Recruit careers pages are anonymous and unauthenticated, and any member of
 * a business can edit the customization row, so `header_html` is untrusted
 * input rendered to every visitor. This is a hand-rolled allowlist: it is
 * weaker than DOMPurify (no mutation-XSS round-trip guard beyond escaping,
 * no namespace checks beyond dropping svg/math) and should be replaced by
 * DOMPurify if the workspace ever gains that dependency.
 *
 * The shape follows DOMPurify: parse with the browser's own parser so entity
 * and nesting quirks are handled by the spec implementation, then re-serialise
 * only allowlisted nodes and attributes. Raw-text and foreign-namespace
 * elements are dropped with their subtree, which is what keeps the output
 * stable when the browser re-parses it through innerHTML.
 */

const ALLOWED_TAGS = new Set([
  'a',
  'abbr',
  'b',
  'blockquote',
  'br',
  'caption',
  'cite',
  'code',
  'col',
  'colgroup',
  'dd',
  'del',
  'div',
  'dl',
  'dt',
  'em',
  'figcaption',
  'figure',
  'h1',
  'h2',
  'h3',
  'h4',
  'h5',
  'h6',
  'hr',
  'i',
  'img',
  'ins',
  'kbd',
  'li',
  'mark',
  'nav',
  'ol',
  'p',
  'pre',
  'q',
  's',
  'samp',
  'section',
  'small',
  'span',
  'strong',
  'sub',
  'sup',
  'table',
  'tbody',
  'td',
  'tfoot',
  'th',
  'thead',
  'tr',
  'u',
  'ul',
  'var',
  'wbr',
]);

const VOID_TAGS = new Set(['br', 'col', 'hr', 'img', 'wbr']);

const ALLOWED_ATTRS = new Set([
  'alt',
  'class',
  'colspan',
  'dir',
  'height',
  'href',
  'lang',
  'rel',
  'rowspan',
  'src',
  'target',
  'title',
  'width',
]);

const URL_ATTRS = new Set(['href', 'src']);

const SAFE_PROTOCOLS = new Set(['http:', 'https:', 'mailto:', 'tel:']);

/**
 * Elements removed together with everything inside them. Unwrapping them
 * instead would either surface their source as visible text (`<script>`) or
 * re-enter the parser in a different context: these are the elements whose
 * content the HTML parser treats as raw text, plus the foreign-namespace and
 * embedded-content elements that allow SVG/MathML parser confusion.
 */
const DROP_SUBTREE = new Set([
  'annotation-xml',
  'audio',
  'base',
  'body',
  'button',
  'desc',
  'embed',
  'fieldset',
  'form',
  'foreignobject',
  'frame',
  'frameset',
  'head',
  'html',
  'iframe',
  'input',
  'link',
  'math',
  'meta',
  'mi',
  'mn',
  'mo',
  'ms',
  'mtext',
  'noembed',
  'noframes',
  'noscript',
  'object',
  'option',
  'plaintext',
  'script',
  'select',
  'style',
  'svg',
  'template',
  'textarea',
  'title',
  'video',
  'xmp',
]);

/**
 * Returns the URL when it is safe to keep as a link target, else null.
 *
 * The probe strips ASCII control characters and whitespace first because
 * browsers do the same when resolving a URL: `java\tscript:alert(1)` and
 * `java&Tab;script:...` both reach the scheme parser as `javascript:`. The
 * parser has already decoded entities by the time we see the value, so
 * `java&#115;cript:` is caught here too.
 *
 * The scheme is matched case-insensitively and compared lowercased, because
 * URL schemes are case-insensitive: `JaVaScRiPt:` is `javascript:` to the
 * browser. Matching case-sensitively lets a mixed-case scheme fall through to
 * the relative-URL branch and survive as a live href.
 */
function safeUrl(raw: string): string | null {
  const value = raw.trim();
  const probe = Array.from(value)
    .filter((ch) => ch.charCodeAt(0) > 0x20)
    .join('');
  if (!probe) return null;

  const scheme = probe.match(/^([a-z][a-z0-9+.-]*):/i);
  if (!scheme) {
    // Protocol-relative URLs are absolute in disguise; they are not needed
    // for a careers header and are easier to reason about when rejected.
    return probe.startsWith('//') ? null : value;
  }
  return SAFE_PROTOCOLS.has(`${scheme[1].toLowerCase()}:`) ? value : null;
}

function escapeText(value: string): string {
  return value.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
}

function escapeAttr(value: string): string {
  return escapeText(value).replace(/"/g, '&quot;');
}

function serialize(node: Node, out: string[]): void {
  if (node.nodeType === 3) {
    out.push(escapeText(node.nodeValue ?? ''));
    return;
  }
  // Comments, doctypes, CDATA and processing instructions carry no display
  // value but can hide conditional-comment payloads. Dropped outright.
  if (node.nodeType !== 1) return;

  const el = node as Element;
  const tag = el.localName.toLowerCase();
  if (DROP_SUBTREE.has(tag)) return;

  if (!ALLOWED_TAGS.has(tag)) {
    // Unknown but harmless wrapper (`<font>`, a stray `<center>`): keep its
    // text so businesses do not silently lose copy.
    for (const child of Array.from(node.childNodes)) serialize(child, out);
    return;
  }

  out.push(`<${tag}`);
  const target = el.getAttribute('target');
  for (const name of el.getAttributeNames()) {
    const attr = name.toLowerCase();
    if (!ALLOWED_ATTRS.has(attr)) continue;
    // `rel` and `target` are decided after the loop so an attacker-authored
    // value cannot win over the noopener we force onto _blank links, and so
    // neither attribute is emitted twice.
    if (attr === 'rel' || attr === 'target') continue;
    const value = el.getAttribute(name) ?? '';
    if (URL_ATTRS.has(attr)) {
      const url = safeUrl(value);
      if (url) out.push(` ${attr}="${escapeAttr(url)}"`);
      continue;
    }
    if (value) out.push(` ${attr}="${escapeAttr(value)}"`);
  }
  if (target === '_blank' || target === '_self') out.push(` target="${target}"`);
  if (target === '_blank') {
    out.push(' rel="noopener noreferrer"');
  } else {
    const rel = el.getAttribute('rel');
    if (rel && /^[a-zA-Z\s-]*$/.test(rel)) out.push(` rel="${escapeAttr(rel)}"`);
  }
  out.push('>');

  if (VOID_TAGS.has(tag)) return;
  for (const child of Array.from(node.childNodes)) serialize(child, out);
  out.push(`</${tag}>`);
}

/**
 * sanitizeHtml returns an allowlisted, re-escaped copy of `html`, or '' when
 * there is nothing safe to render. Fails closed if no HTML parser is
 * available rather than falling back to raw markup.
 */
export function sanitizeHtml(html: string | null | undefined): string {
  if (!html || typeof DOMParser === 'undefined') return '';

  const parsed = new DOMParser().parseFromString(`<body>${html}</body>`, 'text/html');
  const out: string[] = [];
  for (const child of Array.from(parsed.body.childNodes)) serialize(child, out);
  return out.join('');
}
