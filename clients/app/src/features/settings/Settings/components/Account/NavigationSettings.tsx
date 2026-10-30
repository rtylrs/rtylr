import type { ComponentType } from 'react';
import { useTranslation } from 'react-i18next';
import { ArrowDown, ArrowUp, Eye, EyeOff } from 'lucide-react';
import { toast } from 'sonner';
import { getOneKitErrorMessage } from '@voxire/api';
import { useUserStore } from '../../../../../shared/stores/userStore';
import { useAddons, isAddonEnabled } from '../../../../../shared/hooks/queries/useAddons';
import { canAccessPortalPath } from '../../../../../shared/utils/canAccessPortalPath';
import { rawMenuSubItems, rawNavItems } from '../../../../../shell/navigation';
import { RAIL_BOTTOM, railItems } from '../../../../../shell/railItems';
import {
  usePortalLayout,
  useResetPortalLayout,
  useSavePortalLayout,
  type PortalArrangement,
} from '../../../../../shell/portalLayout';
import {
  Button,
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '../../../../../shared/components/ui';
import './navigation-settings.css';

interface Entry {
  id: string;
  label: string;
  icon: ComponentType<{ size?: number }>;
}

type Area = 'rail' | 'workspace';

// Items that can move but never hide, as on the server.
const REQUIRED = new Set(['home', 'dashboard']);

interface Placed {
  id: string;
  group: 'top' | 'bottom';
}

/**
 * The person's arrangement with every id in place, including ones this
 * screen doesn't show (an app they can't open), so saving keeps them where
 * they were. A saved arrangement comes back without its empty lists, so a
 * missing bottom group means empty, not the default.
 */
function placedIds(entries: Entry[], arrangement: PortalArrangement | undefined, bottom: string[]) {
  const saved = arrangement?.top !== undefined || arrangement?.bottom !== undefined;
  const top = saved
    ? [...(arrangement?.top ?? [])]
    : entries.filter((e) => !bottom.includes(e.id)).map((e) => e.id);
  const low = saved
    ? (arrangement?.bottom ?? []).filter((id) => !top.includes(id))
    : entries.filter((e) => bottom.includes(e.id)).map((e) => e.id);
  for (const entry of entries) {
    if (!top.includes(entry.id) && !low.includes(entry.id))
      (bottom.includes(entry.id) ? low : top).push(entry.id);
  }
  return { top, bottom: low, hidden: [...(arrangement?.hidden ?? [])] };
}

function ArrangeList({
  area,
  title,
  description,
  entries,
  arrangement,
  bottom,
  saving,
  onSave,
  onReset,
}: {
  area: Area;
  title: string;
  description: string;
  entries: Entry[];
  arrangement: PortalArrangement | undefined;
  bottom: string[];
  saving: boolean;
  onSave: (next: PortalArrangement) => void;
  onReset: () => void;
}) {
  const { t } = useTranslation();
  const ids = placedIds(entries, arrangement, bottom);
  const byId = new Map(entries.map((entry) => [entry.id, entry]));
  const rows: Placed[] = [
    ...ids.top.map((id) => ({ id, group: 'top' as const })),
    ...ids.bottom.map((id) => ({ id, group: 'bottom' as const })),
  ].filter((row) => byId.has(row.id));
  const hidden = new Set(ids.hidden);
  const hasBottom = area === 'rail';

  // Moves past the neighbour this screen shows; at the edge of the top
  // group an app bar item crosses to the bottom group, and back.
  const move = (index: number, step: -1 | 1) => {
    const row = rows[index];
    const neighbour = rows[index + step];
    const next = { top: [...ids.top], bottom: [...ids.bottom], hidden: ids.hidden };
    const from = row.group === 'top' ? next.top : next.bottom;
    from.splice(from.indexOf(row.id), 1);
    if (neighbour && neighbour.group === row.group) {
      const at = from.indexOf(neighbour.id);
      from.splice(step < 0 ? at : at + 1, 0, row.id);
    } else if (step > 0) {
      next.bottom.unshift(row.id);
    } else {
      next.top.push(row.id);
    }
    onSave(next);
  };

  const toggle = (id: string) =>
    onSave({
      top: ids.top,
      bottom: ids.bottom,
      hidden: hidden.has(id) ? ids.hidden.filter((other) => other !== id) : [...ids.hidden, id],
    });

  return (
    <section className="nav-arrange" aria-labelledby={`nav-arrange-${area}`}>
      <div className="nav-arrange-head">
        <div>
          <h3 id={`nav-arrange-${area}`}>{title}</h3>
          <p>{description}</p>
        </div>
        <Button
          variant="outline"
          size="sm"
          disabled={!arrangement?.customized || saving}
          onClick={onReset}
        >
          {t('settings.navigation.restore')}
        </Button>
      </div>
      <ol className="nav-arrange-list">
        {rows.map((row, index) => {
          const entry = byId.get(row.id)!;
          const startsBottom =
            hasBottom && row.group === 'bottom' && rows[index - 1]?.group !== 'bottom';
          const isHidden = hidden.has(row.id);
          const canUp = index > 0 || (hasBottom && row.group === 'bottom');
          const canDown = index < rows.length - 1 || (hasBottom && row.group === 'top');
          return (
            <li key={row.id} className={isHidden ? 'is-hidden' : undefined}>
              {startsBottom && (
                <span className="nav-arrange-divider">{t('settings.navigation.bottomGroup')}</span>
              )}
              <span className="nav-arrange-row">
                <span className="nav-arrange-icon" aria-hidden="true">
                  <entry.icon size={16} />
                </span>
                <span className="nav-arrange-label">
                  {entry.label}
                  {isHidden && <small>{t('settings.navigation.hiddenTag')}</small>}
                </span>
                <Button
                  variant="ghost"
                  size="icon"
                  disabled={!canUp}
                  onClick={() => move(index, -1)}
                  aria-label={t('settings.navigation.moveUp', { item: entry.label })}
                >
                  <ArrowUp size={15} />
                </Button>
                <Button
                  variant="ghost"
                  size="icon"
                  disabled={!canDown}
                  onClick={() => move(index, 1)}
                  aria-label={t('settings.navigation.moveDown', { item: entry.label })}
                >
                  <ArrowDown size={15} />
                </Button>
                <Button
                  variant="ghost"
                  size="icon"
                  disabled={REQUIRED.has(row.id)}
                  onClick={() => toggle(row.id)}
                  aria-pressed={isHidden}
                  aria-label={t(
                    isHidden ? 'settings.navigation.show' : 'settings.navigation.hide',
                    { item: entry.label }
                  )}
                >
                  {isHidden ? <EyeOff size={15} /> : <Eye size={15} />}
                </Button>
              </span>
            </li>
          );
        })}
        {hasBottom && !rows.some((row) => row.group === 'bottom') && (
          <li>
            <span className="nav-arrange-divider">{t('settings.navigation.bottomGroup')}</span>
          </li>
        )}
      </ol>
    </section>
  );
}

/** Arrange the app bar and the Home sidebar by hand. Caret changes the same
 *  arrangement when asked. */
export default function NavigationSettings() {
  const { t } = useTranslation();
  const user = useUserStore((state) => state.user);
  const { data: addons = [] } = useAddons();
  const layout = usePortalLayout(user?.id);
  const save = useSavePortalLayout(user?.id);
  const reset = useResetPortalLayout(user?.id);
  const saving = save.isPending || reset.isPending;
  const failed = (error: unknown) =>
    toast.error(getOneKitErrorMessage(error, t('settings.navigation.saveFailed')));

  const workspace: Entry[] = [
    ...rawNavItems,
    ...(isAddonEnabled(addons, 'menu') ? rawMenuSubItems : []),
  ]
    .filter((item) => canAccessPortalPath(user, item.path))
    .map((item) => ({
      id: item.key ?? item.path,
      label: t(`common.nav.${item.key}`),
      icon: item.icon,
    }));

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t('settings.navigation.title')}</CardTitle>
        <CardDescription>{t('settings.navigation.description')}</CardDescription>
      </CardHeader>
      <CardContent className="pt-0 nav-arrange-areas">
        <ArrangeList
          area="rail"
          title={t('settings.navigation.railTitle')}
          description={t('settings.navigation.railDescription')}
          entries={railItems(user, t)}
          arrangement={layout?.rail}
          bottom={RAIL_BOTTOM}
          saving={saving}
          onSave={(rail) => save.mutate({ rail }, { onError: failed })}
          onReset={() => reset.mutate('rail', { onError: failed })}
        />
        <ArrangeList
          area="workspace"
          title={t('settings.navigation.workspaceTitle')}
          description={t('settings.navigation.workspaceDescription')}
          entries={workspace}
          arrangement={layout?.workspace}
          bottom={[]}
          saving={saving}
          onSave={(next) => save.mutate({ workspace: next }, { onError: failed })}
          onReset={() => reset.mutate('workspace', { onError: failed })}
        />
      </CardContent>
    </Card>
  );
}
