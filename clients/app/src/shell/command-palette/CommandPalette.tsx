import { useEffect, useId, useMemo, useRef, useState, type KeyboardEvent } from 'react';
import { useNavigate, useRouterState } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { ArrowUpRight, CornerDownLeft, Search, X } from 'lucide-react';
import { Dialog, DialogContent, DialogDescription, DialogTitle } from '@voxire/ui';
import { FLAG_KEYS, useFeatureFlag } from '@voxire/config';
import { useCaretStore } from '@voxire/caret';
import { useCaretChatCommands } from '../../features/caret/useCaretChatCommands';
import { useFlowCommands } from '../../modules/flow/useFlowCommands';
import { useFlowStore } from '../../modules/flow/shared/stores/flowStore';
import { useCommandPalette } from '../../shared/stores/commandPaletteStore';
import { useUserStore } from '../../shared/stores/userStore';
import {
  COMMAND_GROUPS,
  getPortalCommands,
  searchPortalCommands,
  type PortalCommand,
} from './commands';
import './command-palette.css';

export default function CommandPalette() {
  const { t, i18n } = useTranslation();
  const { open, setOpen } = useCommandPalette();
  const user = useUserStore((state) => state.user);
  const navigate = useNavigate();
  const [query, setQuery] = useState('');
  const pathname = useRouterState({ select: (state) => state.location.pathname });
  const flowPreview = useFeatureFlag(FLAG_KEYS.APP_FLOW_PREVIEW_ROUTES);
  const flowSearch = useFlowCommands(user, open && pathname.startsWith('/flow/'), query);
  const onCaret = pathname === '/caret' || pathname.startsWith('/caret/');
  // Chats are searchable everywhere, but only listed up front inside Caret.
  const chatSearch = useCaretChatCommands(user, open && (onCaret || query.trim().length > 0));
  const [activeId, setActiveId] = useState<string | null>(null);
  const inputRef = useRef<HTMLInputElement>(null);
  const listRef = useRef<HTMLDivElement>(null);
  const returnFocusRef = useRef<HTMLElement | null>(null);
  const id = useId();
  const commands = useMemo(
    () => [
      ...getPortalCommands(user, t, flowPreview),
      ...flowSearch.commands,
      ...chatSearch.commands,
    ],
    [user, t, flowPreview, flowSearch.commands, chatSearch.commands]
  );
  const filtered = useMemo(
    () =>
      searchPortalCommands(commands, query, (group) => t(`common.commandPalette.groups.${group}`)),
    [commands, query, t]
  );
  const groups = useMemo(
    () =>
      onCaret
        ? ['chats' as const, ...COMMAND_GROUPS.filter((group) => group !== 'chats')]
        : COMMAND_GROUPS,
    [onCaret]
  );
  const ordered = useMemo(
    () => groups.flatMap((group) => filtered.filter((command) => command.group === group)),
    [groups, filtered]
  );
  const active = ordered.find((command) => command.id === activeId) ?? ordered[0];

  useEffect(() => {
    if (!open) {
      setQuery('');
      setActiveId(null);
    }
  }, [open]);

  useEffect(() => {
    if (open && active) {
      listRef.current
        ?.querySelector(`[id="${id}-${active.id}"]`)
        ?.scrollIntoView({ block: 'nearest' });
    }
  }, [open, active?.id, id]);

  const run = async (command: PortalCommand) => {
    setOpen(false);
    if ('path' in command.destination) navigate({ to: command.destination.path });
    else if ('href' in command.destination)
      window.open(command.destination.href, '_blank', 'noopener,noreferrer');
    else if ('newFlowTask' in command.destination) {
      useFlowStore.getState().setQuickCreateOpen(true);
    } else if ('openFlowTask' in command.destination) {
      const taskId = command.destination.openFlowTask;
      if (pathname.startsWith('/flow/'))
        navigate({
          to: '.',
          search: (previous: Record<string, unknown>) => ({ ...previous, openTask: taskId }),
        });
      else navigate({ to: '/flow/my-issues', search: { openTask: taskId } });
    } else if ('openCaretChat' in command.destination) {
      const { openCaretChat: sessionId, title, projectId } = command.destination;
      useCaretStore.getState().openChat({ sessionId, title, projectId });
      if (!onCaret) navigate({ to: '/caret' });
    }
  };

  const handleKeys = (event: KeyboardEvent<HTMLInputElement>) => {
    if (event.nativeEvent.isComposing) return;
    if (event.key === 'ArrowDown' || event.key === 'ArrowUp') {
      event.preventDefault();
      if (!ordered.length) return;
      const offset = event.key === 'ArrowDown' ? 1 : -1;
      const index = ordered.findIndex((command) => command.id === active?.id);
      setActiveId(ordered[(index + offset + ordered.length) % ordered.length].id);
    } else if (event.key === 'Enter') {
      event.preventDefault();
      if (active) run(active);
    }
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogContent
        className="portal-command-dialog"
        overlayClassName="portal-command-overlay"
        closeLabel={t('common.closeDialog')}
        dir={i18n.dir()}
        onOpenAutoFocus={(event) => {
          event.preventDefault();
          returnFocusRef.current =
            document.activeElement instanceof HTMLElement ? document.activeElement : null;
          inputRef.current?.focus();
        }}
        onCloseAutoFocus={(event) => {
          event.preventDefault();
          const target = returnFocusRef.current?.isConnected
            ? returnFocusRef.current
            : document.querySelector<HTMLButtonElement>('[data-portal-command-trigger]');
          target?.focus();
        }}
      >
        <DialogTitle className="sr-only">{t('common.commandPalette.label')}</DialogTitle>
        <DialogDescription className="sr-only">
          {t('common.commandPalette.description')}
        </DialogDescription>
        <div className="portal-command-search">
          <Search size={20} aria-hidden="true" />
          <input
            ref={inputRef}
            value={query}
            onChange={(event) => {
              setQuery(event.target.value);
              setActiveId(null);
            }}
            onKeyDown={handleKeys}
            placeholder={t('common.commandPalette.placeholder')}
            aria-label={t('common.commandPalette.placeholder')}
            role="combobox"
            aria-autocomplete="list"
            aria-expanded={open}
            aria-controls={`${id}-results`}
            aria-activedescendant={active ? `${id}-${active.id}` : undefined}
            autoComplete="off"
            spellCheck={false}
          />
          {query && (
            <button
              type="button"
              aria-label={t('common.clearSearch')}
              onClick={() => {
                setQuery('');
                setActiveId(null);
                inputRef.current?.focus();
              }}
            >
              <X size={15} />
            </button>
          )}
        </div>
        <div
          ref={listRef}
          id={`${id}-results`}
          role="listbox"
          aria-label={t('common.commandPalette.resultsLabel')}
          className="portal-command-results"
        >
          {flowSearch.loading && (
            <p role="status" className="px-5 py-2 text-sm">
              {t('common.commandPalette.loadingFlow')}
            </p>
          )}
          {flowSearch.failed && (
            <p role="alert" className="px-5 py-2 text-sm">
              {t('common.commandPalette.flowSearchFailed')}
            </p>
          )}
          {chatSearch.failed && (
            <p role="alert" className="px-5 py-2 text-sm">
              {t('common.commandPalette.chatSearchFailed')}
            </p>
          )}
          {groups.map((group) => {
            const groupCommands = filtered.filter((command) => command.group === group);
            if (!groupCommands.length) return null;
            return (
              <div
                role="group"
                aria-labelledby={`${id}-group-${group}`}
                key={group}
                className={`portal-command-group${group === 'apps' && !query.trim() ? ' is-app-grid' : ''}`}
              >
                <p id={`${id}-group-${group}`}>{t(`common.commandPalette.groups.${group}`)}</p>
                <div className="portal-command-options">
                  {groupCommands.map((command) => {
                    const Icon = command.icon;
                    const external = 'href' in command.destination;
                    return (
                      <button
                        key={command.id}
                        id={`${id}-${command.id}`}
                        type="button"
                        role="option"
                        tabIndex={-1}
                        aria-selected={active?.id === command.id}
                        aria-label={command.label}
                        aria-description={
                          external ? t('common.commandPalette.newTab') : command.description
                        }
                        className="portal-command-option"
                        onPointerMove={(event) => {
                          if (event.pointerType !== 'touch') setActiveId(command.id);
                        }}
                        onClick={() => run(command)}
                      >
                        <Icon size={20} className="portal-command-icon" />
                        <span className="portal-command-option-copy">
                          <span>{command.label}</span>
                          {command.description && <small>{command.description}</small>}
                        </span>
                        {external ? (
                          <ArrowUpRight
                            size={14}
                            className="portal-command-destination"
                            aria-label={t('common.commandPalette.newTab')}
                          />
                        ) : (
                          <CornerDownLeft
                            size={14}
                            className="portal-command-destination"
                            aria-hidden="true"
                          />
                        )}
                      </button>
                    );
                  })}
                </div>
              </div>
            );
          })}
        </div>
        {!filtered.length && (
          <div className="portal-command-empty" role="status">
            <Search size={24} aria-hidden="true" />
            <strong>{t('common.commandPalette.noResults', { query: query.trim() })}</strong>
            <span>{t('common.commandPalette.searchHint')}</span>
            <button
              type="button"
              onClick={() => {
                setQuery('');
                setActiveId(null);
                inputRef.current?.focus();
              }}
            >
              {t('common.clearSearch')}
            </button>
          </div>
        )}
        <footer className="portal-command-footer">
          <span>{t('common.commandPalette.results', { count: filtered.length })}</span>
          <div>
            <span>
              <kbd>↑ ↓</kbd>
              {t('common.commandPalette.move')}
            </span>
            <span>
              <kbd>↵</kbd>
              {t('common.commandPalette.open')}
            </span>
            <span>
              <kbd>esc</kbd>
              {t('common.close')}
            </span>
          </div>
        </footer>
      </DialogContent>
    </Dialog>
  );
}
