import React from 'react';
import { useTranslation } from 'react-i18next';

interface PermissionProfilePickerProps {
  label: string;
  profiles: { id: string; name: string }[];
  selectedIds: string[];
  onChange: (ids: string[]) => void;
  disabled?: boolean;
}

export function PermissionProfilePicker({
  label,
  profiles,
  selectedIds,
  onChange,
  disabled = false,
}: PermissionProfilePickerProps): React.JSX.Element {
  const { t } = useTranslation();
  return (
    <fieldset disabled={disabled} className="space-y-2 disabled:opacity-50">
      <legend className="text-xs font-medium" style={{ color: 'var(--dash-text)' }}>
        {label}
      </legend>
      {profiles.length === 0 ? (
        <p className="text-xs" style={{ color: 'var(--dash-text-dim)' }}>
          {t('team.authorization.noProfilesAvailable')}
        </p>
      ) : (
        profiles.map((profile) => (
          <label
            key={profile.id}
            className="flex items-center gap-2 rounded-lg px-3 py-2 text-xs"
            style={{
              background: 'var(--dash-bg)',
              border: '1px solid var(--dash-border)',
              color: 'var(--dash-text)',
            }}
          >
            <input
              type="checkbox"
              checked={selectedIds.includes(profile.id)}
              onChange={(event) =>
                onChange(
                  event.target.checked
                    ? [...new Set([...selectedIds, profile.id])]
                    : selectedIds.filter((id) => id !== profile.id)
                )
              }
            />
            <span>{profile.name}</span>
          </label>
        ))
      )}
    </fieldset>
  );
}
