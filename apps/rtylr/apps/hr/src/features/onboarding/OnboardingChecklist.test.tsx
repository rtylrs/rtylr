import React from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { EmployeeChecklistItemPayload } from '@voxire/voxie/hr/onboarding/v1/types';

const mocks = vi.hoisted(() => ({
  useGetEmployeeChecklist: vi.fn(),
  useListChecklistTemplates: vi.fn(),
  completeMutate: vi.fn(),
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('../../shared/hooks/queries/useOnboarding', () => ({
  useGetEmployeeChecklist: mocks.useGetEmployeeChecklist,
  useListChecklistTemplates: mocks.useListChecklistTemplates,
  useCompleteChecklistItem: () => ({ mutate: mocks.completeMutate }),
}));

import OnboardingChecklist from './OnboardingChecklist';

const item = (id: string, checklistItemId: string): EmployeeChecklistItemPayload => ({
  id,
  checklistItemId,
  status: 'incomplete',
  dueDate: '2999-01-01',
});

const items = [item('11', '1'), item('12', '2')];

describe('OnboardingChecklist attachment gate', () => {
  beforeEach(() => {
    mocks.completeMutate.mockReset();
    mocks.useGetEmployeeChecklist.mockReturnValue({
      data: { checklists: [{ id: '5', templateId: '7', items }] },
      isLoading: false,
      refetch: vi.fn(),
    });
    mocks.useListChecklistTemplates.mockReturnValue({
      data: {
        templates: [
          {
            id: '7',
            items: [
              { id: '1', title: 'Sign the handbook', requiresAttachment: false },
              { id: '2', title: 'Upload certification', requiresAttachment: true },
            ],
          },
        ],
      },
    });
  });

  const completeButtons = (): HTMLElement[] =>
    screen.getAllByRole('button', { name: 'common.complete' });

  it('completes an item the template does not require an attachment for', () => {
    render(<OnboardingChecklist employeeId={7} />);

    fireEvent.click(completeButtons()[0]);

    expect(mocks.completeMutate).toHaveBeenCalledWith({ itemId: 11 });
    expect(
      screen.queryByText('onboardingMgmt.checklist.attachmentRequired')
    ).not.toBeInTheDocument();
  });

  it('opens the attachment modal for an item the template requires one for', () => {
    render(<OnboardingChecklist employeeId={7} />);

    fireEvent.click(completeButtons()[1]);

    expect(screen.getByText('onboardingMgmt.checklist.attachmentRequired')).toBeInTheDocument();
    expect(mocks.completeMutate).not.toHaveBeenCalled();
  });

  it('keeps the attachment step when the template is unavailable', () => {
    mocks.useListChecklistTemplates.mockReturnValue({ data: undefined });

    render(<OnboardingChecklist employeeId={7} />);

    fireEvent.click(completeButtons()[0]);

    expect(screen.getByText('onboardingMgmt.checklist.attachmentRequired')).toBeInTheDocument();
    expect(mocks.completeMutate).not.toHaveBeenCalled();
  });
});
