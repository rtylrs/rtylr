import React, { useEffect, useState } from 'react';
import { MapPin, Monitor, Plus, Loader2 } from 'lucide-react';
import { createOneKitClient } from '@voxire/api';
import { AccessServiceClient } from '@voxire/voxie/pos/access/v1/client';
import { LocationServiceClient } from '@voxire/voxie/pos/location/v1/client';
import { TerminalServiceClient } from '@voxire/voxie/pos/terminal/v1/client';
import { toast } from 'sonner';
import { useUserStore } from '../../shared/stores/userStore';
import { RtylrWordmark } from '../../shared/components/RtylrWordmark';
import {
  createdLocationId,
  createdTerminalId,
  filterTerminalsForLocation,
  mapBranchLocations,
  mergeLocationOptions,
  type LocationOption,
  type TerminalOption,
} from './locationSelection';

const accessClient = createOneKitClient(AccessServiceClient, 'pos');
const locationClient = createOneKitClient(LocationServiceClient, 'pos');
const terminalClient = createOneKitClient(TerminalServiceClient, 'pos');

interface Props {
  onComplete: (locationId: number, terminalId: number) => void;
}

type Step = 'location' | 'create' | 'terminal';

const LocationSelect: React.FC<Props> = ({ onComplete }) => {
  // Only the business owner/admin may create POS locations; store managers and
  // staff work at locations they're assigned to.
  const role = useUserStore((s) => s.role);
  const canCreateLocation = role === 'owner' || role === 'admin';
  const [step, setStep] = useState<Step>('location');
  const [locations, setLocations] = useState<LocationOption[]>([]);
  const [terminals, setTerminals] = useState<TerminalOption[]>([]);
  const [selectedLocation, setSelectedLocation] = useState<LocationOption | null>(null);
  const [loading, setLoading] = useState(true);
  const [creatingTerminal, setCreatingTerminal] = useState(false);
  const [creatingLocation, setCreatingLocation] = useState(false);
  const [newTerminalName, setNewTerminalName] = useState('');
  const [newLocationName, setNewLocationName] = useState('');
  const [newLocationAddress, setNewLocationAddress] = useState('');
  const [newLocationPhone, setNewLocationPhone] = useState('');
  const [locationError, setLocationError] = useState<string | null>(null);
  const [terminalError, setTerminalError] = useState<string | null>(null);

  // Load both assigned (staff) and owned (business owner) locations.
  // Owners typically have no member-location assignments, so the assigned
  // list alone is not enough - merge both sources.
  useEffect(() => {
    (async () => {
      const [assigned, owned] = await Promise.allSettled([
        accessClient.getMyLocations({}),
        locationClient.listLocations({}),
      ]);

      const lists: LocationOption[][] = [];
      for (const result of [assigned, owned]) {
        if (result.status === 'fulfilled') {
          lists.push(mapBranchLocations(result.value.locations ?? []));
        }
      }

      if (lists.length === 0) {
        setLocations([]);
        setLocationError('Unable to load locations. Check your connection and try again.');
        setLoading(false);
        return;
      }

      const merged = mergeLocationOptions(...lists);
      setLocations(merged);
      setLocationError(null);

      if (merged.length === 1) {
        await handleLocationSelect(merged[0]);
      } else if (merged.length === 0 && canCreateLocation) {
        // Owner/admin first-run onboarding. Staff/store managers with no
        // assignments stay on the location step and see the "ask your owner" note.
        setStep('create');
      }
      setLoading(false);
    })();
  }, []);

  const handleLocationSelect = async (loc: LocationOption) => {
    setSelectedLocation(loc);
    setStep('terminal');
    setLoading(true);
    setTerminalError(null);
    try {
      const data = await terminalClient.listTerminals({});
      const filtered = filterTerminalsForLocation(data.terminals ?? [], loc.locationId);
      setTerminals(filtered);
      if (filtered.length === 0) {
        setNewTerminalName((prev) => prev || 'Register 1');
      }
    } catch {
      setTerminals([]);
      setTerminalError('Unable to load terminals. You can retry or register a new terminal.');
    }
    setLoading(false);
  };

  const handleCreateLocation = async () => {
    if (!canCreateLocation) return;
    if (!newLocationName.trim()) return;
    setCreatingLocation(true);
    try {
      const data = await locationClient.createLocation({
        name: newLocationName.trim(),
        address: newLocationAddress.trim(),
        phone: newLocationPhone.trim(),
        isActive: true,
      });
      const loc: LocationOption = {
        locationId: createdLocationId(data.locationId ?? ''),
        name: newLocationName.trim(),
        address: newLocationAddress.trim(),
        phone: newLocationPhone.trim(),
      };
      setLocations((prev) => mergeLocationOptions(prev, [loc]));
      await handleLocationSelect(loc);
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Failed to create location');
    } finally {
      setCreatingLocation(false);
    }
  };

  const handleTerminalSelect = (terminal: TerminalOption) => {
    if (!selectedLocation) return;
    onComplete(selectedLocation.locationId, terminal.id);
  };

  const handleCreateTerminal = async () => {
    if (!selectedLocation || !newTerminalName.trim()) return;
    setCreatingTerminal(true);
    try {
      const data = await terminalClient.createTerminal({
        branchId: String(selectedLocation.locationId),
        name: newTerminalName.trim(),
        deviceId: '',
      });
      onComplete(selectedLocation.locationId, createdTerminalId(data.terminalId ?? ''));
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Failed to register terminal');
    } finally {
      setCreatingTerminal(false);
    }
  };

  const isFirstLocation = locations.length === 0;

  const subtitle =
    step === 'location'
      ? 'Select your location'
      : step === 'create'
        ? isFirstLocation
          ? "Welcome! Let's set up your first location"
          : 'Add a new location'
        : `Select terminal at ${selectedLocation?.name}`;

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
          <RtylrWordmark suffix="POS" className="text-2xl" />
          <p className="text-sm text-pos-muted mt-1">{subtitle}</p>
        </div>

        {step === 'location' && (
          <div className="grid gap-3">
            {locationError && (
              <div className="text-center py-12">
                <MapPin className="w-10 h-10 mx-auto text-pos-muted mb-3" />
                <p className="text-sm text-destructive">{locationError}</p>
              </div>
            )}
            {!locationError && locations.length === 0 && !canCreateLocation && (
              <div className="text-center py-12">
                <MapPin className="w-10 h-10 mx-auto text-pos-muted mb-3" />
                <p className="text-sm text-pos-text font-medium">No locations assigned yet</p>
                <p className="text-xs text-pos-muted mt-1">
                  Ask your business owner to assign you to a location.
                </p>
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
            {!locationError && canCreateLocation && (
              <button
                onClick={() => setStep('create')}
                className="w-full flex items-center justify-center gap-2 p-4 rounded-xl border border-dashed border-pos-border text-pos-muted hover:text-pos-text hover:border-blue-500/50 transition-colors text-sm"
              >
                <Plus className="w-4 h-4" />
                Add location
              </button>
            )}
          </div>
        )}

        {step === 'create' && (
          <div className="space-y-4">
            {locations.length > 0 && (
              <button
                onClick={() => setStep('location')}
                className="text-xs text-pos-muted hover:text-pos-text transition-colors"
              >
                &larr; Back to locations
              </button>
            )}

            <div className="p-5 rounded-xl bg-pos-card border border-pos-border space-y-3">
              {isFirstLocation && (
                <p className="text-xs text-pos-muted">
                  Create the location this register runs at. You can add terminals, menu items
                  and staff right after.
                </p>
              )}
              <input
                type="text"
                value={newLocationName}
                onChange={(e) => setNewLocationName(e.target.value)}
                placeholder="Location name (e.g. Downtown Cafe)"
                className="w-full h-11 px-3 rounded-lg text-sm bg-pos-bg border border-pos-border text-pos-text placeholder:text-pos-muted outline-none focus:border-blue-500/50"
              />
              <input
                type="text"
                value={newLocationAddress}
                onChange={(e) => setNewLocationAddress(e.target.value)}
                placeholder="Address (optional)"
                className="w-full h-11 px-3 rounded-lg text-sm bg-pos-bg border border-pos-border text-pos-text placeholder:text-pos-muted outline-none focus:border-blue-500/50"
              />
              <input
                type="text"
                value={newLocationPhone}
                onChange={(e) => setNewLocationPhone(e.target.value)}
                placeholder="Phone (optional)"
                className="w-full h-11 px-3 rounded-lg text-sm bg-pos-bg border border-pos-border text-pos-text placeholder:text-pos-muted outline-none focus:border-blue-500/50"
              />
              <button
                onClick={handleCreateLocation}
                disabled={!newLocationName.trim() || creatingLocation}
                className="w-full h-11 rounded-lg text-sm font-semibold bg-blue-500 text-white disabled:opacity-50 flex items-center justify-center"
              >
                {creatingLocation ? (
                  <Loader2 className="w-4 h-4 animate-spin" />
                ) : (
                  'Create location'
                )}
              </button>
            </div>

            {isFirstLocation && (
              <p className="text-center text-xs text-pos-muted">
                Staff member? Ask your business owner to assign you to a location instead.
              </p>
            )}
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
                  <p className="text-sm font-medium text-pos-text">
                    {terminals.length === 0 ? 'Register this terminal' : 'Register new terminal'}
                  </p>
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
