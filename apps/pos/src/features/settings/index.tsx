import React, { useEffect, useState } from 'react';
import { toast } from 'sonner';
import { unwrap } from '@voxire/api';
import ScreenLayout from '../../layouts/ScreenLayout';
import { useSettingsStore } from '../../shared/stores/settingsStore';
import type { POSBusinessProfile, POSFeatureKey } from '../../shared/stores/settingsStore';
import {
  normalizePOSTheme,
  POS_THEME_OPTIONS,
  useThemeStore,
} from '../../shared/stores/themeStore';
import { useCartStore } from '../../shared/stores/cartStore';
import { useTerminalContextStore } from '../../shared/stores/terminalContextStore';
import {
  posService,
  type POSBusinessSettingsResponse,
  type POSTheme,
} from '../../shared/services/api/posService';
import { listPrinters, listSerialPorts } from '../../shared/services/hardware/hardwareDiscovery';
import type { PrinterInfo, SerialPortInfo } from '../../shared/services/hardware/hardwareDiscovery';
import appConfig from '../../app.json';
import { HardwareHealthCheck } from './HardwareHealthCheck';

const INPUT_CLS =
  'w-full h-11 px-3 rounded-lg bg-white/8 border border-pos-border text-pos-text outline-none focus:border-white/25';
const SELECT_CLS =
  'w-full h-11 px-3 rounded-lg bg-white/8 border border-pos-border text-pos-text outline-none focus:border-white/25';

const PROFILE_OPTIONS: { value: POSBusinessProfile; label: string }[] = [
  { value: 'retail', label: 'Retail' },
  { value: 'quick_service', label: 'Quick service' },
  { value: 'restaurant', label: 'Restaurant' },
];

const FEATURE_OPTIONS: { key: POSFeatureKey; label: string }[] = [
  { key: 'tables', label: 'Tables' },
  { key: 'kitchenDisplay', label: 'Kitchen display' },
  { key: 'modifiers', label: 'Modifiers' },
  { key: 'combos', label: 'Combos' },
  { key: 'tips', label: 'Tips' },
  { key: 'giftCards', label: 'Gift cards' },
  { key: 'customers', label: 'Customers' },
  { key: 'inventory', label: 'Inventory' },
  { key: 'cashDrawer', label: 'Cash drawer' },
];

