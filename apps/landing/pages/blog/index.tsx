import React, { useState, useMemo } from 'react';
import { motion } from 'motion/react';
import Link from 'next/link';
import { Search, Clock } from 'lucide-react';
import AppLayout from '../../components/Layout/AppLayout';
import MainLayout from '../../layouts/Main/MainLayout';
import HelmetContainer from '../../components/HelmetContainer';
import { blurFadeIn, staggerBlur, viewportOnce } from '../../lib/animations';
import styles from './blog.module.scss';

// ── Blog post index ────────────────────────────────────────────────────────
// In a real CMS these would come from MDX or a headless CMS. For now they
// live here as the source of truth.
interface Post {
  slug: string;
  title: string;
  excerpt: string;
  category: 'Product' | 'Engineering' | 'Playbook' | 'Story' | 'Announcement';
  author: string;
  date: string;
  readingMinutes: number;
  featured?: boolean;
}

const POSTS: Post[] = [
  {
    slug: 'per-seat-daily-proration',
    title: 'Why we bill by the day, not by the month.',
    excerpt:
      'Restaurant turnover is 75% annually. Charging a full month for a server who lasted nine days is a tax on small business. Here\u2019s how daily proration actually works under the hood.',
    category: 'Product',
    author: 'Mohammad Homsi',
    date: 'Apr 14, 2026',
    readingMinutes: 7,
    featured: true,
  },
  {
    slug: 'one-platform-not-seven-tools',
    title: 'One platform, not seven tools: the case against bundles.',
    excerpt:
      'Most "all-in-one" systems are a UI veneer over seven separate databases. When POS says an item is in stock and ERP says it isn\u2019t, someone\u2019s lying. We chose a harder architecture.',
    category: 'Engineering',
    author: 'Abed El-Fattah Amouneh',
    date: 'Apr 07, 2026',
    readingMinutes: 9,
  },
  {
    slug: 'offline-pos-design',
    title: 'Designing a POS that survives a dead WiFi router.',
    excerpt:
      'Service doesn\u2019t pause because the internet did. A breakdown of the CRDT-based offline queue, the conflict-resolution rules, and why we chose eventual consistency over strict ordering.',
    category: 'Engineering',
    author: 'Abed El-Fattah Amouneh',
    date: 'Mar 28, 2026',
    readingMinutes: 12,
  },
  {
    slug: 'restaurant-opening-playbook',
    title: 'The first 30 days: a new restaurant\u2019s rtylr playbook.',
    excerpt:
      'How to configure menu, schedule, payroll, and suppliers in your first week so nothing drags on opening night. Real checklists from three of our early customers.',
    category: 'Playbook',
    author: 'Laila Ahmed',
    date: 'Mar 20, 2026',
    readingMinutes: 8,
  },
  {
    slug: 'ai-built-in-not-bolted-on',
    title: 'AI that understands your business, not just your prompts.',
    excerpt:
      'Generic chatbots know the world. rtylr AI knows your menu, your suppliers, your payroll runs, and your busy nights. Here\u2019s how that changes what the assistant can actually do.',
    category: 'Product',
    author: 'Mohammad Homsi',
    date: 'Mar 12, 2026',
    readingMinutes: 6,
  },
  {
    slug: 'why-we-left-square',
    title: 'Why Byblos Kitchen left Square after three years.',
    excerpt:
      '\u201CI was paying per location, per register, per module, per payroll run, per SMS. I couldn\u2019t tell you what my software cost me.\u201D A customer story.',
    category: 'Story',
    author: 'Laila Ahmed',
    date: 'Mar 05, 2026',
    readingMinutes: 5,
  },
  {
    slug: 'rtylr-beta-announcement',
    title: 'rtylr AI is in beta: design your business from a sentence.',
    excerpt:
      'Tell the AI what your business is and it drafts your menu, inventory items, team roles, schedule template, and invoice templates. You review, approve, launch.',
    category: 'Announcement',
    author: 'Voxire Team',
    date: 'Feb 26, 2026',
    readingMinutes: 3,
  },
  {
    slug: 'multi-location-accounting',
    title: 'Multi-location accounting without the month-end panic.',
    excerpt:
      'Four locations, four register reconciliations, one set of books. The concrete bookkeeping pattern we recommend (and the one we\u2019ve seen go wrong).',
    category: 'Playbook',
    author: 'Sam Patel',
    date: 'Feb 18, 2026',
    readingMinutes: 10,
  },
];

const CATEGORIES = ['All', 'Product', 'Engineering', 'Playbook', 'Story', 'Announcement'] as const;

