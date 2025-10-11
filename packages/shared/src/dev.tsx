import React from 'react';
import { createRoot } from 'react-dom/client';
import { Button } from './shared';
import { FaPlus, FaDownload, FaHeart, FaArrowRight } from 'react-icons/fa';
import './shared/shared.scss';
import './shared/components/Button/Button.scss';

const DevApp: React.FC = () => {
  return (
    <div style={{ padding: '40px', maxWidth: '1200px', margin: '0 auto' }}>
      <h1>Button Component Variants</h1>

      {/* Primary Variants */}
      <div style={{ marginBottom: '40px' }}>
        <h2>Primary Buttons</h2>
        <div style={{ display: 'flex', gap: '16px', flexWrap: 'wrap', marginBottom: '20px' }}>
          <Button variant="primary" size="sm">Small Primary</Button>
          <Button variant="primary" size="md">Medium Primary</Button>
          <Button variant="primary" size="lg">Large Primary</Button>
        </div>
        <div style={{ display: 'flex', gap: '16px', flexWrap: 'wrap' }}>
          <Button variant="primary" prefix={<FaPlus />}>With Prefix</Button>
          <Button variant="primary" suffix={<FaArrowRight />}>With Suffix</Button>
          <Button variant="primary" loading>Loading</Button>
        </div>
      </div>

      {/* Secondary Variants */}
      <div style={{ marginBottom: '40px' }}>
        <h2>Secondary Buttons</h2>
        <div style={{ display: 'flex', gap: '16px', flexWrap: 'wrap', marginBottom: '20px' }}>
          <Button variant="secondary" size="sm">Small Secondary</Button>
          <Button variant="secondary" size="md">Medium Secondary</Button>
          <Button variant="secondary" size="lg">Large Secondary</Button>
        </div>
        <div style={{ display: 'flex', gap: '16px', flexWrap: 'wrap' }}>
          <Button variant="secondary" prefix={<FaDownload />}>Download</Button>
          <Button variant="secondary" suffix={<FaArrowRight />}>Continue</Button>
          <Button variant="secondary" disabled>Disabled</Button>
        </div>
      </div>

      {/* Outline Variants */}
      <div style={{ marginBottom: '40px' }}>
        <h2>Outline Buttons</h2>
        <div style={{ display: 'flex', gap: '16px', flexWrap: 'wrap', marginBottom: '20px' }}>
          <Button variant="outline" size="sm">Small Outline</Button>
          <Button variant="outline" size="md">Medium Outline</Button>
          <Button variant="outline" size="lg">Large Outline</Button>
        </div>
        <div style={{ display: 'flex', gap: '16px', flexWrap: 'wrap' }}>
          <Button variant="outline" prefix={<FaHeart />}>Like</Button>
          <Button variant="outline" suffix={<FaArrowRight />}>Next</Button>
          <Button variant="outline" loading>Processing</Button>
        </div>
      </div>

      {/* Ghost Variants */}
      <div style={{ marginBottom: '40px' }}>
        <h2>Ghost Buttons</h2>
        <div style={{ display: 'flex', gap: '16px', flexWrap: 'wrap', marginBottom: '20px' }}>
          <Button variant="ghost" size="sm">Small Ghost</Button>
          <Button variant="ghost" size="md">Medium Ghost</Button>
          <Button variant="ghost" size="lg">Large Ghost</Button>
        </div>
        <div style={{ display: 'flex', gap: '16px', flexWrap: 'wrap' }}>
          <Button variant="ghost" prefix={<FaPlus />}>Add</Button>
          <Button variant="ghost" suffix={<FaArrowRight />}>View More</Button>
          <Button variant="ghost" disabled>Disabled</Button>
        </div>
      </div>

      {/* Soft Variants - NEW */}
      <div style={{ marginBottom: '40px' }}>
        <h2>Soft Buttons (New)</h2>
        <div style={{ display: 'flex', gap: '16px', flexWrap: 'wrap', marginBottom: '20px' }}>
          <Button variant="soft" size="sm">Small Soft</Button>
          <Button variant="soft" size="md">Medium Soft</Button>
          <Button variant="soft" size="lg">Large Soft</Button>
        </div>
        <div style={{ display: 'flex', gap: '16px', flexWrap: 'wrap' }}>
          <Button variant="soft" prefix={<FaPlus />}>Create</Button>
          <Button variant="soft" suffix={<FaArrowRight />}>Continue</Button>
          <Button variant="soft" loading>Saving</Button>
        </div>
      </div>

      {/* Dark Outline Variants - NEW */}
      <div style={{ marginBottom: '40px' }}>
        <h2>Dark Outline Buttons (New)</h2>
        <div style={{ display: 'flex', gap: '16px', flexWrap: 'wrap', marginBottom: '20px' }}>
          <Button variant="dark-outline" size="sm">Small Dark</Button>
          <Button variant="dark-outline" size="md">Medium Dark</Button>
          <Button variant="dark-outline" size="lg">Large Dark</Button>
        </div>
        <div style={{ display: 'flex', gap: '16px', flexWrap: 'wrap' }}>
          <Button variant="dark-outline" prefix={<FaDownload />}>Download</Button>
          <Button variant="dark-outline" suffix={<FaArrowRight />}>Get Started</Button>
          <Button variant="dark-outline" disabled>Disabled</Button>
        </div>
      </div>

      {/* Full Width Examples */}
      <div style={{ marginBottom: '40px' }}>
        <h2>Full Width Examples</h2>
        <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
          <Button variant="primary" fullWidth>Full Width Primary</Button>
          <Button variant="soft" fullWidth prefix={<FaPlus />}>Full Width Soft with Icon</Button>
          <Button variant="dark-outline" fullWidth suffix={<FaArrowRight />}>Full Width Dark with Icon</Button>
        </div>
      </div>
    </div>
  );
};

// Mount the app
const container = document.getElementById('root');
if (container) {
  const root = createRoot(container);
  root.render(<DevApp />);
}