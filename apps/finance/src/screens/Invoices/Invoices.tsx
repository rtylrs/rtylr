import React, { useEffect, useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { Plus, Send, CheckCircle, Eye, X, FileText } from 'lucide-react';
import { financeApi, Invoice } from '../../services/financeApi';

const ACCENT = '#10b981';

type InvoiceStatus = 'Draft' | 'Sent' | 'Paid' | 'Overdue';

const tabs: (InvoiceStatus | 'All')[] = ['All', 'Draft', 'Sent', 'Paid', 'Overdue'];

const statusStyle: Record<InvoiceStatus, React.CSSProperties> = {
  Draft:   { background: 'rgba(113,113,122,0.15)', color: '#71717a' },
  Sent:    { background: 'rgba(59,130,246,0.12)',  color: '#3b82f6' },
  Paid:    { background: 'rgba(16,185,129,0.12)',  color: '#10b981' },
  Overdue: { background: 'rgba(239,68,68,0.12)',   color: '#ef4444' },
};

const fmt = (n: number) =>
  n.toLocaleString('en-US', { style: 'currency', currency: 'USD', minimumFractionDigits: 0 });

const cardStyle: React.CSSProperties = {
  background: 'var(--hub-card)',
  border: '1px solid var(--hub-border)',
  borderRadius: 12,
};

// ── New Invoice Modal ──────────────────────────────────────────
interface ModalProps { onClose: () => void }

const NewInvoiceModal: React.FC<ModalProps> = ({ onClose }) => {
  return (
    <div
      style={{
        position: 'fixed', inset: 0, zIndex: 50,
        background: 'rgba(0,0,0,0.6)',
        display: 'flex', alignItems: 'center', justifyContent: 'center',
        backdropFilter: 'blur(4px)',
      }}
      onClick={onClose}
    >
      <motion.div
        initial={{ opacity: 0, scale: 0.95, y: 12 }}
        animate={{ opacity: 1, scale: 1, y: 0 }}
        exit={{ opacity: 0, scale: 0.95, y: 8 }}
        transition={{ duration: 0.2 }}
        onClick={(e) => e.stopPropagation()}
        style={{
          background: 'var(--hub-surface)',
          border: '1px solid var(--hub-border)',
          borderRadius: 16,
          padding: 28,
          width: 480,
          maxWidth: '90vw',
        }}
      >
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 20 }}>
          <h2 style={{ margin: 0, fontSize: 16, fontWeight: 700, color: 'var(--hub-text)' }}>New Invoice</h2>
          <button onClick={onClose} style={{ background: 'none', border: 'none', cursor: 'pointer', color: 'var(--hub-text-muted)', padding: 4 }}>
            <X size={18} />
          </button>
        </div>

        <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
          {[
            { label: 'Client Name', placeholder: 'e.g. Acme Corporation', type: 'text' },
            { label: 'Invoice #',   placeholder: 'INV-019',               type: 'text' },
            { label: 'Amount ($)',  placeholder: '0.00',                  type: 'number' },
            { label: 'Due Date',    placeholder: '',                      type: 'date' },
          ].map(({ label, placeholder, type }) => (
            <div key={label} style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
              <label style={{ fontSize: 12, fontWeight: 600, color: 'var(--hub-text-muted)', letterSpacing: '0.03em' }}>{label}</label>
              <input
                type={type}
                placeholder={placeholder}
                style={{
                  padding: '8px 12px',
                  borderRadius: 8,
                  border: '1px solid var(--hub-border)',
                  background: 'var(--hub-input-bg)',
                  color: 'var(--hub-text)',
                  fontSize: 13,
                  outline: 'none',
                  width: '100%',
                  boxSizing: 'border-box',
                }}
              />
            </div>
          ))}
          <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
            <label style={{ fontSize: 12, fontWeight: 600, color: 'var(--hub-text-muted)', letterSpacing: '0.03em' }}>Notes</label>
            <textarea
              rows={3}
              placeholder="Optional notes for the client…"
              style={{
                padding: '8px 12px',
                borderRadius: 8,
                border: '1px solid var(--hub-border)',
                background: 'var(--hub-input-bg)',
                color: 'var(--hub-text)',
                fontSize: 13,
                outline: 'none',
                width: '100%',
                boxSizing: 'border-box',
                resize: 'vertical',
                fontFamily: 'inherit',
              }}
            />
          </div>
        </div>

        <div style={{ display: 'flex', gap: 10, marginTop: 24, justifyContent: 'flex-end' }}>
          <button
            onClick={onClose}
            style={{ padding: '8px 18px', borderRadius: 8, border: '1px solid var(--hub-border)', background: 'var(--hub-hover)', color: 'var(--hub-text-muted)', fontSize: 13, cursor: 'pointer', fontWeight: 500 }}
          >
            Cancel
          </button>
          <button
            onClick={onClose}
            style={{ padding: '8px 18px', borderRadius: 8, border: 'none', background: ACCENT, color: '#fff', fontSize: 13, cursor: 'pointer', fontWeight: 600 }}
          >
            Create Invoice
          </button>
        </div>
      </motion.div>
    </div>
  );
};

