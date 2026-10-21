import { PageHeader } from '../../../../../shared/components/dash/PageHeader';
import React, { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { HelmetContainer } from '@voxire/components';
import { Globe, Copy, Check, AlertCircle } from 'lucide-react';
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardDescription,
  Input,
  Label,
  Button,
  Badge,
  Separator,
} from '../../../../../shared/components/ui';
import { cn } from '../../../../../shared/lib/utils';
import { useUserStore } from '../../../../../shared/stores/userStore';
import { useConfigStore } from '../../../../../shared/stores/configStore';
import { PORTAL_URL, buildPublicMenuUrl } from '@voxire/config';

import {
  useCustomDomain,
  useRegisterCustomDomain,
  useRemoveCustomDomain,
  useVerifyCustomDomain,
} from '../../../../../shared/hooks/queries/useBusiness';
import { toast } from 'sonner';

const DOMAIN_REGEX = /^(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,63}$/;
const DNS_TTL = '3600';
const PORTAL_HOST = new URL(PORTAL_URL).hostname;

interface DnsRecord {
  type: string;
  name: string;
  value: string;
}

const CopyValue: React.FC<{ value: string }> = ({ value }) => {
  const { t } = useTranslation();
  const [copied, setCopied] = useState(false);
  const copy = async () => {
    try {
      await navigator.clipboard.writeText(value);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      toast.error(t('common.copyFailed', 'Copy failed'));
    }
  };
  return (
    <button
      type="button"
      onClick={copy}
      aria-label={t('settings.domain.copy')}
      className="inline-flex items-center gap-1.5 font-mono break-all text-start"
      style={{ color: 'var(--dash-text)' }}
    >
      <span>{value}</span>
      {copied ? <Check size={12} className="shrink-0" /> : <Copy size={12} className="shrink-0" />}
    </button>
  );
};

