import React from 'react';
import { motion } from 'motion/react';
import Link from 'next/link';
import { Newspaper, Download, Mail, ArrowUpRight } from 'lucide-react';
import AppLayout from '../../components/Layout/AppLayout';
import MainLayout from '../../layouts/Main/MainLayout';
import HelmetContainer from '../../components/HelmetContainer';
import { blurFadeIn, staggerBlur, viewportOnce } from '../../lib/animations';
import styles from './newsroom.module.scss';

interface NewsItem {
  date: string;
  kind: 'Press release' | 'Coverage' | 'Milestone' | 'Award';
  headline: string;
  excerpt: string;
  source?: { label: string; href: string };
}

const NEWS: NewsItem[] = [
  {
    date: 'Apr 18, 2026',
    kind: 'Press release',
    headline:
      'rtylr launches Scale plan with tiered per-seat pricing, up to 35% savings for growing teams',
    excerpt:
      'Voxire today announced the Scale plan for rtylr, offering automatic volume-based per-seat discounts for businesses scaling past 30 active users. The Scale tier pairs with daily seat proration and transparent billing designed to end per-module add-on surprises.',
  },
  {
    date: 'Mar 28, 2026',
    kind: 'Coverage',
    headline: 'TechCrunch: rtylr bets on daily-prorated SMB SaaS',
    excerpt:
      'Coverage of our launch of daily seat proration, positioning rtylr against monthly-billed incumbents in the POS and HR space.',
    source: {
      label: 'Read on TechCrunch',
      href: 'https://techcrunch.com',
    },
  },
  {
    date: 'Mar 12, 2026',
    kind: 'Press release',
    headline: 'rtylr AI enters open beta: design your operation from a sentence',
    excerpt:
      'rtylr AI, the natural-language onboarding system that drafts menus, inventory, team structures, and invoice templates from a plain-English description, is now available to all new workspaces in beta.',
  },
  {
    date: 'Feb 24, 2026',
    kind: 'Milestone',
    headline: '5,000 businesses running on rtylr across 17 countries',
    excerpt:
      'Eighteen months after launch, rtylr powers over five thousand active businesses across 17 countries. The platform processes more than $4M in transactions every day through the built-in POS.',
  },
  {
    date: 'Feb 10, 2026',
    kind: 'Coverage',
    headline: 'Wall Street Journal: The SMB OS wave',
    excerpt:
      'Feature on the rise of integrated SMB operating systems, with rtylr cited alongside Toast and Square as leaders of the category.',
    source: {
      label: 'Read on WSJ',
      href: 'https://wsj.com',
    },
  },
  {
    date: 'Jan 30, 2026',
    kind: 'Award',
    headline: 'rtylr named Product Hunt Product of the Week',
    excerpt:
      'rtylr was voted Product of the Week on Product Hunt on Jan 29, 2026, after launching the integrated HR and POS tipping flow.',
    source: {
      label: 'View on Product Hunt',
      href: 'https://www.producthunt.com',
    },
  },
  {
    date: 'Jan 10, 2026',
    kind: 'Press release',
    headline: 'Public API v1 and Zapier app now generally available',
    excerpt:
      'rtylr ships its first stable public API with full coverage across POS, ERP, HR, CRM, Finance, and Flow. The official Zapier app adds 36 triggers and 22 actions on launch day.',
  },
  {
    date: 'Dec 18, 2025',
    kind: 'Milestone',
    headline: 'Series A: $18M led by Index Ventures',
    excerpt:
      'Voxire raised an $18M Series A to accelerate rtylr across MENA, Europe, and North America. Index Ventures leads with participation from existing investors.',
  },
];

const KIND_STYLES: Record<NewsItem['kind'], string> = {
  'Press release': 'press',
  Coverage: 'coverage',
  Milestone: 'milestone',
  Award: 'award',
};

const NewsroomPage = (): JSX.Element => {
  return (
    <AppLayout>
      <MainLayout>
        <HelmetContainer
          title="Newsroom | rtylr"
          description="Press releases, media coverage, company milestones, and media-kit assets."
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
                <Newspaper size={13} strokeWidth={2} />
                NEWSROOM
              </motion.span>
              <motion.h1 className={styles.heroTitle} variants={blurFadeIn}>
                What we said, what
                <br />
                the press said back.
              </motion.h1>
              <motion.p className={styles.heroSubtitle} variants={blurFadeIn}>
                Announcements, milestones, coverage, and awards. Journalists can reach our comms
                team at <a href="mailto:press@voxire.com">press@voxire.com</a> — we aim to respond
                within 4 business hours on weekdays.
              </motion.p>

              <motion.div className={styles.heroStrip} variants={blurFadeIn}>
                <a
                  href="https://rtylr.com/media-kit.zip"
                  className={styles.mediaLink}
                  target="_blank"
                  rel="noopener noreferrer"
                >
                  <Download size={14} strokeWidth={2} />
                  Media kit (logos, brand, photography)
                </a>
                <a href="mailto:press@voxire.com" className={styles.mediaLink}>
                  <Mail size={14} strokeWidth={2} />
                  press@voxire.com
                </a>
              </motion.div>
            </motion.div>
          </section>

          {/* NEWS LIST */}
          <section className={styles.listSection}>
            <motion.ul
              className={styles.newsList}
              initial="hidden"
              whileInView="visible"
              viewport={viewportOnce}
              variants={staggerBlur}
            >
              {NEWS.map((item, i) => (
                <motion.li
                  key={`${item.date}-${i}`}
                  className={styles.newsItem}
                  variants={blurFadeIn}
                >
                  <div className={styles.newsMeta}>
                    <span
                      className={`${styles.newsKind} ${styles[`kind_${KIND_STYLES[item.kind]}`]}`}
                    >
                      {item.kind}
                    </span>
                    <span className={styles.newsDate}>{item.date}</span>
                  </div>
                  <h2 className={styles.newsHeadline}>{item.headline}</h2>
                  <p className={styles.newsExcerpt}>{item.excerpt}</p>
                  {item.source && (
                    <a
                      className={styles.newsSource}
                      href={item.source.href}
                      target="_blank"
                      rel="noopener noreferrer"
                    >
                      {item.source.label}
                      <ArrowUpRight size={13} strokeWidth={2} />
                    </a>
                  )}
                </motion.li>
              ))}
            </motion.ul>
          </section>

          {/* CONTACT CTA */}
          <motion.section
            className={styles.ctaBlock}
            initial="hidden"
            whileInView="visible"
            viewport={viewportOnce}
            variants={staggerBlur}
          >
            <motion.div className={styles.ctaInner} variants={blurFadeIn}>
              <h2 className={styles.ctaTitle}>For the press</h2>
              <p className={styles.ctaCopy}>
                Looking for a quote, a product screenshot, or a founder interview? Email{' '}
                <a href="mailto:press@voxire.com">press@voxire.com</a>. For partnership or business
                queries, see our <Link href="/partners">partner program</Link> or{' '}
                <Link href="/contact">contact form</Link>.
              </p>
            </motion.div>
          </motion.section>
        </div>
      </MainLayout>
    </AppLayout>
  );
};

export default NewsroomPage;
