import { useEffect, useMemo, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useQuery } from '@tanstack/react-query';
import { CalendarClock, Download, PanelLeft, Settings } from 'lucide-react';
import { resolveApiBase } from '@voxire/config';
import { HelmetContainer } from '@voxire/components';
import { CaretIcon, CaretPanel, useCaretStore, useCaretUIStore } from '@voxire/caret';
import { useUserStore } from '../../shared/stores/userStore';
import {
  buildPortalPageContext,
  buildPortalSystemContext,
} from '../../shared/lib/portalCaretContext';
import { usePortalHeaderActions } from '../../shell/PortalHeaderActions';
import { useCaretInstall } from './useCaretInstall';
import BuddyRoutines, { buddyClient, buddySystemContext } from './Buddy';
import BuddyPerch from './BuddyPerch';
import BuddySprite from './BuddySprite';
import { useBuddyMood } from './useBuddyMood';
import { BUDDY_COLORS, useBuddyColor } from './useBuddyColor';
import './caret-workspace.css';

export default function CaretWorkspace() {
  const { t } = useTranslation();
  const user = useUserStore((state) => state.user);
  const actions = usePortalHeaderActions();
  const title = useCaretStore(
    (state) => state.tabs.find((tab) => tab.id === state.activeTabId)?.title
  );
  const activeSessionId = useCaretStore(
    (state) => state.tabs.find((tab) => tab.id === state.activeTabId)?.sessionId
  );
  const buddyActive = Boolean(activeSessionId?.startsWith('buddy-'));
  const buddyMood = useBuddyMood(buddyActive);
  const buddyScope = `${user?.businessId}:${user?.id}`;
  const seenKey = `caret-buddy-seen:${buddyScope}`;
  const [buddyColorName] = useBuddyColor(buddyScope);
  const buddyColor = BUDDY_COLORS[buddyColorName];
  const thread = useQuery({
    queryKey: ['caret-buddy', buddyScope, 'thread'],
    queryFn: () => buddyClient.getBuddyThread({}),
    refetchInterval: 60_000,
    enabled: Boolean(user?.id),
  });
  const lastPostAt = thread.data?.lastPostAt ?? '';
  const [seenAt, setSeenAt] = useState(() => localStorage.getItem(seenKey) ?? '');
  useEffect(() => {
    if (!buddyActive) return;
    const now = new Date().toISOString();
    localStorage.setItem(seenKey, now);
    setSeenAt(now);
  }, [buddyActive, lastPostAt, seenKey]);
  const buddyUnread =
    !buddyActive &&
    lastPostAt !== '' &&
    (seenAt === '' || Date.parse(lastPostAt) > Date.parse(seenAt));
  const setView = useCaretUIStore((state) => state.setView);
  const { canInstall, install } = useCaretInstall();
  const view = useCaretUIStore((state) => state.view);
  const [routinesOpen, setRoutinesOpen] = useState(false);
  const [buddyError, setBuddyError] = useState(false);
  const isAdmin = Boolean(user?.effectiveAccess?.isOwner || user?.effectiveAccess?.isAdmin);
  const openBuddy = async () => {
    setBuddyError(false);
    try {
      const { sessionId } = await buddyClient.getBuddyThread({});
      if (!sessionId) throw new Error('missing buddy chat');
      useCaretStore.getState().openChat({ sessionId, title: t('common.caret.buddy.title') });
      setView('chat');
      if (window.matchMedia('(max-width: 700px)').matches) setSidebarOpen(false);
    } catch {
      setBuddyError(true);
    }
  };
  const [sidebarOpen, setSidebarOpen] = useState(
    () => !window.matchMedia('(max-width: 700px)').matches
  );
  const sidebarToggleRef = useRef<HTMLButtonElement>(null);
  const headerToggleRef = useRef<HTMLButtonElement>(null);
  const restoreToggleFocus = useRef(false);
  const context = useMemo(
    () => ({
      pathname: '/caret',
      businessName: user?.businessName,
      apps: user?.effectiveAccess?.apps
        .filter((app) => app.licensed && app.hasAccess)
        .map((app) => app.appId),
    }),
    [user?.businessName, user?.effectiveAccess]
  );

  useEffect(() => {
    setView('chat');
  }, [setView]);
  useEffect(() => {
    // The installed app's "New chat" shortcut opens /caret?new=1.
    const params = new URLSearchParams(window.location.search);
    if (params.get('new') !== '1') return;
    useCaretStore.getState().newTab();
    params.delete('new');
    const query = params.toString();
    window.history.replaceState(null, '', `/caret${query ? `?${query}` : ''}`);
  }, []);
  useEffect(() => {
    if (restoreToggleFocus.current) {
      (sidebarOpen ? sidebarToggleRef : headerToggleRef).current?.focus();
      restoreToggleFocus.current = false;
    }
  }, [sidebarOpen]);
  useEffect(() => {
    const close = (event: KeyboardEvent) => {
      if (
        !event.defaultPrevented &&
        event.key === 'Escape' &&
        window.matchMedia('(max-width: 700px)').matches
      ) {
        restoreToggleFocus.current = true;
        setSidebarOpen(false);
      }
    };
    window.addEventListener('keydown', close);
    return () => window.removeEventListener('keydown', close);
  }, []);

  const toggle = (location: 'sidebar' | 'header') => (
    <button
      ref={location === 'sidebar' ? sidebarToggleRef : headerToggleRef}
      type="button"
      className="portal-icon-button"
      onClick={() => {
        restoreToggleFocus.current = true;
        setSidebarOpen((open) => !open);
      }}
      aria-label={t('common.caret.toggleHistory')}
      aria-expanded={sidebarOpen}
      aria-controls="caret-chat-history"
    >
      <PanelLeft size={18} />
    </button>
  );

  return (
    <div className="portal-caret-workspace portal-caret">
      <HelmetContainer title="Caret · rtylr" />
      <CaretPanel
        variant="fullscreen"
        shortcutsEnabled={false}
        appId="dash"
        userId={Number(user?.id) || 0}
        apiBaseUrl={resolveApiBase()}
        systemContext={
          buddyActive
            ? `${buildPortalSystemContext(context)}\n\n${buddySystemContext()}`
            : buildPortalSystemContext(context)
        }
        pageContext={buildPortalPageContext(context)}
        workspace={{
          sidebarOpen,
          pollSessionId: buddyActive ? (activeSessionId ?? undefined) : undefined,
          composerAccessory: buddyActive ? (
            <BuddyPerch mood={buddyMood} label={t('common.caret.buddy.title')} color={buddyColor} />
          ) : undefined,
          onSelectChat: () => {
            setRoutinesOpen(false);
            if (window.matchMedia('(max-width: 700px)').matches) setSidebarOpen(false);
          },
          sidebarHeader: (
            <div className="caret-workspace-brand">
              <span>
                <CaretIcon size={24} />
                Caret
              </span>
              {toggle('sidebar')}
            </div>
          ),
          sidebarContent: (
            <section className="caret-bots" aria-label={t('common.caret.buddy.bots')}>
              <h2 className="caret-bots-heading">{t('common.caret.buddy.bots')}</h2>
              <button
                type="button"
                className="caret-bot-row"
                aria-current={buddyActive ? 'page' : undefined}
                onClick={() => void openBuddy()}
              >
                <BuddySprite
                  size={22}
                  color={buddyColor}
                  mood={buddyActive ? buddyMood : buddyUnread ? 'alert' : 'idle'}
                  track={false}
                />
                {t('common.caret.buddy.title')}
                {buddyUnread && (
                  <span
                    className="caret-bot-unread"
                    role="status"
                    aria-label={t('common.caret.buddy.unread')}
                  />
                )}
              </button>
              {buddyError && (
                <p role="alert" className="caret-bots-error">
                  {t('common.caret.buddy.openFailed')}
                </p>
              )}
            </section>
          ),
          header: (
            <header className="caret-workspace-header">
              <div>
                {!sidebarOpen && toggle('header')}
                {buddyActive && (
                  <BuddySprite size={24} mood={buddyMood} track={false} color={buddyColor} />
                )}
                <span>{buddyActive ? t('common.caret.buddy.title') : title || 'Caret'}</span>
              </div>
              <div className="caret-workspace-header-actions">
                {buddyActive && (
                  <button
                    type="button"
                    className="portal-icon-button"
                    onClick={() => setRoutinesOpen((open) => !open)}
                    aria-label={t('common.caret.buddy.routines')}
                    aria-pressed={routinesOpen}
                    title={t('common.caret.buddy.routines')}
                  >
                    <CalendarClock size={18} />
                  </button>
                )}
                {canInstall && (
                  <button
                    type="button"
                    className="portal-icon-button"
                    onClick={() => void install()}
                    aria-label={t('common.caret.installApp')}
                    title={t('common.caret.installApp')}
                  >
                    <Download size={18} />
                  </button>
                )}
                <button
                  type="button"
                  className="portal-icon-button"
                  onClick={() => setView(view === 'settings' ? 'chat' : 'settings')}
                  aria-label={t('common.caret.settings')}
                  aria-pressed={view === 'settings'}
                  title={t('common.caret.settings')}
                >
                  <Settings size={18} />
                </button>
                {actions}
              </div>
            </header>
          ),
          welcome: buddyActive ? (
            (send: (prompt: string) => void) => (
              <div className="caret-workspace-welcome">
                <BuddySprite
                  size={96}
                  color={buddyColor}
                  mood={buddyMood}
                  pokeable
                  title={t('common.caret.buddy.title')}
                  className="caret-buddy-hero"
                />
                <h1>{t('common.caret.buddy.welcome')}</h1>
                <p>{t('common.caret.buddy.welcomeHint')}</p>
                <div className="caret-buddy-starters">
                  {(
                    ['starterCheckIn', 'starterSupport', 'starterInvoices', 'starterWhat'] as const
                  ).map((key) => (
                    <button
                      key={key}
                      type="button"
                      onClick={() => send(t(`common.caret.buddy.${key}`))}
                    >
                      {t(`common.caret.buddy.${key}`)}
                    </button>
                  ))}
                </div>
              </div>
            )
          ) : (
            <div className="caret-workspace-welcome">
              <CaretIcon size={38} />
              <h1>{t('common.caret.welcome')}</h1>
              <p>{t('common.caret.welcomeHint')}</p>
            </div>
          ),
          composerPlaceholder: buddyActive
            ? t('common.caret.buddy.composer')
            : t('common.caret.composer'),
          historyLabels: {
            heading: t('common.caret.recents'),
            description: '',
            newChat: t('common.caret.newChat'),
            loading: t('common.caret.loading'),
            empty: t('common.caret.empty'),
            loadFailed: t('common.caret.loadFailed'),
            retry: t('common.refresh'),
            deleteChat: t('common.caret.deleteChat'),
            deleteTitle: t('common.caret.deleteChatTitle'),
            deleteMessage: t('common.caret.deleteChatMessage'),
            confirmDelete: t('common.caret.confirmDeleteChat'),
            cancel: t('common.cancel'),
            chatActions: t('common.caret.chatActions'),
            renameChat: t('common.caret.renameChat'),
            renameTitle: t('common.caret.renameChatTitle'),
            renameHint: t('common.caret.renameChatHint'),
            renameFailed: t('common.caret.renameChatFailed'),
            save: t('common.save'),
            close: t('common.close'),
            pinned: t('common.caret.pinnedChat'),
            pinChat: t('common.caret.pinChat'),
            unpinChat: t('common.caret.unpinChat'),
          },
          projectLabels: {
            heading: t('common.caret.projects'),
            allChats: t('common.caret.allChats'),
            create: t('common.caret.createProject'),
            createTitle: t('common.caret.createProject'),
            editTitle: t('common.caret.editProject'),
            name: t('common.caret.projectName'),
            context: t('common.caret.projectContext'),
            contextHint: t('common.caret.projectContextHint'),
            icon: t('common.caret.projectIcon'),
            color: t('common.caret.projectColor'),
            save: t('common.caret.saveProject'),
            cancel: t('common.cancel'),
            archive: t('common.caret.archiveProject'),
            archiveTitle: t('common.caret.archiveProjectTitle'),
            archiveMessage: t('common.caret.archiveProjectMessage'),
            confirmArchive: t('common.caret.confirmArchiveProject'),
            empty: t('common.caret.noProjects'),
            loadFailed: t('common.caret.projectsLoadFailed'),
            required: t('common.caret.projectNameRequired'),
            loading: t('common.caret.projectsLoading'),
            saving: t('common.caret.projectSaving'),
            saveFailed: t('common.caret.projectSaveFailed'),
            archiveFailed: t('common.caret.projectArchiveFailed'),
          },
        }}
      />
      {buddyActive && routinesOpen && (
        <BuddyRoutines
          key={buddyScope}
          scope={buddyScope}
          isAdmin={isAdmin}
          onClose={() => setRoutinesOpen(false)}
        />
      )}
    </div>
  );
}
