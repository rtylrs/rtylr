import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Button } from '@voxire/shared';
import { useRouter } from 'next/router';
import { useConfigStore } from '../../../stores/configStore';

function Faq() {
  const { t } = useTranslation();
  const router = useRouter();
  const [activeIndex, setActiveIndex] = useState(null);
  const { faqs } = useConfigStore();

  const toggleAnswer = (index) => {
    setActiveIndex(activeIndex === index ? null : index);
  };

  return (
    <div className="faq-section">
      <h2 className="vbitez-title">{t('QandA.title')}</h2>
      <div className="faq-container">
        {faqs &&
          faqs.slice(0, 8).map((item: any, index: number) => (
            <div key={index} className="faq-item">
              <div className="faq-question" onClick={() => toggleAnswer(index)}>
                <span>Q. {item.question}</span>
                <span className="faq-icon">
                  {activeIndex === index ? '−' : '+'}
                </span>
              </div>
              {activeIndex === index && (
                <div className="faq-answer">
                  <p>{item.answer}</p>
                </div>
              )}
            </div>
          ))}
      </div>
      <div
        style={{
          marginTop: '40px',
          display: 'flex',
          justifyContent: 'center',
          marginBottom: '20px',
        }}
      >
        <Button variant="secondary" onClick={() => router.push('/faqs')}>
          View All
        </Button>
      </div>
    </div>
  );
}
export default Faq;
