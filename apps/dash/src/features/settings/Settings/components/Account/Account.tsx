import React, { useState, useEffect } from 'react';
import { signOut } from '@voxire/auth';
import { useMutation } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import { toast } from 'sonner';
import { safe } from '../../../../../shared/lib/api';
import { useUserStore } from '../../../../../shared/stores/userStore';
import { ConfirmDialog } from '../../../../../shared/components/dash';

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

interface LoginHistoryItem {
  email: string;
  method: string;
  ip: string;
  timeAgo: string;
}

const Account: React.FC = () => {
  const { user } = useUserStore();
  const [formData, setFormData] = useState({ name: '', email: '' });
  const [isEditing, setIsEditing] = useState(false);
  const [loginHistory, setLoginHistory] = useState<LoginHistoryItem[]>([]);
  const [confirmInvalidateOpen, setConfirmInvalidateOpen] = useState(false);

  useEffect(() => {
    void unwrap(
      safe.get<{
        history?: Array<{ email: string; method: string; ip?: string; createdAt: string }>;
      }>('/security/login-history')
    )
      .then((res) => {
        const history = res.history ?? [];
        setLoginHistory(
          history.map((h) => ({
            email: h.email,
            method: h.method === 'email' ? 'Logged in via email' : `Logged in via ${h.method}`,
            ip: h.ip ? `from ${h.ip}` : '',
            timeAgo: formatTimeAgo(h.createdAt),
          }))
        );
      })
      .catch(() => {});
  }, []);

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
    mutationFn: () =>
      unwrap(
        safe.patch('/account/profile', {
          name: formData.name,
        })
      ),
    onSuccess: () => {
      toast.success('Profile saved');
      setIsEditing(false);
    },
    onError: () => toast.error('Failed to save profile'),
  });

  const invalidateSessions = useMutation({
    mutationFn: () => unwrap(safe.post('/security/invalidate-sessions')),
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
        <CardHeader>
          <CardTitle>Login history</CardTitle>
          <CardDescription>Recent login activity for your account</CardDescription>
        </CardHeader>
        <CardContent className="pt-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Email</TableHead>
                <TableHead>Method</TableHead>
                <TableHead>IP Address</TableHead>
                <TableHead>Time</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {loginHistory.map((item, i) => (
                <TableRow key={i}>
                  <TableCell>{item.email}</TableCell>
                  <TableCell className="text-gray-500">{item.method}</TableCell>
                  <TableCell className="text-gray-500">{item.ip}</TableCell>
                  <TableCell className="text-gray-400">{item.timeAgo}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  );
};

export default Account;
