import React, { useState, useCallback, useMemo, useRef, useEffect } from 'react';
import { createPortal } from 'react-dom';
import { ResponsiveGridLayout, type Layout, type ResponsiveLayouts } from 'react-grid-layout';
import { motion, AnimatePresence } from 'motion/react';
import { Settings2, X, Plus, GripVertical } from 'lucide-react';
import { ErrorBoundary } from '@voxire/ui';
import { HelmetContainer } from '@voxire/components';
import { useUserStore } from '../../../shared/stores/userStore';
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

// react-grid-layout's ResponsiveGridLayoutProps misses draggableHandle, isDraggable, isResizable
// eslint-disable-next-line @typescript-eslint/no-explicit-any
const GridLayout = ResponsiveGridLayout as React.ComponentType<any>;

// ─── Admin Dashboard ────────────────────────────────────────────────────────

const AdminDashboard: React.FC = () => {
  const { user } = useUserStore();
  const businessName = (user?.businessName as string | undefined) || 'Dashboard';
  const containerRef = useRef<HTMLDivElement>(null);
  const [containerWidth, setContainerWidth] = useState(0);

  useEffect(() => {
    const el = containerRef.current;
    if (!el) return;
    const ro = new ResizeObserver((entries) => {
      for (const entry of entries) {
        setContainerWidth(entry.contentRect.width);
      }
    });
    ro.observe(el);
    setContainerWidth(el.clientWidth);
    return () => ro.disconnect();
  }, []);

  const [layouts, setLayouts] = useState(() => ({ lg: loadLayout() }));
  const [activeWidgets, setActiveWidgets] = useState<string[]>(loadActiveWidgets);
  const [showCatalog, setShowCatalog] = useState(false);

  const handleLayoutChange = useCallback((_: Layout[], allLayouts: ResponsiveLayouts) => {
    setLayouts(allLayouts as unknown as { lg: Record<string, unknown>[] });
    if (allLayouts.lg) saveLayout(allLayouts.lg as unknown as Record<string, unknown>[]);
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

  const availableToAdd = useMemo(() => {
    return WIDGET_CATALOG.filter((w) => !activeWidgets.includes(w.id));
  }, [activeWidgets]);

  return (
    <div ref={containerRef} className="space-y-4 overflow-hidden">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-lg font-semibold" style={{ color: 'var(--dash-text)' }}>
            {businessName}
          </h1>
          <p className="text-xs mt-0.5" style={{ color: 'var(--dash-text-dim)' }}>
            {today}
          </p>
        </div>
        <button
          onClick={() => setShowCatalog(true)}
          className="h-8 w-8 rounded-lg flex items-center justify-center transition-colors"
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

      {/* Grid */}
      {containerWidth > 0 && (
        <GridLayout
          className="layout"
          layouts={layouts}
          breakpoints={{ lg: 0 }}
          cols={{ lg: 4 }}
          rowHeight={80}
          width={containerWidth}
          onLayoutChange={handleLayoutChange}
          {...({ isDraggable: showCatalog, isResizable: showCatalog } as Record<string, unknown>)}
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
        </GridLayout>
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

const StaffDashboard: React.FC<{ userApps: string[] }> = () => (
  <div className="flex items-center justify-center h-64">
    <p className="text-sm" style={{ color: 'var(--dash-text-muted)' }}>
      Redirecting...
    </p>
  </div>
);

// ─── Root ───────────────────────────────────────────────────────────────────

const VendorDashboard: React.FC = () => {
  const { user, role } = useUserStore();
  const isAdmin = role === 'admin' || role === 'owner' || !role;
  const userApps: string[] = (user?.apps as string[] | undefined) ?? [];
  if (!isAdmin) return <StaffDashboard userApps={userApps} />;
  return (
    <>
      <HelmetContainer title="Dashboard · rtylr" />
      <AdminDashboard />
    </>
  );
};

export default VendorDashboard;
