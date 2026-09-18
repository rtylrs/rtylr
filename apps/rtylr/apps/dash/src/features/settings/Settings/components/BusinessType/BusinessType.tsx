import React, { useState, useEffect, useMemo } from 'react';
import { HelmetContainer } from '@voxire/components';
import { useTranslation } from 'react-i18next';
import { Search, Loader2, Check } from 'lucide-react';
import { AnimatePresence, motion } from 'motion/react';
import { EmptyState } from '@voxire/ui';
import { useBusinessStore, type BusinessType } from '../../../../../shared/stores/businessStore';
import { useUpdateBusiness } from '../../../../../shared/hooks/queries/useBusiness';
import { useIndustries } from '../../../../../shared/hooks/queries/useIndustries';
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  Button,
  Input,
} from '../../../../../shared/components/ui';
import { cn } from '../../../../../shared/lib/utils';

const BusinessTypeSettings: React.FC = () => {
  const { t } = useTranslation();
  const { businessType, setBusinessType } = useBusinessStore();
  const updateBusiness = useUpdateBusiness();
  const { data, isLoading } = useIndustries();

  const [selected, setSelected] = useState<string>(businessType || '');
  const [search, setSearch] = useState('');

  useEffect(() => {
    setSelected(businessType || '');
  }, [businessType]);

  const groups = useMemo(() => data?.groups ?? [], [data?.groups]);

  const filteredGroups = useMemo(() => {
    const q = search.toLowerCase().trim();
    if (!q) return groups;
    return groups
      .map((g) => ({
        ...g,
        types: g.types.filter(
          (t) => t.label.toLowerCase().includes(q) || t.slug.toLowerCase().includes(q)
        ),
      }))
      .filter((g) => g.types.length > 0);
  }, [groups, search]);

  const selectedLabel = useMemo(() => {
    for (const g of groups) {
      const t = g.types.find((t) => t.slug === selected);
      if (t) return t.label;
    }
    return selected;
  }, [groups, selected]);

  const handleSave = () => {
    if (!selected) return;

    updateBusiness.mutate(
      { data: { businessType: selected } },
      { onSuccess: () => setBusinessType(selected as BusinessType) }
    );
  };

  const hasChanges = selected !== (businessType || '');

  return (
    <>
      <HelmetContainer title="Industry · rtylr" />
      <div className={cn('space-y-6', hasChanges && 'pb-24')}>
        <div>
          <h1 className="text-xl font-bold" style={{ color: 'var(--dash-text)' }}>
            Industry
          </h1>
          <p className="text-sm mt-0.5" style={{ color: 'var(--dash-text-muted)' }}>
            Select the category that best describes your business.
          </p>
        </div>

        {/* Search */}
        <div className="relative">
          <Search
            size={14}
            className="absolute left-3 top-1/2 -translate-y-1/2"
            style={{ color: 'var(--dash-text-dim)' }}
          />
          <Input
            placeholder="Search industries..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="pl-9"
          />
        </div>

        {isLoading ? (
          <div className="flex items-center justify-center py-16">
            <Loader2 size={20} className="animate-spin" style={{ color: 'var(--dash-text-dim)' }} />
          </div>
        ) : filteredGroups.length === 0 ? (
          <EmptyState
            icon={Search}
            title={t('settings.businessType.noIndustriesMatchTitle', { search })}
            description={t('settings.businessType.noIndustriesMatchDescription')}
            action={{ label: t('common.clearSearch'), onClick: () => setSearch('') }}
          />
        ) : (
          <div className="space-y-5">
            {filteredGroups.map((group) => (
              <Card key={group.category}>
                <CardHeader className="pb-3">
                  <CardTitle className="text-sm">{group.category}</CardTitle>
                </CardHeader>
                <CardContent className="pt-0">
                  <div className="grid grid-cols-2 gap-2">
                    {group.types.map((type) => {
                      const isActive = selected === type.slug;
                      return (
                        <button
                          type="button"
                          key={type.slug}
                          onClick={() => setSelected(type.slug)}
                          className={cn(
                            'px-3 py-2.5 rounded-lg text-sm text-left transition-all font-medium flex items-center gap-2'
                          )}
                          style={{
                            background: isActive ? 'rgba(69,131,196,0.12)' : 'var(--dash-hover)',
                            border: isActive
                              ? '1px solid rgba(69,131,196,0.35)'
                              : '1px solid transparent',
                            color: isActive ? '#4583C4' : 'var(--dash-text-muted)',
                          }}
                        >
                          {type.icon && <span>{type.icon}</span>}
                          {type.label}
                        </button>
                      );
                    })}
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        )}

        <AnimatePresence>
          {hasChanges && (
            <motion.div
              key="save-footer"
              initial={{ y: 96, opacity: 0 }}
              animate={{ y: 0, opacity: 1 }}
              exit={{ y: 96, opacity: 0 }}
              transition={{ type: 'spring', stiffness: 380, damping: 30, mass: 0.6 }}
              className="fixed bottom-4 left-1/2 -translate-x-1/2 z-40"
            >
              <div
                className="flex items-center gap-4 pl-4 pr-2 py-2 rounded-full shadow-lg"
                style={{
                  background: 'var(--dash-card)',
                  border: '1px solid var(--dash-border)',
                  boxShadow: '0 10px 28px -8px rgba(0,0,0,0.18)',
                  backdropFilter: 'blur(8px)',
                }}
              >
                <div className="flex items-center gap-2 min-w-0">
                  <span
                    className="h-6 w-6 rounded-full flex items-center justify-center shrink-0"
                    style={{ background: 'var(--accent-muted)', color: 'var(--accent-on-muted)' }}
                  >
                    <Check size={13} />
                  </span>
                  <p className="text-sm truncate" style={{ color: 'var(--dash-text-muted)' }}>
                    Selected:{' '}
                    <span style={{ color: 'var(--dash-text)', fontWeight: 600 }}>
                      {selectedLabel}
                    </span>
                  </p>
                </div>
                <Button
                  onClick={handleSave}
                  disabled={updateBusiness.isPending}
                  className="rounded-full"
                >
                  {updateBusiness.isPending ? 'Saving...' : 'Save changes'}
                </Button>
              </div>
            </motion.div>
          )}
        </AnimatePresence>
      </div>
    </>
  );
};

export default BusinessTypeSettings;
