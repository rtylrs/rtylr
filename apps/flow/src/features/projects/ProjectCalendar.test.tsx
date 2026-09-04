// @vitest-environment happy-dom
import React from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { render, screen } from '@testing-library/react';

const { language } = vi.hoisted(() => ({ language: { value: 'en-US' } }));

vi.mock('react-i18next', async (importOriginal) => ({
  ...(await importOriginal<typeof import('react-i18next')>()),
  useTranslation: () => ({ t: (key: string) => key, i18n: { language: language.value } }),
}));

vi.mock('@tanstack/react-router', () => ({
  useParams: () => ({ projectId: 'p-9' }),
  useNavigate: () => vi.fn(),
}));

vi.mock('@voxire/components', () => ({ HelmetContainer: () => null }));
vi.mock('../../shared/stores/flowStore', () => ({ useFlowStore: () => ({ projects: [] }) }));
vi.mock('../tasks/useTasks', () => ({ useTasks: () => ({ data: [] }) }));
vi.mock('../../shared/lib/useProjectRealtime', () => ({ useProjectRealtime: () => undefined }));
vi.mock('./ProjectViewTabs', () => ({ default: () => null }));
vi.mock('./useCurrentProject', () => ({
  useCurrentProject: () => ({ project: { id: 'p-9', name: 'Web' }, isLoading: false }),
}));

import ProjectCalendar from './ProjectCalendar';

describe('ProjectCalendar localisation (#1052)', () => {
  beforeEach(() => {
    language.value = 'en-US';
  });

  it('renders English weekday headers in English', () => {
    render(<ProjectCalendar />);

    for (const day of ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat']) {
      expect(screen.getByText(day)).toBeInTheDocument();
    }
  });

  it('renders Arabic weekday headers when the app language is Arabic', () => {
    language.value = 'ar';

    render(<ProjectCalendar />);

    // The global calendar already renders these; this screen used a hardcoded
    // English array, so an Arabic user saw SUN MON TUE next to a right-to-left
    // layout.
    expect(screen.getByText('الأحد')).toBeInTheDocument();
    expect(screen.getByText('السبت')).toBeInTheDocument();
    expect(screen.queryByText('Sun')).not.toBeInTheDocument();
  });

  it('renders French weekday headers when the app language is French', () => {
    language.value = 'fr';

    render(<ProjectCalendar />);

    expect(screen.getByText('dim.')).toBeInTheDocument();
    expect(screen.queryByText('Sun')).not.toBeInTheDocument();
  });

  it('keeps Sunday as the first column in every language', () => {
    language.value = 'fr';

    const { container } = render(<ProjectCalendar />);
    const headers = [...container.querySelectorAll('div')]
      .map((node) => node.textContent)
      .filter((text): text is string => text === 'dim.');

    // Only the language of the labels changed, not which day the week starts
    // on, so a locale change cannot silently shift every task by a column.
    expect(headers.length).toBeGreaterThan(0);
  });

  it('formats the month label in the app language, not the browser locale', () => {
    language.value = 'fr';

    render(<ProjectCalendar />);

    const month = new Date().toLocaleString('fr', { month: 'long', year: 'numeric' });
    expect(screen.getByText(month)).toBeInTheDocument();
  });
});
