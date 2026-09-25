export interface Customer {
  id: string;
  name: string;
  email?: string;
  phone?: string;
  loyaltyPoints: number;
  totalSpent: number;
  visitCount: number;
  lastVisit?: string;
  tags?: string[];
  createdAt: string;
}

export type CampaignStatus = 'draft' | 'active' | 'paused' | 'completed';
export type CampaignChannel = 'email' | 'sms' | 'push';

export interface Campaign {
  id: string;
  name: string;
  channel: CampaignChannel;
  status: CampaignStatus;
  audience: string;
  sentCount?: number;
  openRate?: number;
  createdAt: string;
}

export interface LoyaltyTier {
  name: string;
  minPoints: number;
  discount: number;
  color: string;
}
