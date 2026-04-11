import React from 'react';
import { useRouter } from 'next/router';
import { Button } from '@voxire/components';
import AppLayout from '../../components/Layout/AppLayout';
import MainLayout from '../../layouts/Main/MainLayout';
import HelmetContainer from '../../components/HelmetContainer';
import styles from './hardware.module.scss';

interface HardwareItem {
  name: string;
  description: string;
  imageUrl: string;
}

const hardwareItems: HardwareItem[] = [
  {
    name: 'Digital Scale – HS-117',
    description:
      'Precision digital scale perfect for weighing ingredients, products, and packages. Features high accuracy and easy-to-read display.',
    imageUrl: 'https://via.placeholder.com/400x300/232220/ffffff?text=Digital+Scale+HS-117',
  },
  {
    name: 'Cash Drawer',
    description:
      'Secure cash drawer solution for point-of-sale systems. Durable construction with secure locking mechanism and smooth operation.',
    imageUrl: 'https://via.placeholder.com/400x300/232220/ffffff?text=Cash+Drawer',
  },
  {
    name: 'All-in-One Touch Screen HS-9401',
    description:
      'Complete POS solution with integrated touch screen display. Combines power and functionality in a sleek, space-saving design.',
    imageUrl: 'https://via.placeholder.com/400x300/232220/ffffff?text=Touch+Screen+HS-9401',
  },
  {
    name: 'Label Barcode Printer',
    description:
      'High-speed thermal label printer for barcode labels and product tags. Ideal for inventory management and retail operations.',
    imageUrl: 'https://via.placeholder.com/400x300/232220/ffffff?text=Label+Barcode+Printer',
  },
  {
    name: 'Thermal Receipt Printer HS-8250',
    description:
      'Fast and reliable thermal receipt printer for POS transactions. Silent operation with excellent print quality and durability.',
    imageUrl: 'https://via.placeholder.com/400x300/232220/ffffff?text=Thermal+Printer+HS-8250',
  },
  {
    name: 'Bluetooth Printer',
    description:
      'Wireless Bluetooth-enabled printer for flexible printing solutions. Connect seamlessly to mobile devices and POS systems.',
    imageUrl: 'https://via.placeholder.com/400x300/232220/ffffff?text=Bluetooth+Printer',
  },
  {
    name: 'Single Screen with Software',
    description:
      'Complete single-screen POS system with integrated software. Perfect for small businesses looking for an all-in-one solution.',
    imageUrl: 'https://via.placeholder.com/400x300/232220/ffffff?text=Single+Screen+POS',
  },
  {
    name: 'Tablet 10.1" (Windows 10, Intel N4020)',
    description:
      'Versatile 10.1-inch tablet running Windows 10. Powered by Intel N4020 processor, ideal for mobile POS and order management.',
    imageUrl: 'https://via.placeholder.com/400x300/232220/ffffff?text=Tablet+10.1',
  },
  {
    name: 'Barcode Reader',
    description:
      'High-performance barcode scanner for quick and accurate product scanning. Supports multiple barcode formats and connection types.',
    imageUrl: 'https://via.placeholder.com/400x300/232220/ffffff?text=Barcode+Reader',
  },
  {
    name: 'Barcode Reader (Alternative)',
    description:
      'Alternative barcode scanning solution with enhanced features. Ergonomic design and advanced scanning technology for retail environments.',
    imageUrl: 'https://via.placeholder.com/400x300/232220/ffffff?text=Barcode+Reader+Alt',
  },
  {
    name: 'Barcode Reader (Another Model)',
    description:
      'Another reliable barcode scanner option with compact design. Perfect for tight spaces and high-volume scanning operations.',
    imageUrl: 'https://via.placeholder.com/400x300/232220/ffffff?text=Barcode+Reader+Model',
  },
  {
    name: 'Price Checker',
    description:
      'Customer-facing price checker display for retail environments. Helps customers verify prices quickly and improves transparency.',
    imageUrl: 'https://via.placeholder.com/400x300/232220/ffffff?text=Price+Checker',
  },
];

const HardwarePage = (): JSX.Element => {
  const router = useRouter();

  const handleBookDemo = () => {
    router.push('/demo');
  };

  const handleContactSales = () => {
    router.push('/contact');
  };

  return (
    <AppLayout>
      <MainLayout>
        <HelmetContainer title="Hardware | Rtylr" />
        <div className={styles.hardwarePage}>
          <div className={styles.container}>
            <div className={styles.header}>
              <h1 className={styles.title}>Hardware Solutions</h1>
              <p className={styles.subtitle}>
                Complete hardware ecosystem designed to integrate seamlessly with Rtylr POS. All
                equipment is tested and optimized for peak performance.
              </p>
            </div>

            <div className={styles.grid}>
              {hardwareItems.map((item, index) => (
                <div key={index} className={styles.card}>
                  <div className={styles.imageContainer}>
                    <img src={item.imageUrl} alt={item.name} className={styles.image} />
                  </div>
                  <div className={styles.content}>
                    <h3 className={styles.cardTitle}>{item.name}</h3>
                    <p className={styles.cardDescription}>{item.description}</p>
                  </div>
                </div>
              ))}
            </div>

            <div className={styles.ctaSection}>
              <h2 className={styles.ctaTitle}>Ready to Get Started?</h2>
              <p className={styles.ctaDescription}>
                Connect with our team to discuss hardware solutions tailored to your business needs.
              </p>
              <div className={styles.ctaButtons}>
                <Button variant="dark-outline" onClick={handleBookDemo} size="lg">
                  Book a Demo
                </Button>
                <Button variant="soft" onClick={handleContactSales} size="lg">
                  Contact Sales
                </Button>
              </div>
            </div>
          </div>
        </div>
      </MainLayout>
    </AppLayout>
  );
};

export default HardwarePage;