const BlogPage = (): JSX.Element => {
  const [query, setQuery] = useState('');
  const [activeCategory, setActiveCategory] = useState<(typeof CATEGORIES)[number]>('All');

  const { featured, rest } = useMemo(() => {
    const q = query.trim().toLowerCase();
    const filtered = POSTS.filter((p) => {
      if (activeCategory !== 'All' && p.category !== activeCategory) return false;
      if (!q) return true;
      return (
        p.title.toLowerCase().includes(q) ||
        p.excerpt.toLowerCase().includes(q) ||
        p.author.toLowerCase().includes(q)
      );
    });
    const featuredPost = filtered.find((p) => p.featured);
    const others = filtered.filter((p) => p !== featuredPost);
    return { featured: featuredPost, rest: others };
  }, [query, activeCategory]);

  return (
    <AppLayout>
      <MainLayout>
        <HelmetContainer
          title="Blog | rtylr"
          description="Product, engineering, and playbooks from the rtylr team. How we build, how to run your business on it, and why."
          canonical="https://rtylr.com/blog/"
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
                BLOG
              </motion.span>
              <motion.h1 className={styles.heroTitle} variants={blurFadeIn}>
                How we build rtylr,
                <br />
                and how to build on it.
              </motion.h1>
              <motion.p className={styles.heroSubtitle} variants={blurFadeIn}>
                Engineering deep-dives, operator playbooks, customer stories, and the occasional
                opinion on how SMB software should work.
              </motion.p>

              <motion.div className={styles.controls} variants={blurFadeIn}>
                <div className={styles.searchBox}>
                  <Search size={15} strokeWidth={2} className={styles.searchIcon} />
                  <input
                    type="search"
                    placeholder="Search posts, authors, topics…"
                    value={query}
                    onChange={(e) => setQuery(e.target.value)}
                    className={styles.searchInput}
                  />
                </div>
                <div className={styles.categoryTabs}>
                  {CATEGORIES.map((cat) => (
                    <button
                      key={cat}
                      type="button"
                      onClick={() => setActiveCategory(cat)}
                      className={`${styles.categoryTab} ${activeCategory === cat ? styles.categoryTabActive : ''}`}
                    >
                      {cat}
                    </button>
                  ))}
                </div>
              </motion.div>
            </motion.div>
          </section>

          {/* FEATURED POST */}
          {featured && (
            <motion.section
              className={styles.featuredSection}
              initial="hidden"
              whileInView="visible"
              viewport={viewportOnce}
              variants={staggerBlur}
            >
              <Link href={`/blog/${featured.slug}`} className={styles.featured}>
                <div className={styles.featuredLeft}>
                  <span className={styles.featuredBadge}>FEATURED</span>
                  <h2 className={styles.featuredTitle}>{featured.title}</h2>
                  <p className={styles.featuredExcerpt}>{featured.excerpt}</p>
                  <div className={styles.featuredMeta}>
                    <span className={styles.postCategory}>{featured.category}</span>
                    <span className={styles.postDot}>·</span>
                    <span>{featured.author}</span>
                    <span className={styles.postDot}>·</span>
                    <span>{featured.date}</span>
                    <span className={styles.postDot}>·</span>
                    <span className={styles.postReadTime}>
                      <Clock size={11} strokeWidth={2} />
                      {featured.readingMinutes} min read
                    </span>
                  </div>
                </div>
                <div className={styles.featuredRight} aria-hidden="true">
                  <div className={styles.featuredArtwork}>
                    <div className={styles.artworkBar} style={{ width: '72%' }} />
                    <div className={styles.artworkBar} style={{ width: '48%' }} />
                    <div className={styles.artworkBar} style={{ width: '88%' }} />
                    <div className={styles.artworkBar} style={{ width: '34%' }} />
                  </div>
                </div>
              </Link>
            </motion.section>
          )}

          {/* POST GRID */}
          <section className={styles.gridSection}>
            {rest.length === 0 && !featured ? (
              <div className={styles.empty}>
                No posts match. Try a different filter or search term.
              </div>
            ) : (
              <motion.div
                className={styles.grid}
                initial="hidden"
                whileInView="visible"
                viewport={viewportOnce}
                variants={staggerBlur}
                key={`${query}-${activeCategory}`}
              >
                {rest.map((post) => (
                  <motion.div key={post.slug} variants={blurFadeIn}>
                    <Link href={`/blog/${post.slug}`} className={styles.postCard}>
                      <div className={styles.postMetaTop}>
                        <span className={styles.postCategory}>{post.category}</span>
                      </div>
                      <h3 className={styles.postTitle}>{post.title}</h3>
                      <p className={styles.postExcerpt}>{post.excerpt}</p>
                      <div className={styles.postMetaBottom}>
                        <span>{post.author}</span>
                        <span className={styles.postDot}>·</span>
                        <span>{post.date}</span>
                        <span className={styles.postDot}>·</span>
                        <span className={styles.postReadTime}>
                          <Clock size={11} strokeWidth={2} />
                          {post.readingMinutes} min
                        </span>
                      </div>
                    </Link>
                  </motion.div>
                ))}
              </motion.div>
            )}
          </section>
        </div>
      </MainLayout>
    </AppLayout>
  );
};

export default BlogPage;
