import React from 'react';
import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { ListTodo } from 'lucide-react';
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
  EmptyState,
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

    it('carries a 44px hit area expressed as a literal, not in spacing units', () => {
      render(<Switch checked={false} onCheckedChange={vi.fn()} label="Enable feature" />);
      const cls = screen.getByRole('switch').className;

      // Asserting class strings is usually a poor test. It is the right one here
      // because the defect this guards against is a plausible tidy-up rather
      // than a logic error, and no unit environment loads the compiled CSS to
      // measure the pseudo-element for real.
      //
      // `after:h-11` looks equivalent and is not: it compiles to
      // calc(var(--spacing) * 11), and Flow re-points --spacing at a runtime
      // density multiplier, so the 0.75 step would take this target to 33px. A
      // minimum has to be a minimum. Measured in a browser, not assumed.
      expect(cls).toContain('after:h-[44px]');
      expect(cls).toContain('after:w-[44px]');
      expect(cls).not.toMatch(/after:[hw]-\d/);

      // The visual track deliberately stays 20x36: growing it would move the
      // switch in every consumer's layout.
      expect(cls).toContain('h-5');
      expect(cls).toContain('w-9');
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

  describe('Button — accent contrast and dense size', () => {
    it('does not put white on the accent fill, which measures 1.9:1', () => {
      render(<Button variant="accent">Save</Button>);

      const cls = screen.getByRole('button', { name: 'Save' }).className;
      expect(cls).not.toContain('text-white');
      expect(cls).toContain('text-accent-foreground');
    });

    it('leaves the other variants on white, where the fill is dark enough', () => {
      render(<Button variant="primary">Primary</Button>);

      expect(screen.getByRole('button', { name: 'Primary' }).className).toContain('text-white');
    });

    it('offers an xs size for toolbar and table-row actions', () => {
      render(
        <Button variant="accent" size="xs">
          Export
        </Button>
      );

      expect(screen.getByRole('button', { name: 'Export' }).className).toContain('h-[30px]');
    });

    it('keeps md as the default size', () => {
      render(<Button>Default</Button>);

      expect(screen.getByRole('button', { name: 'Default' }).className).toContain('h-10');
    });
  });

  describe('Badge — dense size', () => {
    it('offers a sm size at 10px for dense table rows', () => {
      render(<Badge size="sm">Pending</Badge>);

      expect(screen.getByText('Pending').className).toContain('text-[10px]');
    });

    it('stays at text-xs by default, so existing badges do not change', () => {
      render(<Badge>Default</Badge>);

      expect(screen.getByText('Default').className).toContain('text-xs');
    });
  });

  describe('Avatar — dense size', () => {
    it('offers an xs size at 24px with 10px initials', () => {
      render(<Avatar initials="MK" size="xs" />);

      const cls = screen.getByText('MK').parentElement?.className ?? '';
      expect(cls).toContain('h-6');
      expect(cls).toContain('w-6');
      expect(cls).toContain('text-[10px]');
    });

    it('keeps md as the default size, so existing avatars do not change', () => {
      render(<Avatar initials="SO" />);

      expect(screen.getByText('SO').parentElement?.className ?? '').toContain('h-9');
    });
  });

  describe('EmptyState', () => {
    it('renders the icon, title, description and a required call-to-action', () => {
      const handleAction = vi.fn();
      render(
        <EmptyState
          icon={ListTodo}
          title="No tasks assigned"
          description="Nothing is on your plate right now."
          action={{ label: 'Add your first task', onClick: handleAction }}
        />
      );

      expect(screen.getByText('No tasks assigned')).toBeDefined();
      expect(screen.getByText('Nothing is on your plate right now.')).toBeDefined();

      const cta = screen.getByRole('button', { name: 'Add your first task' });
      fireEvent.click(cta);
      expect(handleAction).toHaveBeenCalledTimes(1);
    });

    it('keeps the CTA at a 44px touch target on mobile, both sizes', () => {
      render(
        <EmptyState
          icon={ListTodo}
          title="No tasks"
          action={{ label: 'Add task', onClick: vi.fn() }}
        />
      );
      render(
        <EmptyState
          icon={ListTodo}
          title="No activity"
          action={{ label: 'Refresh', onClick: vi.fn() }}
          size="compact"
        />
      );

      expect(screen.getByRole('button', { name: 'Add task' }).className).toContain('h-11');
      expect(screen.getByRole('button', { name: 'Refresh' }).className).toContain('h-11');
    });

    it('renders without a description when none is given', () => {
      render(
        <EmptyState
          icon={ListTodo}
          title="No results"
          action={{ label: 'Clear filters', onClick: vi.fn() }}
        />
      );

      expect(screen.getByText('No results')).toBeDefined();
      expect(screen.getByRole('button', { name: 'Clear filters' })).toBeDefined();
    });

    it('shrinks the illustration, text and button for compact placements', () => {
      render(
        <EmptyState
          icon={ListTodo}
          title="No activity yet"
          action={{ label: 'Refresh', onClick: vi.fn() }}
          size="compact"
        />
      );

      expect(screen.getByText('No activity yet').className).toContain('text-xs');
      expect(screen.getByRole('button', { name: 'Refresh' }).className).toContain('h-[30px]');
    });
  });
});
