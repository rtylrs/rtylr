import React, { useState } from 'react';
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

import { useUpdateBusiness } from '../../../../../shared/hooks/queries/useBusiness';
import { toast } from 'sonner';

const DOMAIN_REGEX = /^(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,63}$/;

const Domain: React.FC = () => {
  const user = useUserStore((s) => s.user);
  const websiteData = useConfigStore((s) => s.websiteData) as
    | { slug?: string; domain?: string; customDomain?: string }
    | undefined;
  const updateBusiness = useUpdateBusiness();
  const fallbackSlug =
    user?.businessName
      ?.toLowerCase()
      .replace(/\s+/g, '-')
      .replace(/[^a-z0-9-]/g, '') ?? 'your-business';
  const slug = websiteData?.slug || fallbackSlug;
  const menuUrl = `menu.rtylr.com/${slug}`;

  const configuredDomain = websiteData?.customDomain || websiteData?.domain;
  const isConfigured = Boolean(configuredDomain && DOMAIN_REGEX.test(configuredDomain));

  const [copied, setCopied] = useState(false);
  const [dnsOpen, setDnsOpen] = useState(true);
  const [customDomainInput, setCustomDomainInput] = useState(configuredDomain || '');

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
      toast.error('Please enter a valid domain (e.g. menu.yourbrand.com)');
      return;
    }
    updateBusiness.mutate(
      {
        data: { customDomain: domain, website: domain },
      },
      {
        onSuccess: () => {
          toast.success('Custom domain saved. DNS propagation may take up to 24 hours.');
        },
      }
    );
  };

  return (
    <>
      <HelmetContainer title="Domain · rtylr" />
      <div className="space-y-6">
        {/* Header */}
        <div>
          <h2 className="text-xl font-bold" style={{ color: 'var(--dash-text)' }}>
            Custom Domain
          </h2>
          <p className="text-sm mt-0.5" style={{ color: 'var(--dash-text-muted)' }}>
            Connect a custom domain to your customer-facing menu.
          </p>
        </div>

        {/* Section 1, Your menu URL */}
        <Card>
          <CardHeader>
            <CardTitle>Your menu URL</CardTitle>
            <CardDescription>
              This is the default public URL for your menu, always available.
            </CardDescription>
          </CardHeader>
          <CardContent className="pt-0">
            <Label className="mb-1.5 block">Menu link</Label>
            <div className="flex items-center gap-2">
              <div className="flex-1 relative">
                <Globe
                  size={14}
                  className="absolute left-3 top-1/2 -translate-y-1/2"
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
                  <CardTitle>Custom Domain</CardTitle>
                  <CardDescription>Point your own domain at your rtylr menu.</CardDescription>
                </div>
                <Badge variant={isConfigured ? 'default' : 'outline'}>
                  {isConfigured ? 'Configured' : 'Not configured'}
                </Badge>
              </div>
            </CardHeader>
            <CardContent className="pt-0 space-y-4">
              {/* Domain input + connect button */}
              <div className="space-y-1.5">
                <Label htmlFor="custom-domain">Domain</Label>
                <div className="flex items-center gap-2">
                  <Input
                    id="custom-domain"
                    placeholder="menu.mybusiness.com"
                    value={customDomainInput}
                    onChange={(e) => setCustomDomainInput(e.target.value)}
                    className="flex-1"
                  />
                  <Button size="sm" onClick={handleConnectDomain} className="shrink-0">
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
                  <span>DNS Setup Instructions</span>
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
                                className="px-3 py-2 text-left font-semibold"
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
                            <td className="px-3 py-2.5 font-mono" style={{ color: '#4583C4' }}>
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
                      <span>DNS changes can take up to 48 hours to propagate.</span>
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
