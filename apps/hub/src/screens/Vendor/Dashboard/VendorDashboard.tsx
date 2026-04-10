import React, { useState, useCallback, useMemo, useRef } from 'react';
import { ResponsiveGridLayout, useContainerWidth } from 'react-grid-layout';
import { motion, AnimatePresence } from 'framer-motion';
import { Pencil, X, Plus, GripVertical, Check } from 'lucide-react';
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

// ─── Admin Dashboard (customizable) ─────────────────────────────────────────

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
    // Find the lowest y position to place new widget at bottom
    let maxY = 0;
    for (const item of currentLayout) {
      const bottom = item.y + item.h;
      if (bottom > maxY) maxY = bottom;
    }

    const newItem = {
      i: widgetId,
      x: 0,
      y: maxY,
      w: config.defaultW,
      h: config.defaultH,
      minW: config.minW,
      minH: config.minH,
      maxW: config.maxW,
      maxH: config.maxH,
    };

    const newLayout = [...currentLayout, newItem];
    const newActive = [...activeWidgets, widgetId];

    setLayouts({ lg: newLayout });
    setActiveWidgets(newActive);
    saveLayout(newLayout);
    saveActiveWidgets(newActive);
    setShowCatalog(false);
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
    <div ref={containerRef} className="space-y-4">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-lg font-semibold" style={{ color: 'var(--hub-text)' }}>{businessName}</h1>
          <p className="text-xs mt-0.5" style={{ color: 'var(--hub-text-dim)' }}>{today}</p>
        </div>
        <button
          onClick={() => { setEditing(!editing); if (editing) setShowCatalog(false); }}
          className="flex items-center gap-2 px-3 h-8 rounded-lg text-xs font-medium transition-colors"
          style={{
            background: editing ? '#4583C4' : 'var(--hub-hover)',
            color: editing ? '#fff' : 'var(--hub-text-muted)',
            border: editing ? 'none' : '1px solid var(--hub-border)',
          }}
        >
          {editing ? <><Check size={13} /> Done</> : <><Pencil size={13} /> Edit</>}
        </button>
      </div>

      {/* Grid */}
      <ResponsiveGridLayout
        className="layout"
        layouts={layouts}
        breakpoints={{ lg: 0 }}
        cols={{ lg: 4 }}
        rowHeight={80}
        width={containerWidth || 800}
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
              className="rounded-xl overflow-hidden"
              style={{
                background: 'var(--hub-card)',
                border: editing ? '2px dashed var(--hub-border)' : '1px solid var(--hub-border)',
              }}
            >
              {editing && (
                <div className="absolute top-0 left-0 right-0 z-10 flex items-center justify-between px-2 py-1" style={{ background: 'var(--hub-surface)' }}>
                  <div className="widget-drag-handle cursor-grab flex items-center gap-1 text-[10px] font-medium" style={{ color: 'var(--hub-text-dim)' }}>
                    <GripVertical size={12} /> Drag
                  </div>
                  <button
                    onClick={() => removeWidget(widgetId)}
                    className="h-5 w-5 rounded flex items-center justify-center hover:bg-[var(--hub-hover)]"
                    style={{ color: '#ef4444' }}
                  >
                    <X size={12} />
                  </button>
                </div>
              )}
              <div className={editing ? 'pt-6 h-full' : 'h-full'}>
                <Widget />
              </div>
            </div>
          );
        })}
      </ResponsiveGridLayout>

      {/* Add widget button (only in edit mode) */}
      {editing && (
        <motion.div initial={{ opacity: 0, y: 8 }} animate={{ opacity: 1, y: 0 }}>
          <button
            onClick={() => setShowCatalog(!showCatalog)}
            className="flex items-center gap-2 w-full px-4 py-3 rounded-xl text-sm font-medium transition-colors"
            style={{
              background: 'var(--hub-card)',
              border: '2px dashed var(--hub-border)',
              color: 'var(--hub-text-muted)',
            }}
          >
            <Plus size={16} />
            Add widget
          </button>
        </motion.div>
      )}

      {/* Widget catalog */}
      <AnimatePresence>
        {showCatalog && (
          <motion.div
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: 10 }}
            className="rounded-xl p-4"
            style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
          >
            <p className="text-sm font-semibold mb-3" style={{ color: 'var(--hub-text)' }}>Available widgets</p>
            {availableToAdd.length === 0 ? (
              <p className="text-xs" style={{ color: 'var(--hub-text-muted)' }}>All widgets are already on your dashboard</p>
            ) : (
              <div className="grid grid-cols-2 gap-2">
                {availableToAdd.map((w) => (
                  <button
                    key={w.id}
                    onClick={() => addWidget(w.id)}
                    className="flex items-start gap-3 p-3 rounded-lg text-left transition-colors hover:bg-[var(--hub-hover)]"
                    style={{ border: '1px solid var(--hub-border)' }}
                  >
                    <Plus size={14} className="mt-0.5 shrink-0" style={{ color: '#4583C4' }} />
                    <div>
                      <p className="text-xs font-semibold" style={{ color: 'var(--hub-text)' }}>{w.label}</p>
                      <p className="text-[11px] mt-0.5" style={{ color: 'var(--hub-text-dim)' }}>{w.description}</p>
                    </div>
                  </button>
                ))}
              </div>
            )}
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
};

// ─── Staff Dashboard ────────────────────────────────────────────────────────
// Staff users don't see this — they get StaffView from BusinessRoute.
// This is here as a fallback.

const StaffDashboard: React.FC<{ userApps: string[] }> = () => {
  return (
    <div className="flex items-center justify-center h-64">
      <p className="text-sm" style={{ color: 'var(--hub-text-muted)' }}>Redirecting...</p>
    </div>
  );
};

// ─── Root ───────────────────────────────────────────────────────────────────

const VendorDashboard: React.FC = () => {
  const { user, role } = useUserStore();
  const isAdmin = role === 'admin' || role === 'owner' || !role;
  const userApps: string[] = (user as any)?.apps ?? [];

  if (!isAdmin) return <StaffDashboard userApps={userApps} />;
  return <AdminDashboard />;
};

export default VendorDashboard;
