import React from 'react';
import styles from './POSLayout.module.scss';
import { useUserStore } from '../../stores/userStore';
import { useThemeStore } from '../../stores/themeStore';

interface POSLayoutProps {
  children: React.ReactNode;
}

const POSLayout: React.FC<POSLayoutProps> = ({ children }) => {
  const logout = useUserStore((s) => s.logout);
  const user = useUserStore((s) => s.user);
  const { theme, toggleTheme } = useThemeStore();

  return (
    <div className={styles.layout} data-theme={theme}>
      <header className={styles.header}>
        <div className={styles.header__brand}>
          <span className={styles.header__logo}>⬡</span>
          <span className={styles.header__name}>
            {user?.name ?? 'POS Terminal'}
          </span>
        </div>
        <div className={styles.header__actions}>
          <button
            className={styles.header__btn}
            onClick={toggleTheme}
            aria-label="Toggle theme"
            title="Toggle theme"
          >
            {theme === 'dark' ? '☀︎' : '◑'}
          </button>
          <button
            className={styles.header__btn}
            onClick={logout}
            aria-label="Log out"
            title="Log out"
          >
            ⏻
          </button>
        </div>
      </header>
      <main className={styles.main}>{children}</main>
    </div>
  );
};

export default POSLayout;
