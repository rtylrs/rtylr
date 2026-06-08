import React, { useMemo, useState } from 'react';
import { useNavigate } from '@tanstack/react-router';
import { AlertCircle, CheckCircle2 } from 'lucide-react';
import Logo from '../../shared/components/Logo';
import { Button } from '@voxire/components';
import { unwrap } from '@voxire/api';
import type { ResetPasswordConfirmRequest } from '@voxire/contracts/common/auth/v1/reset_password_confirm';
import { safe } from '../../shared/lib/api';

const ResetPasswordConfirm: React.FC = () => {
  const navigate = useNavigate();
  const [password, setPassword] = useState('');
  const [confirm, setConfirm] = useState('');
  const [loading, setLoading] = useState(false);
  const [done, setDone] = useState(false);
  const [error, setError] = useState('');

  const token = useMemo(() => {
    return new URLSearchParams(window.location.search).get('token') ?? '';
  }, []);

  const mismatch = confirm.length > 0 && password !== confirm;

  const handleSubmit = async (): Promise<void> => {
    if (!token) {
      setError('Invalid reset link. Please request a new one.');
      return;
    }
    if (password !== confirm) {
      setError('Passwords do not match.');
      return;
    }
    setLoading(true);
    setError('');
    try {
      const request: ResetPasswordConfirmRequest = { token, password };
      await unwrap(safe.post('/auth/password-reset/confirm', request));
      setDone(true);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to reset password.';
      setError(msg);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen w-full flex items-center justify-center bg-white px-4">
      <div className="w-full max-w-sm">
        <div className="flex justify-center mb-8">
          <Logo variant="dark" className="h-8 w-auto select-none" />
        </div>

        <h2 className="text-xl font-semibold text-[#111] mb-1">Set a new password</h2>
        <p className="text-sm text-[#666] mb-6">Choose a strong password for your account.</p>

        {done ? (
          <div className="flex flex-col gap-4">
            <div
              className="flex items-start gap-2 rounded-lg border border-green-200 bg-green-50 px-3 py-2.5"
              role="alert"
            >
              <CheckCircle2 className="h-4 w-4 text-green-600 mt-0.5 shrink-0" />
              <p className="text-xs text-green-700 leading-relaxed">
                Password updated. You can now sign in.
              </p>
            </div>
            <Button
              type="button"
              variant="primary"
              fullWidth
              height={40}
              onClick={() => navigate({ to: '/' })}
            >
              Go to sign in
            </Button>
          </div>
        ) : (
          <form
            onSubmit={(e) => {
              e.preventDefault();
              void handleSubmit();
            }}
            className="flex flex-col gap-4"
          >
            <div className="flex flex-col gap-1.5">
              <label htmlFor="new-password" className="text-sm font-medium text-[#111]">
                New password
              </label>
              <input
                id="new-password"
                type="password"
                placeholder="At least 8 characters"
                autoComplete="new-password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
                autoFocus
                className="w-full h-10 rounded-lg border border-[#e5e5e5] bg-white px-3 text-sm text-[#111] placeholder:text-[#999] outline-none transition-colors focus:border-[#4583C4] focus:ring-2 focus:ring-[#4583C4]/20"
              />
            </div>

            <div className="flex flex-col gap-1.5">
              <label htmlFor="confirm-password" className="text-sm font-medium text-[#111]">
                Confirm password
              </label>
              <input
                id="confirm-password"
                type="password"
                placeholder="Repeat password"
                autoComplete="new-password"
                value={confirm}
                onChange={(e) => setConfirm(e.target.value)}
                required
                className="w-full h-10 rounded-lg border border-[#e5e5e5] bg-white px-3 text-sm text-[#111] placeholder:text-[#999] outline-none transition-colors focus:border-[#4583C4] focus:ring-2 focus:ring-[#4583C4]/20"
              />
              {mismatch && <p className="text-xs text-red-500">Passwords do not match</p>}
            </div>

            {error && (
              <div
                className="flex items-start gap-2 rounded-lg border border-red-200 bg-red-50 px-3 py-2.5"
                role="alert"
              >
                <AlertCircle className="h-4 w-4 text-red-500 mt-0.5 shrink-0" />
                <p className="text-xs text-red-600 leading-relaxed">{error}</p>
              </div>
            )}

            <Button
              type="submit"
              variant="primary"
              fullWidth
              loading={loading}
              disabled={!password || !!mismatch}
              height={40}
            >
              Reset password
            </Button>
          </form>
        )}
      </div>
    </div>
  );
};

export default ResetPasswordConfirm;
