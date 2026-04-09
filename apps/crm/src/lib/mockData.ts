import { Customer, Segment, Campaign, LoyaltyMember, LoyaltyReward, LoyaltyTier, Automation, Template, FeedbackItem, TimelineEntry, Settings, CampaignStats } from '../services/crmApi';

export const mockCustomers: Customer[] = [
  { id: 1, name: 'Sarah Johnson', email: 'sarah.j@email.com', phone: '+1 555-0101', totalSpent: 4820, visits: 34, loyaltyPoints: 4820, lastVisit: '2 days ago', initials: 'SJ', tags: ['VIP', 'Frequent'], notes: 'Prefers oat milk.', createdAt: '2023-06-12', consent: { email: true, sms: true, marketing: true } },
  { id: 2, name: 'Michael Chen', email: 'm.chen@email.com', phone: '+1 555-0182', totalSpent: 3640, visits: 28, loyaltyPoints: 3640, lastVisit: '5 days ago', initials: 'MC', tags: ['VIP'], notes: '', createdAt: '2023-08-01', consent: { email: true, sms: false, marketing: true } },
  { id: 3, name: 'Emily Rodriguez', email: 'e.rodriguez@email.com', phone: '+1 555-0247', totalSpent: 2910, visits: 22, loyaltyPoints: 2910, lastVisit: '1 week ago', initials: 'ER', tags: ['Birthday'], notes: 'Birthday May 4', createdAt: '2024-01-14', consent: { email: true, sms: true, marketing: false } },
  { id: 4, name: 'James Wilson', email: 'j.wilson@email.com', phone: '+1 555-0391', totalSpent: 1870, visits: 17, loyaltyPoints: 870, lastVisit: '2 weeks ago', initials: 'JW', tags: [], notes: '', createdAt: '2024-03-02', consent: { email: true, sms: false, marketing: true } },
  { id: 5, name: 'Aisha Patel', email: 'a.patel@email.com', phone: '+1 555-0412', totalSpent: 1420, visits: 14, loyaltyPoints: 620, lastVisit: '3 weeks ago', initials: 'AP', tags: ['New'], notes: '', createdAt: '2024-05-18', consent: { email: true, sms: true, marketing: true } },
  { id: 6, name: 'David Kim', email: 'd.kim@email.com', phone: '+1 555-0563', totalSpent: 980, visits: 9, loyaltyPoints: 320, lastVisit: '1 month ago', initials: 'DK', tags: [], notes: '', createdAt: '2024-07-22', consent: { email: true, sms: false, marketing: false } },
  { id: 7, name: 'Olivia Martinez', email: 'o.martinez@email.com', phone: '+1 555-0674', totalSpent: 540, visits: 5, loyaltyPoints: 180, lastVisit: '6 weeks ago', initials: 'OM', tags: ['At Risk'], notes: '', createdAt: '2024-09-10', consent: { email: true, sms: true, marketing: true } },
  { id: 8, name: 'Noah Thompson', email: 'n.thompson@email.com', phone: '+1 555-0788', totalSpent: 210, visits: 2, loyaltyPoints: 60, lastVisit: '2 months ago', initials: 'NT', tags: ['At Risk'], notes: '', createdAt: '2024-12-01', consent: { email: false, sms: false, marketing: false } },
];

export const mockSegments: Segment[] = [
  { id: 1, name: 'VIP Customers', description: 'Top 10% by spend', customerCount: 482, color: '#e879f9', rules: ['Spend > $2000', 'Visits > 20'], trend: '+12%', trendUp: true },
  { id: 2, name: 'At Risk', description: 'Not visited in 60+ days', customerCount: 127, color: '#ef4444', rules: ['Last visit > 60 days'], trend: '-5%', trendUp: false },
  { id: 3, name: 'New This Month', description: 'Joined in last 30 days', customerCount: 184, color: '#10b981', rules: ['Joined in last 30 days'], trend: '+8%', trendUp: true },
  { id: 4, name: 'Birthday Month', description: 'Birthday this month', customerCount: 36, color: '#f59e0b', rules: ['Birthday = this month'], trend: '+2%', trendUp: true },
];

