import React from 'react';
import { Trans, useTranslation } from 'react-i18next';
import { useRouter } from 'next/router';

import AppLayout from '../Layout/AppLayout';
import MainLayout from '../../layouts/Main/MainLayout';
import HelmetContainer from '../HelmetContainer';
import { webPageSchema } from '../../lib/structured-data';
import { localeFromPath } from '../../lib/locales';
import styles from '../../pages/legal/legal.module.scss';
import { LegalArticle, legalTransComponents } from './LegalArticle';

interface LegalPageLayoutProps {
  /** i18n namespace prefix for one legal document, e.g. 'legal.terms'. */
  i18nKey: string;
  /** Bare, unlocalised path, e.g. '/legal/terms/'. */
  canonicalPath: string;
}

/**
 * The shell every legal page shares: Helmet, header, table of contents,
 * the document body (LegalArticle), and the closing contact strip. Header
 * and ToC are read from the same `${i18nKey}.sections` object the body
 * uses, so the ToC can never list a section the body doesn't have or vice
 * versa - one list, not two kept in step by hand.
 */
export function LegalPageLayout({ i18nKey, canonicalPath }: LegalPageLayoutProps): React.JSX.Element {
  const { t } = useTranslation();
  const router = useRouter();
  const locale = localeFromPath(router.asPath).code;
  const sections = t(`${i18nKey}.sections`, { returnObjects: true }) as Record<string, string>;

  return (
    <AppLayout>
      <MainLayout>
        <HelmetContainer
          title={t(`${i18nKey}.meta.title`)}
          description={t(`${i18nKey}.meta.description`)}
          canonical={`https://rtylr.com${canonicalPath}`}
          structuredData={webPageSchema({
            name: t(`${i18nKey}.title`),
            path: canonicalPath,
            description: t(`${i18nKey}.meta.schemaDescription`),
          })}
        />

        <main className={styles.page}>
          <div className={styles.container}>
            <header className={styles.header}>
              <span className={styles.eyebrow}>{t(`${i18nKey}.eyebrow`)}</span>
              <h1 className={styles.title}>{t(`${i18nKey}.title`)}</h1>
              <p className={styles.meta}>{t(`${i18nKey}.lastUpdated`)}</p>
            </header>

            <div className={styles.layout}>
              <nav className={styles.toc} aria-label={t(`${i18nKey}.tocAriaLabel`)}>
                {/* Label for the nav, not a section of the document: as an
                    h4 it jumped straight from the h1. */}
                <p className={styles.tocTitle}>{t(`${i18nKey}.tocTitle`)}</p>
                <ol>
                  {Object.entries(sections).map(([id, label]) => (
                    <li key={id}>
                      <a href={`#${id}`}>{label}</a>
                    </li>
                  ))}
                </ol>
              </nav>

              <article className={styles.body}>
                <LegalArticle i18nKey={i18nKey} />

                <div className={styles.contactStrip}>
                  <h3>{t(`${i18nKey}.contactStrip.title`)}</h3>
                  <p>
                    <Trans
                      i18nKey={`${i18nKey}.contactStrip.body`}
                      components={legalTransComponents(locale)}
                    />
                  </p>
                </div>
              </article>
            </div>
          </div>
        </main>
      </MainLayout>
    </AppLayout>
  );
}
