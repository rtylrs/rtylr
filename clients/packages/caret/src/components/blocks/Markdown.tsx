import React, { useEffect, useMemo, useRef, useState } from "react";
import { Check, Copy } from "lucide-react";
import ReactMarkdown, {
  type Components,
  type UrlTransform,
} from "react-markdown";
import remarkGfm from "remark-gfm";
import "../../styles/markdown.css";
import ChartBlock, { parseChartSpec, type ChartSpec } from "./ChartBlock";
import {
  isAppPath,
  parsePageSpec,
  useCaretHost,
  type CaretPageSpec,
} from "../../host";
import type { HighlightNode } from "../../lib/highlightCode";

export interface MarkdownProps {
  source: string;
}

function sanitizeMarkdownUrl(url: string): string {
  const trimmed = url.trim();
  if (/^https?:\/\//i.test(trimmed) || trimmed.startsWith("/")) return trimmed;
  return "";
}

const safeUrlTransform: UrlTransform = (url) => sanitizeMarkdownUrl(url);

function tableHasBodyRows(children: React.ReactNode): boolean {
  return React.Children.toArray(children).some((section) => {
    if (!React.isValidElement<{ children?: React.ReactNode }>(section))
      return false;
    if (section.type !== "tbody") return false;
    return React.Children.toArray(section.props.children).some((row) =>
      React.isValidElement(row),
    );
  });
}

/** A ```chart block whose JSON describes a usable chart. */
function chartFromCode(children: React.ReactNode): ChartSpec | null {
  const code = React.Children.toArray(children)[0];
  if (
    !React.isValidElement<{ className?: string; children?: React.ReactNode }>(
      code,
    )
  )
    return null;
  if (!/\blanguage-chart\b/.test(code.props.className ?? "")) return null;
  const raw = React.Children.toArray(code.props.children).join("");
  return parseChartSpec(raw);
}

/** A ```page block naming a page of the app Caret is running in. */
function pageFromCode(children: React.ReactNode): CaretPageSpec | null {
  const code = React.Children.toArray(children)[0];
  if (
    !React.isValidElement<{ className?: string; children?: React.ReactNode }>(
      code,
    )
  )
    return null;
  if (!/\blanguage-page\b/.test(code.props.className ?? "")) return null;
  return parsePageSpec(React.Children.toArray(code.props.children).join(""));
}

const PageBlock: React.FC<{ spec: CaretPageSpec }> = ({ spec }) => {
  const host = useCaretHost();
  if (host.renderPage) return <>{host.renderPage(spec)}</>;
  return (
    <p className="caret-p">
      <a href={spec.path} className="caret-link">
        {spec.title ?? spec.path}
      </a>
    </p>
  );
};

/** In-app links open through the host, so the chat stays where it is. */
const AppLink: React.FC<
  React.AnchorHTMLAttributes<HTMLAnchorElement> & { href: string }
> = ({ href, ...props }) => {
  const { openPage } = useCaretHost();
  return (
    <a
      {...props}
      href={href}
      className="caret-link"
      onClick={(event) => {
        if (
          !openPage ||
          event.metaKey ||
          event.ctrlKey ||
          event.shiftKey ||
          event.button !== 0
        )
          return;
        event.preventDefault();
        openPage({ path: href });
      }}
    />
  );
};

type Highlighter = (code: string, language: string) => HighlightNode[] | null;

// The highlighter and its grammars load on first use, so a chat with no code
// doesn't pay for them. Later blocks reuse the loaded module.
let loadedHighlighter: Highlighter | null = null;
let highlighterLoad: Promise<Highlighter> | null = null;

function loadHighlighter(): Promise<Highlighter> {
  highlighterLoad ??= import("../../lib/highlightCode").then((module) => {
    loadedHighlighter = module.highlightCode;
    return module.highlightCode;
  });
  return highlighterLoad;
}

function useHighlighter(enabled: boolean): Highlighter | null {
  const [highlighter, setHighlighter] = useState<Highlighter | null>(
    () => loadedHighlighter,
  );
  useEffect(() => {
    if (!enabled || highlighter) return;
    let live = true;
    loadHighlighter()
      .then((loaded) => {
        if (live) setHighlighter(() => loaded);
      })
      .catch(() => undefined);
    return (): void => {
      live = false;
    };
  }, [enabled, highlighter]);
  return highlighter;
}

function renderHighlight(nodes: HighlightNode[]): React.ReactNode {
  return nodes.map((node, index) =>
    node.type === "text" ? (
      node.value
    ) : (
      <span key={index} className={node.properties?.className?.join(" ")}>
        {renderHighlight(node.children)}
      </span>
    ),
  );
}

/** Inside a fenced block: tokens are coloured once the highlighter has loaded. */
const CodeContent: React.FC<{
  language: string;
  className?: string;
  children?: React.ReactNode;
}> = ({ language, className, children }) => {
  const text = React.Children.toArray(children).join("");
  const highlighter = useHighlighter(true);
  const highlighted = useMemo(
    () => (highlighter ? highlighter(text, language) : null),
    [highlighter, text, language],
  );
  return (
    <code
      className={["caret-inline-code", className].filter(Boolean).join(" ")}
    >
      {highlighted ? renderHighlight(highlighted) : children}
    </code>
  );
};

/** A fenced block with a copy button that shows on hover or keyboard focus. */
const CodeBlockFrame: React.FC<{
  children?: React.ReactNode;
  preProps: React.HTMLAttributes<HTMLPreElement>;
}> = ({ children, preProps }) => {
  const preRef = useRef<HTMLPreElement>(null);
  const timer = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);
  const [copied, setCopied] = useState(false);

  useEffect(() => (): void => clearTimeout(timer.current), []);

  const copy = async (): Promise<void> => {
    const text = (preRef.current?.textContent ?? "").replace(/\n$/, "");
    try {
      await navigator.clipboard.writeText(text);
    } catch {
      // The Clipboard API can be blocked in iframes or when unfocused.
      return;
    }
    setCopied(true);
    clearTimeout(timer.current);
    timer.current = setTimeout(() => setCopied(false), 1500);
  };

  return (
    <div className={`caret-codeblock-frame${copied ? " is-copied" : ""}`}>
      <button
        type="button"
        className="caret-codeblock-copy"
        aria-label={copied ? "Copied" : "Copy code"}
        title={copied ? "Copied" : "Copy code"}
        onClick={() => void copy()}
      >
        {copied ? (
          <Check size={14} aria-hidden="true" />
        ) : (
          <Copy size={14} aria-hidden="true" />
        )}
        <span aria-live="polite">{copied ? "Copied" : "Copy"}</span>
      </button>
      <pre ref={preRef} className="caret-codeblock" {...preProps}>
        {children}
      </pre>
    </div>
  );
};

