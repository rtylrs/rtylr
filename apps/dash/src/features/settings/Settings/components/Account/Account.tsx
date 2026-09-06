import React, { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { HelmetContainer } from '@voxire/components';
import { signOut } from '@voxire/auth';
import { keepPreviousData, useMutation, useQuery } from '@tanstack/react-query';
import type {
  LoginEntry as ApiLoginEntry,
} from '@voxire/voxie/hub/security/v1/types';
import { toast } from 'sonner';
import {
  CheckCircle2,
  ChevronLeft,
  ChevronRight,
  History,
  Laptop,
  Monitor,
  Smartphone,
  Tablet,
  XCircle,
} from 'lucide-react';
import { EmptyState } from '@voxire/ui';
import { securityClient, userClient } from '../../../../../shared/lib/clients';
import { useUserStore } from '../../../../../shared/stores/userStore';
import { ConfirmDialog } from '../../../../../shared/components/dash';
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardDescription,
  Input,
  Label,
  Button,
  Table,
  TableHeader,
  TableBody,
  TableRow,
  TableHead,
  TableCell,
} from '../../../../../shared/components/ui';

function formatTimeAgo(dateStr: string): string {
  const now = Date.now();
  const then = new Date(dateStr).getTime();
  const diff = now - then;
  const mins = Math.floor(diff / 60000);
  if (mins < 1) return 'just now';
  if (mins < 60) return `${mins}m ago`;
  const hours = Math.floor(mins / 60);
  if (hours < 24) return `${hours}h ago`;
  const days = Math.floor(hours / 24);
  if (days < 7) return `${days}d ago`;
  const weeks = Math.floor(days / 7);
  return `${weeks}w ago`;
}

function formatDateTime(dateStr: string): string {
  const date = new Date(dateStr);
  return Number.isNaN(date.getTime()) ? '' : date.toLocaleString();
}

type DeviceKind = 'mobile' | 'tablet' | 'desktop' | 'unknown';

interface DeviceDetails {
  kind: DeviceKind;
  label: string;
  description: string;
}

function getDeviceDetails(userAgent: string): DeviceDetails {
  const ua = userAgent.toLowerCase();
  if (!ua) {
    return {
      kind: 'unknown',
      label: 'Unknown device',
      description: 'Device information unavailable',
    };
  }

  const kind: DeviceKind = /ipad|tablet/.test(ua)
    ? 'tablet'
    : /iphone|android.*mobile|mobile/.test(ua)
      ? 'mobile'
      : 'desktop';

  const browser = /edg\//.test(ua)
    ? 'Edge'
    : /opr\//.test(ua)
      ? 'Opera'
      : /electron/.test(ua)
        ? 'Electron'
        : /firefox|fxios/.test(ua)
          ? 'Firefox'
          : /chrome|crios/.test(ua)
            ? 'Chrome'
            : /safari/.test(ua)
              ? 'Safari'
              : 'Browser';

  const platform = /iphone/.test(ua)
    ? 'iPhone'
    : /ipad/.test(ua)
      ? 'iPad'
      : /android/.test(ua)
        ? 'Android'
        : /mac os|macintosh/.test(ua)
          ? 'macOS'
          : /windows/.test(ua)
            ? 'Windows'
            : /linux/.test(ua)
              ? 'Linux'
              : '';

  return {
    kind,
    label: platform ? `${browser} on ${platform}` : browser,
    description: kind[0].toUpperCase() + kind.slice(1),
  };
}

function getDeviceIcon(kind: DeviceKind) {
  switch (kind) {
    case 'mobile':
      return Smartphone;
    case 'tablet':
      return Tablet;
    case 'desktop':
      return Laptop;
    default:
      return Monitor;
  }
}

function formatMethod(method: string): string {
  if (!method) return 'Unknown';
  return method.charAt(0).toUpperCase() + method.slice(1);
}

interface LoginHistoryItem {
  id: string;
  email: string;
  method: string;
  ip: string;
  userAgent: string;
  success: boolean;
  createdAt: string;
  timeAgo: string;
  device: DeviceDetails;
}

interface LoginHistoryPage {
  entries: LoginHistoryItem[];
  total: number;
  page: number;
  pageSize: number;
  hasMore: boolean;
}

const LOGIN_HISTORY_PAGE_SIZE = 20;

