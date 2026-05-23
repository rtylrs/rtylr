import React from 'react';
import {
  Utensils,
  Package,
  Users,
  UserCheck,
  DollarSign,
  GitMerge,
  Settings,
  Info,
  BookOpen,
  CreditCard,
  Globe,
  Layers,
  LucideProps,
} from 'lucide-react';
import { WindowType } from '../stores/useDesktopStore';

const WINDOW_ICONS: Record<WindowType, React.FC<LucideProps>> = {
  pos: Utensils,
  erp: Package,
  crm: Users,
  hr: UserCheck,
  finance: DollarSign,
  flow: GitMerge,
  settings: Settings,
  about: Info,
  docs: BookOpen,
  pricing: CreditCard,
  landing: Globe,
};

interface AppIconProps extends LucideProps {
  type: WindowType;
}

export const AppIcon: React.FC<AppIconProps> = ({ type, ...props }) => {
  const Icon = WINDOW_ICONS[type] ?? Layers;
  return <Icon {...props} />;
};
