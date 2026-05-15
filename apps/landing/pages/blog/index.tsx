import React from 'react';
import { motion } from 'motion/react';
import { PenLine } from 'lucide-react';
import AppLayout from '../../components/Layout/AppLayout';
import MainLayout from '../../layouts/Main/MainLayout';
import HelmetContainer from '../../components/HelmetContainer';
import { blurFadeIn, staggerBlur } from '../../lib/animations';
import styles from './blog.module.scss';

const BlogPage = (): JSX.Element => {
  return (
    <AppLayout>
      <MainLayout>
        <HelmetContainer
          title="Blog | rtylr"
          description="Product, engineering, and playbooks from the rtylr team."
          canonical="https://rtylr.com/blog/"
        />

        <div className={styles.page}>
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
                Coming soon.
              </motion.h1>
              <motion.p className={styles.heroSubtitle} variants={blurFadeIn}>
                Engineering deep-dives, operator playbooks, customer stories, and the occasional
                opinion on how SMB software should work. Check back soon.
              </motion.p>
              <motion.div className={styles.comingSoonIcon} variants={blurFadeIn}>
                <PenLine size={40} strokeWidth={1.5} />
              </motion.div>
            </motion.div>
          </section>
        </div>
      </MainLayout>
    </AppLayout>
  );
};

export default BlogPage;
