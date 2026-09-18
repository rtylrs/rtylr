import './index.scss';

// Existing components
export { Button } from './Button';
export { default as Input } from './Input';
export { default as Modal } from './Modal';
export { default as LoadingSpinner } from './LoadingSpinner';
export { default as FullScreenLoader } from './FullScreenLoader';
export { default as AnimatedRoutes } from './AnimatedRoutes';
export { default as Icon } from './Icon';
export { default as Select } from './Select';
export { HelmetContainer } from './HelmetContainer';

// New components (SCSS modules, replaces MUI)
export { Switch } from './Switch';
export { Spinner } from './Spinner';
export { Alert } from './Alert';
export { Chip } from './Chip';

// Self-host license notice (admin-only; no-op for the hosted product)
export { LicenseBanner, useLicenseState, resolveLicenseNotice, isLicenseAdmin } from './license';
export type {
  LicenseBannerProps,
  LicenseNotice,
  LicenseNoticeKind,
  LicenseNoticeSeverity,
  LicenseState,
  LicenseStatus,
  LicenseMode,
} from './license';