// ── Main component ────────────────────────────────────────────
const Invoices: React.FC = () => {
  const [activeTab, setActiveTab] = useState<InvoiceStatus | 'All'>('All');
  const [showModal, setShowModal] = useState(false);
  const [invoices, setInvoices] = useState<Invoice[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    financeApi.getInvoices()
      .then((data) => { if (data?.length) setInvoices(data); })
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  const filtered = activeTab === 'All'
    ? invoices
    : invoices.filter((inv) => inv.status === activeTab);

  const totalFiltered = filtered.reduce((s, inv) => s + inv.amount, 0);

  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={{ duration: 0.25 }}
      style={{ display: 'flex', flexDirection: 'column', gap: 20 }}
    >
      {/* Header row */}
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <div>
          <h2 style={{ margin: 0, fontSize: 18, fontWeight: 700, color: 'var(--hub-text)' }}>Invoices</h2>
          <p style={{ margin: '2px 0 0', fontSize: 13, color: 'var(--hub-text-muted)' }}>
            {filtered.length} invoice{filtered.length !== 1 ? 's' : ''} · {fmt(totalFiltered)} total
          </p>
        </div>
        <button
          onClick={() => setShowModal(true)}
          style={{
            display: 'flex', alignItems: 'center', gap: 6,
            padding: '8px 16px', borderRadius: 8,
            background: ACCENT, color: '#fff',
            border: 'none', fontSize: 13, fontWeight: 600, cursor: 'pointer',
          }}
        >
          <Plus size={15} />
          New Invoice
        </button>
      </div>

      {/* Tabs */}
      <div style={{ display: 'flex', gap: 4, padding: '4px', background: 'var(--hub-surface)', borderRadius: 10, border: '1px solid var(--hub-border)', width: 'fit-content' }}>
        {tabs.map((tab) => (
          <button
            key={tab}
            onClick={() => setActiveTab(tab)}
            style={{
              padding: '5px 14px',
              borderRadius: 7,
              border: 'none',
              fontSize: 12,
              fontWeight: 500,
              cursor: 'pointer',
              background: activeTab === tab ? ACCENT : 'transparent',
              color: activeTab === tab ? '#fff' : 'var(--hub-text-muted)',
              transition: 'all 0.15s',
            }}
          >
            {tab}
          </button>
        ))}
      </div>

      {/* Table */}
      <div style={{ ...cardStyle, overflow: 'hidden' }}>
        <table style={{ width: '100%', borderCollapse: 'collapse' }}>
          <thead>
            <tr style={{ borderBottom: '1px solid var(--hub-border)' }}>
              {['Invoice #', 'Client', 'Amount', 'Status', 'Due Date', 'Actions'].map((h) => (
                <th
                  key={h}
                  style={{
                    padding: '10px 16px',
                    textAlign: 'left',
                    fontSize: 11,
                    fontWeight: 600,
                    textTransform: 'uppercase',
                    letterSpacing: '0.07em',
                    color: 'var(--hub-text-dim)',
                  }}
                >
                  {h}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {filtered.map((inv, i) => (
              <tr
                key={inv.id}
                style={{
                  borderBottom: i < filtered.length - 1 ? '1px solid var(--hub-border)' : 'none',
                  transition: 'background 0.12s',
                }}
                onMouseEnter={(e) => (e.currentTarget.style.background = 'var(--hub-hover)')}
                onMouseLeave={(e) => (e.currentTarget.style.background = 'transparent')}
              >
                <td style={{ padding: '12px 16px', fontSize: 13, fontWeight: 600, color: 'var(--hub-text)' }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                    <FileText size={13} style={{ color: 'var(--hub-text-dim)' }} />
                    {inv.id}
                  </div>
                </td>
                <td style={{ padding: '12px 16px', fontSize: 13, color: 'var(--hub-text)' }}>{inv.client}</td>
                <td style={{ padding: '12px 16px', fontSize: 13, fontWeight: 600, color: 'var(--hub-text)', fontVariantNumeric: 'tabular-nums' }}>
                  {fmt(inv.amount)}
                </td>
                <td style={{ padding: '12px 16px' }}>
                  <span style={{
                    padding: '3px 10px',
                    borderRadius: 20,
                    fontSize: 11,
                    fontWeight: 600,
                    ...statusStyle[inv.status],
                  }}>
                    {inv.status}
                  </span>
                </td>
                <td style={{ padding: '12px 16px', fontSize: 12, color: inv.status === 'Overdue' ? '#ef4444' : 'var(--hub-text-muted)' }}>
                  {inv.dueDate}
                </td>
                <td style={{ padding: '12px 16px' }}>
                  <div style={{ display: 'flex', gap: 6 }}>
                    <button style={actionBtn}>
                      <Eye size={12} />
                      View
                    </button>
                    {inv.status === 'Draft' && (
                      <button style={{ ...actionBtn, color: '#3b82f6', borderColor: 'rgba(59,130,246,0.3)' }}>
                        <Send size={12} />
                        Send
                      </button>
                    )}
                    {(inv.status === 'Sent' || inv.status === 'Overdue') && (
                      <button style={{ ...actionBtn, color: '#10b981', borderColor: 'rgba(16,185,129,0.3)' }}>
                        <CheckCircle size={12} />
                        Mark Paid
                      </button>
                    )}
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Modal */}
      <AnimatePresence>
        {showModal && <NewInvoiceModal onClose={() => setShowModal(false)} />}
      </AnimatePresence>
    </motion.div>
  );
};

const actionBtn: React.CSSProperties = {
  display: 'flex',
  alignItems: 'center',
  gap: 4,
  padding: '4px 10px',
  borderRadius: 6,
  border: '1px solid var(--hub-border)',
  background: 'transparent',
  color: 'var(--hub-text-muted)',
  fontSize: 11,
  fontWeight: 500,
  cursor: 'pointer',
};

export default Invoices;
