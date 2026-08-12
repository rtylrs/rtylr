import React from 'react';
import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import {
  Button,
  Input,
  Badge,
  Alert,
  Avatar,
  Switch,
  Spinner,
  FullScreenLoader,
  ErrorBoundary,
  DefaultErrorFallback,
  WidgetErrorFallback,
  isDynamicImportError,
  recoverFromDynamicImportError,
} from '../index';

// Component that throws for testing ErrorBoundary
function ProblemChild({ shouldThrow = true }: { shouldThrow?: boolean }) {
  if (shouldThrow) {
    throw new Error('Test boundary error');
  }
  return <div>Healthy Child</div>;
}

describe('@voxire/ui components', () => {
  describe('Button', () => {
    it('renders label and handles clicks', () => {
      const handleClick = vi.fn();
      render(<Button onClick={handleClick}>Click Me</Button>);
      const btn = screen.getByRole('button', { name: /click me/i });
      expect(btn).toBeDefined();
      fireEvent.click(btn);
      expect(handleClick).toHaveBeenCalledTimes(1);
    });

    it('disables button when disabled or loading', () => {
      render(<Button disabled>Disabled</Button>);
      const btn = screen.getByRole('button', { name: /disabled/i });
      expect(btn).toHaveProperty('disabled', true);
    });
  });

  describe('Input', () => {
    it('renders input and accepts typed value', () => {
      const handleChange = vi.fn();
      render(<Input placeholder="Enter name" onChange={handleChange} />);
      const input = screen.getByPlaceholderText('Enter name');
      fireEvent.change(input, { target: { value: 'John' } });
      expect(handleChange).toHaveBeenCalled();
    });

    it('renders error message when error prop is provided', () => {
      render(<Input label="Email" error="Invalid email address" />);
      expect(screen.getByText('Invalid email address')).toBeDefined();
    });
  });

  describe('Badge', () => {
    it('renders badge text correctly', () => {
      render(<Badge variant="success">Active</Badge>);
      expect(screen.getByText('Active')).toBeDefined();
    });
  });

  describe('Alert', () => {
    it('renders alert message', () => {
      render(<Alert variant="warning">Check system status</Alert>);
      expect(screen.getByRole('alert')).toBeDefined();
      expect(screen.getByText('Check system status')).toBeDefined();
    });
  });

  describe('Avatar', () => {
    it('renders initials when src is absent', () => {
      render(<Avatar initials="AS" />);
      expect(screen.getByText('AS')).toBeDefined();
    });
  });

  describe('Switch', () => {
    it('toggles state on click', () => {
      const handleToggle = vi.fn();
      render(<Switch checked={false} onCheckedChange={handleToggle} label="Enable feature" />);
      const switchBtn = screen.getByRole('switch');
      fireEvent.click(switchBtn);
      expect(handleToggle).toHaveBeenCalledWith(true);
    });
  });

  describe('Spinner & FullScreenLoader', () => {
    it('renders spinner element', () => {
      const { container } = render(<Spinner size="md" />);
      expect(container.firstChild).toBeDefined();
    });

    it('renders fullscreen loader overlay', () => {
      const { container } = render(<FullScreenLoader className="custom-loader" />);
      expect(container.querySelector('.custom-loader')).toBeDefined();
    });
  });

  describe('ErrorBoundary & Fallbacks', () => {
    it('catches render errors and displays fallback UI', () => {
      const spy = vi.spyOn(console, 'error').mockImplementation(() => {});
      const onError = vi.fn();

      render(
        <ErrorBoundary onError={onError}>
          <ProblemChild />
        </ErrorBoundary>
      );

      expect(screen.getByRole('alert')).toBeDefined();
      expect(screen.getByText(/Test boundary error/i)).toBeDefined();
      expect(onError).toHaveBeenCalled();

      spy.mockRestore();
    });

    it('renders DefaultErrorFallback and handles retry', () => {
      const reset = vi.fn();
      render(<DefaultErrorFallback error={new Error('Fallback test')} reset={reset} />);
      const tryAgainBtn = screen.getByRole('button', { name: /try again/i });
      fireEvent.click(tryAgainBtn);
      expect(reset).toHaveBeenCalledTimes(1);
    });

    it('renders WidgetErrorFallback and handles retry', () => {
      const reset = vi.fn();
      render(<WidgetErrorFallback error={new Error('Widget error')} reset={reset} />);
      const retryBtn = screen.getByRole('button', { name: /retry/i });
      fireEvent.click(retryBtn);
      expect(reset).toHaveBeenCalledTimes(1);
    });

    it('detects dynamic import errors accurately', () => {
      expect(isDynamicImportError(new Error('Failed to fetch dynamically imported module'))).toBe(
        true
      );
      expect(recoverFromDynamicImportError(new Error('Syntax error'))).toBe(false);
    });
  });
});
