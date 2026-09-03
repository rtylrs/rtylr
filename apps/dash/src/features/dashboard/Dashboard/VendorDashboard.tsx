import React, { useState, useCallback, useRef, useEffect } from 'react';
import { createPortal } from 'react-dom';
import { useNavigate } from '@tanstack/react-router';
import {
  Responsive as ResponsiveGridLayout,
  type Layout,
  type ResponsiveLayouts,
} from 'react-grid-layout/legacy';
import { motion, AnimatePresence } from 'motion/react';
import { Settings2, X, Plus, GripVertical } from 'lucide-react';
import { ErrorBoundary } from '@voxire/ui';
import { HelmetContainer } from '@voxire/components';
import { toFiniteNumber } from '@voxire/api';
import { canEffectiveAccess } from '@voxire/permissions';
import { useTranslation } from 'react-i18next';
import { useUserStore } from '../../../shared/stores/userStore';
import { Button } from '../../../shared/components/ui';
import {
  getWidgetComponent,
  WIDGET_CATALOG,
  loadLayout,
  saveLayout,
  loadActiveWidgets,
  saveActiveWidgets,
  DEFAULT_LAYOUT,
} from './widgets';

import 'react-grid-layout/css/styles.css';
import 'react-resizable/css/styles.css';

const today = new Date().toLocaleDateString('en-US', {
  weekday: 'long',
  year: 'numeric',
  month: 'long',
  day: 'numeric',
});

type SavedLayoutItem = Record<string, unknown> & {
  i: string;
  x: number;
  y: number;
  w: number;
  h: number;
  minW?: number;
  minH?: number;
  maxW?: number;
  maxH?: number;
};
type SavedLayout = SavedLayoutItem[];
type LayoutSignatureItem = Partial<
  Pick<SavedLayoutItem, 'i' | 'x' | 'y' | 'w' | 'h' | 'minW' | 'minH' | 'maxW' | 'maxH'>
>;

function layoutSignature(layout: ReadonlyArray<LayoutSignatureItem> | undefined): string {
  return JSON.stringify(
    (layout ?? []).map((item) => ({
      i: String(item.i ?? ''),
      x: toFiniteNumber(item.x),
      y: toFiniteNumber(item.y),
      w: toFiniteNumber(item.w),
      h: toFiniteNumber(item.h),
      minW: toFiniteNumber(item.minW),
      minH: toFiniteNumber(item.minH),
      maxW: toFiniteNumber(item.maxW),
      maxH: toFiniteNumber(item.maxH),
    }))
  );
}

// ─── Admin Dashboard ────────────────────────────────────────────────────────

