import React from 'react';
import { cn } from '../../lib/cn';
import { motion } from 'motion/react';
import { Trans, useTranslation } from 'react-i18next';
import Link from 'next/link';
import { useRouter } from 'next/router';
import {
  ShieldCheck,
  Lock,
  Globe,
  FileCheck,
  Activity,
  Server,
  KeyRound,
  CheckCircle2,
} from 'lucide-react';
import AppLayout from '../../components/Layout/AppLayout';
import MainLayout from '../../layouts/Main/MainLayout';
import HelmetContainer from '../../components/HelmetContainer';
import { blurFadeIn, staggerBlur, viewportOnce } from '../../lib/animations';
import { webPageSchema } from '../../lib/structured-data';
import Lines from '../../components/text/Lines';
import { localeFromPath, navPath, navPathWithHash } from '../../lib/locales';
import styles from './trust.module.scss';

// Icon and anchor only. Title and copy resolve from trust.pillars,
// matched by position. An anchor is a URL, not copy.
// Which compliance rows are achieved. A boolean drives a tick and a style,
// so it is state rather than copy and stays out of the locale files.
const COMPLIANCE_READY: boolean[] = [true, true, true, false, false, false];

const PILLAR_META: { icon: typeof ShieldCheck; link: string }[] = [
  { icon: ShieldCheck, link: '/legal/security#program' },
  { icon: Lock, link: '/legal/security#data-protection' },
  { icon: KeyRound, link: '/legal/security#access-control' },
  { icon: Server, link: '/legal/security#infrastructure' },
  { icon: Activity, link: '/legal/security#monitoring' },
  { icon: FileCheck, link: '/legal/security#vulnerability' },
];

const statusLinkComponent = (
  <a href="https://status.rtylr.com" target="_blank" rel="noopener noreferrer" />
);
const secMailComponent = <a href="mailto:security@voxire.com" />;



