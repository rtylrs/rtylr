import type { IconProps } from "phosphor-react";
import {
  BookOpen as BookOpen,
  GearSix as Settings,
  CreditCard,
  FolderSimple as Folder,
  MagicWand as Wand2,
  Phone,
  Plus,
  QrCode,
} from "phosphor-react";
import { Link } from "react-router-dom";

function QuickLinks() {
  return (
    <div className="quick-links">
      <QuickLink icon={BookOpen} title="Menu" to="/app/business/menu" />
      <QuickLink icon={Folder} title="My Menus" to="/app/business/menus" />
      {/* <QuickLink icon={faChartPie} title="Reports" to="/app/business/reports" /> */}
      <QuickLink
        icon={QrCode}
        title="Qr Codes"
        to="/app/business/qr-codes/"
      />
      <QuickLink
        icon={Plus}
        title="Create Qr"
        to="/app/business/qr-codes/create"
      />
      <QuickLink
        icon={Wand2}
        title="Appearance"
        to="/app/business/design"
      />
      <QuickLink
        icon={CreditCard}
        title="Billing"
        to="/app/business/billing"
      />
      <QuickLink icon={Settings} title="Settings" to="/settings" />
      <QuickLink icon={Phone} title="Support" to="/contact" />
    </div>
  );
}

const QuickLink = ({ icon: Icon, title, to }: { icon: React.ComponentType<IconProps>; title: string; to: string }) => (
  <Link className="quick-link" to={to}>
    <Icon size={20} />
    <span>{title}</span>
  </Link>
);

export default QuickLinks;
