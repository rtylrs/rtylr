import { useEffect, useMemo, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Bot, Download, PanelLeft, Settings } from 'lucide-react';
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
import Buddy from './Buddy';
import './caret-workspace.css';

export default function CaretWorkspace() {
  const { t } = useTranslation();
  const user = useUserStore((state) => state.user);
  const actions = usePortalHeaderActions();
  const title = useCaretStore(
    (state) => state.tabs.find((tab) => tab.id === state.activeTabId)?.title
  );
  const setView = useCaretUIStore((state) => state.setView);
  const { canInstall, install } = useCaretInstall();
  const view = useCaretUIStore((state) => state.view);
  const [buddyOpen, setBuddyOpen] = useState(false);
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

  if (buddyOpen)
    return (
      <div className="portal-caret-workspace">
        <Buddy
          key={`${user?.businessId}:${user?.id}`}
          scope={`${user?.businessId}:${user?.id}`}
          onBack={() => setBuddyOpen(false)}
        />
      </div>
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
        systemContext={buildPortalSystemContext(context)}
        pageContext={buildPortalPageContext(context)}
        workspace={{
          sidebarOpen,
          onSelectChat: () => {
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
          header: (
            <header className="caret-workspace-header">
              <div>
                {!sidebarOpen && toggle('header')}
                <span>{title || 'Caret'}</span>
              </div>
              <div className="caret-workspace-header-actions">
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
                {(user?.effectiveAccess?.isOwner || user?.effectiveAccess?.isAdmin) && (
                  <button
                    type="button"
                    className="portal-icon-button"
                    onClick={() => setBuddyOpen(true)}
                    aria-label={t('common.caret.buddy.title')}
                    title={t('common.caret.buddy.title')}
                  >
                    <Bot size={18} />
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
          welcome: (
            <div className="caret-workspace-welcome">
              <CaretIcon size={38} />
              <h1>{t('common.caret.welcome')}</h1>
              <p>{t('common.caret.welcomeHint')}</p>
            </div>
          ),
          composerPlaceholder: t('common.caret.composer'),
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
    </div>
  );
}