const AdminDashboard: React.FC = () => {
  const { user } = useUserStore();
  const navigate = useNavigate();
  const { t } = useTranslation();
  const businessName = (user?.businessName as string | undefined) || 'Dashboard';
  const canCreateRole = user?.effectiveAccess
    ? canEffectiveAccess(user.effectiveAccess, 'hub:roles:create')
    : user?.role?.toLowerCase() === 'admin' || user?.role?.toLowerCase() === 'owner';
  const containerRef = useRef<HTMLDivElement>(null);
  const [containerWidth, setContainerWidth] = useState(0);

  useEffect(() => {
    const el = containerRef.current;
    if (!el) return;
    const ro = new ResizeObserver((entries) => {
      for (const entry of entries) {
        const nextWidth = Math.round(entry.contentRect.width);
        setContainerWidth((current) => (current === nextWidth ? current : nextWidth));
      }
    });
    ro.observe(el);
    setContainerWidth((current) => (current === el.clientWidth ? current : el.clientWidth));
    return () => ro.disconnect();
  }, []);

  const [layouts, setLayouts] = useState<{ lg: SavedLayout }>(() => ({ lg: loadLayout() }));
  const [activeWidgets, setActiveWidgets] = useState<string[]>(loadActiveWidgets);
  const [showCatalog, setShowCatalog] = useState(false);

  const handleLayoutChange = useCallback((_: Layout, allLayouts: ResponsiveLayouts) => {
    const nextLayout = (allLayouts.lg ?? []) as unknown as SavedLayout;
    setLayouts((current) => {
      if (layoutSignature(current.lg) === layoutSignature(nextLayout)) return current;
      saveLayout(nextLayout);
      return { lg: nextLayout };
    });
  }, []);

  const addWidget = useCallback(
    (widgetId: string) => {
      const config = WIDGET_CATALOG.find((w) => w.id === widgetId);
      if (!config) return;

      const currentLayout = layouts.lg || [];
      let maxY = 0;
      for (const item of currentLayout) {
        const bottom = item.y + item.h;
        if (bottom > maxY) maxY = bottom;
      }

      const newLayout = [
        ...currentLayout,
        {
          i: widgetId,
          x: 0,
          y: maxY,
          w: config.defaultW,
          h: config.defaultH,
          minW: config.minW,
          minH: config.minH,
          maxW: config.maxW,
          maxH: config.maxH,
        },
      ];
      const newActive = [...activeWidgets, widgetId];

      setLayouts({ lg: newLayout });
      setActiveWidgets(newActive);
      saveLayout(newLayout);
      saveActiveWidgets(newActive);
    },
    [layouts, activeWidgets]
  );

  const removeWidget = useCallback(
    (widgetId: string) => {
      const newLayout = (layouts.lg || []).filter((l: { i: string }) => l.i !== widgetId);
      const newActive = activeWidgets.filter((id) => id !== widgetId);
      setLayouts({ lg: newLayout });
      setActiveWidgets(newActive);
      saveLayout(newLayout);
      saveActiveWidgets(newActive);
    },
    [layouts, activeWidgets]
  );

  const resetLayout = useCallback(() => {
    const ids = DEFAULT_LAYOUT.map((l) => l.i);
    setLayouts({ lg: DEFAULT_LAYOUT });
    setActiveWidgets(ids);
    saveLayout(DEFAULT_LAYOUT);
    saveActiveWidgets(ids);
    setShowCatalog(false);
  }, []);

  return (
    <div ref={containerRef} className="space-y-4 overflow-hidden">
      {/* Header */}
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-lg font-semibold" style={{ color: 'var(--dash-text)' }}>
            {businessName}
          </h1>
          <p className="text-xs mt-0.5" style={{ color: 'var(--dash-text-dim)' }}>
            {today}
          </p>
        </div>
        <div className="flex w-full items-center gap-2 sm:w-auto">
          {canCreateRole && (
            <Button
              type="button"
              size="lg"
              className="min-h-11 flex-1 sm:flex-none"
              onClick={() => void navigate({ to: '/roles', search: { create: 1 } })}
              aria-label={t('dashboard.createCustomRole')}
            >
              <Plus size={16} />
              {t('dashboard.createCustomRole')}
            </Button>
          )}
          <button
            type="button"
            onClick={() => setShowCatalog(true)}
            className="flex min-h-11 min-w-11 items-center justify-center rounded-lg transition-colors"
            style={{ color: 'var(--dash-text-muted)', border: '1px solid var(--dash-border)' }}
            onMouseEnter={(e) => {
              (e.currentTarget as HTMLElement).style.background = 'var(--dash-hover)';
            }}
            onMouseLeave={(e) => {
              (e.currentTarget as HTMLElement).style.background = 'transparent';
            }}
            title="Customize dashboard"
          >
            <Settings2 size={15} />
          </button>
        </div>
      </div>

      {/* Grid */}
      {containerWidth > 0 && (
        <ResponsiveGridLayout
          className="layout"
          layouts={layouts}
          breakpoints={{ lg: 1200, md: 996, sm: 768, xs: 480, xxs: 0 }}
          cols={{ lg: 4, md: 3, sm: 2, xs: 1, xxs: 1 }}
          rowHeight={80}
          width={containerWidth}
          onLayoutChange={handleLayoutChange}
          isDraggable={showCatalog}
          isResizable={showCatalog}
          draggableHandle=".widget-drag-handle"
          containerPadding={[0, 0]}
          margin={[12, 12]}
        >
          {activeWidgets.map((widgetId) => {
            const Widget = getWidgetComponent(widgetId);
            if (!Widget) return <div key={widgetId} />;

            return (
              <div
                key={widgetId}
                className="rounded-xl overflow-hidden relative"
                style={{
                  background: 'var(--dash-card)',
                  border: showCatalog
                    ? '1px dashed rgba(69,131,196,0.4)'
                    : '1px solid var(--dash-border)',
                }}
              >
                {showCatalog && (
                  <div className="absolute top-1.5 right-1.5 z-10 flex items-center gap-1">
                    <div
                      className="widget-drag-handle cursor-grab h-6 w-6 rounded-md flex items-center justify-center"
                      style={{ background: 'var(--dash-hover)', color: 'var(--dash-text-dim)' }}
                    >
                      <GripVertical size={12} />
                    </div>
                    <button
                      type="button"
                      onClick={() => removeWidget(widgetId)}
                      className="h-6 w-6 rounded-md flex items-center justify-center"
                      style={{
                        background: 'var(--status-cancelled-bg)',
                        color: 'var(--status-cancelled)',
                      }}
                    >
                      <X size={12} />
                    </button>
                  </div>
                )}
                <div className="h-full">
                  <ErrorBoundary>
                    <Widget />
                  </ErrorBoundary>
                </div>
              </div>
            );
          })}
        </ResponsiveGridLayout>
      )}

      {/* Widget catalog modal (portaled to body) */}
      {createPortal(
        <AnimatePresence>
          {showCatalog && (
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              className="fixed inset-0 z-50 flex items-center justify-center"
              style={{ background: 'rgba(0,0,0,0.4)' }}
              onClick={() => setShowCatalog(false)}
            >
              <motion.div
                initial={{ scale: 0.95, opacity: 0 }}
                animate={{ scale: 1, opacity: 1 }}
                exit={{ scale: 0.95, opacity: 0 }}
                onClick={(e) => e.stopPropagation()}
                className="rounded-2xl w-full max-w-lg max-h-[80vh] overflow-hidden flex flex-col"
                style={{
                  background: 'var(--dash-surface)',
                  border: '1px solid var(--dash-border)',
                }}
              >
                {/* Modal header */}
                <div
                  className="flex items-center justify-between px-6 py-4"
                  style={{ borderBottom: '1px solid var(--dash-border)' }}
                >
                  <div>
                    <h2 className="text-base font-semibold" style={{ color: 'var(--dash-text)' }}>
                      Customize dashboard
                    </h2>
                    <p className="text-xs mt-0.5" style={{ color: 'var(--dash-text-muted)' }}>
                      Add, remove, or rearrange widgets
                    </p>
                  </div>
                  <button
                    type="button"
                    onClick={() => setShowCatalog(false)}
                    className="h-7 w-7 rounded-lg flex items-center justify-center hover:bg-[var(--dash-hover)]"
                    style={{ color: 'var(--dash-text-muted)' }}
                  >
                    <X size={16} />
                  </button>
                </div>

                {/* Widget list */}
                <div className="flex-1 overflow-y-auto px-6 py-4 space-y-2">
                  <p
                    className="text-[11px] font-semibold uppercase tracking-wider mb-2"
                    style={{ color: 'var(--dash-text-dim)' }}
                  >
                    Widgets
                  </p>
                  {WIDGET_CATALOG.map((w) => {
                    const isActive = activeWidgets.includes(w.id);
                    return (
                      <div
                        key={w.id}
                        className="flex items-center justify-between py-2.5 px-3 rounded-lg"
                        style={{ background: isActive ? 'rgba(69,131,196,0.05)' : 'transparent' }}
                      >
                        <div className="flex-1 min-w-0">
                          <p className="text-sm font-medium" style={{ color: 'var(--dash-text)' }}>
                            {w.label}
                          </p>
                          <p className="text-[11px]" style={{ color: 'var(--dash-text-dim)' }}>
                            {w.description}
                          </p>
                        </div>
                        {isActive ? (
                          <button
                            type="button"
                            onClick={() => removeWidget(w.id)}
                            className="shrink-0 text-xs font-medium px-3 py-1.5 rounded-md"
                            style={{
                              color: 'var(--status-cancelled)',
                              background: 'var(--status-cancelled-bg)',
                            }}
                          >
                            Remove
                          </button>
                        ) : (
                          <button
                            type="button"
                            onClick={() => addWidget(w.id)}
                            className="shrink-0 text-xs font-medium px-3 py-1.5 rounded-md"
                            style={{
                              color: 'var(--accent)',
                              background: 'var(--accent-muted)',
                            }}
                          >
                            <Plus size={12} className="inline mr-1" />
                            Add
                          </button>
                        )}
                      </div>
                    );
                  })}
                </div>

                {/* Footer */}
                <div
                  className="px-6 py-3 flex justify-end"
                  style={{ borderTop: '1px solid var(--dash-border)' }}
                >
                  <button
                    type="button"
                    onClick={resetLayout}
                    className="text-xs font-medium px-3 py-1.5 rounded-md transition-colors"
                    style={{ color: 'var(--dash-text-muted)', background: 'var(--dash-hover)' }}
                  >
                    Reset to default
                  </button>
                </div>
              </motion.div>
            </motion.div>
          )}
        </AnimatePresence>,
        document.body
      )}
    </div>
  );
};