const TrustPage = (): React.JSX.Element => {
  const { t } = useTranslation();
  const router = useRouter();
  const locale = localeFromPath(router.asPath).code;
  const pillars = t('trust.pillars', { returnObjects: true }) as unknown as {
    title: string;
    copy: string;
  }[];
  const compliance = t('trust.compliance', { returnObjects: true }) as unknown as {
    label: string;
    status: string;
  }[];
  return (
    <AppLayout>
      <MainLayout>
        <HelmetContainer
          title={t('trust.meta.title')}
          description={t('trust.meta.description')}
          canonical="https://rtylr.com/trust/"
          structuredData={webPageSchema({
            name: 'Trust Center',
            path: '/trust/',
            description: 'Security, privacy, compliance, uptime, and vendor transparency.',
          })}
        />

        <div className={styles.page}>
          {/* HERO */}
          <section className={styles.hero}>
            <motion.div
              className={styles.heroInner}
              initial="hidden"
              animate="visible"
              variants={staggerBlur}
            >
              <motion.span className={styles.heroEyebrow} variants={blurFadeIn}>
                <ShieldCheck size={13} strokeWidth={2} />
                {t('trust.hero.eyebrow')}
              </motion.span>
              <motion.h1 className={styles.heroTitle} variants={blurFadeIn}>
                <Lines text={t('trust.hero.title')} />
              </motion.h1>
              <motion.p className={styles.heroSubtitle} variants={blurFadeIn}>
                {t('trust.hero.subtitle')}
              </motion.p>
              <motion.div className={styles.heroLinks} variants={blurFadeIn}>
                <Link href={navPath('/legal/security', locale)} className={styles.primaryBtn}>
                  {t('trust.hero.securityDocs')}
                </Link>
                <Link href={navPath('/legal/privacy-policy', locale)} className={styles.outlineBtn}>
                  {t('trust.hero.privacyPolicy')}
                </Link>
              </motion.div>
            </motion.div>
          </section>

          {/* STATUS STRIP */}
          <motion.section
            className={styles.statusSection}
            initial="hidden"
            whileInView="visible"
            viewport={viewportOnce}
            variants={staggerBlur}
          >
            <motion.div className={styles.statusGrid} variants={blurFadeIn}>
              <div className={styles.statusCell}>
                <span className={styles.statusDot} />
                <div>
                  <div className={styles.statusLabel}>{t('trust.status.operational')}</div>
                  <div className={styles.statusSub}>
                    <Trans i18nKey="trust.status.viewLive" components={{ statusLink: statusLinkComponent }} />
                  </div>
                </div>
              </div>
              <div className={styles.statusCell}>
                <div className={styles.statusBig} dir="ltr">{t('trust.status.uptimeValue')}</div>
                <div className={styles.statusSub}>{t('trust.status.uptimeLabel')}</div>
              </div>
              <div className={styles.statusCell}>
                <div className={styles.statusBig}>{t('trust.status.breachValue')}</div>
                <div className={styles.statusSub}>{t('trust.status.breachLabel')}</div>
              </div>
              <div className={styles.statusCell}>
                <div className={styles.statusBig} dir="ltr">1</div>
                <div className={styles.statusSub}>{t('trust.status.penTestLabel')}</div>
              </div>
            </motion.div>
          </motion.section>

          {/* PILLARS */}
          <section className={styles.pillarsSection}>
            <motion.div
              className={styles.pillarsHeader}
              initial="hidden"
              whileInView="visible"
              viewport={viewportOnce}
              variants={staggerBlur}
            >
              <motion.span className={styles.sectionEyebrow} variants={blurFadeIn}>
                {t('trust.pillarsSection.eyebrow')}
              </motion.span>
              <motion.h2 className={styles.sectionTitle} variants={blurFadeIn}>
                <Lines text={t('trust.pillarsSection.title')} />
              </motion.h2>
            </motion.div>
            <motion.div
              className={styles.pillarsGrid}
              initial="hidden"
              whileInView="visible"
              viewport={viewportOnce}
              variants={staggerBlur}
            >
              {PILLAR_META.map(({ icon: Icon, link }, pi) => (
                <motion.div key={pillars[pi].title} className={styles.pillarCard} variants={blurFadeIn}>
                  <div className={styles.pillarIcon}>
                    <Icon size={18} strokeWidth={1.8} />
                  </div>
                  <h3 className={styles.pillarTitle}>{pillars[pi].title}</h3>
                  <p className={styles.pillarCopy}>{pillars[pi].copy}</p>
                  <Link href={navPathWithHash(link, locale)} className={styles.pillarLink}>
                    {t('trust.pillarsSection.learnMore')}
                  </Link>
                </motion.div>
              ))}
            </motion.div>
          </section>

          {/* COMPLIANCE */}
          <motion.section
            className={styles.complianceSection}
            initial="hidden"
            whileInView="visible"
            viewport={viewportOnce}
            variants={staggerBlur}
          >
            <motion.div className={styles.complianceHeader} variants={blurFadeIn}>
              <span className={styles.sectionEyebrow}>{t('trust.complianceSection.eyebrow')}</span>
              <h2 className={styles.sectionTitle}>{t('trust.complianceSection.title')}</h2>
              <p className={styles.complianceCopy}>
                {t('trust.complianceSection.copy')}
              </p>
              <div className={styles.contactStrip}>
                <h3>{t('trust.packet.title')}</h3>
                <p>
                  <Trans i18nKey="trust.packet.body" components={{ secMail: secMailComponent }} />
                </p>
              </div>
            </motion.div>
            <motion.ul className={styles.complianceList} variants={blurFadeIn}>
              {compliance.map((item, i) => (
                <li key={item.label} className={styles.complianceItem}>
                  <span className={styles.complianceLabel}>{item.label}</span>
                  <span
                    className={cn(
                      styles.complianceStatus,
                      COMPLIANCE_READY[i] && styles.complianceStatusReady
                    )}
                  >
                    {COMPLIANCE_READY[i] && <CheckCircle2 size={14} strokeWidth={2.2} />}
                    {item.status}
                  </span>
                </li>
              ))}
            </motion.ul>
          </motion.section>

          {/* DATA LOCATION */}
          <motion.section
            className={styles.dataSection}
            initial="hidden"
            whileInView="visible"
            viewport={viewportOnce}
            variants={staggerBlur}
          >
            <motion.div className={styles.dataInner} variants={blurFadeIn}>
              <Globe size={22} strokeWidth={1.8} className={styles.dataIcon} />
              <h2 className={styles.dataTitle}>
                {t('trust.residency.title')}
              </h2>
              <p className={styles.dataCopy}>{t('trust.residency.body')}</p>
            </motion.div>
          </motion.section>
        </div>
      </MainLayout>
    </AppLayout>
  );
};

export default TrustPage;
