import React, { useState, useEffect } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { Search, ChevronDown, ChevronUp, Phone, Mail, Star } from 'lucide-react';
import { crmApi } from '../../services/crmApi';

const ACCENT = '#3b82f6';

type Tier = 'Bronze' | 'Silver' | 'Gold' | 'Platinum';

interface Customer {
  id: number;
  name: string;
  email: string;
  phone: string;
  totalSpent: number;
  visits: number;
  loyaltyPoints: number;
  lastVisit: string;
  initials: string;
}

const mockCustomers: Customer[] = [
  { id: 1, name: 'Sarah Johnson',    email: 'sarah.j@email.com',    phone: '+1 555-0101', totalSpent: 4820, visits: 34, loyaltyPoints: 4820, lastVisit: '2 days ago',  initials: 'SJ' },
  { id: 2, name: 'Michael Chen',     email: 'm.chen@email.com',     phone: '+1 555-0182', totalSpent: 3640, visits: 28, loyaltyPoints: 3640, lastVisit: '5 days ago',  initials: 'MC' },
  { id: 3, name: 'Emily Rodriguez',  email: 'e.rodriguez@email.com', phone: '+1 555-0247', totalSpent: 2910, visits: 22, loyaltyPoints: 2910, lastVisit: '1 week ago',  initials: 'ER' },
  { id: 4, name: 'James Wilson',     email: 'j.wilson@email.com',   phone: '+1 555-0391', totalSpent: 1870, visits: 17, loyaltyPoints: 870,  lastVisit: '2 weeks ago', initials: 'JW' },
  { id: 5, name: 'Aisha Patel',      email: 'a.patel@email.com',    phone: '+1 555-0412', totalSpent: 1420, visits: 14, loyaltyPoints: 620,  lastVisit: '3 weeks ago', initials: 'AP' },
  { id: 6, name: 'David Kim',        email: 'd.kim@email.com',      phone: '+1 555-0563', totalSpent: 980,  visits: 9,  loyaltyPoints: 320,  lastVisit: '1 month ago', initials: 'DK' },
  { id: 7, name: 'Olivia Martinez',  email: 'o.martinez@email.com', phone: '+1 555-0674', totalSpent: 540,  visits: 5,  loyaltyPoints: 180,  lastVisit: '6 weeks ago', initials: 'OM' },
  { id: 8, name: 'Noah Thompson',    email: 'n.thompson@email.com', phone: '+1 555-0788', totalSpent: 210,  visits: 2,  loyaltyPoints: 60,   lastVisit: '2 months ago', initials: 'NT' },
];

function getTier(points: number): Tier {
  if (points >= 2500) return 'Platinum';
  if (points >= 1000) return 'Gold';
  if (points >= 500)  return 'Silver';
  return 'Bronze';
}

const tierColor: Record<Tier, string> = {
  Platinum: '#e879f9',
  Gold:     '#f59e0b',
  Silver:   '#94a3b8',
  Bronze:   '#b45309',
};

const tierFilters: Array<Tier | 'All'> = ['All', 'Platinum', 'Gold', 'Silver', 'Bronze'];

function TierBadge({ tier }: { tier: Tier }) {
  return (
    <span
      style={{
        display: 'inline-flex',
        alignItems: 'center',
        gap: '4px',
        padding: '2px 8px',
        borderRadius: '9999px',
        fontSize: '10px',
        fontWeight: 600,
        background: `${tierColor[tier]}18`,
        color: tierColor[tier],
        whiteSpace: 'nowrap',
      }}
    >
      <Star size={9} />
      {tier}
    </span>
  );
}

