import { describe, expect, it, vi } from "vitest";

const sentryMock = vi.hoisted(() => {
  const browserSession = { name: "BrowserSession" };
  const baseSend = vi.fn(() => Promise.resolve({ statusCode: 200 }));
  const baseFlush = vi.fn(() => Promise.resolve(true));
  let browserSessionInstalled = false;

  const client = {
    getIntegrationByName: vi.fn((name: string) =>
      name === browserSession.name && browserSessionInstalled
        ? browserSession
        : undefined,
    ),
  };

  const sdk = {
    init: vi.fn(),
    addIntegration: vi.fn((integration: { name: string }) => {
      if (integration.name === browserSession.name)
        browserSessionInstalled = true;
    }),
    browserSessionIntegration: vi.fn(() => browserSession),
    captureSession: vi.fn(),
    endSession: vi.fn(),
    startSession: vi.fn(),
    getClient: vi.fn(() => client),
    makeFetchTransport: vi.fn(() => ({ send: baseSend, flush: baseFlush })),
    replayIntegration: vi.fn(() => ({ name: "Replay" })),
  };

  return { baseSend, sdk };
});

vi.mock("@sentry/react", () => sentryMock.sdk);

interface TestEnvelope {
  sent_at: string;
}

type TestItem = [{ type: string }, unknown];
type TestTransport = {
  send: (
    envelope: [TestEnvelope, TestItem[]],
  ) => PromiseLike<{ statusCode: number }>;
};

interface InitOptions {
  integrations: (
    integrations: Array<{ name: string }>,
  ) => Array<{ name: string }>;
  tracesSampler: (context: object) => number;
  transport: (options: object) => TestTransport;
  beforeSendTransaction: (event: {
    tags: Record<string, string>;
  }) => { tags: Record<string, string> } | null;
}

describe("Sentry consent lifecycle", () => {
  it("does not send optional telemetry before consent or after withdrawal", async () => {
    const { createSentryOptions } = await import("./sentry");
    let consentGranted = false;
    const options = createSentryOptions(
      { app: "flow" },
      "https://public@example.ingest.sentry.io/1",
      "test",
      () => consentGranted,
    ) as unknown as InitOptions;

    expect(
      options.integrations([
        { name: "BrowserSession" },
        { name: "GlobalHandlers" },
      ]),
    ).toEqual([{ name: "GlobalHandlers" }]);
    expect(options.tracesSampler({})).toBe(0);
    expect(options.beforeSendTransaction({ tags: {} })).toBeNull();

    const transport = options.transport({});
    const sessionEnvelope: [TestEnvelope, TestItem[]] = [
      { sent_at: "2026-09-23T00:00:00.000Z" },
      [[{ type: "session" }, {}]],
    ];
    await transport.send(sessionEnvelope);
    expect(sentryMock.baseSend).not.toHaveBeenCalled();

    consentGranted = true;
    expect(options.tracesSampler({})).toBe(0.1);
    const transaction = { tags: {} };
    expect(options.beforeSendTransaction(transaction)).toBe(transaction);
    await transport.send(sessionEnvelope);
    expect(sentryMock.baseSend).toHaveBeenCalledOnce();

    sentryMock.baseSend.mockClear();
    consentGranted = false;
    expect(options.tracesSampler({})).toBe(0);
    expect(options.beforeSendTransaction({ tags: {} })).toBeNull();
    await transport.send(sessionEnvelope);
    expect(sentryMock.baseSend).not.toHaveBeenCalled();

    await transport.send([
      { sent_at: "2026-09-23T00:00:00.000Z" },
      [[{ type: "event" }, {}]],
    ]);
    expect(sentryMock.baseSend).toHaveBeenCalledOnce();
  });
});