export const mockCampaigns: Campaign[] = [
  { id: 1, name: 'Spring Rewards', description: 'Bonus points campaign', channel: 'Email', status: 'Active', audience: 1420, sent: 1420, openRate: '42%', clickRate: '12%', revenue: '$8,940', updatedAt: '2 days ago', subject: 'Double points this week' },
  { id: 2, name: 'VIP Appreciation', description: 'Thank you to top customers', channel: 'Email', status: 'Active', audience: 482, sent: 482, openRate: '58%', clickRate: '18%', revenue: '$12,340', updatedAt: '5 days ago', subject: 'You are one of our favorites' },
  { id: 3, name: 'Win Back', description: 'Re-engage inactive customers', channel: 'SMS', status: 'Draft', audience: 320, sent: 0, openRate: '-', clickRate: '-', revenue: '-', updatedAt: '1 week ago' },
  { id: 4, name: 'Birthday Club', description: 'Monthly birthday offers', channel: 'Email', status: 'Scheduled', audience: 36, sent: 0, openRate: '-', clickRate: '-', revenue: '-', updatedAt: 'yesterday' },
];

export const mockLoyaltyMembers: LoyaltyMember[] = mockCustomers.map((c) => ({
  id: c.id,
  name: c.name,
  initials: c.initials,
  points: c.loyaltyPoints,
  tier: c.loyaltyPoints >= 2500 ? 'Platinum' : c.loyaltyPoints >= 1000 ? 'Gold' : c.loyaltyPoints >= 500 ? 'Silver' : 'Bronze',
  visits: c.visits,
  joinedAt: c.createdAt || '2024-01-01',
  lastActive: c.lastVisit,
}));

export const mockLoyaltyRewards: LoyaltyReward[] = [
  { id: 1, name: 'Free Coffee', description: 'Any drink on the house', pointsCost: 200, category: 'Drinks', stock: 999, active: true },
  { id: 2, name: '10% Off Next Visit', description: 'One time discount', pointsCost: 350, category: 'Discount', stock: 999, active: true },
  { id: 3, name: 'Free Dessert', description: 'Any dessert', pointsCost: 450, category: 'Food', stock: 50, active: true },
  { id: 4, name: '$10 Store Credit', description: 'Redeem at checkout', pointsCost: 1000, category: 'Credit', stock: 999, active: true },
  { id: 5, name: 'Branded Mug', description: 'Limited edition', pointsCost: 1500, category: 'Merch', stock: 20, active: true },
  { id: 6, name: 'VIP Tasting Event', description: 'Exclusive event invite', pointsCost: 3000, category: 'Experience', stock: 10, active: true },
];

export const mockLoyaltyTiers: LoyaltyTier[] = [
  { id: 1, name: 'Bronze', threshold: 0, multiplier: 1, perks: ['Earn points on every purchase'], color: '#b45309' },
  { id: 2, name: 'Silver', threshold: 500, multiplier: 1.25, perks: ['1.25x points', 'Birthday reward'], color: '#94a3b8' },
  { id: 3, name: 'Gold', threshold: 1000, multiplier: 1.5, perks: ['1.5x points', 'Free drink monthly', 'Priority support'], color: '#f59e0b' },
  { id: 4, name: 'Platinum', threshold: 2500, multiplier: 2, perks: ['2x points', 'VIP events', 'Personal concierge'], color: '#e879f9' },
];

export const mockAutomations: Automation[] = [
  { id: 1, name: 'Welcome Series', description: '3-email onboarding', trigger: 'New customer', status: 'active', enrolled: 184, completed: 142, createdAt: '2024-02-01' },
  { id: 2, name: 'Birthday Greeting', description: 'Bonus points on birthday', trigger: 'Birthday', status: 'active', enrolled: 36, completed: 28, createdAt: '2024-01-15' },
  { id: 3, name: 'Win-back', description: 'Re-engage after 60 days', trigger: '60 days inactive', status: 'active', enrolled: 127, completed: 42, createdAt: '2024-03-10' },
  { id: 4, name: 'Review Request', description: 'Ask for a review post-visit', trigger: 'Post visit', status: 'paused', enrolled: 0, completed: 0, createdAt: '2024-04-22' },
];

