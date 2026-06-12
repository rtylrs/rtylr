import React, { useEffect, useState } from 'react';
import { MapPin, Monitor, Plus, Loader2 } from 'lucide-react';
import { unwrap } from '@voxire/api';
import { toast } from 'sonner';
import { posService } from '../../shared/services/api/posService';
import {
  extractCreatedTerminalId,
  filterTerminalsForLocation,
  mapBranchLocations,
  type LocationOption,
  type TerminalOption,
} from './locationSelection';

interface Props {
  onComplete: (locationId: number, terminalId: number) => void;
}

const LocationSelect: React.FC<Props> = ({ onComplete }) => {
  const [step, setStep] = useState<'location' | 'terminal'>('location');
  const [locations, setLocations] = useState<LocationOption[]>([]);
  const [terminals, setTerminals] = useState<TerminalOption[]>([]);
  const [selectedLocation, setSelectedLocation] = useState<LocationOption | null>(null);
  const [loading, setLoading] = useState(true);
  const [creatingTerminal, setCreatingTerminal] = useState(false);
  const [newTerminalName, setNewTerminalName] = useState('');
  const [locationError, setLocationError] = useState<string | null>(null);
  const [terminalError, setTerminalError] = useState<string | null>(null);

  // Load locations
  useEffect(() => {
    (async () => {
      try {
        const data = await unwrap(posService.getMyLocations());
        const locs = mapBranchLocations(
          ((data as Record<string, unknown>)?.locations as unknown[]) ?? []
        );
        setLocations(locs);
        setLocationError(null);
        // Auto-select if only one location
        if (locs.length === 1) {
          handleLocationSelect(locs[0]);
        }
      } catch {
        // Owner/admin sees all locations
        try {
          const data = await unwrap(posService.getBranches());
          const rawBranches = Array.isArray(data)
            ? data
            : ((data as Record<string, unknown>)?.locations ?? []);
          const mapped = mapBranchLocations(rawBranches as unknown[]);
          setLocations(mapped);
          setLocationError(null);
          if (mapped.length === 1) handleLocationSelect(mapped[0]);
        } catch {
          setLocations([]);
          setLocationError('Unable to load locations. Check your connection and try again.');
        }
      } finally {
        setLoading(false);
      }
    })();
  }, []);

  const handleLocationSelect = async (loc: LocationOption) => {
    setSelectedLocation(loc);
    setStep('terminal');
    setLoading(true);
    setTerminalError(null);
    try {
      const data = await unwrap(posService.getTerminals());
      const rawTerminals =
        (data as Record<string, unknown>)?.terminals ?? (Array.isArray(data) ? data : []);
      const filtered = filterTerminalsForLocation(rawTerminals as unknown[], loc.locationId);
      setTerminals(filtered);
    } catch {
      setTerminals([]);
      setTerminalError('Unable to load terminals. You can retry or register a new terminal.');
    }
    setLoading(false);
  };

  const handleTerminalSelect = (terminal: TerminalOption) => {
    if (!selectedLocation) return;
    onComplete(selectedLocation.locationId, terminal.id);
  };

  const handleCreateTerminal = async () => {
    if (!selectedLocation || !newTerminalName.trim()) return;
    setCreatingTerminal(true);
    try {
      const data = await unwrap(
        posService.createTerminal({
          branchId: selectedLocation.locationId,
          name: newTerminalName.trim(),
          type: 'pos',
        })
      );
      onComplete(selectedLocation.locationId, extractCreatedTerminalId(data));
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Failed to register terminal');
    } finally {
      setCreatingTerminal(false);
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-pos-bg">
        <Loader2 className="w-8 h-8 animate-spin text-pos-muted" />
      </div>
    );
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-pos-bg p-6">
      <div className="w-full max-w-lg">
        {/* Header */}
        <div className="text-center mb-8">
          <h1 className="text-2xl font-bold text-pos-text">rtylr POS</h1>
          <p className="text-sm text-pos-muted mt-1">
            {step === 'location'
              ? 'Select your location'
              : `Select terminal at ${selectedLocation?.name}`}
          </p>
        </div>

        {step === 'location' && (
          <div className="grid gap-3">
            {locations.length === 0 && (
              <div className="text-center py-12">
                <MapPin className="w-10 h-10 mx-auto text-pos-muted mb-3" />
                {locationError ? (
                  <p className="text-sm text-destructive">{locationError}</p>
                ) : (
                  <>
                    <p className="text-sm text-pos-muted">No locations assigned to your account.</p>
                    <p className="text-xs text-pos-muted mt-1">
                      Ask your business owner to assign you to a location.
                    </p>
                  </>
                )}
              </div>
            )}
            {locations.map((loc) => (
              <button
                key={loc.locationId}
                onClick={() => handleLocationSelect(loc)}
                className="w-full flex items-center gap-4 p-5 rounded-xl bg-pos-card border border-pos-border hover:border-blue-500/50 transition-colors text-left"
              >
                <div className="w-12 h-12 rounded-xl bg-blue-500/10 flex items-center justify-center shrink-0">
                  <MapPin className="w-5 h-5 text-blue-500" />
                </div>
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-semibold text-pos-text">{loc.name}</p>
                  {loc.address && (
                    <p className="text-xs text-pos-muted truncate mt-0.5">{loc.address}</p>
                  )}
                </div>
              </button>
            ))}
          </div>
        )}

        {step === 'terminal' && (
          <div className="space-y-4">
            {/* Back button */}
            {locations.length > 1 && (
              <button
                onClick={() => {
                  setStep('location');
                  setSelectedLocation(null);
                }}
                className="text-xs text-pos-muted hover:text-pos-text transition-colors"
              >
                &larr; Back to locations
              </button>
            )}

            <div className="grid gap-3">
              {terminalError && (
                <div className="rounded-lg border border-destructive/40 bg-destructive/10 px-3 py-2 text-xs text-destructive">
                  {terminalError}
                </div>
              )}
              {terminals.map((t) => (
                <button
                  key={t.id}
                  onClick={() => handleTerminalSelect(t)}
                  className="w-full flex items-center gap-4 p-5 rounded-xl bg-pos-card border border-pos-border hover:border-blue-500/50 transition-colors text-left"
                >
                  <div className="w-12 h-12 rounded-xl bg-emerald-500/10 flex items-center justify-center shrink-0">
                    <Monitor className="w-5 h-5 text-emerald-500" />
                  </div>
                  <div className="flex-1">
                    <p className="text-sm font-semibold text-pos-text">{t.name}</p>
                    <p className="text-xs text-pos-muted capitalize">{t.type} terminal</p>
                  </div>
                </button>
              ))}

              {/* Register new terminal */}
              <div className="p-5 rounded-xl bg-pos-card border border-dashed border-pos-border">
                <div className="flex items-center gap-3 mb-3">
                  <Plus className="w-4 h-4 text-pos-muted" />
                  <p className="text-sm font-medium text-pos-text">Register new terminal</p>
                </div>
                <div className="flex gap-2">
                  <input
                    type="text"
                    value={newTerminalName}
                    onChange={(e) => setNewTerminalName(e.target.value)}
                    placeholder="Terminal name (e.g. Register 1)"
                    className="flex-1 h-10 px-3 rounded-lg text-sm bg-pos-bg border border-pos-border text-pos-text placeholder:text-pos-muted outline-none focus:border-blue-500/50"
                  />
                  <button
                    onClick={handleCreateTerminal}
                    disabled={!newTerminalName.trim() || creatingTerminal}
                    className="h-10 px-4 rounded-lg text-sm font-medium bg-blue-500 text-white disabled:opacity-50"
                  >
                    {creatingTerminal ? <Loader2 className="w-4 h-4 animate-spin" /> : 'Add'}
                  </button>
                </div>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default LocationSelect;
