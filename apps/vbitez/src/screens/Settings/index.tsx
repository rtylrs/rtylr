import React from 'react';
import { toast } from 'sonner';
import ScreenLayout from '../../layouts/ScreenLayout';
import { useSettingsStore } from '../../stores/settingsStore';

const SettingsScreen: React.FC = () => {
  const s = useSettingsStore();

  const updateTip = (index: number, value: number) => {
    const next = [...s.tipPresets];
    next[index] = value / 100;
    s.setTipPresets(next);
  };

  return (
    <ScreenLayout title="Settings">
      <div className="p-5 max-w-3xl space-y-5">
        <Section title="Tax">
          <label className="text-xs text-pos-muted">Tax rate (%)</label>
          <input
            type="number"
            step="0.01"
            value={(s.taxRate * 100).toFixed(2)}
            onChange={(e) => s.setTaxRate(Number(e.target.value) / 100)}
            className="w-full h-11 px-3 rounded-lg bg-white/8 border border-pos-border text-pos-text outline-none focus:border-white/25"
          />
        </Section>

        <Section title="Tip Presets">
          <div className="grid grid-cols-4 gap-2">
            {s.tipPresets.map((pct, i) => (
              <input
                key={i}
                type="number"
                step="1"
                value={Math.round(pct * 100)}
                onChange={(e) => updateTip(i, Number(e.target.value))}
                className="h-11 px-3 rounded-lg bg-white/8 border border-pos-border text-pos-text outline-none focus:border-white/25 text-center"
              />
            ))}
          </div>
        </Section>

        <Section title="Receipt">
          <label className="text-xs text-pos-muted">Footer</label>
          <textarea
            value={s.receiptFooter}
            onChange={(e) => s.setReceiptFooter(e.target.value)}
            rows={3}
            className="w-full px-3 py-2 rounded-lg bg-white/8 border border-pos-border text-pos-text outline-none focus:border-white/25 resize-none"
          />
        </Section>

        <Section title="Printer">
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="text-xs text-pos-muted">Name</label>
              <input
                value={s.printer.name}
                onChange={(e) => s.setPrinter({ name: e.target.value })}
                className="w-full h-11 px-3 rounded-lg bg-white/8 border border-pos-border text-pos-text outline-none focus:border-white/25"
              />
            </div>
            <div>
              <label className="text-xs text-pos-muted">Copies</label>
              <input
                type="number"
                min={1}
                max={5}
                value={s.printer.copies}
                onChange={(e) => s.setPrinter({ copies: Number(e.target.value) })}
                className="w-full h-11 px-3 rounded-lg bg-white/8 border border-pos-border text-pos-text outline-none focus:border-white/25"
              />
            </div>
            <label className="flex items-center gap-2 col-span-2 text-sm text-pos-text">
              <input
                type="checkbox"
                checked={s.printer.cashDrawerKick}
                onChange={(e) => s.setPrinter({ cashDrawerKick: e.target.checked })}
                className="w-5 h-5"
              />
              Kick cash drawer on print
            </label>
          </div>
        </Section>

        <Section title="Workstation">
          <div className="space-y-3">
            <div>
              <label className="text-xs text-pos-muted">Auto logout after (minutes)</label>
              <input
                type="number"
                min={0}
                value={s.autoLogoutMinutes}
                onChange={(e) => s.setAutoLogoutMinutes(Number(e.target.value))}
                className="w-full h-11 px-3 rounded-lg bg-white/8 border border-pos-border text-pos-text outline-none focus:border-white/25"
              />
            </div>
            <label className="flex items-center gap-2 text-sm text-pos-text">
              <input
                type="checkbox"
                checked={s.speedMode}
                onChange={(e) => s.setSpeedMode(e.target.checked)}
                className="w-5 h-5"
              />
              Speed mode (one tap to add)
            </label>
            <label className="flex items-center gap-2 text-sm text-pos-text">
              <input
                type="checkbox"
                checked={s.barMode}
                onChange={(e) => s.setBarMode(e.target.checked)}
                className="w-5 h-5"
              />
              Bar mode UI
            </label>
            <label className="flex items-center gap-2 text-sm text-pos-text">
              <input
                type="checkbox"
                checked={s.quickAccessSidebar}
                onChange={(e) => s.setQuickAccessSidebar(e.target.checked)}
                className="w-5 h-5"
              />
              Quick access sidebar
            </label>
          </div>
        </Section>

        <button
          onClick={() => {
            s.reset();
            toast.success('Settings reset');
          }}
          className="h-12 px-5 rounded-lg bg-destructive text-white font-semibold"
        >
          Reset to defaults
        </button>
      </div>
    </ScreenLayout>
  );
};

const Section: React.FC<{ title: string; children: React.ReactNode }> = ({ title, children }) => (
  <section className="rounded-xl bg-pos-card border border-pos-border p-4 space-y-3">
    <h2 className="text-sm font-bold text-pos-text">{title}</h2>
    {children}
  </section>
);

export default SettingsScreen;