function ExpandedDetail({ customer }: { customer: Customer }) {
  const tier = getTier(customer.loyaltyPoints);
  return (
    <motion.tr
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      exit={{ opacity: 0 }}
      transition={{ duration: 0.15 }}
    >
      <td colSpan={8} style={{ padding: 0 }}>
        <div
          style={{
            padding: '20px 24px',
            background: 'rgba(59,130,246,0.04)',
            borderTop: '1px solid var(--hub-border)',
            borderBottom: '1px solid var(--hub-border)',
            display: 'grid',
            gridTemplateColumns: 'repeat(3, 1fr)',
            gap: '20px',
          }}
        >
          <div>
            <p style={{ fontSize: '11px', color: 'var(--hub-text-dim)', marginBottom: '4px' }}>Contact Info</p>
            <div style={{ display: 'flex', alignItems: 'center', gap: '6px', marginBottom: '6px' }}>
              <Mail size={12} style={{ color: ACCENT }} />
              <span style={{ fontSize: '13px', color: 'var(--hub-text)' }}>{customer.email}</span>
            </div>
            <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
              <Phone size={12} style={{ color: ACCENT }} />
              <span style={{ fontSize: '13px', color: 'var(--hub-text)' }}>{customer.phone}</span>
            </div>
          </div>
          <div>
            <p style={{ fontSize: '11px', color: 'var(--hub-text-dim)', marginBottom: '4px' }}>Loyalty Status</p>
            <TierBadge tier={tier} />
            <p style={{ fontSize: '13px', color: 'var(--hub-text)', marginTop: '6px' }}>
              {customer.loyaltyPoints.toLocaleString()} points
            </p>
            <p style={{ fontSize: '11px', color: 'var(--hub-text-muted)' }}>
              {tier !== 'Platinum'
                ? `${(tier === 'Gold' ? 2500 : tier === 'Silver' ? 1000 : 500) - customer.loyaltyPoints} pts to next tier`
                : 'Maximum tier reached'}
            </p>
          </div>
          <div>
            <p style={{ fontSize: '11px', color: 'var(--hub-text-dim)', marginBottom: '4px' }}>Summary</p>
            <p style={{ fontSize: '13px', color: 'var(--hub-text)' }}>${customer.totalSpent.toLocaleString()} lifetime spend</p>
            <p style={{ fontSize: '11px', color: 'var(--hub-text-muted)' }}>{customer.visits} total visits</p>
            <p style={{ fontSize: '11px', color: 'var(--hub-text-muted)', marginTop: '2px' }}>Last seen {customer.lastVisit}</p>
          </div>
        </div>
      </td>
    </motion.tr>
  );
}