const Account: React.FC = () => {
  const { t } = useTranslation();
  const { user } = useUserStore();
  const [formData, setFormData] = useState({ name: '', email: '' });
  const [isEditing, setIsEditing] = useState(false);
  const [confirmInvalidateOpen, setConfirmInvalidateOpen] = useState(false);
  const [loginHistoryPage, setLoginHistoryPage] = useState(1);

  const {
    data: loginHistory,
    isLoading: isLoginHistoryLoading,
    isError: isLoginHistoryError,
    isFetching: isLoginHistoryFetching,
    refetch: refetchLoginHistory,
  } = useQuery<LoginHistoryPage>({
    queryKey: ['security', 'login-history', { page: loginHistoryPage }],
    placeholderData: keepPreviousData,
    queryFn: async () => {
      const response = await securityClient.loginHistory({
        page: loginHistoryPage,
        pageSize: LOGIN_HISTORY_PAGE_SIZE,
      });
      const entries = (response.history ?? []).map((entry: ApiLoginEntry, index) => {
        const createdAt = entry.createdAt ?? '';
        const userAgent = entry.userAgent ?? '';
        return {
          id: entry.id ?? `${createdAt}-${index}`,
          email: entry.email ?? '—',
          method: formatMethod(entry.method ?? ''),
          ip: entry.ip ?? '',
          userAgent,
          // Generated Go booleans use `omitempty`, so a persisted false is
          // absent on the wire. Treat missing as failed rather than hiding a
          // failed sign-in as successful.
          success: entry.success ?? false,
          createdAt,
          timeAgo: createdAt ? formatTimeAgo(createdAt) : '—',
          device: getDeviceDetails(userAgent),
        };
      });

      return {
        entries,
        total: response.total ?? entries.length,
        page: response.page ?? loginHistoryPage,
        pageSize: response.pageSize ?? LOGIN_HISTORY_PAGE_SIZE,
        hasMore: response.hasMore ?? false,
      };
    },
  });

  const loginHistoryData: LoginHistoryPage = loginHistory ?? {
    entries: [],
    total: 0,
    page: loginHistoryPage,
    pageSize: LOGIN_HISTORY_PAGE_SIZE,
    hasMore: false,
  };

  useEffect(() => {
    if (user) {
      setFormData({
        name: user.fullName || '',
        email: user.email || '',
      });
    }
  }, [user]);

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setFormData((prev) => ({ ...prev, [name]: value }));
    setIsEditing(true);
  };

  const updateProfile = useMutation({
    mutationFn: () => userClient.updateProfile({ name: formData.name }),
    onSuccess: () => {
      toast.success('Profile saved');
      setIsEditing(false);
    },
    onError: () => toast.error('Failed to save profile'),
  });

  const invalidateSessions = useMutation({
    mutationFn: () => securityClient.invalidateSessions({}),
    onSuccess: () => signOut('dash'),
    onError: () => toast.error('Failed to invalidate sessions'),
  });

  const handleSave = (): void => {
    updateProfile.mutate();
  };

  const handleCancel = () => {
    if (user) {
      setFormData({
        name: user.fullName || '',
        email: user.email || '',
      });
    }
    setIsEditing(false);
  };

  return (
    <>
      <HelmetContainer title="Account · rtylr" />
      <div className="space-y-6">
        <div>
          <h1 className="text-xl font-bold text-primary">Account</h1>
          <p className="text-sm text-gray-400 mt-0.5">Manage your personal account information</p>
        </div>

        {/* Profile */}
        <Card>
          <CardHeader>
            <CardTitle>Profile</CardTitle>
            <CardDescription>Update your personal details</CardDescription>
          </CardHeader>
          <CardContent className="pt-0 space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-1.5">
                <Label htmlFor="name">Full name</Label>
                <Input id="name" name="name" value={formData.name} onChange={handleInputChange} />
              </div>
              <div className="space-y-1.5">
                <Label htmlFor="email">Email</Label>
                <Input
                  id="email"
                  name="email"
                  type="email"
                  value={formData.email}
                  disabled
                  readOnly
                />
              </div>
            </div>
            <div className="flex justify-end gap-2 pt-1">
              <Button
                variant="outline"
                onClick={handleCancel}
                disabled={!isEditing || updateProfile.isPending}
              >
                Cancel
              </Button>
              <Button onClick={handleSave} disabled={!isEditing || updateProfile.isPending}>
                {updateProfile.isPending ? 'Saving…' : 'Save changes'}
              </Button>
            </div>
          </CardContent>
        </Card>

        {/* Security */}
        <Card>
          <CardHeader>
            <CardTitle>Security</CardTitle>
            <CardDescription>Manage your account sessions</CardDescription>
          </CardHeader>
          <CardContent className="pt-0">
            <p className="text-sm text-gray-500 mb-3">
              Invalidate all sessions for your account. This will log you out everywhere.
            </p>
            <Button variant="destructive" size="sm" onClick={() => setConfirmInvalidateOpen(true)}>
              Invalidate all sessions
            </Button>
          </CardContent>
        </Card>

        <ConfirmDialog
          open={confirmInvalidateOpen}
          onOpenChange={setConfirmInvalidateOpen}
          title="Invalidate all sessions?"
          description="This will sign you out of Dash and revoke every active session for your account."
          confirmLabel="Invalidate sessions"
          destructive
          isLoading={invalidateSessions.isPending}
          onConfirm={async () => {
            await invalidateSessions.mutateAsync();
          }}
        />

        {/* Login History */}
        <Card>
          <CardHeader className="pb-4">
            <CardTitle>Login history</CardTitle>
            <CardDescription>
              {loginHistoryData.total > 0
                ? `${loginHistoryData.total} sign-in${loginHistoryData.total === 1 ? '' : 's'} recorded for your account`
                : 'Recent login activity for your account'}
            </CardDescription>
          </CardHeader>
          <CardContent className="pt-0">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Email</TableHead>
                  <TableHead>Device</TableHead>
                  <TableHead>Method</TableHead>
                  <TableHead>IP Address</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Time</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {loginHistoryData.entries.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={6} className="py-10 text-center text-gray-500">
                      {isLoginHistoryLoading ? (
                        'Loading login history…'
                      ) : isLoginHistoryError ? (
                        'Unable to load login history.'
                      ) : (
                        <EmptyState
                          icon={History}
                          title={t('settings.account.noLoginActivityTitle')}
                          description={t('settings.account.noLoginActivityDescription')}
                          action={{
                            label: t('common.refresh'),
                            onClick: () => void refetchLoginHistory(),
                          }}
                          size="compact"
                        />
                      )}
                    </TableCell>
                  </TableRow>
                ) : (
                  loginHistoryData.entries.map((item) => {
                    const DeviceIcon = getDeviceIcon(item.device.kind);
                    return (
                      <TableRow key={item.id}>
                        <TableCell className="whitespace-nowrap">{item.email}</TableCell>
                        <TableCell>
                          <div
                            className="flex items-center gap-2.5 min-w-[190px]"
                            title={item.userAgent || undefined}
                          >
                            <div className="h-8 w-8 rounded-lg bg-black/5 dark:bg-white/10 flex items-center justify-center shrink-0">
                              <DeviceIcon size={15} style={{ color: 'var(--dash-text-muted)' }} />
                            </div>
                            <div className="min-w-0">
                              <p
                                className="font-medium truncate"
                                style={{ color: 'var(--dash-text)' }}
                              >
                                {item.device.label}
                              </p>
                              <p className="text-xs" style={{ color: 'var(--dash-text-dim)' }}>
                                {item.device.description}
                              </p>
                            </div>
                          </div>
                        </TableCell>
                        <TableCell
                          className="whitespace-nowrap"
                          style={{ color: 'var(--dash-text-muted)' }}
                        >
                          {item.method}
                        </TableCell>
                        <TableCell
                          className="whitespace-nowrap"
                          style={{ color: 'var(--dash-text-muted)' }}
                        >
                          {item.ip || '—'}
                        </TableCell>
                        <TableCell className="whitespace-nowrap">
                          <span
                            className={`inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-medium ${
                              item.success
                                ? 'bg-emerald-500/15 text-emerald-600 dark:text-emerald-400'
                                : 'bg-red-500/15 text-red-600 dark:text-red-400'
                            }`}
                          >
                            {item.success ? <CheckCircle2 size={12} /> : <XCircle size={12} />}
                            {item.success ? 'Successful' : 'Failed'}
                          </span>
                        </TableCell>
                        <TableCell
                          className="text-gray-400 whitespace-nowrap"
                          title={formatDateTime(item.createdAt)}
                        >
                          {item.timeAgo}
                        </TableCell>
                      </TableRow>
                    );
                  })
                )}
              </TableBody>
            </Table>

            {loginHistoryData.total > 0 && (
              <div className="flex items-center justify-between gap-3 border-t border-gray-100 pt-4 mt-4">
                <p className="text-xs text-gray-400">
                  Showing {(loginHistoryData.page - 1) * loginHistoryData.pageSize + 1}
                  {'–'}
                  {Math.min(
                    loginHistoryData.page * loginHistoryData.pageSize,
                    loginHistoryData.total
                  )}{' '}
                  of {loginHistoryData.total}
                </p>
                <div className="flex items-center gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setLoginHistoryPage((page) => Math.max(1, page - 1))}
                    disabled={loginHistoryPage === 1 || isLoginHistoryFetching}
                  >
                    <ChevronLeft size={14} className="rtl:rotate-180" />
                    Previous
                  </Button>
                  <span className="text-xs text-gray-500 whitespace-nowrap">
                    Page {loginHistoryData.page}
                  </span>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setLoginHistoryPage((page) => page + 1)}
                    disabled={!loginHistoryData.hasMore || isLoginHistoryFetching}
                  >
                    Next
                    <ChevronRight size={14} className="rtl:rotate-180" />
                  </Button>
                </div>
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </>
  );
};

export default Account;