export const mockTemplates: Template[] = [
  { id: 1, name: 'Welcome Email', channel: 'email', category: 'Onboarding', subject: 'Welcome to {{business_name}}', body: 'Hi {{first_name}}, thanks for joining!', updatedAt: '1 week ago' },
  { id: 2, name: 'Birthday Wishes', channel: 'email', category: 'Engagement', subject: 'Happy Birthday {{first_name}}', body: 'Enjoy a free drink on us!', updatedAt: '2 weeks ago' },
  { id: 3, name: 'Win-back Offer', channel: 'email', category: 'Re-engagement', subject: 'We miss you, {{first_name}}', body: 'Come back for 20% off', updatedAt: '3 days ago' },
  { id: 4, name: 'Abandoned Cart', channel: 'email', category: 'Recovery', subject: 'You left something behind', body: 'Finish your order now', updatedAt: '1 day ago' },
  { id: 5, name: 'SMS Promo', channel: 'sms', category: 'Promotion', body: 'Flash sale 20% off today only. Reply STOP to unsubscribe.', updatedAt: '5 days ago' },
  { id: 6, name: 'Review Request', channel: 'email', category: 'Feedback', subject: 'How was your visit?', body: 'Leave us a review and earn 50 points', updatedAt: 'yesterday' },
];

export const mockFeedback: FeedbackItem[] = [
  { id: 1, customer: 'Sarah Johnson', rating: 5, nps: 10, comment: 'Always a great experience. Staff is amazing.', createdAt: '2 days ago', source: 'Email survey' },
  { id: 2, customer: 'Michael Chen', rating: 4, nps: 9, comment: 'Love the loyalty program.', createdAt: '5 days ago', source: 'In-store' },
  { id: 3, customer: 'Emily Rodriguez', rating: 5, nps: 10, comment: 'Best coffee in town!', createdAt: '1 week ago', source: 'Email survey' },
  { id: 4, customer: 'David Kim', rating: 3, nps: 6, comment: 'Wait times have been longer lately.', createdAt: '2 weeks ago', source: 'SMS survey' },
  { id: 5, customer: 'Olivia Martinez', rating: 2, nps: 4, comment: 'Order was wrong twice in a row.', createdAt: '3 weeks ago', source: 'Email survey' },
];

export const mockSettings: Settings = {
  defaultConsent: true,
  timezone: 'America/Los_Angeles',
  fromEmail: 'hello@rtylr.com',
  fromName: 'Rtylr Team',
  smsSenderId: 'RTYLR',
  businessName: 'Rtylr',
  atRiskDays: 60,
  currency: 'USD',
};

export const mockTimeline: TimelineEntry[] = [
  { id: 1, type: 'visit', title: 'Visited store', description: 'Morning visit', at: '2 days ago' },
  { id: 2, type: 'purchase', title: 'Purchase $24.50', description: '2 drinks, 1 pastry', at: '2 days ago' },
  { id: 3, type: 'points', title: 'Earned 25 points', at: '2 days ago' },
  { id: 4, type: 'email', title: 'Opened "Spring Rewards"', at: '4 days ago' },
  { id: 5, type: 'sms', title: 'Received SMS: Flash sale', at: '5 days ago' },
  { id: 6, type: 'reward', title: 'Redeemed Free Coffee', description: '-200 points', at: '1 week ago' },
  { id: 7, type: 'note', title: 'Staff note', description: 'Prefers oat milk', at: '2 weeks ago' },
  { id: 8, type: 'visit', title: 'Visited store', at: '2 weeks ago' },
];

export const mockCampaignStats: CampaignStats = {
  id: 1,
  sent: 1420,
  delivered: 1398,
  opened: 589,
  clicked: 172,
  unsubscribed: 4,
  bounced: 22,
  revenue: 8940,
  timeline: [
    { day: 'Mon', opens: 120, clicks: 34 },
    { day: 'Tue', opens: 180, clicks: 52 },
    { day: 'Wed', opens: 140, clicks: 40 },
    { day: 'Thu', opens: 80, clicks: 22 },
    { day: 'Fri', opens: 69, clicks: 24 },
  ],
};
