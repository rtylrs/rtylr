import React, { useState, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import { LogIn, LogOut, QrCode, KeyRound, MapPin, Delete, Settings } from 'lucide-react';
import { toast } from 'sonner';
import { useUserStore } from '../../../../shared/stores/userStore';
import {
  useAttendanceLocations,
  useClockIn,
  useClockOut,
  useSetPin,
} from '../../shared/hooks/queries/useAttendance';
import { useHRConfigStore } from '../../shared/stores/hrConfigStore';
import type { AttendanceRecordPayload } from '@voxire/voxie/hr/attendance/v1/types';
import { Card, Button, Pill, Modal, ACCENT } from '../../shared/components/ui';

interface ClockInWidgetProps {
  activeRecord?: AttendanceRecordPayload;
}

function formatTime(timestamp: number): string {
  return new Date(timestamp * 1000).toLocaleTimeString(undefined, {
    hour: '2-digit',
    minute: '2-digit',
  });
}

/* ── PIN Numpad Modal ─────────────────────────────────────────── */

interface PinNumpadProps {
  open: boolean;
  onClose: () => void;
  mode: 'clockIn' | 'setPin' | 'confirmPin';
  onSubmit: (pin: string) => void;
  isPending: boolean;
  error?: string;
}

const PIN_MIN = 4;
const PIN_MAX = 6;

const PinNumpad = ({
  open,
  onClose,
  mode,
  onSubmit,
  isPending,
  error,
}: PinNumpadProps): React.JSX.Element => {
  const { t } = useTranslation('hrCommon');
  const [pin, setPin] = useState('');

  const titleKey =
    mode === 'clockIn'
      ? 'hrAttendance.pinModal.title'
      : mode === 'setPin'
        ? 'hrAttendance.pinModal.setTitle'
        : 'hrAttendance.pinModal.confirmTitle';

  const subtitleKey =
    mode === 'clockIn'
      ? 'hrAttendance.pinModal.subtitle'
      : mode === 'setPin'
        ? 'hrAttendance.pinModal.setSubtitle'
        : 'hrAttendance.pinModal.confirmSubtitle';

  const submitLabel =
    mode === 'clockIn' ? t('hrAttendance.pinModal.submit') : t('hrAttendance.pinModal.savePin');

  const handleDigit = useCallback((digit: string): void => {
    setPin((prev) => (prev.length < PIN_MAX ? prev + digit : prev));
  }, []);

  const handleBackspace = useCallback((): void => {
    setPin((prev) => prev.slice(0, -1));
  }, []);

  const handleSubmit = useCallback((): void => {
    if (pin.length < PIN_MIN) return;
    onSubmit(pin);
    setPin('');
  }, [pin, onSubmit]);

  const handleClose = useCallback((): void => {
    setPin('');
    onClose();
  }, [onClose]);

  const isValid = pin.length >= PIN_MIN && pin.length <= PIN_MAX;

  return (
    <Modal open={open} onClose={handleClose} title={t(titleKey)} width={340}>
      <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 16 }}>
        <p style={{ margin: 0, fontSize: 12, color: 'var(--hr-text-muted)', textAlign: 'center' }}>
          {t(subtitleKey)}
        </p>

        {/* PIN dots */}
        <div
          style={{
            display: 'flex',
            gap: 10,
            justifyContent: 'center',
            minHeight: 28,
            alignItems: 'center',
          }}
        >
          {Array.from({ length: PIN_MAX }, (_, i) => (
            <div
              key={i}
              style={{
                width: 14,
                height: 14,
                borderRadius: '50%',
                border: `2px solid ${i < pin.length ? ACCENT : 'var(--hr-border)'}`,
                background: i < pin.length ? ACCENT : 'transparent',
                transition: 'all 0.15s',
              }}
            />
          ))}
        </div>

        {error && (
          <p style={{ margin: 0, fontSize: 11, color: 'var(--hr-error)', textAlign: 'center' }}>
            {error}
          </p>
        )}

        {/* Numpad grid */}
        <div
          style={{
            display: 'grid',
            gridTemplateColumns: 'repeat(3, 1fr)',
            gap: 8,
            width: '100%',
            maxWidth: 240,
          }}
        >
          {['1', '2', '3', '4', '5', '6', '7', '8', '9'].map((digit) => (
            <button
              key={digit}
              type="button"
              onClick={() => handleDigit(digit)}
              disabled={isPending}
              style={{
                height: 48,
                minWidth: 48,
                borderRadius: 10,
                border: '1px solid var(--hr-border)',
                background: 'var(--hr-input-bg)',
                color: 'var(--hr-text)',
                fontSize: 18,
                fontWeight: 600,
                cursor: 'pointer',
                transition: 'background 0.15s',
              }}
              aria-label={digit}
            >
              {digit}
            </button>
          ))}
          <button
            type="button"
            onClick={handleBackspace}
            disabled={isPending || pin.length === 0}
            style={{
              height: 48,
              minWidth: 48,
              borderRadius: 10,
              border: '1px solid var(--hr-border)',
              background: 'transparent',
              color: 'var(--hr-text-muted)',
              cursor: 'pointer',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
            }}
            aria-label={t('hrCommon.delete')}
          >
            <Delete size={18} />
          </button>
          <button
            type="button"
            onClick={() => handleDigit('0')}
            disabled={isPending}
            style={{
              height: 48,
              minWidth: 48,
              borderRadius: 10,
              border: '1px solid var(--hr-border)',
              background: 'var(--hr-input-bg)',
              color: 'var(--hr-text)',
              fontSize: 18,
              fontWeight: 600,
              cursor: 'pointer',
            }}
            aria-label="0"
          >
            0
          </button>
          <div />
        </div>

        <Button
          variant="primary"
          onClick={handleSubmit}
          disabled={!isValid || isPending}
          style={{ width: '100%', maxWidth: 240 }}
        >
          {isPending ? t('hrCommon.loading') : submitLabel}
        </Button>
      </div>
    </Modal>
  );
};

