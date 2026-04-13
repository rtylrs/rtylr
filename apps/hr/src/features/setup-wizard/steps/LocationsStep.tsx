import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Plus, Trash2 } from 'lucide-react';
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

const LocationsStep = ({ onNext, onBack, isSubmitting }: LocationsStepProps): React.JSX.Element => {
  const { t } = useTranslation();
  const [singleLocation, setSingleLocation] = useState(false);
  const [locations, setLocations] = useState<LocationRow[]>([]);
  const [newName, setNewName] = useState('');
  const [newAddress, setNewAddress] = useState('');

  const addLocation = (): void => {
    const name = newName.trim();
    const address = newAddress.trim();
    if (!name || !address) return;
    setLocations((prev) => [...prev, { name, address }]);
    setNewName('');
    setNewAddress('');
  };

  const removeLocation = (index: number): void => {
    setLocations((prev) => prev.filter((_, i) => i !== index));
  };

  const handleSubmit = (): void => {
    onNext({ singleLocation, locations });
  };

  const inputStyle: React.CSSProperties = {
    width: '100%',
    padding: '10px 12px',
    borderRadius: '8px',
    border: '1px solid var(--hr-border)',
    background: 'var(--hr-input-bg)',
    color: 'var(--hr-text)',
    fontSize: '14px',
    outline: 'none',
  };

  return (
    <StepContainer
      title={t('setup.locations.title')}
      description={t('setup.locations.description')}
      onBack={onBack}
      onNext={handleSubmit}
      isSubmitting={isSubmitting}
    >
      <div className="flex flex-col gap-4">
        <label className="flex items-center gap-2 cursor-pointer">
          <input
            type="checkbox"
            checked={singleLocation}
            onChange={(e) => setSingleLocation(e.target.checked)}
            className="h-4 w-4 rounded"
          />
          <span className="text-sm" style={{ color: 'var(--hr-text)' }}>
            {t('setup.locations.singleLocation')}
          </span>
        </label>

        {!singleLocation && (
          <>
            <div className="flex flex-col gap-2 sm:flex-row">
              <input
                type="text"
                placeholder={t('setup.locations.locationName')}
                value={newName}
                onChange={(e) => setNewName(e.target.value)}
                style={inputStyle}
              />
              <input
                type="text"
                placeholder={t('setup.locations.locationAddress')}
                value={newAddress}
                onChange={(e) => setNewAddress(e.target.value)}
                style={inputStyle}
              />
              <button
                type="button"
                onClick={addLocation}
                className="flex shrink-0 items-center gap-1 rounded-lg px-3 py-2 text-sm font-medium"
                style={{ background: 'var(--accent)', color: '#fff' }}
                aria-label={t('setup.locations.addLocation')}
              >
                <Plus size={14} />
                {t('common.add')}
              </button>
            </div>

            {locations.length === 0 ? (
              <p className="text-center text-sm" style={{ color: 'var(--hr-text-dim)' }}>
                {t('setup.locations.noLocations')}
              </p>
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
                    </div>
                    <button
                      type="button"
                      onClick={() => removeLocation(index)}
                      className="rounded p-1 transition-colors hover:bg-red-500/10"
                      style={{ color: '#ef4444' }}
                      aria-label={t('common.delete')}
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
