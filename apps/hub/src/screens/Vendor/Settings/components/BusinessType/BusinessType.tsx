import React, { useState, useEffect } from 'react';
import { useBusinessStore } from '../../../../../stores/businessStore';
import { useUpdateBusiness } from '../../../../../hooks/queries/useBusiness';
import { useUserStore } from '../../../../../stores/userStore';
import { Card, CardContent, CardHeader, CardTitle, CardDescription, Button, Input, Label } from '../../../../../components/ui';
import { cn } from '../../../../../lib/utils';

type IndustryGroup = {
  label: string;
  emoji: string;
  types: { value: string; label: string }[];
};

const INDUSTRY_GROUPS: IndustryGroup[] = [
  {
    label: 'Food & Beverage',
    emoji: '🍽️',
    types: [
      { value: 'restaurant',  label: 'Restaurant' },
      { value: 'cafe',        label: 'Cafe / Coffee Shop' },
      { value: 'bar',         label: 'Bar / Pub' },
      { value: 'food_truck',  label: 'Food Truck' },
      { value: 'bakery',      label: 'Bakery' },
      { value: 'juice_bar',   label: 'Juice Bar' },
      { value: 'catering',    label: 'Catering' },
      { value: 'fast_food',   label: 'Fast Food' },
    ],
  },
  {
    label: 'Retail',
    emoji: '🛍️',
    types: [
      { value: 'retail_clothing',  label: 'Clothing & Apparel' },
      { value: 'retail_grocery',   label: 'Grocery / Supermarket' },
      { value: 'retail_pharmacy',  label: 'Pharmacy' },
      { value: 'retail_electronics', label: 'Electronics' },
      { value: 'retail_gifts',     label: 'Gifts & Souvenirs' },
      { value: 'retail_books',     label: 'Books & Stationery' },
    ],
  },
  {
    label: 'Health & Beauty',
    emoji: '💆',
    types: [
      { value: 'salon',      label: 'Hair Salon' },
      { value: 'barber',     label: 'Barbershop' },
      { value: 'spa',        label: 'Spa & Wellness' },
      { value: 'gym',        label: 'Gym / Fitness' },
      { value: 'nail_salon', label: 'Nail Salon' },
      { value: 'clinic',     label: 'Medical Clinic' },
    ],
  },
  {
    label: 'Hospitality',
    emoji: '🏨',
    types: [
      { value: 'hotel',      label: 'Hotel' },
      { value: 'hostel',     label: 'Hostel' },
      { value: 'resort',     label: 'Resort' },
      { value: 'guesthouse', label: 'Guesthouse / B&B' },
    ],
  },
  {
    label: 'Other',
    emoji: '🏢',
    types: [
      { value: 'events',     label: 'Events & Venue' },
      { value: 'education',  label: 'Education' },
      { value: 'services',   label: 'Professional Services' },
      { value: 'other',      label: 'Other' },
    ],
  },
];

const ALL_TYPES = INDUSTRY_GROUPS.flatMap((g) => g.types);

const BusinessTypeSettings: React.FC = () => {
  const { businessType, setBusinessType } = useBusinessStore();
  const { userId } = useUserStore();
  const updateBusiness = useUpdateBusiness();
  const [selected, setSelected] = useState<string>(businessType || '');
  const [customType, setCustomType] = useState('');

  useEffect(() => { if (businessType) setSelected(businessType); }, [businessType]);

  const handleSave = async () => {
    if (!userId || !selected) return;
    setBusinessType(selected as any);
    const data: any = { businessType: selected };
    if (selected === 'other' && customType) data.customBusinessType = customType;
    updateBusiness.mutate({ businessId: userId, data });
  };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-xl font-bold" style={{ color: 'var(--hub-text)' }}>Industry</h1>
        <p className="text-sm mt-0.5" style={{ color: 'var(--hub-text-muted)' }}>
          Select the category that best describes your business. This personalizes your dashboard experience.
        </p>
      </div>

      <div className="space-y-5">
        {INDUSTRY_GROUPS.map((group) => (
          <Card key={group.label}>
            <CardHeader className="pb-3">
              <CardTitle className="flex items-center gap-2 text-sm">
                <span>{group.emoji}</span>
                {group.label}
              </CardTitle>
            </CardHeader>
            <CardContent className="pt-0">
              <div className="grid grid-cols-2 gap-2">
                {group.types.map((type) => {
                  const isActive = selected === type.value;
                  return (
                    <button
                      key={type.value}
                      onClick={() => setSelected(type.value)}
                      className={cn(
                        'px-3 py-2.5 rounded-lg text-sm text-left transition-all font-medium',
                      )}
                      style={{
                        background: isActive ? 'rgba(69,131,196,0.12)' : 'var(--hub-hover)',
                        border: isActive ? '1px solid rgba(69,131,196,0.35)' : '1px solid transparent',
                        color: isActive ? '#4583C4' : 'var(--hub-text-muted)',
                      }}
                    >
                      {type.label}
                    </button>
                  );
                })}
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      {selected === 'other' && (
        <Card>
          <CardContent className="pt-5">
            <div className="space-y-1.5">
              <Label>Describe your business</Label>
              <Input
                value={customType}
                onChange={(e) => setCustomType(e.target.value)}
                placeholder="e.g. Pet grooming, Florist, Coworking space..."
              />
            </div>
          </CardContent>
        </Card>
      )}

      {selected && (
        <div className="flex items-center justify-between pt-1">
          <p className="text-sm" style={{ color: 'var(--hub-text-muted)' }}>
            Selected: <span style={{ color: 'var(--hub-text)', fontWeight: 500 }}>
              {ALL_TYPES.find((t) => t.value === selected)?.label ?? selected}
            </span>
          </p>
          <Button onClick={handleSave} disabled={updateBusiness.isPending}>
            {updateBusiness.isPending ? 'Saving...' : 'Save'}
          </Button>
        </div>
      )}
    </div>
  );
};

export default BusinessTypeSettings;