const Customers: React.FC = () => {
  const [customers, setCustomers] = useState<Customer[]>(mockCustomers);
  const [search, setSearch] = useState('');
  const [tierFilter, setTierFilter] = useState<Tier | 'All'>('All');
  const [expandedId, setExpandedId] = useState<number | null>(null);

  useEffect(() => {
    crmApi.getCustomers().then((data) => {
      if (data?.length) setCustomers(data);
    }).catch(() => {/* keep mock data */});
  }, []);

  const filtered = customers.filter((c) => {
    const matchSearch =
      c.name.toLowerCase().includes(search.toLowerCase()) ||
      c.email.toLowerCase().includes(search.toLowerCase());
    const matchTier = tierFilter === 'All' || getTier(c.loyaltyPoints) === tierFilter;
    return matchSearch && matchTier;
  });

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '20px', maxWidth: '1280px' }}>
      {/* Header */}
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <div>
          <h2 style={{ fontSize: '18px', fontWeight: 700, color: 'var(--hub-text)', margin: 0 }}>Customers</h2>
          <p style={{ fontSize: '13px', color: 'var(--hub-text-muted)', margin: '2px 0 0' }}>{customers.length} total customers</p>
        </div>
        <button
          style={{
            padding: '8px 16px',
            borderRadius: '8px',
            fontSize: '13px',
            fontWeight: 600,
            background: ACCENT,
            color: '#fff',
            border: 'none',
            cursor: 'pointer',
          }}
        >
          Export CSV
        </button>
      </div>

      {/* Filters */}
      <div style={{ display: 'flex', alignItems: 'center', gap: '12px', flexWrap: 'wrap' }}>
        {/* Search */}
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: '8px',
            padding: '0 12px',
            height: '36px',
            borderRadius: '8px',
            background: 'var(--hub-input-bg)',
            border: '1px solid var(--hub-border)',
            flex: '1',
            minWidth: '220px',
            maxWidth: '340px',
          }}
        >
          <Search size={14} style={{ color: 'var(--hub-text-muted)', flexShrink: 0 }} />
          <input
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Search customers..."
            style={{
              flex: 1,
              background: 'transparent',
              border: 'none',
              outline: 'none',
              fontSize: '13px',
              color: 'var(--hub-text)',
            }}
          />
        </div>

        {/* Tier filter tabs */}
        <div style={{ display: 'flex', gap: '4px', background: 'var(--hub-card)', border: '1px solid var(--hub-border)', borderRadius: '8px', padding: '3px' }}>
          {tierFilters.map((t) => (
            <button
              key={t}
              onClick={() => setTierFilter(t)}
              style={{
                padding: '4px 12px',
                borderRadius: '6px',
                fontSize: '12px',
                fontWeight: 500,
                cursor: 'pointer',
                border: 'none',
                background: tierFilter === t ? (t === 'All' ? ACCENT : tierColor[t as Tier]) : 'transparent',
                color: tierFilter === t ? '#fff' : 'var(--hub-text-muted)',
                transition: 'all 0.15s',
              }}
            >
              {t}
            </button>
          ))}
        </div>
      </div>

      {/* Table */}
      <motion.div
        initial={{ opacity: 0, y: 8 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.2, ease: 'easeOut' }}
        style={{
          background: 'var(--hub-card)',
          border: '1px solid var(--hub-border)',
          borderRadius: '12px',
          overflow: 'hidden',
        }}
      >
        <table style={{ width: '100%', borderCollapse: 'collapse' }}>
          <thead>
            <tr style={{ borderBottom: '1px solid var(--hub-border)' }}>
              {['Customer', 'Email', 'Phone', 'Total Spent', 'Visits', 'Loyalty Pts', 'Last Visit', ''].map((h) => (
                <th
                  key={h}
                  style={{
                    padding: '11px 16px',
                    textAlign: 'left',
                    fontSize: '11px',
                    fontWeight: 600,
                    textTransform: 'uppercase',
                    letterSpacing: '0.05em',
                    color: 'var(--hub-text-dim)',
                    whiteSpace: 'nowrap',
                  }}
                >
                  {h}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            <AnimatePresence>
              {filtered.map((c) => {
                const tier = getTier(c.loyaltyPoints);
                const isExpanded = expandedId === c.id;
                return (
                  <React.Fragment key={c.id}>
                    <tr
                      onClick={() => setExpandedId(isExpanded ? null : c.id)}
                      style={{
                        borderBottom: isExpanded ? 'none' : '1px solid var(--hub-border)',
                        cursor: 'pointer',
                        transition: 'background 0.12s',
                        background: isExpanded ? 'rgba(59,130,246,0.04)' : 'transparent',
                      }}
                      onMouseEnter={(e) => {
                        if (!isExpanded) (e.currentTarget as HTMLElement).style.background = 'var(--hub-hover)';
                      }}
                      onMouseLeave={(e) => {
                        if (!isExpanded) (e.currentTarget as HTMLElement).style.background = 'transparent';
                      }}
                    >
                      <td style={{ padding: '12px 16px' }}>
                        <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
                          <div
                            style={{
                              width: '32px',
                              height: '32px',
                              borderRadius: '8px',
                              background: 'rgba(59,130,246,0.12)',
                              color: ACCENT,
                              display: 'flex',
                              alignItems: 'center',
                              justifyContent: 'center',
                              fontSize: '11px',
                              fontWeight: 700,
                              flexShrink: 0,
                            }}
                          >
                            {c.initials}
                          </div>
                          <div>
                            <div style={{ fontSize: '13px', fontWeight: 500, color: 'var(--hub-text)' }}>{c.name}</div>
                            <TierBadge tier={tier} />
                          </div>
                        </div>
                      </td>
                      <td style={{ padding: '12px 16px', fontSize: '12px', color: 'var(--hub-text-muted)' }}>{c.email}</td>
                      <td style={{ padding: '12px 16px', fontSize: '12px', color: 'var(--hub-text-muted)' }}>{c.phone}</td>
                      <td style={{ padding: '12px 16px', fontSize: '13px', fontWeight: 600, color: 'var(--hub-text)' }}>
                        ${c.totalSpent.toLocaleString()}
                      </td>
                      <td style={{ padding: '12px 16px', fontSize: '13px', color: 'var(--hub-text)' }}>{c.visits}</td>
                      <td style={{ padding: '12px 16px' }}>
                        <div style={{ display: 'flex', alignItems: 'center', gap: '4px' }}>
                          <span style={{ fontSize: '13px', color: 'var(--hub-text)', fontWeight: 500 }}>
                            {c.loyaltyPoints.toLocaleString()}
                          </span>
                          <span style={{ fontSize: '11px', color: 'var(--hub-text-muted)' }}>pts</span>
                        </div>
                      </td>
                      <td style={{ padding: '12px 16px', fontSize: '12px', color: 'var(--hub-text-muted)' }}>{c.lastVisit}</td>
                      <td style={{ padding: '12px 16px', textAlign: 'right' }}>
                        {isExpanded
                          ? <ChevronUp size={14} style={{ color: ACCENT }} />
                          : <ChevronDown size={14} style={{ color: 'var(--hub-text-dim)' }} />}
                      </td>
                    </tr>
                    <AnimatePresence>
                      {isExpanded && <ExpandedDetail customer={c} />}
                    </AnimatePresence>
                  </React.Fragment>
                );
              })}
            </AnimatePresence>
            {filtered.length === 0 && (
              <tr>
                <td colSpan={8} style={{ padding: '40px', textAlign: 'center', color: 'var(--hub-text-muted)', fontSize: '13px' }}>
                  No customers match your search.
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </motion.div>
    </div>
  );
};

export default Customers;
