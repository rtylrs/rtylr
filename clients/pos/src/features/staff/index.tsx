import { createOneKitClient } from '@voxire/api';
import { UserServiceClient } from '@voxire/voxie/pos/user/v1/client';
import React, { useCallback, useEffect, useState } from 'react';
import { toast } from 'sonner';
import { Loader2, ShieldCheck, UserX, KeyRound } from 'lucide-react';
import ScreenLayout from '../../layouts/ScreenLayout';
import ManagerRequired from '../../shared/components/ManagerRequired';
import { useCapabilities } from '../../shared/hooks/useCapabilities';
import { useTerminalContextStore } from '../../shared/stores/terminalContextStore';
import {
  mapPOSUsers,
  buildStaffUpdate,
  type StaffUser,
  type StaffEdits,
  type StaffRole,
} from './staffWorkflow';

const userClient = createOneKitClient(UserServiceClient, 'pos');

const ROLE_OPTIONS: StaffRole[] = ['manager', 'cashier', 'server', 'bartender', 'kitchen'];

const CAP_FIELDS: { key: keyof StaffEdits; label: string }[] = [
  { key: 'canVoid', label: 'Void orders' },
  { key: 'canRefund', label: 'Issue refunds' },
  { key: 'canApplyDiscount', label: 'Apply discounts' },
  { key: 'canOpenDrawer', label: 'Open drawer' },
];

const editsFromUser = (u: StaffUser): StaffEdits => ({
  role: u.role,
  isActive: u.isActive,
  maxDiscountPct: u.maxDiscountPct,
  canVoid: u.canVoid,
  canRefund: u.canRefund,
  canOpenDrawer: u.canOpenDrawer,
  canApplyDiscount: u.canApplyDiscount,
  newPin: '',
});

