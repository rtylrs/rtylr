import React from 'react';
import { cn } from '../../../lib/cn';
import { motion } from 'motion/react';
import { fadeInUp, viewportOnce } from '../../../lib/animations';
import styles from './TestimonialsSection.module.scss';

const testimonials = [
  {
    quote: 'كنا نقفل الحسابات في الثامن من كل شهر. هلق بنقفلها بالثاني، ما في شي لازم نسوي تسوية.',
    author: 'ليلى أحمد',
    role: 'صاحبة مطعم، مطبخ بيبلوس',
    company: 'مطعم · ٣ فروع',
    image: 'https://placehold.co/80x80/d6b88e/1a1612?text=LA',
  },
  {
    quote:
      'النادل وصل لـ٢٠ زيارة وrtylr كتب إيميل الشكر لحاله. مديري سألني إذا أنا ضبطتها. ما ضبطتها.',
    author: 'كريم منصور',
    role: 'مدير عام، مطعم التراس',
    company: 'مطعم',
    image: 'https://placehold.co/80x80/c9a48a/1a1612?text=KM',
  },
  {
    quote:
      'انقطع الإنترنت ليلة السبت وعنا ٢٠٠ طاولة. الكاشير ما وقف. ما كنت أتوقع برنامج يقدر يعمل هيك.',
    author: 'طارق الزين',
    role: 'صاحب، كافيه الجبل',
    company: 'مطعم',
    image: 'https://placehold.co/80x80/b89a8a/1a1612?text=TZ',
  },
  {
    quote:
      'الجدول كان على واتساب. هلق الدوام بينشر بأربع ضغطات والرواتب بتتسوى تلقائي كل أحد بالليل.',
    author: 'سارة الحلبي',
    role: 'مسؤولة عمليات، سلسلة Urban',
    company: 'ريتيل · ٥ فروع',
    image: 'https://placehold.co/80x80/a89680/1a1612?text=SH',
  },
  {
    quote:
      'التحويل من النظام القديم وفّر علينا تقريباً ٦٠٠ دولار على كل فرع كل شهر. وهاد قبل ما أحسب الإضافات.',
    author: 'نادين فرحات',
    role: 'مؤسسة، ماركت الطازج',
    company: 'بقالة · ٤ فروع',
    image: 'https://placehold.co/80x80/d6b88e/1a1612?text=NF',
  },
  {
    quote:
      'المهمة مربوطة بالفاتورة، بالزبون، بالدوام. بطّلت أفتح أربع تابات لأربع برامج، بس بفتح Flow.',
    author: 'جاد معلوف',
    role: 'مدير، Tech Store Pro',
    company: 'ريتيل',
    image: 'https://placehold.co/80x80/c9a48a/1a1612?text=JM',
  },
];

const TestimonialsSection: React.FC = () => {
  const duplicatedTestimonials = [...testimonials, ...testimonials];

  return (
    <section className={styles.testimonialsSection}>
      <div className={styles.container}>
        <motion.h2
          className={styles.sectionTitle}
          initial="hidden"
          whileInView="visible"
          viewport={viewportOnce}
          variants={fadeInUp}
        >
          What operators say about it.
        </motion.h2>
        <motion.div
          className={styles.tickersWrapper}
          initial={{ opacity: 0 }}
          whileInView={{ opacity: 1 }}
          viewport={{ once: true, margin: '-50px' }}
          transition={{ duration: 0.8, delay: 0.2 }}
        >
          <div className={styles.tickerContainer}>
            <div className={cn(styles.tickerTrack, styles.tickerLeft)}>
              {duplicatedTestimonials.map((testimonial, index) => (
                <div key={index} className={styles.testimonialCard}>
                  <p className={styles.quote}>&ldquo;{testimonial.quote}&rdquo;</p>
                  <div className={styles.authorInfo}>
                    <img
                      src={testimonial.image}
                      alt={testimonial.author}
                      className={styles.authorImage}
                    />
                    <div className={styles.authorDetails}>
                      <div className={styles.authorName}>{testimonial.author}</div>
                      <div className={styles.authorRole}>{testimonial.role}</div>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>
          <div className={styles.tickerContainer}>
            <div className={cn(styles.tickerTrack, styles.tickerRight)}>
              {duplicatedTestimonials.map((testimonial, index) => (
                <div key={index} className={styles.testimonialCard}>
                  <p className={styles.quote}>&ldquo;{testimonial.quote}&rdquo;</p>
                  <div className={styles.authorInfo}>
                    <img
                      src={testimonial.image}
                      alt={testimonial.author}
                      className={styles.authorImage}
                    />
                    <div className={styles.authorDetails}>
                      <div className={styles.authorName}>{testimonial.author}</div>
                      <div className={styles.authorRole}>{testimonial.role}</div>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </motion.div>
      </div>
    </section>
  );
};

export default TestimonialsSection;