// ─── Staff (fallback) ───────────────────────────────────────────────────────

const StaffDashboard: React.FC<{ userApps: string[] }> = ({ userApps }) => {
  React.useEffect(() => {
    if (userApps.length > 0) {
      const firstApp = userApps[0];
      const validApps = ['pos', 'erp', 'hr', 'crm', 'finance', 'flow', 'insights'];
      if (validApps.includes(firstApp)) {
        window.location.href = `https://${firstApp}.rtylr.com`;
      }
    }
  }, [userApps]);

  return (
    <div className="max-w-xl mx-auto py-12 text-center space-y-4">
      <h2 className="text-xl font-bold" style={{ color: 'var(--dash-text)' }}>
        Welcome to rtylr
      </h2>
      <p className="text-sm" style={{ color: 'var(--dash-text-muted)' }}>
        Select an app below to access your workspace.
      </p>
      <div className="flex flex-wrap justify-center gap-3 pt-2">
        {userApps.length > 0 ? (
          userApps.map((app) => (
            <a
              key={app}
              href={`https://${app}.rtylr.com`}
              className="px-4 py-2 rounded-lg text-xs font-semibold uppercase tracking-wider transition-colors hover:bg-[var(--dash-hover)]"
              style={{
                background: 'var(--dash-card)',
                border: '1px solid var(--dash-border)',
                color: 'var(--dash-text)',
              }}
            >
              Open {app}
            </a>
          ))
        ) : (
          <p className="text-xs" style={{ color: 'var(--dash-text-dim)' }}>
            No apps assigned to your staff account. Please contact your workspace administrator.
          </p>
        )}
      </div>
    </div>
  );
};

// ─── Root ───────────────────────────────────────────────────────────────────

const VendorDashboard: React.FC = () => {
  const { user } = useUserStore();
  const role = (user?.role ?? '').toLowerCase();
  const isAdmin = user?.effectiveAccess
    ? user.effectiveAccess.isOwner || user.effectiveAccess.isAdmin
    : role === 'admin' || role === 'owner';
  const userApps: string[] = user?.effectiveAccess
    ? user.effectiveAccess.apps
        .filter((app) => app.licensed && app.hasAccess)
        .map((app) => app.appId)
    : (user?.apps ?? []);
  if (!isAdmin) return <StaffDashboard userApps={userApps} />;
  return (
    <>
      <HelmetContainer title="Dashboard · rtylr" />
      <AdminDashboard />
    </>
  );
};

export default VendorDashboard;
