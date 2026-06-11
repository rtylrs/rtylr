import React from 'react';
import { cn } from '../../lib/cn';
import { motion } from 'motion/react';
import Link from 'next/link';
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
import styles from './trust.module.scss';

const PILLARS = [
  {
    icon: ShieldCheck,
    title: 'Security program',
    copy: 'Formal information-security program aligned with ISO 27001 and SOC 2. Reviewed quarterly.',
    link: '/legal/security#program',
  },
  {
    icon: Lock,
    title: 'Encryption',
    copy: 'TLS 1.2+ in transit. AES-256 at rest. Separate backup regions. Secrets managed in a vault.',
    link: '/legal/security#data-protection',
  },
  {
    icon: KeyRound,
    title: 'Access control',
    copy: 'SSO + mandatory MFA for production. Least-privilege access, quarterly reviews.',
    link: '/legal/security#access-control',
  },
  {
    icon: Server,
    title: 'Infrastructure',
    copy: 'Cloudflare edge, Firebase / GCP managed services, multi-region backups with PITR.',
    link: '/legal/security#infrastructure',
  },
  {
    icon: Activity,
    title: '24/7 monitoring',
    copy: 'Continuous monitoring and alerting. Documented incident-response runbook. 72-hour breach notifications.',
    link: '/legal/security#monitoring',
  },
  {
    icon: FileCheck,
    title: 'Vulnerability management',
    copy: 'Dependencies scanned on every build. Annual third-party pen test. Critical CVEs patched within 7 days.',
    link: '/legal/security#vulnerability',
  },
];

const COMPLIANCE = [
  { label: 'GDPR / UK GDPR', status: 'Compliant', ready: true },
  { label: 'CCPA / CPRA', status: 'Compliant', ready: true },
  { label: 'PCI DSS', status: 'Via Paddle (Level 1)', ready: true },
  { label: 'SOC 2 Type II', status: 'In progress · Q4 2026', ready: false },
  { label: 'ISO 27001', status: 'Aligned, not yet certified', ready: false },
  { label: 'HIPAA', status: 'Not applicable', ready: false },
];

const TrustPage = (): JSX.Element => {
  return (
    <AppLayout>
      <MainLayout>
        <HelmetContainer
          title="Trust Center | rtylr"
          description="Security, privacy, compliance, uptime, and vendor transparency, everything customers need before they commit."
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
                TRUST CENTER
              </motion.span>
              <motion.h1 className={styles.heroTitle} variants={blurFadeIn}>
                Run your business
                <br />
                on something trustworthy.
              </motion.h1>
              <motion.p className={styles.heroSubtitle} variants={blurFadeIn}>
                Security, privacy, uptime, and compliance &mdash; the stuff that should be
                straightforward, made transparent. Everything you need to evaluate rtylr before you
                commit.
              </motion.p>
              <motion.div className={styles.heroLinks} variants={blurFadeIn}>
                <Link href="/legal/security" className={styles.primaryBtn}>
                  Security documentation
                </Link>
                <Link href="/legal/privacy-policy" className={styles.outlineBtn}>
                  Privacy policy
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
                  <div className={styles.statusLabel}>All systems operational</div>
                  <div className={styles.statusSub}>
                    View live status at{' '}
                    <a href="https://status.rtylr.com" target="_blank" rel="noopener noreferrer">
                      status.rtylr.com
                    </a>
                  </div>
                </div>
              </div>
              <div className={styles.statusCell}>
                <div className={styles.statusBig}>99.99%</div>
                <div className={styles.statusSub}>Rolling 90-day uptime</div>
              </div>
              <div className={styles.statusCell}>
                <div className={styles.statusBig}>&lt; 72h</div>
                <div className={styles.statusSub}>Breach notification commitment</div>
              </div>
              <div className={styles.statusCell}>
                <div className={styles.statusBig}>1</div>
                <div className={styles.statusSub}>Annual third-party pen test</div>
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
                HOW WE PROTECT YOUR DATA
              </motion.span>
              <motion.h2 className={styles.sectionTitle} variants={blurFadeIn}>
                Six things we take
                <br />
                seriously every day.
              </motion.h2>
            </motion.div>
            <motion.div
              className={styles.pillarsGrid}
              initial="hidden"
              whileInView="visible"
              viewport={viewportOnce}
              variants={staggerBlur}
            >
              {PILLARS.map(({ icon: Icon, title, copy, link }) => (
                <motion.div key={title} className={styles.pillarCard} variants={blurFadeIn}>
                  <div className={styles.pillarIcon}>
                    <Icon size={18} strokeWidth={1.8} />
                  </div>
                  <h3 className={styles.pillarTitle}>{title}</h3>
                  <p className={styles.pillarCopy}>{copy}</p>
                  <Link href={link} className={styles.pillarLink}>
                    Learn more →
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
              <span className={styles.sectionEyebrow}>COMPLIANCE POSTURE</span>
              <h2 className={styles.sectionTitle}>Where we are.</h2>
              <p className={styles.complianceCopy}>
                We publish our compliance status transparently, certified, aligned, or in progress.
              </p>
              <div className={styles.contactStrip}>
                <h3>Need our security packet, SIG, or DPA?</h3>
                <p>
                  Email <a href="mailto:security@voxire.com">security@voxire.com</a> and we&apos;ll
                  send the current packet under NDA within 7 business days, including the latest
                  status of our SOC 2 Type II audit.
                </p>
              </div>
            </motion.div>
            <motion.ul className={styles.complianceList} variants={blurFadeIn}>
              {COMPLIANCE.map((item) => (
                <li key={item.label} className={styles.complianceItem}>
                  <span className={styles.complianceLabel}>{item.label}</span>
                  <span
                    className={cn(
                      styles.complianceStatus,
                      item.ready && styles.complianceStatusReady
                    )}
                  >
                    {item.ready && <CheckCircle2 size={14} strokeWidth={2.2} />}
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
                Where your data lives, and how you can take it with you.
              </h2>
              <p className={styles.dataCopy}>
                Customer Data is hosted in the cloud region that best matches your workspace
                settings. EU-region hosting is available on request for customers with
                data-residency requirements. You can export your full workspace (CSV, XLSX, or API
                dump) at any time from settings. On account termination, your data is retained for
                30 days for recovery, then deleted within 60 days from live systems and within 90
                days from backups.
              </p>
            </motion.div>
          </motion.section>
        </div>
      </MainLayout>
    </AppLayout>
  );
};

export default TrustPage;
