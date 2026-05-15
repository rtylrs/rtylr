import React, { useState } from 'react';
import { Globe, Copy, Check, Clock } from 'lucide-react';
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardDescription,
  Input,
  Label,
  Button,
} from '../../../../../shared/components/ui';
import { cn } from '../../../../../shared/lib/utils';
import { useUserStore } from '../../../../../shared/stores/userStore';

const Domain: React.FC = () => {
  const user = useUserStore((s) => s.user);
  const subdomain = user?.businessName?.toLowerCase().replace(/\s+/g, '-') ?? 'your-business';
  const menuUrl = `${subdomain}.rtylr.com`;

  const [copied, setCopied] = useState(false);

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

  return (
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

      {/* Coming Soon Banner */}
      <div
        className="flex items-center gap-3 rounded-xl px-4 py-3.5"
        style={{
          background: 'rgba(69,131,196,0.08)',
          border: '1px solid rgba(69,131,196,0.2)',
        }}
      >
        <Clock size={15} style={{ color: '#4583C4', flexShrink: 0 }} />
        <div>
          <p className="text-sm font-semibold" style={{ color: '#4583C4' }}>
            Coming Soon
          </p>
          <p className="text-xs mt-0.5" style={{ color: 'var(--dash-text-muted)' }}>
            Custom domain support is on our roadmap. We'll notify you when it's ready.
          </p>
        </div>
      </div>

      {/* Menu URL */}
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

      {/* Custom Domain — disabled */}
      <div className="opacity-40 pointer-events-none select-none">
        <Card>
          <CardHeader>
            <CardTitle>Custom Domain</CardTitle>
            <CardDescription>Point your own domain at your rtylr menu.</CardDescription>
          </CardHeader>
          <CardContent className="pt-0 space-y-4">
            <div className="space-y-1.5">
              <Label>Domain</Label>
              <div className="flex items-center gap-2">
                <Input placeholder="menu.mybusiness.com" className="flex-1" disabled />
                <Button size="sm" disabled className="shrink-0">
                  Connect Domain
                </Button>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
};

export default Domain;