const markdownComponents: Components = {
  h1: ({ node: _node, ...props }) => <h1 className="caret-h" {...props} />,
  h2: ({ node: _node, ...props }) => <h2 className="caret-h" {...props} />,
  h3: ({ node: _node, ...props }) => <h3 className="caret-h" {...props} />,
  h4: ({ node: _node, ...props }) => <h4 className="caret-h" {...props} />,
  h5: ({ node: _node, ...props }) => <h5 className="caret-h" {...props} />,
  h6: ({ node: _node, ...props }) => <h6 className="caret-h" {...props} />,
  p: ({ node: _node, ...props }) => <p className="caret-p" {...props} />,
  ul: ({ node: _node, ...props }) => <ul className="caret-list" {...props} />,
  ol: ({ node: _node, ...props }) => <ol className="caret-list" {...props} />,
  blockquote: ({ node: _node, ...props }) => (
    <blockquote className="caret-quote" {...props} />
  ),
  hr: ({ node: _node, ...props }) => <hr className="caret-hr" {...props} />,
  code: ({ node: _node, className, children, ...props }) => {
    const language = /\blanguage-([\w+#-]+)/.exec(className ?? "")?.[1];
    if (language && language !== "chart" && language !== "page") {
      return (
        <CodeContent language={language} className={className}>
          {children}
        </CodeContent>
      );
    }
    return (
      <code
        className={["caret-inline-code", className].filter(Boolean).join(" ")}
        {...props}
      >
        {children}
      </code>
    );
  },
  pre: ({ node: _node, children, ...props }) => {
    const chart = chartFromCode(children);
    if (chart) return <ChartBlock spec={chart} />;
    const page = pageFromCode(children);
    if (page) return <PageBlock spec={page} />;
    return <CodeBlockFrame preProps={props}>{children}</CodeBlockFrame>;
  },
  table: ({ node: _node, children, ...props }) =>
    tableHasBodyRows(children) ? (
      <div className="caret-table-wrap">
        <table className="caret-table" {...props}>
          {children}
        </table>
      </div>
    ) : (
      <p className="caret-table-empty">No matching records.</p>
    ),
  a: ({ node: _node, href, ...props }) => {
    const safeHref = href ? sanitizeMarkdownUrl(href) : "";
    if (!safeHref) return <span>{props.children}</span>;

    const isExternal = /^https?:\/\//i.test(safeHref);
    if (!isExternal && isAppPath(safeHref))
      return <AppLink {...props} href={safeHref} />;
    return (
      <a
        {...props}
        href={safeHref}
        target={isExternal ? "_blank" : undefined}
        rel={isExternal ? "noopener noreferrer" : undefined}
        className="caret-link"
      />
    );
  },
};

/** Render model Markdown with GitHub-flavored tables, lists, and task items. */
export const Markdown: React.FC<MarkdownProps> = ({ source }) => {
  return (
    <div className="caret-md">
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}
        components={markdownComponents}
        urlTransform={safeUrlTransform}
      >
        {source}
      </ReactMarkdown>
    </div>
  );
};

export default Markdown;