const SettingsScreen: React.FC = () => {
  const s = useSettingsStore();
  const theme = useThemeStore((state) => state.theme);
  const setTheme = useThemeStore((state) => state.setTheme);
  const { locationId, terminalId, clearTerminalContext } = useTerminalContextStore();
  const clearCart = useCartStore((state) => state.clearCart);

  const [printers, setPrinters] = useState<PrinterInfo[]>([]);
  const [serialPorts, setSerialPorts] = useState<SerialPortInfo[]>([]);
  const [discovering, setDiscovering] = useState(false);
  const [savingPOSSettings, setSavingPOSSettings] = useState(false);

  const discover = async () => {
    setDiscovering(true);
    try {
      const [p, sp] = await Promise.all([listPrinters(), listSerialPorts()]);
      setPrinters(p);
      setSerialPorts(sp);
      if (p.length === 0 && sp.length === 0) {
        toast.info('No hardware detected');
      } else {
        toast.success(`Found ${p.length} printer(s), ${sp.length} serial port(s)`);
      }
    } catch {
      toast.error('Hardware discovery failed');
    } finally {
      setDiscovering(false);
    }
  };

  useEffect(() => {
    discover();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const updateTip = (index: number, value: number) => {
    const next = [...s.tipPresets];
    next[index] = value / 100;
    s.setTipPresets(next);
  };

  const savePOSSettings = async () => {
    setSavingPOSSettings(true);
    try {
      const selectedTheme: POSTheme = normalizePOSTheme(theme);
      const response = await unwrap(
        posService.updatePOSSettings({
          business_profile: s.businessProfile,
          theme: selectedTheme,
          features: s.features,
          receipt_footer: s.receiptFooter,
        })
      );
      const settings = (response as POSBusinessSettingsResponse).settings;
      if (settings) {
        s.setBusinessProfile(settings.business_profile);
        s.setFeatures(settings.features);
        s.setReceiptFooter(settings.receipt_footer);
        setTheme(settings.theme);
      }
      toast.success('POS settings saved');
    } catch {
      toast.error('Failed to save POS settings');
    } finally {
      setSavingPOSSettings(false);
    }
  };

  return (
    <ScreenLayout title="Settings">
      <div className="p-5 max-w-3xl space-y-5">
        <Section title="Business">
          <div className="grid grid-cols-1 gap-3">
            <div>
              <label className="text-xs text-pos-muted">Business name (appears on receipt)</label>
              <input
                value={s.business.name}
                onChange={(e) => s.setBusiness({ name: e.target.value })}
                className={INPUT_CLS}
              />
            </div>
            <div>
              <label className="text-xs text-pos-muted">Address</label>
              <input
                value={s.business.address}
                onChange={(e) => s.setBusiness({ address: e.target.value })}
                className={INPUT_CLS}
              />
            </div>
            <div>
              <label className="text-xs text-pos-muted">Phone</label>
              <input
                value={s.business.phone}
                onChange={(e) => s.setBusiness({ phone: e.target.value })}
                className={INPUT_CLS}
              />
            </div>
          </div>
        </Section>

        <Section title="POS Experience">
          <div className="space-y-4">
            <div>
              <label className="text-xs text-pos-muted">Workflow</label>
              <select
                value={s.businessProfile}
                onChange={(e) => s.setBusinessProfile(e.target.value as POSBusinessProfile)}
                className={SELECT_CLS}
              >
                {PROFILE_OPTIONS.map((profile) => (
                  <option key={profile.value} value={profile.value}>
                    {profile.label}
                  </option>
                ))}
              </select>
            </div>
            <div>
              <label className="text-xs text-pos-muted">Default theme</label>
              <select
                value={normalizePOSTheme(theme)}
                onChange={(e) => setTheme(e.target.value)}
                className={SELECT_CLS}
              >
                {POS_THEME_OPTIONS.map((option) => (
                  <option key={option.value} value={option.value}>
                    {option.label}
                  </option>
                ))}
              </select>
            </div>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
              {FEATURE_OPTIONS.map((feature) => (
                <label
                  key={feature.key}
                  className="flex items-center gap-2 rounded-lg bg-white/5 border border-pos-border p-3 text-sm text-pos-text"
                >
                  <input
                    type="checkbox"
                    checked={s.features[feature.key]}
                    onChange={(e) => s.setFeature(feature.key, e.target.checked)}
                    className="w-5 h-5"
                  />
                  {feature.label}
                </label>
              ))}
            </div>
            <button
              type="button"
              onClick={savePOSSettings}
              disabled={savingPOSSettings}
              className="h-10 px-4 rounded-lg bg-pos-text text-pos-bg text-sm font-bold disabled:opacity-50"
            >
              {savingPOSSettings ? 'Saving...' : 'Save Business Settings'}
            </button>
          </div>
        </Section>

        <Section title="Tax">
          <label className="text-xs text-pos-muted">Tax rate (%)</label>
          <input
            type="number"
            step="0.01"
            value={(s.taxRate * 100).toFixed(2)}
            onChange={(e) => s.setTaxRate(Number(e.target.value) / 100)}
            className={INPUT_CLS}
          />
        </Section>

        {s.features.tips && (
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
        )}

        <Section title="Receipt">
          <label className="text-xs text-pos-muted">Footer message</label>
          <textarea
            value={s.receiptFooter}
            onChange={(e) => s.setReceiptFooter(e.target.value)}
            rows={3}
            className="w-full px-3 py-2 rounded-lg bg-white/8 border border-pos-border text-pos-text outline-none focus:border-white/25 resize-none"
          />
        </Section>

        <Section title="Printer & Cash Drawer">
          <div className="flex items-center justify-between">
            <p className="text-xs text-pos-muted">
              {discovering
                ? 'Scanning hardware...'
                : `${printers.length} printer(s) · ${serialPorts.length} serial port(s) detected`}
            </p>
            <button
              type="button"
              onClick={discover}
              disabled={discovering}
              className="h-8 px-3 rounded-lg bg-white/10 text-pos-text text-xs font-semibold disabled:opacity-50"
            >
              Refresh
            </button>
          </div>

          {/* Transport */}
          <div>
            <label className="text-xs text-pos-muted">Transport</label>
            <select
              value={s.printer.transport}
              onChange={(e) =>
                s.setPrinter({ transport: e.target.value as 'windows' | 'serial' | 'network' })
              }
              className={SELECT_CLS}
            >
              <option value="windows">Windows Spooler</option>
              <option value="serial">Serial / USB-Serial</option>
              <option value="network">Network (TCP/IP)</option>
            </select>
          </div>

          {/* Transport-specific fields */}
          {s.printer.transport === 'windows' && (
            <div>
              <label className="text-xs text-pos-muted">Windows printer name</label>
              {printers.length > 0 ? (
                <select
                  value={s.printer.name}
                  onChange={(e) => s.setPrinter({ name: e.target.value })}
                  className={SELECT_CLS}
                >
                  {printers.map((p) => (
                    <option key={p.name} value={p.name}>
                      {p.name}
                      {p.isDefault ? ' (default)' : ''} - {p.connectionType}
                    </option>
                  ))}
                </select>
              ) : (
                <input
                  value={s.printer.name}
                  placeholder="e.g. EPSON TM-T88VI"
                  onChange={(e) => s.setPrinter({ name: e.target.value })}
                  className={INPUT_CLS}
                />
              )}
            </div>
          )}

          {s.printer.transport === 'serial' && (
            <div className="grid grid-cols-2 gap-3">
              <div className="col-span-2">
                <label className="text-xs text-pos-muted">Serial port</label>
                {serialPorts.length > 0 ? (
                  <select
                    value={s.printer.serialPort}
                    onChange={(e) => s.setPrinter({ serialPort: e.target.value })}
                    className={SELECT_CLS}
                  >
                    <option value="">- select port -</option>
                    {serialPorts.map((p) => (
                      <option key={p.portName} value={p.portName}>
                        {p.portName}
                        {p.product ? ` (${p.product})` : ''}
                      </option>
                    ))}
                  </select>
                ) : (
                  <input
                    value={s.printer.serialPort}
                    placeholder="e.g. /dev/tty.usbserial-1420 or COM3"
                    onChange={(e) => s.setPrinter({ serialPort: e.target.value })}
                    className={INPUT_CLS}
                  />
                )}
              </div>
              <div>
                <label className="text-xs text-pos-muted">Baud rate</label>
                <select
                  value={s.printer.serialBaudRate}
                  onChange={(e) => s.setPrinter({ serialBaudRate: Number(e.target.value) })}
                  className={SELECT_CLS}
                >
                  {[9600, 19200, 38400, 57600, 115200].map((b) => (
                    <option key={b} value={b}>
                      {b}
                    </option>
                  ))}
                </select>
              </div>
            </div>
          )}

          {s.printer.transport === 'network' && (
            <div className="grid grid-cols-3 gap-3">
              <div className="col-span-2">
                <label className="text-xs text-pos-muted">Printer host / IP</label>
                <input
                  value={s.printer.networkHost}
                  placeholder="e.g. 192.168.1.100"
                  onChange={(e) => s.setPrinter({ networkHost: e.target.value })}
                  className={INPUT_CLS}
                />
              </div>
              <div>
                <label className="text-xs text-pos-muted">TCP port</label>
                <input
                  type="number"
                  value={s.printer.networkPort}
                  onChange={(e) => s.setPrinter({ networkPort: Number(e.target.value) })}
                  className={INPUT_CLS}
                />
              </div>
            </div>
          )}

          {/* Copies */}
          <div>
            <label className="text-xs text-pos-muted">Copies</label>
            <input
              type="number"
              min={1}
              max={5}
              value={s.printer.copies}
              onChange={(e) => s.setPrinter({ copies: Number(e.target.value) })}
              className={INPUT_CLS}
            />
          </div>

          {s.features.cashDrawer && (
            <div className="space-y-2 pt-1">
              <label className="flex items-center gap-2 text-sm text-pos-text">
                <input
                  type="checkbox"
                  checked={s.printer.cashDrawerKick}
                  onChange={(e) => s.setPrinter({ cashDrawerKick: e.target.checked })}
                  className="w-5 h-5"
                />
                Kick cash drawer on print
              </label>
              {s.printer.cashDrawerKick && (
                <div>
                  <label className="text-xs text-pos-muted">Drawer pin</label>
                  <select
                    value={s.printer.drawerPin}
                    onChange={(e) =>
                      s.setPrinter({ drawerPin: Number(e.target.value) as 2 | 5 })
                    }
                    className={SELECT_CLS}
                  >
                    <option value={2}>Pin 2 (most printers)</option>
                    <option value={5}>Pin 5</option>
                  </select>
                </div>
              )}
            </div>
          )}
        </Section>

        {s.features.kitchenDisplay && (
          <Section title="KDS Hardware Display">
            <p className="text-xs text-pos-muted">
              Connect a dedicated hardware KDS display (VFD, kitchen screen) via serial or network.
              Set to None to use the browser-based KDS only.
            </p>
            <div>
              <label className="text-xs text-pos-muted">Transport</label>
              <select
                value={s.kdsHardware.transport}
                onChange={(e) =>
                  s.setKdsHardware({
                    transport: e.target.value as 'none' | 'serial' | 'network',
                  })
                }
                className={SELECT_CLS}
              >
                <option value="none">None (software KDS only)</option>
                <option value="serial">Serial / USB-Serial</option>
                <option value="network">Network (TCP/IP)</option>
              </select>
            </div>
            {s.kdsHardware.transport === 'serial' && (
              <div className="grid grid-cols-2 gap-3">
                <div className="col-span-2">
                  <label className="text-xs text-pos-muted">Serial port</label>
                  {serialPorts.length > 0 ? (
                    <select
                      value={s.kdsHardware.serialPort}
                      onChange={(e) => s.setKdsHardware({ serialPort: e.target.value })}
                      className={SELECT_CLS}
                    >
                      <option value="">- select port -</option>
                      {serialPorts.map((p) => (
                        <option key={p.portName} value={p.portName}>
                          {p.portName}
                          {p.product ? ` (${p.product})` : ''}
                        </option>
                      ))}
                    </select>
                  ) : (
                    <input
                      value={s.kdsHardware.serialPort}
                      placeholder="e.g. /dev/tty.usbserial-1420"
                      onChange={(e) => s.setKdsHardware({ serialPort: e.target.value })}
                      className={INPUT_CLS}
                    />
                  )}
                </div>
                <div>
                  <label className="text-xs text-pos-muted">Baud rate</label>
                  <select
                    value={s.kdsHardware.serialBaudRate}
                    onChange={(e) => s.setKdsHardware({ serialBaudRate: Number(e.target.value) })}
                    className={SELECT_CLS}
                  >
                    {[9600, 19200, 38400, 115200].map((b) => (
                      <option key={b} value={b}>
                        {b}
                      </option>
                    ))}
                  </select>
                </div>
              </div>
            )}
            {s.kdsHardware.transport === 'network' && (
              <div className="grid grid-cols-3 gap-3">
                <div className="col-span-2">
                  <label className="text-xs text-pos-muted">Host / IP</label>
                  <input
                    value={s.kdsHardware.networkHost}
                    placeholder="e.g. 192.168.1.200"
                    onChange={(e) => s.setKdsHardware({ networkHost: e.target.value })}
                    className={INPUT_CLS}
                  />
                </div>
                <div>
                  <label className="text-xs text-pos-muted">Port</label>
                  <input
                    type="number"
                    value={s.kdsHardware.networkPort}
                    onChange={(e) => s.setKdsHardware({ networkPort: Number(e.target.value) })}
                    className={INPUT_CLS}
                  />
                </div>
              </div>
            )}
          </Section>
        )}

        <Section title="Hardware Status">
          <HardwareHealthCheck
            printer={s.printer}
            kds={s.kdsHardware}
            showCashDrawer={s.features.cashDrawer}
            showKds={s.features.kitchenDisplay}
          />
        </Section>

        <Section title="Workstation">
          <div className="space-y-3">
            <div className="rounded-lg bg-white/5 border border-pos-border p-3 space-y-2">
              <div className="flex items-center justify-between gap-3">
                <div>
                  <div className="text-xs text-pos-muted">Assigned terminal</div>
                  <div className="text-sm text-pos-text">
                    {locationId && terminalId
                      ? `Location ${locationId} / Terminal ${terminalId}`
                      : 'Not assigned'}
                  </div>
                </div>
                <button
                  type="button"
                  disabled={!locationId || !terminalId}
                  onClick={() => {
                    clearCart();
                    clearTerminalContext();
                    toast.success('Terminal assignment cleared');
                  }}
                  className="h-10 px-3 rounded-lg bg-white/10 text-pos-text text-xs font-semibold disabled:opacity-50"
                >
                  Reassign
                </button>
              </div>
            </div>
            <div>
              <label className="text-xs text-pos-muted">Auto logout after (minutes)</label>
              <input
                type="number"
                min={0}
                value={s.autoLogoutMinutes}
                onChange={(e) => s.setAutoLogoutMinutes(Number(e.target.value))}
                className={INPUT_CLS}
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
            toast.success('Settings reset to defaults');
          }}
          className="h-12 px-5 rounded-lg bg-destructive text-white font-semibold"
        >
          Reset to defaults
        </button>

        <Section title="About">
          {[
            { label: 'Application', value: appConfig.name },
            { label: 'Version', value: `v${appConfig.version}` },
          ].map(({ label, value }) => (
            <div
              key={label}
              className="flex justify-between items-center py-1 border-b border-pos-border"
            >
              <span className="text-xs text-pos-muted">{label}</span>
              <span className="text-xs font-mono text-pos-text">{value}</span>
            </div>
          ))}
        </Section>
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
