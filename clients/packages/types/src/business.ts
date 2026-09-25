export interface Business {
  id: string;
  name: string;
  storeName?: string;
  industry?: string;
  planType?: 'free' | 'starter' | 'pro' | 'enterprise';
  addons?: Record<string, boolean>;
  domain?: string;
  email?: string;
}