/* ── ClockInWidget ────────────────────────────────────────────── */

const ClockInWidget = ({ activeRecord }: ClockInWidgetProps): React.JSX.Element => {
  const { t } = useTranslation('hrCommon');
  const config = useHRConfigStore((s) => s.config);
  const clockIn = useClockIn();
  const clockOut = useClockOut();
  const setPin = useSetPin();
  const {
    data: locations = [],
    isPending: locationsPending,
    isError: locationsError,
  } = useAttendanceLocations();
  const userId = Number(useUserStore((s) => s.user?.id)) || 0;
  const location = locations.find((item) => item.isDefault) ?? locations[0];
  const locationId = location?.id ?? '';

  const [pinModalOpen, setPinModalOpen] = useState(false);
  const [pinMode, setPinMode] = useState<'clockIn' | 'setPin' | 'confirmPin'>('clockIn');
  const [pinError, setPinError] = useState<string | undefined>();
  const [pendingNewPin, setPendingNewPin] = useState('');
  const [gpsLoading, setGpsLoading] = useState(false);

  const isClockedIn = activeRecord && !activeRecord.clockOutAt;

  const enabledMethods = config?.clockInMethods ?? ['app'];
  const isAppEnabled = enabledMethods.includes('app');
  const isQrEnabled = enabledMethods.includes('qr');
  const isPinEnabled = enabledMethods.includes('pin');
  const isGpsEnabled = enabledMethods.includes('gps');
  const clockInDisabled = !locationId || locationsPending || locationsError;

  const handleClockIn = useCallback((): void => {
    if (!locationId) return;
    clockIn.mutate({ method: 'app', locationId });
  }, [clockIn, locationId]);

  const handleClockOut = useCallback((): void => {
    if (!activeRecord?.id) return;
    clockOut.mutate({ attendanceRecordId: activeRecord.id });
  }, [activeRecord, clockOut]);

  /* ── PIN flow ─────────────────────────────────────────────── */

  const handleOpenPinClockIn = useCallback((): void => {
    setPinError(undefined);
    setPinMode('clockIn');
    setPinModalOpen(true);
  }, []);

  const handleOpenSetPin = useCallback((): void => {
    setPinError(undefined);
    setPinMode('setPin');
    setPendingNewPin('');
    setPinModalOpen(true);
  }, []);

  const handlePinSubmit = useCallback(
    (enteredPin: string): void => {
      if (pinMode === 'clockIn') {
        if (!locationId) return;
        clockIn.mutate(
          { method: 'pin', pin: enteredPin, locationId },
          { onSuccess: () => setPinModalOpen(false) }
        );
      } else if (pinMode === 'setPin') {
        setPendingNewPin(enteredPin);
        setPinError(undefined);
        setPinMode('confirmPin');
      } else if (pinMode === 'confirmPin') {
        if (enteredPin !== pendingNewPin) {
          setPinError(t('hrAttendance.pinModal.pinMismatch'));
          return;
        }
        setPin.mutate(
          { employeeId: userId, pin: enteredPin },
          {
            onSuccess: () => {
              setPinModalOpen(false);
              setPendingNewPin('');
            },
          }
        );
      }
    },
    [pinMode, pendingNewPin, clockIn, setPin, t, userId, locationId]
  );

  /* ── GPS flow ─────────────────────────────────────────────── */

  const handleGpsClockIn = useCallback((): void => {
    if (!locationId) return;
    if (!navigator.geolocation) {
      toast.error(t('hrAttendance.gps.locationUnavailable'));
      return;
    }

    setGpsLoading(true);

    navigator.geolocation.getCurrentPosition(
      (position) => {
        setGpsLoading(false);
        clockIn.mutate({
          method: 'gps',
          locationId,
          lat: position.coords.latitude,
          lng: position.coords.longitude,
        });
      },
      (error) => {
        setGpsLoading(false);
        switch (error.code) {
          case error.PERMISSION_DENIED:
            toast.error(t('hrAttendance.gps.permissionDenied'));
            break;
          case error.POSITION_UNAVAILABLE:
            toast.error(t('hrAttendance.gps.locationUnavailable'));
            break;
          case error.TIMEOUT:
            toast.error(t('hrAttendance.gps.locationTimeout'));
            break;
          default:
            toast.error(t('hrAttendance.gps.locationUnavailable'));
        }
      },
      { enableHighAccuracy: true, timeout: 15000, maximumAge: 0 }
    );
  }, [clockIn, locationId, t]);

  return (
    <>
      <Card
        style={{
          display: 'flex',
          flexDirection: 'column',
          gap: 12,
          padding: 16,
          border: isClockedIn ? `1px solid ${ACCENT}40` : undefined,
        }}
      >
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <span style={{ fontSize: 14, fontWeight: 600, color: 'var(--hr-text)' }}>
            {t('hrAttendance.clockWidget.title')}
          </span>
          <Pill color={isClockedIn ? 'var(--hr-status-active)' : 'var(--hr-status-inactive)'}>
            {isClockedIn
              ? t('hrAttendance.clockWidget.clockedIn')
              : t('hrAttendance.clockWidget.clockedOut')}
          </Pill>
        </div>

        {isClockedIn && activeRecord ? (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
            <div style={{ fontSize: 12, color: 'var(--hr-text-muted)' }}>
              {t('hrAttendance.clockWidget.since', {
                time: formatTime(Number(activeRecord.clockInAt ?? 0)),
              })}
              {activeRecord.isLate && (
                <Pill color="var(--hr-error)" style={{ marginLeft: 8, fontSize: 9 }}>
                  {t('hrAttendance.late', { minutes: activeRecord.lateMinutes })}
                </Pill>
              )}
            </div>
            <Button
              variant="danger"
              onClick={handleClockOut}
              disabled={clockOut.isPending}
              icon={<LogOut size={14} />}
            >
              {t('hrAttendance.clockWidget.clockOutBtn')}
            </Button>
          </div>
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
            <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
              {isAppEnabled && (
                <Button
                  variant="primary"
                  onClick={handleClockIn}
                  disabled={clockIn.isPending || clockInDisabled}
                  icon={<LogIn size={14} />}
                >
                  {t('hrAttendance.clockWidget.clockInApp')}
                </Button>
              )}
              {isQrEnabled && (
                <Button
                  variant="secondary"
                  onClick={() => toast.info(t('hrAttendance.qrComingSoon'))}
                  icon={<QrCode size={14} />}
                >
                  {t('hrAttendance.clockWidget.clockInQR')}
                </Button>
              )}
              {isPinEnabled && (
                <Button
                  variant="secondary"
                  onClick={handleOpenPinClockIn}
                  disabled={clockIn.isPending || clockInDisabled}
                  icon={<KeyRound size={14} />}
                >
                  {t('hrAttendance.clockWidget.clockInPin')}
                </Button>
              )}
              {isGpsEnabled && (
                <Button
                  variant="secondary"
                  onClick={handleGpsClockIn}
                  disabled={clockIn.isPending || gpsLoading || clockInDisabled}
                  icon={<MapPin size={14} />}
                >
                  {gpsLoading
                    ? t('hrAttendance.gps.acquiring')
                    : t('hrAttendance.clockWidget.clockInGps')}
                </Button>
              )}
            </div>
            {locationsError ? (
              <p role="alert" style={{ margin: 0, fontSize: 11, color: 'var(--hr-error)' }}>
                {t('hrErrors.generic')}
              </p>
            ) : !locationsPending && !locationId ? (
              <p style={{ margin: 0, fontSize: 11, color: 'var(--hr-text-muted)' }}>
                {t('hrSetup.locations.noLocations')}
              </p>
            ) : null}
            {isPinEnabled && (
              <button
                type="button"
                onClick={handleOpenSetPin}
                style={{
                  display: 'inline-flex',
                  alignItems: 'center',
                  gap: 4,
                  background: 'none',
                  border: 'none',
                  padding: '4px 0',
                  cursor: 'pointer',
                  fontSize: 11,
                  color: 'var(--hr-text-dim)',
                }}
                aria-label={t('hrAttendance.pinModal.setPin')}
              >
                <Settings size={11} />
                {t('hrAttendance.pinModal.setPin')}
              </button>
            )}
          </div>
        )}
      </Card>

      <PinNumpad
        open={pinModalOpen}
        onClose={() => setPinModalOpen(false)}
        mode={pinMode}
        onSubmit={handlePinSubmit}
        isPending={clockIn.isPending || setPin.isPending}
        error={pinError}
      />
    </>
  );
};

export default ClockInWidget;
