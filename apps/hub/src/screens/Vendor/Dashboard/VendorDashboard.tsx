import React, { useState, useCallback, useMemo, useRef } from 'react';
import { ResponsiveGridLayout, useContainerWidth } from 'react-grid-layout';
import { motion, AnimatePresence } from 'framer-motion';
import { Settings2, X, Plus, GripVertical, Check } from 'lucide-react';
import { useUserStore } from '../../../stores/userStore';
import {
  getWidgetComponent,
  WIDGET_CATALOG,
  loadLayout,
  saveLayout,
  loadActiveWidgets,
  saveActiveWidgets,
} from './widgets';

import 'react-grid-layout/css/styles.css';
import 'react-resizable/css/styles.css';

const today = new Date().toLocaleDateString('en-US', {
  weekday: 'long', year: 'numeric', month: 'long', day: 'numeric',
});

// ─── Admin Dashboard ────────────────────────────────────────────────────────

const AdminDashboard: React.FC = () => {
  const { user } = useUserStore();
  const businessName = (user as any)?.businessName || 'Dashboard';
  const containerRef = useRef<HTMLDivElement>(null);
  const { width: containerWidth } = useContainerWidth(containerRef);

  const [layouts, setLayouts] = useState(() => ({ lg: loadLayout() }));
  const [activeWidgets, setActiveWidgets] = useState<string[]>(loadActiveWidgets);
  const [editing, setEditing] = useState(false);
  const [showCatalog, setShowCatalog] = useState(false);

  const handleLayoutChange = useCallback((_: any, allLayouts: any) => {
    setLayouts(allLayouts);
    if (allLayouts.lg) saveLayout(allLayouts.lg);
  }, []);

  const addWidget = useCallback((widgetId: string) => {
    const config = WIDGET_CATALOG.find((w) => w.id === widgetId);
    if (!config) return;

    const currentLayout = layouts.lg || [];
    let maxY = 0;
    for (const item of currentLayout) {
      const bottom = item.y + item.h;
      if (bottom > maxY) maxY = bottom;
    }

    const newLayout = [...currentLayout, {
      i: widgetId, x: 0, y: maxY,
      w: config.defaultW, h: config.defaultH,
      minW: config.minW, minH: config.minH,
      maxW: config.maxW, maxH: config.maxH,
    }];
    const newActive = [...activeWidgets, widgetId];

    setLayouts({ lg: newLayout });
    setActiveWidgets(newActive);
    saveLayout(newLayout);
    saveActiveWidgets(newActive);
  }, [layouts, activeWidgets]);

  const removeWidget = useCallback((widgetId: string) => {
    const newLayout = (layouts.lg || []).filter((l: any) => l.i !== widgetId);
    const newActive = activeWidgets.filter((id) => id !== widgetId);
    setLayouts({ lg: newLayout });
    setActiveWidgets(newActive);
    saveLayout(newLayout);
    saveActiveWidgets(newActive);
  }, [layouts, activeWidgets]);

  const availableToAdd = useMemo(() => {
    return WIDGET_CATALOG.filter((w) => !activeWidgets.includes(w.id));
  }, [activeWidgets]);

  return (
    <div ref={containerRef} className="space-y-4 overflow-hidden">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-lg font-semibold" style={{ color: 'var(--hub-text)' }}>{businessName}</h1>
          <p className="text-xs mt-0.5" style={{ color: 'var(--hub-text-dim)' }}>{today}</p>
        </div>
        <button
          onClick={() => {
            if (editing) { setEditing(false); setShowCatalog(false); }
            else setShowCatalog(true);
          }}
          className="h-8 w-8 rounded-lg flex items-center justify-center transition-colors"
          style={{ color: 'var(--hub-text-muted)', border: '1px solid var(--hub-border)' }}
          onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.background = 'var(--hub-hover)'; }}
          onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.background = 'transparent'; }}
          title="Customize dashboard"
        >
          <Settings2 size={15} />
        </button>
      </div>

      {/* Grid */}
      {containerWidth > 0 && (
        <ResponsiveGridLayout
          className="layout"
          layouts={layouts}
          breakpoints={{ lg: 0 }}
          cols={{ lg: 4 }}
          rowHeight={80}
          width={containerWidth}
          onLayoutChange={handleLayoutChange}
          isDraggable={editing}
          isResizable={editing}
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
                  background: 'var(--hub-card)',
                  border: editing ? '1px dashed rgba(69,131,196,0.4)' : '1px solid var(--hub-border)',
                }}
              >
                {editing && (
                  <div className="absolute top-1.5 right-1.5 z-10 flex items-center gap-1">
                    <div className="widget-drag-handle cursor-grab h-6 w-6 rounded-md flex items-center justify-center" style={{ background: 'var(--hub-hover)', color: 'var(--hub-text-dim)' }}>
                      <GripVertical size={12} />
                    </div>
                    <button
                      onClick={() => removeWidget(widgetId)}
                      className="h-6 w-6 rounded-md flex items-center justify-center"
                      style={{ background: 'rgba(239,68,68,0.1)', color: '#ef4444' }}
                    >
                      <X size={12} />
                    </button>
                  </div>
                )}
                <div className="h-full">
                  <Widget />
                </div>
              </div>
            );
          })}
        </ResponsiveGridLayout>
      )}

      {/* Widget catalog modal */}
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
              style={{ background: 'var(--hub-surface)', border: '1px solid var(--hub-border)' }}
            >
              {/* Modal header */}
              <div className="flex items-center justify-between px-6 py-4" style={{ borderBottom: '1px solid var(--hub-border)' }}>
                <div>
                  <h2 className="text-base font-semibold" style={{ color: 'var(--hub-text)' }}>Customize dashboard</h2>
                  <p className="text-xs mt-0.5" style={{ color: 'var(--hub-text-muted)' }}>Add, remove, or rearrange widgets</p>
                </div>
                <button
                  onClick={() => setShowCatalog(false)}
                  className="h-7 w-7 rounded-lg flex items-center justify-center hover:bg-[var(--hub-hover)]"
                  style={{ color: 'var(--hub-text-muted)' }}
                >
                  <X size={16} />
                </button>
              </div>

              {/* Toggle edit mode */}
              <div className="px-6 py-3" style={{ borderBottom: '1px solid var(--hub-border)' }}>
                <button
                  onClick={() => { setEditing(!editing); if (!editing) setShowCatalog(false); }}
                  className="flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium w-full transition-colors"
                  style={{
                    background: editing ? 'rgba(69,131,196,0.1)' : 'var(--hub-hover)',
                    color: editing ? '#4583C4' : 'var(--hub-text-muted)',
                    border: editing ? '1px solid rgba(69,131,196,0.3)' : '1px solid var(--hub-border)',
                  }}
                >
                  {editing ? <><Check size={14} /> Done rearranging</> : <><GripVertical size={14} /> Drag to rearrange</>}
                </button>
              </div>

              {/* Widget list */}
              <div className="flex-1 overflow-y-auto px-6 py-4 space-y-2">
                <p className="text-[11px] font-semibold uppercase tracking-wider mb-2" style={{ color: 'var(--hub-text-dim)' }}>Widgets</p>
                {WIDGET_CATALOG.map((w) => {
                  const isActive = activeWidgets.includes(w.id);
                  return (
                    <div
                      key={w.id}
                      className="flex items-center justify-between py-2.5 px-3 rounded-lg"
                      style={{ background: isActive ? 'rgba(69,131,196,0.05)' : 'transparent' }}
                    >
                      <div className="flex-1 min-w-0">
                        <p className="text-sm font-medium" style={{ color: 'var(--hub-text)' }}>{w.label}</p>
                        <p className="text-[11px]" style={{ color: 'var(--hub-text-dim)' }}>{w.description}</p>
                      </div>
                      {isActive ? (
                        <button
                          onClick={() => removeWidget(w.id)}
                          className="shrink-0 text-xs font-medium px-3 py-1.5 rounded-md"
                          style={{ color: '#ef4444', background: 'rgba(239,68,68,0.08)' }}
                        >
                          Remove
                        </button>
                      ) : (
                        <button
                          onClick={() => addWidget(w.id)}
                          className="shrink-0 text-xs font-medium px-3 py-1.5 rounded-md"
                          style={{ color: '#4583C4', background: 'rgba(69,131,196,0.08)' }}
                        >
                          <Plus size={12} className="inline mr-1" />Add
                        </button>
                      )}
                    </div>
                  );
                })}
              </div>
            </motion.div>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
};

// ─── Staff (fallback) ───────────────────────────────────────────────────────

const StaffDashboard: React.FC<{ userApps: string[] }> = () => (
  <div className="flex items-center justify-center h-64">
    <p className="text-sm" style={{ color: 'var(--hub-text-muted)' }}>Redirecting...</p>
  </div>
);

// ─── Root ───────────────────────────────────────────────────────────────────

const VendorDashboard: React.FC = () => {
  const { user, role } = useUserStore();
  const isAdmin = role === 'admin' || role === 'owner' || !role;
  const userApps: string[] = (user as any)?.apps ?? [];
  if (!isAdmin) return <StaffDashboard userApps={userApps} />;
  return <AdminDashboard />;
};

export default VendorDashboard;
