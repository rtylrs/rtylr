import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { motion } from 'motion/react';
import { Save, ArrowLeft, Sparkles } from 'lucide-react';
import { toast } from 'sonner';
import PageHeader from '../../shared/components/PageHeader';
import ReceiptUpload from '../../shared/components/ReceiptUpload';
import CurrencyInput from '../../shared/components/CurrencyInput';
import { useExpenseCategories } from '../../shared/hooks/useExpenseCategories';
import { cardStyle, ghostBtn, primaryBtn, inputStyle, labelStyle, ACCENT } from '../../shared/lib/tokens';

const NewExpense: React.FC = () => {
  const navigate = useNavigate();
  const { data: categories = [] } = useExpenseCategories();
  const [receipt, setReceipt] = useState<string | null>(null);
  const [description, setDescription] = useState('');
  const [amount, setAmount] = useState(0);
  const [category, setCategory] = useState('');
  const [date, setDate] = useState(new Date().toISOString().slice(0, 10));
  const [notes, setNotes] = useState('');
  const [ocrHint, setOcrHint] = useState<{ merchant?: string; amount?: number; date?: string } | null>(null);

  const applyOcr = () => {
    if (ocrHint?.merchant) setDescription(ocrHint.merchant);
    if (ocrHint?.amount) setAmount(ocrHint.amount);
    if (ocrHint?.date) setDate(ocrHint.date);
    toast.success('Applied OCR hints');
  };

  const handleSave = () => {
    if (!description || !amount || !category) {
      toast.error('Description, amount and category are required');
      return;
    }
    toast.success('Expense submitted');
    navigate('/expenses');
  };

  return (
    <motion.div initial={{ opacity: 0, y: 4 }} animate={{ opacity: 1, y: 0 }} style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
      <PageHeader
        title="Submit Expense"
        breadcrumbs={[{ label: 'Expenses', to: '/expenses' }, { label: 'New' }]}
      />

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 20 }}>
        <div style={{ ...cardStyle, padding: 20, display: 'flex', flexDirection: 'column', gap: 14 }}>
          <div>
            <label style={labelStyle}>Description</label>
            <input value={description} onChange={(e) => setDescription(e.target.value)} style={{ ...inputStyle, marginTop: 6 }} placeholder="Client dinner, Uber ride..." />
          </div>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
            <div>
              <label style={labelStyle}>Amount</label>
              <div style={{ marginTop: 6 }}>
                <CurrencyInput value={amount} onChange={setAmount} />
              </div>
            </div>
            <div>
              <label style={labelStyle}>Date</label>
              <input type="date" value={date} onChange={(e) => setDate(e.target.value)} style={{ ...inputStyle, marginTop: 6 }} />
            </div>
          </div>
          <div>
            <label style={labelStyle}>Category</label>
            <select value={category} onChange={(e) => setCategory(e.target.value)} style={{ ...inputStyle, marginTop: 6 }}>
              <option value="">Select category</option>
              {categories.map((c) => (
                <option key={c.id} value={c.name}>
                  {c.name}
                </option>
              ))}
            </select>
          </div>
          <div>
            <label style={labelStyle}>Notes</label>
            <textarea value={notes} onChange={(e) => setNotes(e.target.value)} rows={3} style={{ ...inputStyle, marginTop: 6, resize: 'vertical', fontFamily: 'inherit' }} />
          </div>
        </div>
        <div style={{ ...cardStyle, padding: 20 }}>
          <div style={{ fontSize: 11, color: 'var(--hub-text-dim)', textTransform: 'uppercase', letterSpacing: '0.08em', marginBottom: 12 }}>
            Receipt
          </div>
          <ReceiptUpload value={receipt} onChange={setReceipt} onOcrHint={setOcrHint} />
          {ocrHint && (
            <div
              style={{
                marginTop: 12,
                padding: 12,
                background: 'rgba(16,185,129,0.08)',
                border: `1px solid ${ACCENT}33`,
                borderRadius: 8,
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
              }}
            >
              <div>
                <div style={{ fontSize: 11, color: ACCENT, display: 'flex', alignItems: 'center', gap: 4 }}>
                  <Sparkles size={11} /> OCR extracted
                </div>
                <div style={{ fontSize: 12, color: 'var(--hub-text-muted)' }}>
                  {ocrHint.merchant} <>&middot;</> ${ocrHint.amount} <>&middot;</> {ocrHint.date}
                </div>
              </div>
              <button style={ghostBtn} onClick={applyOcr}>
                Apply
              </button>
            </div>
          )}
        </div>
      </div>

      <div style={{ display: 'flex', justifyContent: 'space-between' }}>
        <button style={ghostBtn} onClick={() => navigate('/expenses')}>
          <ArrowLeft size={13} /> Cancel
        </button>
        <button style={primaryBtn} onClick={handleSave}>
          <Save size={13} /> Submit
        </button>
      </div>
    </motion.div>
  );
};

export default NewExpense;