const StaffScreen: React.FC = () => {
  const { locationId } = useTerminalContextStore();
  const { role } = useCapabilities();
  const isManager = role === 'manager';

  const [users, setUsers] = useState<StaffUser[]>([]);
  const [loading, setLoading] = useState(true);
  const [selected, setSelected] = useState<StaffUser | null>(null);
  const [edits, setEdits] = useState<StaffEdits | null>(null);
  const [saving, setSaving] = useState(false);

  const load = useCallback(async () => {
    if (!locationId) return;
    setLoading(true);
    try {
      const data = await userClient.listPOSUsers({ locationId: String(locationId) });
      setUsers(mapPOSUsers(data.users ?? []));
    } catch {
      toast.error('Failed to load staff');
    } finally {
      setLoading(false);
    }
  }, [locationId]);

  useEffect(() => {
    if (isManager) void load();
    else setLoading(false);
  }, [isManager, load]);

  const openEditor = (u: StaffUser): void => {
    setSelected(u);
    setEdits(editsFromUser(u));
  };

  const closeEditor = (): void => {
    setSelected(null);
    setEdits(null);
  };

  const save = async (): Promise<void> => {
    if (!selected || !edits) return;
    setSaving(true);
    try {
      await userClient.updatePOSUser(buildStaffUpdate(selected, edits));
      toast.success(edits.newPin ? 'Saved — PIN reset' : 'Staff member updated');
      closeEditor();
      await load();
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Could not save changes');
    } finally {
      setSaving(false);
    }
  };

  const deactivate = async (u: StaffUser): Promise<void> => {
    try {
      await userClient.deactivatePOSUser({ id: String(u.id) });
      toast.success(`${u.name} deactivated`);
      closeEditor();
      await load();
    } catch {
      toast.error('Could not deactivate');
    }
  };

  if (!isManager) {
    return (
      <ScreenLayout title="Staff">
        <ManagerRequired action="manage staff" />
      </ScreenLayout>
    );
  }

  return (
    <ScreenLayout title="Staff" subtitle="Manage PINs and permissions for this location">
      <div className="p-5 max-w-2xl">
        {loading ? (
          <div className="flex justify-center py-12">
            <Loader2 className="w-7 h-7 animate-spin text-pos-muted" />
          </div>
        ) : users.length === 0 ? (
          <p className="text-sm text-pos-muted py-8 text-center">
            No staff have signed in at this location yet. Assign team members to this location in the
            dashboard; they set their PIN on first sign-in.
          </p>
        ) : (
          <div className="grid gap-2">
            {users.map((u) => (
              <button
                key={u.id}
                onClick={() => openEditor(u)}
                className="flex items-center justify-between p-4 rounded-xl bg-pos-card border border-pos-border hover:border-blue-500/50 transition-colors text-left"
              >
                <div className="flex items-center gap-3">
                  <div className="w-10 h-10 rounded-full bg-blue-500/10 flex items-center justify-center font-bold text-blue-400">
                    {u.name.charAt(0).toUpperCase()}
                  </div>
                  <div>
                    <p className="text-sm font-semibold text-pos-text flex items-center gap-1.5">
                      {u.name}
                      {u.role === 'manager' && <ShieldCheck className="w-3.5 h-3.5 text-emerald-500" />}
                      {!u.isActive && <span className="text-[10px] text-pos-muted">(inactive)</span>}
                    </p>
                    <p className="text-[11px] text-pos-muted capitalize">{u.role}</p>
                  </div>
                </div>
              </button>
            ))}
          </div>
        )}
      </div>

      {/* Editor */}
      {selected && edits && (
        <div
          className="fixed inset-0 z-50 flex items-end sm:items-center justify-center bg-black/50 p-4"
          onClick={closeEditor}
        >
          <div
            className="w-full max-w-md bg-pos-card border border-pos-border rounded-2xl p-5 space-y-4"
            onClick={(e) => e.stopPropagation()}
          >
            <div>
              <h2 className="text-base font-bold text-pos-text">{selected.name}</h2>
              <p className="text-[11px] text-pos-muted">Role and permissions at this location</p>
            </div>

            <div className="space-y-1.5">
              <label className="text-xs text-pos-muted">Role</label>
              <select
                value={edits.role}
                onChange={(e) => setEdits({ ...edits, role: e.target.value as StaffRole })}
                className="w-full h-11 px-3 rounded-lg bg-pos-bg border border-pos-border text-pos-text text-sm capitalize outline-none"
              >
                {ROLE_OPTIONS.map((r) => (
                  <option key={r} value={r}>
                    {r}
                  </option>
                ))}
              </select>
            </div>

            <div className="grid grid-cols-2 gap-2">
              {CAP_FIELDS.map((c) => {
                const on = Boolean(edits[c.key]);
                return (
                  <button
                    key={c.key}
                    onClick={() => setEdits({ ...edits, [c.key]: !on })}
                    className="flex items-center justify-between px-3 py-2.5 rounded-lg text-left text-xs"
                    style={{
                      background: on ? 'rgba(16,185,129,0.08)' : 'var(--pos-input, rgba(255,255,255,0.04))',
                      border: `1px solid ${on ? 'rgba(16,185,129,0.4)' : 'transparent'}`,
                    }}
                  >
                    <span className={on ? 'text-emerald-400 font-medium' : 'text-pos-muted'}>
                      {c.label}
                    </span>
                    <span className={on ? 'text-emerald-400' : 'text-pos-muted'}>{on ? 'On' : 'Off'}</span>
                  </button>
                );
              })}
            </div>

            <div className="space-y-1.5">
              <label className="text-xs text-pos-muted">Max discount %</label>
              <input
                type="number"
                inputMode="decimal"
                value={edits.maxDiscountPct}
                onChange={(e) => setEdits({ ...edits, maxDiscountPct: Number(e.target.value) || 0 })}
                className="w-full h-11 px-3 rounded-lg bg-pos-bg border border-pos-border text-pos-text text-sm outline-none"
              />
            </div>

            <div className="space-y-1.5">
              <label className="text-xs text-pos-muted flex items-center gap-1.5">
                <KeyRound className="w-3.5 h-3.5" /> Reset PIN (optional)
              </label>
              <input
                type="password"
                inputMode="numeric"
                maxLength={4}
                value={edits.newPin ?? ''}
                onChange={(e) =>
                  setEdits({ ...edits, newPin: e.target.value.replace(/\D/g, '').slice(0, 4) })
                }
                placeholder="Leave blank to keep current"
                className="w-full h-11 px-3 rounded-lg bg-pos-bg border border-pos-border text-pos-text text-sm outline-none"
              />
            </div>

            <div className="flex items-center justify-between gap-2 pt-1">
              <button
                onClick={() => void deactivate(selected)}
                className="inline-flex items-center gap-1.5 h-10 px-3 rounded-lg text-xs font-medium text-destructive hover:bg-destructive/10"
              >
                <UserX className="w-4 h-4" /> Deactivate
              </button>
              <div className="flex items-center gap-2">
                <button
                  onClick={closeEditor}
                  className="h-10 px-4 rounded-lg text-sm text-pos-muted hover:text-pos-text"
                >
                  Cancel
                </button>
                <button
                  onClick={() => void save()}
                  disabled={saving}
                  className="h-10 px-5 rounded-lg text-sm font-semibold bg-pos-text text-pos-bg disabled:opacity-50"
                >
                  {saving ? 'Saving…' : 'Save'}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </ScreenLayout>
  );
};

export default StaffScreen;
