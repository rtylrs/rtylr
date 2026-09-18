import { useEffect, useState } from 'react';
import { featureClient } from '../lib/clients';
import { storageGet, storageSet } from '../lib/storage';

const AUTH_FEATURE_SUBJECT_KEY = 'auth_feature_subject';

interface FeatureResolution {
  enabled: boolean;
}

interface FeaturesResponse {
  features?: Record<string, FeatureResolution>;
}

function getOrCreateFeatureSubject(): string {
  const existing = storageGet(AUTH_FEATURE_SUBJECT_KEY);
  if (existing) return existing;
  const subject =
    typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function'
      ? crypto.randomUUID()
      : `${Date.now().toString(36)}-${Math.random().toString(36).slice(2)}`;
  storageSet(AUTH_FEATURE_SUBJECT_KEY, subject);
  return subject;
}

function responseKeyFor(key: string): string {
  return key.replace(/_([a-z])/g, (_, c: string) => c.toUpperCase());
}

export function useFeatureGate(key: string, defaultEnabled = true): boolean {
  const [enabled, setEnabled] = useState(defaultEnabled);

  useEffect(() => {
    let mounted = true;
    const resolveGate = async (): Promise<void> => {
      try {
        const subject = getOrCreateFeatureSubject();
        const res = (await featureClient.resolveFeatures({
          subject,
          attributes: { app: 'auth' },
        })) as FeaturesResponse;
        const features = res.features ?? {};
        const next = features[key]?.enabled ?? features[responseKeyFor(key)]?.enabled;
        if (mounted && typeof next === 'boolean') {
          setEnabled(next);
        }
      } catch {
        if (mounted) setEnabled(defaultEnabled);
      }
    };

    void resolveGate();
    return () => {
      mounted = false;
    };
  }, [defaultEnabled, key]);

  return enabled;
}
