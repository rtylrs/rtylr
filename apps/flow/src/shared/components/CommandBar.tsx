import React, { useEffect, useMemo, useRef, useState } from 'react';
import { useNavigate } from '@tanstack/react-router';
import { motion, AnimatePresence } from 'motion/react';
import {
  Search,
  LayoutGrid,
  ListTodo,
  Inbox,
  Calendar,
  Timer,
  Users,
  Milestone,
  Map,
  BarChart3,
  Settings,
  Plus,
  FolderKanban,
  Eye,
  Archive,
} from 'lucide-react';
import { useQuery } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import { useTranslation } from 'react-i18next';
import { FLAG_KEYS, useFeatureFlag } from '@voxire/config';
import { useFlowStore } from '../stores/flowStore';
import { useTasks } from '../../features/tasks/useTasks';
import { useProjects } from '../../features/projects/useProjects';
import { apiTaskToTask } from '../lib/taskMapper';
import { flowApi } from '../services/flowApi';

interface CommandItem {
  id: string;
  label: string;
  hint?: string;
  icon: React.ComponentType<{ size?: number }>;
  action: () => void;
  keywords?: string[];
  group: 'nav' | 'task' | 'project' | 'action';
}

export const CommandBar: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const showPreviewRoutes = useFeatureFlag(FLAG_KEYS.APP_FLOW_PREVIEW_ROUTES);
  const { commandBarOpen, setCommandBarOpen, setQuickCreateOpen } = useFlowStore();
  const { data: apiTasks = [] } = useTasks({});
  const { data: apiProjects = [] } = useProjects();
  const tasks = apiTasks.map(apiTaskToTask);
  const projects = apiProjects;
  const [query, setQuery] = useState('');
  const [debouncedQuery, setDebouncedQuery] = useState('');
  const [selected, setSelected] = useState(0);
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    const t = setTimeout(() => setDebouncedQuery(query), 300);
    return () => clearTimeout(t);
  }, [query]);

  const { data: searchResults } = useQuery({
    queryKey: ['task-search', debouncedQuery],
    queryFn: () => unwrap(flowApi.searchTasks(debouncedQuery)).then((r) => r.tasks ?? []),
    // Search is business-wide now. This also required a non-empty workspace id,
    // which has not existed since that layer was stubbed out, so the command
    // bar's task search never ran at all.
    enabled: debouncedQuery.trim().length >= 3,
  });

  useEffect(() => {
    if (commandBarOpen) {
      setQuery('');
      setSelected(0);
      setTimeout(() => inputRef.current?.focus(), 10);
    }
  }, [commandBarOpen]);

  const items = useMemo<CommandItem[]>(() => {
    // These four are the destinations router.tsx wraps in
    // AppFlagGuard(APP_FLOW_PREVIEW_ROUTES). Offering them with the flag off
    // walked the user into "Preview route unavailable"; the sidebar, the
    // project view tabs, Settings and the shortcuts panel all hide them, so
    // the command bar was the only way in.
    const previewOnly = new Set(['nav-timeline', 'nav-roadmap', 'nav-insights', 'nav-team']);
    const navItems: CommandItem[] = [
      {
        id: 'nav-inbox',
        label: t('flow.commandBar.goToInbox'),
        icon: Inbox,
        action: () => navigate({ to: '/inbox' }),
        group: 'nav',
        hint: 'G I',
      },
      {
        id: 'nav-my',
        label: t('flow.shell.myTasks'),
        icon: ListTodo,
        action: () => navigate({ to: '/my-issues' }),
        group: 'nav',
        hint: 'G M',
      },
      {
        id: 'nav-projects',
        label: t('flow.allProjects.title'),
        icon: LayoutGrid,
        action: () => navigate({ to: '/projects' }),
        group: 'nav',
      },
      {
        id: 'nav-cal',
        label: t('flow.shell.calendar'),
        icon: Calendar,
        action: () => navigate({ to: '/calendar' }),
        group: 'nav',
      },
      {
        id: 'nav-timeline',
        label: t('flow.shell.timeline'),
        icon: Timer,
        action: () => navigate({ to: '/timeline' }),
        group: 'nav',
      },
      {
        id: 'nav-cycles',
        label: t('flow.shell.cycles'),
        icon: Timer,
        action: () => navigate({ to: '/cycles' }),
        group: 'nav',
      },
      {
        id: 'nav-milestones',
        label: t('flow.shell.milestones'),
        icon: Milestone,
        action: () => navigate({ to: '/milestones' }),
        group: 'nav',
      },
      {
        id: 'nav-roadmap',
        label: t('flow.shell.roadmap'),
        icon: Map,
        action: () => navigate({ to: '/roadmap' }),
        group: 'nav',
      },
      {
        id: 'nav-insights',
        label: t('flow.shell.insights'),
        icon: BarChart3,
        action: () => navigate({ to: '/insights' }),
        group: 'nav',
      },
      {
        id: 'nav-team',
        label: t('flow.shell.team'),
        icon: Users,
        action: () => navigate({ to: '/team' }),
        group: 'nav',
      },
      {
        id: 'nav-triage',
        label: t('flow.shell.triage'),
        icon: Eye,
        action: () => navigate({ to: '/triage' }),
        group: 'nav',
      },
      {
        id: 'nav-archive',
        label: t('flow.shell.archive'),
        icon: Archive,
        action: () => navigate({ to: '/archive' }),
        group: 'nav',
      },
      {
        id: 'nav-views',
        label: t('flow.shell.views'),
        icon: Eye,
        action: () => navigate({ to: '/views' }),
        group: 'nav',
      },
      {
        id: 'nav-settings',
        label: t('nav.settings'),
        icon: Settings,
        action: () => navigate({ to: '/settings' }),
        group: 'nav',
      },
    ];
    const actionItems: CommandItem[] = [
      {
        id: 'act-new-task',
        label: t('flow.shell.newTask'),
        icon: Plus,
        hint: 'C',
        action: () => setQuickCreateOpen(true),
        group: 'action',
      },
    ];
    const taskSource =
      query.trim().length >= 3 ? (searchResults?.map(apiTaskToTask) ?? []) : tasks.slice(0, 30);
    const taskItems: CommandItem[] = taskSource.map((t) => ({
      id: `task-${t.id}`,
      label: t.title,
      hint: t.identifier,
      icon: ListTodo,
      action: () => navigate({ to: `/task/${t.id}` as string }),
      keywords: [t.identifier, t.projectName],
      group: 'task',
    }));
    const projectItems: CommandItem[] = projects.map((p) => ({
      id: `proj-${p.id}`,
      label: p.name,
      icon: FolderKanban,
      action: () => navigate({ to: '/p/$projectId', params: { projectId: p.id } }),
      group: 'project',
    }));
    const reachableNavItems = showPreviewRoutes
      ? navItems
      : navItems.filter((item) => !previewOnly.has(item.id));
    return [...actionItems, ...reachableNavItems, ...projectItems, ...taskItems];
  }, [tasks, projects, navigate, setQuickCreateOpen, query, searchResults, t, showPreviewRoutes]);

  const filtered = useMemo(() => {
    if (!query.trim()) return items;
    const q = query.toLowerCase();
    return items.filter(
      (i) =>
        i.label.toLowerCase().includes(q) ||
        i.hint?.toLowerCase().includes(q) ||
        i.keywords?.some((k) => k.toLowerCase().includes(q))
    );
  }, [items, query]);

  useEffect(() => {
    if (selected >= filtered.length) setSelected(0);
  }, [filtered, selected]);

  const run = (item: CommandItem) => {
    item.action();
    setCommandBarOpen(false);
  };

  const handleKey = (e: React.KeyboardEvent) => {
    if (e.key === 'Escape') {
      setCommandBarOpen(false);
    } else if (e.key === 'ArrowDown') {
      e.preventDefault();
      setSelected((s) => Math.min(filtered.length - 1, s + 1));
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      setSelected((s) => Math.max(0, s - 1));
    } else if (e.key === 'Enter') {
      e.preventDefault();
      const item = filtered[selected];
      if (item) run(item);
    }
  };

  const grouped = useMemo(() => {
    const groups: Record<string, CommandItem[]> = { action: [], nav: [], project: [], task: [] };
    filtered.forEach((i) => groups[i.group].push(i));
    return groups;
  }, [filtered]);

  const groupLabels: Record<string, string> = {
    action: t('flow.commandBar.groupActions'),
    nav: t('flow.commandBar.groupNavigation'),
    project: t('nav.projects'),
    task: t('nav.tasks'),
  };

  return (
    <AnimatePresence>
      {commandBarOpen && (
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          transition={{ duration: 0.12 }}
          onClick={() => setCommandBarOpen(false)}
          className="fixed inset-0 z-[100] flex items-start justify-center px-3 pt-4 sm:px-4 sm:pt-[10vh]"
          style={{ background: 'rgba(0,0,0,0.4)', backdropFilter: 'blur(4px)' }}
        >
          <motion.div
            initial={{ opacity: 0, y: -10, scale: 0.98 }}
            animate={{ opacity: 1, y: 0, scale: 1 }}
            exit={{ opacity: 0, y: -10, scale: 0.98 }}
            transition={{ duration: 0.14 }}
            onClick={(e) => e.stopPropagation()}
            onKeyDown={handleKey}
            className="max-h-[min(85dvh,720px)] w-full max-w-xl overflow-hidden rounded-xl"
            style={{
              background: 'var(--flow-card)',
              border: '1px solid var(--flow-border)',
              boxShadow: 'var(--flow-shadow)',
            }}
          >
            {/* Search input */}
            <div
              className="flex items-center gap-2 px-4 py-3"
              style={{ borderBottom: '1px solid var(--flow-border)' }}
            >
              <Search size={15} style={{ color: 'var(--flow-text-dim)' }} />
              <input
                ref={inputRef}
                value={query}
                onChange={(e) => setQuery(e.target.value)}
                placeholder={t('flow.commandBar.searchPlaceholder')}
                className="flex-1 bg-transparent text-body"
                style={{ color: 'var(--flow-text)' }}
              />
              <span
                className="text-meta px-1.5 py-0.5 rounded-sm font-mono"
                style={{
                  background: 'var(--flow-hover)',
                  color: 'var(--flow-text-dim)',
                }}
              >
                ESC
              </span>
            </div>

            {/* Results */}
            <div className="max-h-[min(50vh,360px)] overflow-y-auto py-1 sm:max-h-[60vh]">
              {filtered.length === 0 && (
                <div
                  className="text-meta text-center py-8"
                  style={{ color: 'var(--flow-text-dim)' }}
                >
                  {t('flow.commandBar.noResults')}
                </div>
              )}
              {(['action', 'nav', 'project', 'task'] as const).map((g) => {
                const list = grouped[g];
                if (!list || list.length === 0) return null;
                return (
                  <div key={g} className="mb-1">
                    <div
                      className="px-3 py-1 text-meta font-semibold uppercase tracking-widest"
                      style={{ color: 'var(--flow-text-dim)' }}
                    >
                      {groupLabels[g]}
                    </div>
                    {list.map((item) => {
                      const idx = filtered.indexOf(item);
                      const active = idx === selected;
                      const Icon = item.icon;
                      return (
                        <button
                          type="button"
                          key={item.id}
                          onMouseEnter={() => setSelected(idx)}
                          onClick={() => run(item)}
                          className="w-full flex items-center gap-2.5 px-3 py-2 text-left text-meta"
                          style={{
                            // The accent tint, not the neutral hover grey. This
                            // row is what Enter will run, and a grey that also
                            // means "the pointer is here" does not say so.
                            // The -on-muted pair is the sanctioned one: reusing
                            // --accent-fg on this tint put 40 of 56 light-mode
                            // combinations under 4.5:1 (#970).
                            background: active ? 'var(--accent-muted)' : 'transparent',
                            color: active ? 'var(--accent-on-muted)' : 'var(--flow-text)',
                          }}
                        >
                          <Icon size={13} />
                          <span className="flex-1 truncate">{item.label}</span>
                          {item.hint && (
                            <span
                              className="text-meta font-mono px-1.5 py-0.5 rounded-sm"
                              style={{
                                background: 'var(--flow-active)',
                                color: 'var(--flow-text-dim)',
                              }}
                            >
                              {item.hint}
                            </span>
                          )}
                        </button>
                      );
                    })}
                  </div>
                );
              })}
            </div>
          </motion.div>
        </motion.div>
      )}
    </AnimatePresence>
  );
};

export default CommandBar;
