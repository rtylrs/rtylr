'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { useEffect, useMemo, useState } from 'react';
import type { SearchEntry } from '../lib/content';
import { navGroups } from './navigation-data';

type SiteHeaderProps = {
  searchEntries: SearchEntry[];
};

export function SiteHeader({ searchEntries }: SiteHeaderProps) {
  const pathname = usePathname();
  const [openMenu, setOpenMenu] = useState<string | null>(null);
  const [mobileOpen, setMobileOpen] = useState(false);
  const [searchOpen, setSearchOpen] = useState(false);
  const [query, setQuery] = useState('');

  useEffect(() => {
    setOpenMenu(null);
    setMobileOpen(false);
  }, [pathname]);

  useEffect(() => {
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === '/' && document.activeElement?.tagName !== 'INPUT') {
        event.preventDefault();
        setSearchOpen(true);
      }
      if (event.key === 'Escape') {
        setSearchOpen(false);
        setMobileOpen(false);
        setOpenMenu(null);
      }
    };
    window.addEventListener('keydown', onKeyDown);
    return () => window.removeEventListener('keydown', onKeyDown);
  }, []);

  const results = useMemo(() => {
    const needle = query.trim().toLowerCase().replace(/[-_]/g, ' ');
    if (!needle) return searchEntries.slice(0, 8);
    return searchEntries
      .filter((entry) => `${entry.title} ${entry.description}`.toLowerCase().replace(/[-_]/g, ' ').includes(needle))
      .slice(0, 8);
  }, [query, searchEntries]);

  const toggleMenu = (label: string) => {
    setOpenMenu((current) => (current === label ? null : label));
  };

  return (
    <>
      <header className="site-header">
        <div className="site-header__inner">
          <button
            className="mobile-menu-button"
            aria-label="Toggle navigation bar"
            aria-expanded={mobileOpen}
            onClick={() => setMobileOpen(true)}
          >
            <span aria-hidden="true">☰</span>
          </button>

          <Link href="/" className="site-logo" aria-label="rtylr Docs home">
            <img src="/logo-light.png" alt="rtylr" />
          </Link>

          <nav className="desktop-nav" aria-label="Main navigation">
            {navGroups.map((group) => (
              <div className="nav-group" key={group.label}>
                <button
                  className="nav-group__button"
                  aria-expanded={openMenu === group.label}
                  onClick={() => toggleMenu(group.label)}
                >
                  {group.label} <span aria-hidden="true">⌄</span>
                </button>
                {openMenu === group.label && (
                  <div className="nav-dropdown">
                    <div className="nav-dropdown__eyebrow">{group.label}</div>
                    <div className="nav-dropdown__items">
                      {group.items.map((item) => (
                        <Link href={item.href} key={`${group.label}-${item.label}`} onClick={() => setOpenMenu(null)}>
                          <span>{item.label}</span>
                          <small>{item.detail}</small>
                        </Link>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            ))}
          </nav>

          <div className="site-header__actions">
            <Link className="get-started-link" href="/getting-started/introduction">
              Get started <span aria-hidden="true">⌄</span>
            </Link>
            <button className="theme-button" aria-label="Toggle dark mode">
              ◐
            </button>
            <button className="search-button" aria-label="Open Search" onClick={() => setSearchOpen(true)}>
              <span aria-hidden="true">⌕</span> SEARCH <kbd>/</kbd>
            </button>
          </div>

          <button
            className="mobile-search-button"
            aria-label="Open Search"
            onClick={() => setSearchOpen(true)}
          >
            ⌕ <span>SEARCH</span> <kbd>/</kbd>
          </button>
        </div>

        {mobileOpen && (
          <div className="mobile-drawer" role="dialog" aria-label="Mobile navigation">
            <div className="mobile-drawer__top">
              <Link href="/" className="site-logo" aria-label="rtylr Docs home">
                <img src="/logo-light.png" alt="rtylr" />
              </Link>
              <div className="mobile-drawer__actions">
                <button className="theme-button" aria-label="Toggle dark mode">◔</button>
                <button className="drawer-close" aria-label="Close navigation bar" onClick={() => setMobileOpen(false)}>
                  ×
                </button>
              </div>
            </div>
            <nav aria-label="Mobile navigation">
              {navGroups.map((group) => (
                <div className="mobile-nav-group" key={group.label}>
                  <button onClick={() => toggleMenu(group.label)} aria-expanded={openMenu === group.label}>
                    {group.label} <span aria-hidden="true">⌄</span>
                  </button>
                  {openMenu === group.label && (
                    <div className="mobile-nav-submenu">
                      <button className="mobile-back" onClick={() => setOpenMenu(null)}>← Back to main menu</button>
                      {group.items.map((item) => (
                        <Link href={item.href} key={`${group.label}-${item.label}`}>
                          {item.label}
                        </Link>
                      ))}
                    </div>
                  )}
                </div>
              ))}
            </nav>
          </div>
        )}
      </header>

      {searchOpen && (
        <div className="search-overlay" role="dialog" aria-modal="true" aria-label="Search rtylr docs">
          <button className="search-overlay__backdrop" aria-label="Close search" onClick={() => setSearchOpen(false)} />
          <div className="search-panel">
            <div className="search-panel__top">
              <span>Search rtylr docs</span>
              <button onClick={() => setSearchOpen(false)} aria-label="Close search">×</button>
            </div>
            <input
              autoFocus
              value={query}
              onChange={(event) => setQuery(event.target.value)}
              placeholder="Search documentation..."
              aria-label="Search documentation"
            />
            <div className="search-results">
              {results.map((entry) => (
                <Link href={entry.route} key={entry.route} onClick={() => setSearchOpen(false)}>
                  <strong>{entry.title}</strong>
                  <span>{entry.description}</span>
                </Link>
              ))}
              {!results.length && <p className="search-empty">No documentation matches that search.</p>}
            </div>
          </div>
        </div>
      )}
    </>
  );
}
