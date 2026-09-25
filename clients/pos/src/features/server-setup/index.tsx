import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Cloud, Server } from 'lucide-react';
import { toast } from 'sonner';
import { useServerStore } from '../../shared/stores/serverStore';
import { isValidServerUrl } from '../../shared/services/server/serverConfig';
import { RtylrWordmark } from '../../shared/components/RtylrWordmark';

type Variant = 'first-run' | 'settings';

interface Props {
  variant?: Variant;
}

const INPUT_CLS =
  'w-full h-11 px-3 rounded-lg bg-white/8 border border-pos-border text-pos-text outline-none focus:border-white/25';

const ServerSetup: React.FC<Props> = ({ variant = 'first-run' }) => {
  const { t } = useTranslation('server');
  const mode = useServerStore((s) => s.mode);
  const customApiUrl = useServerStore((s) => s.customApiUrl);
  const useHosted = useServerStore((s) => s.useHosted);
  const useSelfHost = useServerStore((s) => s.useSelfHost);

  const [choice, setChoice] = useState<'hosted' | 'self_host'>(
    mode === 'self_host' ? 'self_host' : 'hosted'
  );
  const [url, setUrl] = useState<string>(customApiUrl ?? '');
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);

  const submit = async (): Promise<void> => {
    setError(null);
    if (choice === 'hosted') {
      useHosted();
      if (variant === 'settings') toast.success(t('saved'));
      return;
    }
    if (!isValidServerUrl(url)) {
      setError(t('invalidUrl'));
      return;
    }
    setSaving(true);
    try {
      const okResult = await useSelfHost(url);
      if (!okResult) {
        setError(t('invalidUrl'));
        return;
      }
      if (variant === 'settings') toast.success(t('saved'));
    } catch {
      toast.error(t('saveFailed'));
    } finally {
      setSaving(false);
    }
  };

  const Choices = (
    <div className="space-y-3">
      <button
        type="button"
        onClick={() => setChoice('hosted')}
        className={`w-full flex items-start gap-3 rounded-xl border p-4 text-left transition-colors ${
          choice === 'hosted'
            ? 'border-white/40 bg-white/10'
            : 'border-pos-border bg-white/5 hover:bg-white/8'
        }`}
      >
        <Cloud className="h-5 w-5 mt-0.5 shrink-0 text-pos-text" aria-hidden="true" />
        <span>
          <span className="block text-sm font-bold text-pos-text">{t('hosted')}</span>
          <span className="block text-xs text-pos-muted">{t('hostedDesc')}</span>
        </span>
      </button>

      <button
        type="button"
        onClick={() => setChoice('self_host')}
        className={`w-full flex items-start gap-3 rounded-xl border p-4 text-left transition-colors ${
          choice === 'self_host'
            ? 'border-white/40 bg-white/10'
            : 'border-pos-border bg-white/5 hover:bg-white/8'
        }`}
      >
        <Server className="h-5 w-5 mt-0.5 shrink-0 text-pos-text" aria-hidden="true" />
        <span>
          <span className="block text-sm font-bold text-pos-text">{t('selfHost')}</span>
          <span className="block text-xs text-pos-muted">{t('selfHostDesc')}</span>
        </span>
      </button>

      {choice === 'self_host' && (
        <div>
          <label className="text-xs text-pos-muted" htmlFor="server-url">
            {t('serverUrlLabel')}
          </label>
          <input
            id="server-url"
            type="url"
            inputMode="url"
            autoComplete="off"
            value={url}
            placeholder={t('serverUrlPlaceholder')}
            onChange={(e) => setUrl(e.target.value)}
            className={INPUT_CLS}
          />
          {error && <p className="mt-1 text-xs text-red-400">{error}</p>}
        </div>
      )}
    </div>
  );

  if (variant === 'settings') {
    return (
      <div className="space-y-3">
        <p className="text-xs text-pos-muted">
          {mode === 'self_host'
            ? t('currentSelfHost', { url: customApiUrl ?? '' })
            : t('currentHosted')}
        </p>
        {Choices}
        <button
          type="button"
          onClick={() => void submit()}
          disabled={saving}
          className="h-10 px-4 rounded-lg bg-pos-text text-pos-bg text-sm font-bold disabled:opacity-50"
        >
          {saving ? t('saving') : t('save')}
        </button>
      </div>
    );
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-pos-bg p-6">
      <div className="w-full max-w-md space-y-6">
        <div className="text-center space-y-2">
          <RtylrWordmark className="text-3xl justify-center" suffix="POS" />
          <h1 className="text-lg font-bold text-pos-text">{t('title')}</h1>
          <p className="text-sm text-pos-muted">{t('subtitle')}</p>
        </div>
        {Choices}
        <button
          type="button"
          onClick={() => void submit()}
          disabled={saving}
          className="w-full h-11 rounded-lg bg-pos-text text-pos-bg text-sm font-bold disabled:opacity-50"
        >
          {saving ? t('saving') : t('continue')}
        </button>
      </div>
    </div>
  );
};

export default ServerSetup;
