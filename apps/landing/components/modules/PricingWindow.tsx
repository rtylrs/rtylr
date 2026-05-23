import React from 'react';
import { Check } from 'lucide-react';

const PLANS = [
  {
    name: 'Starter',
    price: '$49',
    period: '/mo',
    desc: 'For small teams getting started',
    features: ['POS · vbitez', 'Basic ERP', 'Up to 3 users', 'Email support'],
    highlight: false,
  },
  {
    name: 'Growth',
    price: '$149',
    period: '/mo',
    desc: 'Everything you need to scale',
    features: ['All modules', 'Unlimited users', 'AI features', 'Priority support'],
    highlight: true,
  },
  {
    name: 'Enterprise',
    price: 'Custom',
    period: '',
    desc: 'For large organizations',
    features: ['Custom integrations', 'Dedicated CSM', 'SLA guarantee', 'On-prem option'],
    highlight: false,
  },
];

export default function PricingWindow(): React.JSX.Element {
  return (
    <div
      style={{
        height: '100%',
        background: 'var(--app-background)',
        overflowY: 'auto',
        padding: '24px 20px',
      }}
    >
      <div style={{ textAlign: 'center', marginBottom: 24 }}>
        <div style={{ fontSize: 18, fontWeight: 700, color: '#e8e6e1' }}>Simple pricing</div>
        <div style={{ fontSize: 12, color: 'rgba(255,255,255,0.4)', marginTop: 4 }}>
          No hidden fees. Cancel anytime.
        </div>
      </div>
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 12 }}>
        {PLANS.map((plan) => (
          <div
            key={plan.name}
            style={{
              padding: '20px 16px',
              background: plan.highlight ? 'rgba(232,230,225,0.06)' : 'rgba(255,255,255,0.02)',
              border: `1px solid ${plan.highlight ? 'rgba(232,230,225,0.2)' : 'rgba(255,255,255,0.07)'}`,
              borderRadius: 10,
              display: 'flex',
              flexDirection: 'column',
              gap: 12,
            }}
          >
            <div>
              <div
                style={{
                  fontSize: 12,
                  fontWeight: 600,
                  color: 'rgba(255,255,255,0.5)',
                  marginBottom: 4,
                }}
              >
                {plan.name}
              </div>
              <div style={{ display: 'flex', alignItems: 'baseline', gap: 2 }}>
                <span style={{ fontSize: 24, fontWeight: 700, color: '#e8e6e1' }}>
                  {plan.price}
                </span>
                <span style={{ fontSize: 11, color: 'rgba(255,255,255,0.4)' }}>{plan.period}</span>
              </div>
              <div style={{ fontSize: 11, color: 'rgba(255,255,255,0.35)', marginTop: 4 }}>
                {plan.desc}
              </div>
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
              {plan.features.map((f) => (
                <div
                  key={f}
                  style={{
                    fontSize: 11,
                    color: 'rgba(255,255,255,0.6)',
                    display: 'flex',
                    alignItems: 'center',
                    gap: 6,
                  }}
                >
                  <Check size={11} color="#28c840" strokeWidth={2.5} />
                  {f}
                </div>
              ))}
            </div>
            <button
              style={{
                marginTop: 'auto',
                padding: '8px 0',
                background: plan.highlight ? '#e8e6e1' : 'rgba(255,255,255,0.06)',
                color: plan.highlight ? '#0a0a0a' : 'rgba(255,255,255,0.7)',
                border: 'none',
                borderRadius: 6,
                fontSize: 12,
                fontWeight: 600,
                cursor: 'pointer',
              }}
            >
              {plan.name === 'Enterprise' ? 'Contact us' : 'Get started'}
            </button>
          </div>
        ))}
      </div>
    </div>
  );
}
