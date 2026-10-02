import { useEffect, useMemo, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { PanelLeft } from 'lucide-react';
import { resolveApiBase } from '@voxire/config';
import { HelmetContainer } from '@voxire/components';
import { CaretIcon, CaretPanel, useCaretStore, useCaretUIStore } from '@voxire/caret';
import { useUserStore } from '../../shared/stores/userStore';
import {
  buildPortalPageContext,
  buildPortalSystemContext,
} from '../../shared/lib/portalCaretContext';
import { usePortalHeaderActions } from '../../shell/PortalHeaderActions';
import './caret-workspace.css';

export default function CaretWorkspace() {
  const { t } = useTranslation();
  const user = useUserStore((state) => state.user);
  const actions = usePortalHeaderActions();
  const title = useCaretStore(
    (state) => state.tabs.find((tab) => tab.id === state.activeTabId)?.title
  );
  const setView = useCaretUIStore((state) => state.setView);
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
              {actions}
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
            search: t('common.caret.search'),
            clearSearch: t('common.clearSearch'),
            loading: t('common.caret.loading'),
            empty: t('common.caret.empty'),
            noResults: t('common.caret.noResults'),
            loadFailed: t('common.caret.loadFailed'),
            retry: t('common.refresh'),
            deleteChat: t('common.caret.deleteChat'),
            pinChat: t('common.caret.pinChat'),
            unpinChat: t('common.caret.unpinChat'),
          },
        }}
      />
    </div>
  );
}