const DnsTable: React.FC<{ records: DnsRecord[] }> = ({ records }) => {
  const { t } = useTranslation();
  const headers = [
    t('settings.domain.colType'),
    t('settings.domain.colName'),
    t('settings.domain.colValue'),
    t('settings.domain.colTtl'),
  ];
  return (
    <div
      className="rounded-lg overflow-x-auto text-xs"
      style={{ border: '1px solid var(--dash-border)' }}
    >
      <table className="w-full">
        <thead>
          <tr style={{ background: 'var(--dash-hover)' }}>
            {headers.map((h) => (
              <th
                key={h}
                className="px-3 py-2 text-start font-semibold"
                style={{
                  color: 'var(--dash-text-muted)',
                  borderBottom: '1px solid var(--dash-border)',
                }}
              >
                {h}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {records.map((record) => (
            <tr key={`${record.type}-${record.name}`} style={{ background: 'var(--dash-card)' }}>
              <td className="px-3 py-2.5 font-mono" style={{ color: 'var(--dash-text)' }}>
                {record.type}
              </td>
              <td className="px-3 py-2.5">
                <CopyValue value={record.name} />
              </td>
              <td className="px-3 py-2.5">
                <CopyValue value={record.value} />
              </td>
              <td className="px-3 py-2.5 font-mono" style={{ color: 'var(--dash-text-muted)' }}>
                {DNS_TTL}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
};

const Domain: React.FC = () => {
  const { t } = useTranslation();
  const user = useUserStore((s) => s.user);
  const websiteData = useConfigStore((s) => s.websiteData) as { slug?: string } | undefined;
  const { data: customDomain } = useCustomDomain();
  const registerDomain = useRegisterCustomDomain();
  const verifyDomain = useVerifyCustomDomain();
  const removeDomain = useRemoveCustomDomain();
  const fallbackSlug =
    user?.businessName
      ?.toLowerCase()
      .replace(/\s+/g, '-')
      .replace(/[^a-z0-9-]/g, '') ?? 'your-business';
  const slug = websiteData?.slug || fallbackSlug;
  const publicMenuUrl = buildPublicMenuUrl(slug);
  // The field shows the host-and-path form, without the scheme, as before.
  const menuUrl = publicMenuUrl.replace(/^https?:\/\//, '');

  const registeredDomain = customDomain?.domain ?? '';
  const isVerified = customDomain?.status === 'verified';
  const verificationToken = customDomain?.verificationToken ?? '';

  const [copied, setCopied] = useState(false);
  const [customDomainInput, setCustomDomainInput] = useState('');

  // The registered domain arrives with the query, after the input mounted.
  useEffect(() => {
    if (registeredDomain) setCustomDomainInput(registeredDomain);
  }, [registeredDomain]);

  const handleCopy = async () => {
    try {
      // Copy the full URL so it can be pasted straight into a browser.
      await navigator.clipboard.writeText(publicMenuUrl);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      const el = document.createElement('input');
      el.value = publicMenuUrl;
      document.body.appendChild(el);
      el.select();
      document.execCommand('copy');
      document.body.removeChild(el);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  };

  const handleConnectDomain = () => {
    const domain = customDomainInput.trim().toLowerCase();
    if (!domain || !DOMAIN_REGEX.test(domain)) {
      toast.error(t('settings.domain.invalid'));
      return;
    }
    registerDomain.mutate(domain);
  };

  const statusLabel = !registeredDomain
    ? t('settings.domain.notConfigured')
    : isVerified
      ? t('settings.domain.statusVerified')
      : t('settings.domain.statusPending');

  return (
    <>
      <HelmetContainer title={t('settings.domain.pageTitle')} />
      <div className="portal-management-page space-y-6">
        <PageHeader
          title={t('common.nav.customDomain')}
          description={t('common.management.domainDescription')}
        />

        {/* Section 1, Your menu URL */}
        <Card>
          <CardHeader>
            <CardTitle>{t('settings.domain.menuUrlTitle')}</CardTitle>
            <CardDescription>{t('settings.domain.menuUrlDescription')}</CardDescription>
          </CardHeader>
          <CardContent className="pt-0">
            <Label className="mb-1.5 block">{t('settings.domain.menuLink')}</Label>
            <div className="flex items-center gap-2">
              <div className="flex-1 relative">
                <Globe
                  size={14}
                  className="absolute start-3 top-1/2 -translate-y-1/2"
                  style={{ color: 'var(--dash-text-dim)' }}
                />
                <Input
                  readOnly
                  value={menuUrl}
                  className="ps-8 cursor-default select-all"
                  style={{ background: 'var(--dash-hover)', color: 'var(--dash-text-muted)' }}
                />
              </div>
              <Button
                variant="outline"
                size="sm"
                onClick={handleCopy}
                className={cn('shrink-0 gap-1.5 transition-all', copied && 'opacity-80')}
              >
                {copied ? <Check size={13} /> : <Copy size={13} />}
                {copied ? t('settings.domain.copied') : t('settings.domain.copy')}
              </Button>
            </div>
          </CardContent>
        </Card>

        {/* Section 2, Custom Domain */}
        <Card>
          <CardHeader>
            <div className="flex items-start justify-between">
              <div>
                <CardTitle>{t('settings.domain.customTitle')}</CardTitle>
                <CardDescription>{t('settings.domain.customDescription')}</CardDescription>
              </div>
              <Badge variant={isVerified ? 'default' : 'outline'}>{statusLabel}</Badge>
            </div>
          </CardHeader>
          <CardContent className="pt-0 space-y-4">
            <div className="space-y-1.5">
              <Label htmlFor="custom-domain">{t('settings.domain.domainLabel')}</Label>
              <div className="flex items-center gap-2">
                <Input
                  id="custom-domain"
                  placeholder={t('settings.domain.domainPlaceholder')}
                  value={customDomainInput}
                  onChange={(e) => setCustomDomainInput(e.target.value)}
                  className="flex-1"
                />
                <Button
                  size="sm"
                  onClick={handleConnectDomain}
                  disabled={registerDomain.isPending}
                  className="shrink-0"
                >
                  {registeredDomain
                    ? t('settings.domain.updateDomain')
                    : t('settings.domain.connectDomain')}
                </Button>
              </div>
            </div>

            {registeredDomain && (
              <>
                <Separator />

                {!isVerified && verificationToken && (
                  <div className="space-y-3">
                    <h3 className="text-sm font-medium" style={{ color: 'var(--dash-text)' }}>
                      {t('settings.domain.verifyTitle')}
                    </h3>
                    <p className="text-sm" style={{ color: 'var(--dash-text-muted)' }}>
                      {t('settings.domain.verifyIntro')}
                    </p>
                    <DnsTable
                      records={[
                        {
                          type: 'TXT',
                          name: `_rtylr-challenge.${registeredDomain}`,
                          value: verificationToken,
                        },
                      ]}
                    />
                    <div
                      className="flex items-start gap-2 px-3 py-2.5 rounded-lg text-xs"
                      style={{
                        background: 'rgba(251,191,36,0.08)',
                        border: '1px solid rgba(251,191,36,0.2)',
                        color: 'var(--dash-text-muted)',
                      }}
                    >
                      <AlertCircle size={13} className="shrink-0 mt-0.5 text-amber-500" />
                      <span>
                        {t('settings.domain.pendingNotice')} {t('settings.domain.dnsPropagation')}
                      </span>
                    </div>
                    <Button
                      size="sm"
                      onClick={() => verifyDomain.mutate(verificationToken)}
                      disabled={verifyDomain.isPending}
                    >
                      {verifyDomain.isPending
                        ? t('settings.domain.verifying')
                        : t('settings.domain.verifyButton')}
                    </Button>
                  </div>
                )}

                {isVerified && (
                  <p className="text-sm" style={{ color: 'var(--dash-text-muted)' }}>
                    {t('settings.domain.verifiedNotice')}
                  </p>
                )}

                <div className="space-y-3">
                  <h3 className="text-sm font-medium" style={{ color: 'var(--dash-text)' }}>
                    {t('settings.domain.routeTitle')}
                  </h3>
                  <p className="text-sm" style={{ color: 'var(--dash-text-muted)' }}>
                    {t('settings.domain.routeIntro')}
                  </p>
                  <DnsTable
                    records={[{ type: 'CNAME', name: registeredDomain, value: PORTAL_HOST }]}
                  />
                </div>

                <div>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => removeDomain.mutate()}
                    disabled={removeDomain.isPending}
                  >
                    {removeDomain.isPending
                      ? t('settings.domain.removing')
                      : t('settings.domain.removeButton')}
                  </Button>
                </div>
              </>
            )}
          </CardContent>
        </Card>
      </div>
    </>
  );
};

export default Domain;
