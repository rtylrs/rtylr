import React, { useRef, useState } from 'react';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { Plus, Trash2, MapPin } from 'lucide-react';
import { EmptyState } from '@voxire/ui';
import { TIMEZONES } from '@voxire/config';
import StepContainer from '../components/StepContainer';

interface LocationRow {
  name: string;
  address: string;
  timezone?: string;
}

interface LocationsStepProps {
  onNext: (data: { singleLocation: boolean; locations: LocationRow[] }) => void;
  onBack: () => void;
  isSubmitting?: boolean;
}

const inputStyle: React.CSSProperties = {
  border: '1px solid var(--hr-border)',
  background: 'var(--hr-input-bg)',
  color: 'var(--hr-text)',
  fontSize: '16px',
};

const LocationsStep = ({ onNext, onBack, isSubmitting }: LocationsStepProps): React.JSX.Element => {
  const { t } = useTranslation('hrCommon');
  const [singleLocation, setSingleLocation] = useState(false);
  const [locations, setLocations] = useState<LocationRow[]>([]);
  const [newName, setNewName] = useState('');
  const [newAddress, setNewAddress] = useState('');
  const [newTimezone, setNewTimezone] = useState('');
  const nameInputRef = useRef<HTMLInputElement>(null);

  const addLocation = (): void => {
    const name = newName.trim();
    const address = newAddress.trim();
    if (!name || !address) {
      toast.error(t('hrSetup.locations.required'));
      return;
    }
    setLocations((prev) => [...prev, { name, address, timezone: newTimezone || undefined }]);
    setNewName('');
    setNewAddress('');
    setNewTimezone('');
  };

  const removeLocation = (index: number): void => {
    setLocations((prev) => prev.filter((_, i) => i !== index));
  };

  const handleSubmit = (): void => {
    if (singleLocation) {
      const name = newName.trim();
      const address = newAddress.trim();
      const singleLoc = name && address ? [{ name, address }] : [];
      onNext({ singleLocation: true, locations: singleLoc });
    } else {
      onNext({ singleLocation: false, locations });
    }
  };

  return (
    <StepContainer
      title={t('hrSetup.locations.title')}
      description={t('hrSetup.locations.description')}
      onBack={onBack}
      onNext={handleSubmit}
      isSubmitting={isSubmitting}
    >
      <div className="flex flex-col gap-4">
        <label className="flex cursor-pointer items-center gap-2">
          <input
            type="checkbox"
            checked={singleLocation}
            onChange={(e) => setSingleLocation(e.target.checked)}
            className="h-4 w-4 rounded"
          />
          <span className="text-sm" style={{ color: 'var(--hr-text)' }}>
            {t('hrSetup.locations.singleLocation')}
          </span>
        </label>

        {singleLocation ? (
          <div className="flex flex-col gap-3">
            <input
              type="text"
              placeholder={t('hrSetup.locations.locationName')}
              value={newName}
              onChange={(e) => setNewName(e.target.value)}
              className="w-full rounded-lg px-3 py-2.5 outline-none"
              style={inputStyle}
            />
            <input
              type="text"
              placeholder={t('hrSetup.locations.locationAddress')}
              value={newAddress}
              onChange={(e) => setNewAddress(e.target.value)}
              className="w-full rounded-lg px-3 py-2.5 outline-none"
              style={inputStyle}
            />
          </div>
        ) : (
          <>
            <div className="flex flex-col gap-2">
              <div className="flex flex-col gap-2 sm:flex-row">
                <input
                  ref={nameInputRef}
                  type="text"
                  placeholder={t('hrSetup.locations.locationName')}
                  value={newName}
                  onChange={(e) => setNewName(e.target.value)}
                  className="flex-1 rounded-lg px-3 py-2.5 outline-none"
                  style={inputStyle}
                />
                <input
                  type="text"
                  placeholder={t('hrSetup.locations.locationAddress')}
                  value={newAddress}
                  onChange={(e) => setNewAddress(e.target.value)}
                  className="flex-1 rounded-lg px-3 py-2.5 outline-none"
                  style={inputStyle}
                />
              </div>
              <div className="flex gap-2">
                <div className="flex-1">
                  <label
                    className="mb-1 block text-xs font-medium"
                    style={{ color: 'var(--hr-text-muted)' }}
                  >
                    {t('hrSetup.locations.locationTimezone')}
                  </label>
                  <select
                    value={newTimezone}
                    onChange={(e) => setNewTimezone(e.target.value)}
                    className="w-full rounded-lg px-3 py-2.5 outline-none"
                    style={inputStyle}
                  >
                    <option value="">-</option>
                    {TIMEZONES.map((tz) => (
                      <option key={tz.value} value={tz.value}>
                        {tz.label}
                      </option>
                    ))}
                  </select>
                </div>
                <button
                  type="button"
                  onClick={addLocation}
                  className="mt-6 flex shrink-0 items-center gap-1 rounded-lg px-4 py-2.5 text-sm font-medium"
                  style={{ background: 'var(--accent)', color: 'var(--hr-on-accent)' }}
                  aria-label={t('hrSetup.locations.addLocation')}
                >
                  <Plus size={14} />
                  {t('hrCommon.add')}
                </button>
              </div>
            </div>

            {locations.length === 0 ? (
              <EmptyState
                icon={MapPin}
                title={t('hrSetup.locations.noLocations')}
                action={{
                  label: t('hrSetup.locations.addLocation'),
                  onClick: () => nameInputRef.current?.focus(),
                }}
              />
            ) : (
              <div className="flex flex-col gap-2">
                {locations.map((loc, index) => (
                  <div
                    key={index}
                    className="flex items-center justify-between rounded-lg px-4 py-3"
                    style={{ background: 'var(--hr-hover)', border: '1px solid var(--hr-border)' }}
                  >
                    <div>
                      <span className="text-sm font-medium" style={{ color: 'var(--hr-text)' }}>
                        {loc.name}
                      </span>
                      <span className="ml-2 text-xs" style={{ color: 'var(--hr-text-muted)' }}>
                        {loc.address}
                      </span>
                      {loc.timezone && (
                        <span className="ml-2 text-xs" style={{ color: 'var(--hr-text-muted)' }}>
                          ({loc.timezone})
                        </span>
                      )}
                    </div>
                    <button
                      type="button"
                      onClick={() => removeLocation(index)}
                      className="rounded p-1 transition-colors hover:bg-red-500/10"
                      style={{ color: '#ef4444' }}
                      aria-label={t('hrCommon.delete')}
                    >
                      <Trash2 size={14} />
                    </button>
                  </div>
                ))}
              </div>
            )}
          </>
        )}
      </div>
    </StepContainer>
  );
};

export default LocationsStep;
