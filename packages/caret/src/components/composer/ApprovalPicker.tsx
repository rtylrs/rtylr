import React from 'react';
import { ShieldCheck, ShieldAlert, SlidersHorizontal } from 'lucide-react';
import Menu from '../../ui/Menu';
import IconButton from '../../ui/IconButton';
import { useCaretSettingsStore } from '../../stores/settingsStore';

/**
 * Surfaces the auto-approve setting inline in the composer (previously only
 * reachable from the Settings view) - same pattern as Cloudflare's
 * "Ask before editing" / "Automatically edit" picker.
 */
export const ApprovalPicker: React.FC = () => {
  const autoApproveReadOnly = useCaretSettingsStore((s) => s.autoApproveReadOnly);
  const setAutoApproveReadOnly = useCaretSettingsStore((s) => s.setAutoApproveReadOnly);

  return (
    <Menu
      side="top"
      align="end"
      minWidth={230}
      trigger={
        <IconButton
          className="caret-composer-icon"
          size="md"
          label={autoApproveReadOnly ? 'Approval mode: automatic' : 'Approval mode: ask first'}
          aria-haspopup="menu"
        >
          <SlidersHorizontal size={13} />
        </IconButton>
      }
      items={[
        {
          id: 'ask',
          label: 'Ask before acting',
          description: 'Review and approve each change',
          icon: <ShieldAlert size={14} />,
          selected: !autoApproveReadOnly,
          onSelect: () => setAutoApproveReadOnly(false),
        },
        {
          id: 'auto',
          label: 'Auto-approve safe actions',
          description: 'Skip confirmation for read-only actions this conversation',
          icon: <ShieldCheck size={14} />,
          selected: autoApproveReadOnly,
          onSelect: () => setAutoApproveReadOnly(true),
        },
      ]}
    />
  );
};

export default ApprovalPicker;
