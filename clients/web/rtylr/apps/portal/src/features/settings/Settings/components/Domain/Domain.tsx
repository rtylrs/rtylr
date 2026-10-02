import { PageHeader } from '../../../../../shared/components/dash/PageHeader';
import React, { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { HelmetContainer } from '@voxire/components';
import { Globe, Copy, Check, ChevronDown, ChevronUp, ShieldCheck, AlertCircle } from 'lucide-react';
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

import {
  useCustomDomain,
  useRegisterCustomDomain,
} from '../../../../../shared/hooks/queries/useBusiness';
import { toast } from 'sonner';

const DOMAIN_REGEX = /^(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,63}$/;

const Domain: React.FC = () => {
  const { t } = useTranslation();
  const user = useUserStore((s) => s.user);
  const websiteData = useConfigStore((s) => s.websiteData) as { slug?: string } | undefined;
  const { data: registeredDomain } = useCustomDomain();
  const registerDomain = useRegisterCustomDomain();
  const fallbackSlug =
    user?.businessName
      ?.toLowerCase()
      .replace(/\s+/g, '-')
      .replace(/[^a-z0-9-]/g, '') ?? 'your-business';
  const slug = websiteData?.slug || fallbackSlug;
  const menuUrl = `menu.rtylr.com/${slug}`;

  const configuredDomain = registeredDomain || '';
  const isConfigured = Boolean(configuredDomain && DOMAIN_REGEX.test(configuredDomain));

  const [copied, setCopied] = useState(false);
  const [dnsOpen, setDnsOpen] = useState(true);
  const [customDomainInput, setCustomDomainInput] = useState('');

  // The registered domain arrives with the query, after the input mounted.
  useEffect(() => {
    if (registeredDomain) setCustomDomainInput(registeredDomain);
  }, [registeredDomain]);

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(menuUrl);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      const el = document.createElement('input');
      el.value = menuUrl;
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
            <CardDescription>
              This is the default public URL for your menu, always available.
            </CardDescription>
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
                  className="pl-8 cursor-default select-all"
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
                {copied ? 'Copied' : 'Copy'}
              </Button>
            </div>
          </CardContent>
        </Card>

        {/* Section 2, Custom Domain */}
        <div>
          <Card>
            <CardHeader>
              <div className="flex items-start justify-between">
                <div>
                  <CardTitle>{t('settings.domain.customTitle')}</CardTitle>
                  <CardDescription>{t('settings.domain.customDescription')}</CardDescription>
                </div>
                <Badge variant={isConfigured ? 'default' : 'outline'}>
                  {isConfigured ? 'Configured' : 'Not configured'}
                </Badge>
              </div>
            </CardHeader>
            <CardContent className="pt-0 space-y-4">
              {/* Domain input + connect button */}
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
                    Connect Domain
                  </Button>
                </div>
              </div>

              <Separator />

              {/* DNS Instructions, collapsible */}
              <div
                className="rounded-xl overflow-hidden"
                style={{ border: '1px solid var(--dash-border)' }}
              >
                {/* Collapsible header */}
                <button
                  type="button"
                  onClick={() => setDnsOpen((o) => !o)}
                  className="w-full flex items-center justify-between px-4 py-3 text-sm font-medium transition-colors hover:bg-[var(--dash-hover)]"
                  style={{ color: 'var(--dash-text)' }}
                >
                  <span>{t('settings.domain.dnsTitle')}</span>
                  {dnsOpen ? <ChevronUp size={15} /> : <ChevronDown size={15} />}
                </button>

                {dnsOpen && (
                  <div
                    className="px-4 pb-4 space-y-3"
                    style={{ borderTop: '1px solid var(--dash-border)' }}
                  >
                    <p className="text-sm pt-3" style={{ color: 'var(--dash-text-muted)' }}>
                      Add the following CNAME record in your DNS provider's dashboard (e.g.
                      Cloudflare, Namecheap, GoDaddy):
                    </p>

                    {/* DNS table */}
                    <div
                      className="rounded-lg overflow-hidden text-xs"
                      style={{ border: '1px solid var(--dash-border)' }}
                    >
                      <table className="w-full">
                        <thead>
                          <tr style={{ background: 'var(--dash-hover)' }}>
                            {['Type', 'Name', 'Value', 'TTL'].map((h) => (
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
                          <tr style={{ background: 'var(--dash-card)' }}>
                            <td
                              className="px-3 py-2.5 font-mono"
                              style={{ color: 'var(--dash-text)' }}
                            >
                              CNAME
                            </td>
                            <td
                              className="px-3 py-2.5 font-mono"
                              style={{ color: 'var(--dash-text)' }}
                            >
                              menu
                            </td>
                            <td
                              className="px-3 py-2.5 font-mono"
                              style={{ color: 'var(--dash-text)' }}
                            >
                              cname.rtylr.com
                            </td>
                            <td
                              className="px-3 py-2.5 font-mono"
                              style={{ color: 'var(--dash-text-muted)' }}
                            >
                              3600
                            </td>
                          </tr>
                        </tbody>
                      </table>
                    </div>

                    {/* Propagation notice */}
                    <div
                      className="flex items-start gap-2 px-3 py-2.5 rounded-lg text-xs"
                      style={{
                        background: 'rgba(251,191,36,0.08)',
                        border: '1px solid rgba(251,191,36,0.2)',
                        color: 'var(--dash-text-muted)',
                      }}
                    >
                      <AlertCircle size={13} className="shrink-0 mt-0.5 text-amber-500" />
                      <span>{t('settings.domain.dnsPropagation')}</span>
                    </div>
                  </div>
                )}
              </div>

              {/* SSL notice */}
              <div
                className="flex items-start gap-2 px-3 py-2.5 rounded-lg text-xs"
                style={{
                  background: 'rgba(16,185,129,0.07)',
                  border: '1px solid rgba(16,185,129,0.18)',
                  color: 'var(--dash-text-muted)',
                }}
              >
                <ShieldCheck size={13} className="shrink-0 mt-0.5 text-emerald-500" />
                <span>
                  SSL is automatically provisioned once your domain is verified. No manual
                  certificate setup required.
                </span>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </>
  );
};

export default Domain;
