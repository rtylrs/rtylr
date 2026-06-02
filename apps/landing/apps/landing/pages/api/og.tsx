import { ImageResponse } from 'next/og';
import type { NextRequest } from 'next/server';

export const runtime = 'edge';

export default function handler(req: NextRequest) {
  const { searchParams } = new URL(req.url);
  const title = searchParams.get('title') ?? 'rtylr';
  const description = searchParams.get('description') ?? 'Your entire business, one platform.';

  return new ImageResponse(
    <div
      style={{
        width: '100%',
        height: '100%',
        display: 'flex',
        flexDirection: 'column',
        background: 'linear-gradient(135deg, #0e0e1e 0%, #101428 100%)',
        fontFamily: 'Inter, system-ui, sans-serif',
        position: 'relative',
        overflow: 'hidden',
      }}
    >
      {/* Subtle radial glow top-right */}
      <div
        style={{
          position: 'absolute',
          top: -120,
          right: -80,
          width: 520,
          height: 520,
          borderRadius: '50%',
          background: 'radial-gradient(circle, rgba(99,102,241,0.18) 0%, transparent 70%)',
        }}
      />
      {/* Subtle radial glow bottom-left */}
      <div
        style={{
          position: 'absolute',
          bottom: -100,
          left: -60,
          width: 400,
          height: 400,
          borderRadius: '50%',
          background: 'radial-gradient(circle, rgba(34,197,94,0.1) 0%, transparent 70%)',
        }}
      />

      {/* Content */}
      <div
        style={{
          display: 'flex',
          flexDirection: 'column',
          justifyContent: 'space-between',
          height: '100%',
          padding: '64px 72px',
        }}
      >
        {/* Wordmark */}
        <div
          style={{
            fontSize: 28,
            fontWeight: 700,
            letterSpacing: '-0.04em',
            color: '#e8e6e1',
          }}
        >
          rtylr
        </div>

        {/* Main copy */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
          <div
            style={{
              fontSize: title.length > 40 ? 48 : 60,
              fontWeight: 800,
              letterSpacing: '-0.03em',
              lineHeight: 1.08,
              color: '#e8e6e1',
              maxWidth: 860,
            }}
          >
            {title}
          </div>
          {description && (
            <div
              style={{
                fontSize: 24,
                fontWeight: 400,
                color: 'rgba(232,230,225,0.55)',
                maxWidth: 720,
                lineHeight: 1.5,
              }}
            >
              {description}
            </div>
          )}
        </div>

        {/* Footer badge */}
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: 10,
          }}
        >
          <div
            style={{
              width: 8,
              height: 8,
              borderRadius: '50%',
              background: '#22c55e',
            }}
          />
          <div
            style={{
              fontSize: 15,
              fontWeight: 500,
              color: 'rgba(232,230,225,0.4)',
              letterSpacing: '0.04em',
            }}
          >
            rtylr.com
          </div>
        </div>
      </div>
    </div>,
    {
      width: 1200,
      height: 630,
    }
  );
}
