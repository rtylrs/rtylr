import { fireEvent, screen } from '@testing-library/react';

const escape = (value: string): string => value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');

/**
 * The trigger of a FlowSelect. Its accessible name is `<label>: <current value>`,
 * so match on the label prefix and let callers stay independent of the value.
 */
export function flowSelectTrigger(label: string): HTMLElement {
  return screen.getByRole('button', { name: new RegExp(`^${escape(label)}(:|$)`) });
}

/** Opens a FlowSelect and picks an option by its visible label. */
export async function chooseFlowOption(label: string, option: string | RegExp): Promise<void> {
  // Radix opens a dropdown on pointerdown, not click.
  fireEvent.pointerDown(flowSelectTrigger(label), { button: 0, ctrlKey: false });
  fireEvent.click(await screen.findByRole('menuitem', { name: option }));
}
