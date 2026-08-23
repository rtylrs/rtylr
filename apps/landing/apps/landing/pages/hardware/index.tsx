import React from 'react';
import { motion } from 'motion/react';
import { useRouter } from 'next/router';
import {
  Scale,
  Archive,
  Printer,
  ScanLine,
  CreditCard,
  Monitor,
  Tablet,
  MonitorSmartphone,
  Tag,
} from 'lucide-react';
import type { LucideIcon } from 'lucide-react';
import AppLayout from '../../components/Layout/AppLayout';
import MainLayout from '../../layouts/Main/MainLayout';
import HelmetContainer from '../../components/HelmetContainer';
import { productPageSchema } from '../../lib/structured-data';
import { blurFadeIn, staggerBlur, viewportOnce } from '../../lib/animations';
import styles from './hardware.module.scss';

interface HardwareItem {
  name: string;
  model: string;
  description: string;
  icon: LucideIcon;
}

const HARDWARE_ITEMS: HardwareItem[] = [
  {
    name: 'Digital Scale',
    model: 'HS-117',
    description:
      'Precision weighing for ingredients and products. High-accuracy load cell with backlit LCD display.',
    icon: Scale,
  },
  {
    name: 'Cash Drawer',
    model: 'CD-460',
    description:
      'Secure dual-bill, dual-coin POS cash drawer with heavy-duty steel construction and auto-open trigger.',
    icon: Archive,
  },
  {
    name: 'Thermal Receipt Printer',
    model: 'HS-8250',
    description:
      '80mm high-speed receipt printing at 250mm/s. Supports USB, LAN, and Bluetooth connectivity.',
    icon: Printer,
  },
  {
    name: 'Barcode Scanner',
    model: 'BS-200',
    description:
      '1D/2D omnidirectional USB scanner with 360° coverage. Reads damaged and low-contrast barcodes.',
    icon: ScanLine,
  },
  {
    name: 'Card Terminal',
    model: 'CT-500',
    description:
      'Tap, chip & swipe EMV-certified payments. Supports NFC, Apple Pay, Google Pay, and Visa/MC.',
    icon: CreditCard,
  },
  {
    name: 'Kitchen Display System',
    model: 'KDS-15',
    description:
      '15" wall-mount order display with real-time ticket management and multi-station routing.',
    icon: Monitor,
  },
  {
    name: 'Tablet Stand',
    model: 'TS-360',
    description:
      'Adjustable 360° rotating tablet mount. Compatible with 7"–13" tablets. Weighted non-slip base.',
    icon: Tablet,
  },
  {
    name: 'Customer Display',
    model: 'CD-Dual',
    description:
      'Dual-screen customer-facing display. Rear LCD shows order totals, promotions, and loyalty points.',
    icon: MonitorSmartphone,
  },
  {
    name: 'Label Printer',
    model: 'LP-420',
    description:
      '4" thermal label printer for shelf tags, product labels, and expiry stickers. 203 dpi resolution.',
    icon: Tag,
  },
];

const HardwarePage = (): React.JSX.Element => {
  const router = useRouter();

  const handleBookDemo = (): void => {
    void router.push('/demo');
  };

  const handleContactSales = (): void => {
    void router.push('/contact');
  };

  return (
    <AppLayout>
      <MainLayout>
        <HelmetContainer
          title="POS Hardware · Terminals, Printers & Scanners | rtylr"
          description="Tested and certified hardware ecosystem designed for rtylr POS. Every device is optimized out of the box."
          canonical="https://rtylr.com/hardware/"
          structuredData={productPageSchema({
            name: 'Hardware',
            path: '/hardware/',
            description:
              'Tested and certified hardware ecosystem designed for rtylr POS. Every device is optimized out of the box.',
          })}
        />

        <div className={styles.hardwarePage}>
          {/* Hero */}
          <motion.section
            className={styles.hero}
            initial="hidden"
            animate="visible"
            variants={staggerBlur}
          >
            <div className={styles.heroInner}>
              <motion.span className={styles.heroEyebrow} variants={blurFadeIn}>
                <span className={styles.eyebrowDot} />
                RTYLR HARDWARE
              </motion.span>
              <motion.h1 className={styles.heroTitle} variants={blurFadeIn}>
                Plug it in. Ring it up.
              </motion.h1>
              <motion.p className={styles.heroSubtitle} variants={blurFadeIn}>
                Certified printers, card readers, scanners, and kiosks, paired with rtylr POS out of
                the box. No drivers to install, no firmware to sideload, no configuration weekend.
              </motion.p>
            </div>
          </motion.section>

          {/* Hardware grid */}
          <motion.section
            className={styles.gridSection}
            initial="hidden"
            whileInView="visible"
            viewport={viewportOnce}
            variants={staggerBlur}
          >
            <div className={styles.sectionContainer}>
              {/* The device names below are h3. Without an h2 above them the
                  page jumped h1 to h3, which is the one heading rule this
                  repo gates on. */}
              <motion.h2 className={styles.gridTitle} variants={blurFadeIn}>
                What you can plug in
              </motion.h2>
              <div className={styles.hardwareGrid}>
                {HARDWARE_ITEMS.map((item) => {
                  const Icon = item.icon;
                  return (
                    <motion.div
                      key={item.model}
                      className={styles.hardwareCard}
                      variants={blurFadeIn}
                    >
                      <div className={styles.cardAccentBar} aria-hidden="true" />
                      <div className={styles.cardContent}>
                        <div className={styles.cardIconWrap}>
                          <Icon size={24} strokeWidth={1.5} />
                        </div>
                        <div className={styles.cardMeta}>
                          <span className={styles.cardModel}>{item.model}</span>
                          <h3 className={styles.cardName}>{item.name}</h3>
                        </div>
                        <p className={styles.cardDescription}>{item.description}</p>
                      </div>
                    </motion.div>
                  );
                })}
              </div>
            </div>
          </motion.section>

          {/* Integration strip */}
          <motion.section
            className={styles.integrationStrip}
            initial="hidden"
            whileInView="visible"
            viewport={viewportOnce}
            variants={blurFadeIn}
          >
            <div className={styles.stripInner}>
              <span className={styles.stripBadge}>Plug &amp; Play</span>
              <p className={styles.stripText}>
                All hardware integrates with rtylr POS out of the box. Zero drivers, zero
                configuration.
              </p>
            </div>
          </motion.section>

          {/* CTA */}
          <motion.section
            className={styles.ctaSection}
            initial="hidden"
            whileInView="visible"
            viewport={viewportOnce}
            variants={staggerBlur}
          >
            <div className={styles.sectionContainer}>
              <motion.div className={styles.ctaInner} variants={blurFadeIn}>
                <h2 className={styles.ctaTitle}>Get hardware for your business.</h2>
                <p className={styles.ctaSubtitle}>
                  Our team will help you spec the right devices for your setup and ship everything
                  pre-configured for your location.
                </p>
                <div className={styles.ctaButtons}>
                  <button className={styles.ctaPrimary} onClick={handleBookDemo}>
                    Book a demo
                  </button>
                  <button className={styles.ctaSecondary} onClick={handleContactSales}>
                    Contact sales
                  </button>
                </div>
              </motion.div>
            </div>
          </motion.section>
        </div>
      </MainLayout>
    </AppLayout>
  );
};

export default HardwarePage;
