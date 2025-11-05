import React from 'react';
import Link from 'next/link';
import { HiMenu, HiX } from 'react-icons/hi';
import { Button } from '@voxire/shared';
import styles from './Header.module.scss';

interface HeaderMobileProps {
  isOpen: boolean;
  onToggle: () => void;
  onLoginClick: () => void;
  onSignupClick: () => void;
}

const HeaderMobile: React.FC<HeaderMobileProps> = ({
  isOpen,
  onToggle,
  onLoginClick,
  onSignupClick,
}) => {
  return (
    <>
      <button
        className={styles['mobile-menu-toggle']}
        onClick={onToggle}
        aria-label="Toggle mobile menu"
      >
        {isOpen ? <HiX /> : <HiMenu />}
      </button>

      {isOpen && (
        <div className={styles['mobile-menu']}>
          <Link href="/menu" style={{ textDecoration: 'none', color: 'white' }}>
            Menu
          </Link>
          <Link
            href="/deals"
            style={{ textDecoration: 'none', color: 'white' }}
          >
            Deals
          </Link>
          <Link href="/blog" style={{ textDecoration: 'none', color: 'white' }}>
            Blog
          </Link>
          <Link
            href="/hardware"
            style={{ textDecoration: 'none', color: 'white' }}
          >
            Hardware
          </Link>
          <Link
            href="/about"
            style={{ textDecoration: 'none', color: 'white' }}
          >
            About
          </Link>
          <Link
            href="/contact"
            style={{ textDecoration: 'none', color: 'white' }}
          >
            Contact Us
          </Link>
          <Button
            styles={{
              display: 'flex',
              marginBottom: '0px',
            }}
            onClick={onLoginClick}
          >
            <p>Login</p>
          </Button>

          <Button onClick={onSignupClick}>Register Now</Button>
        </div>
      )}
    </>
  );
};

export default HeaderMobile;
