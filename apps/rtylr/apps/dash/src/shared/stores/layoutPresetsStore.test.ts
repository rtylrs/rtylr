import { describe, it, expect, beforeEach } from 'vitest';
import { useLayoutPresetsStore } from './layoutPresetsStore';
import { DEFAULT_LAYOUT_CONFIG } from '@voxire/menu';

const { getState } = useLayoutPresetsStore;

beforeEach(() => {
  useLayoutPresetsStore.setState({
    presets: [
      {
        id: 'default',
        name: 'Default',
        config: DEFAULT_LAYOUT_CONFIG,
        createdAt: Date.now(),
      },
    ],
    activePresetId: 'default',
    draftConfig: DEFAULT_LAYOUT_CONFIG,
  });
});

describe('layoutPresetsStore', () => {
  it('has default preset initially', () => {
    expect(getState().presets).toHaveLength(1);
    expect(getState().presets[0].id).toBe('default');
    expect(getState().presets[0].name).toBe('Default');
  });

  it('has default as activePresetId', () => {
    expect(getState().activePresetId).toBe('default');
  });

  it('draftConfig equals DEFAULT_LAYOUT_CONFIG initially', () => {
    expect(getState().draftConfig).toEqual(DEFAULT_LAYOUT_CONFIG);
  });

  it('createPreset adds a new preset', () => {
    const id = getState().createPreset('Custom');
    expect(getState().presets).toHaveLength(2);
    const created = getState().presets.find((p) => p.id === id);
    expect(created).toBeDefined();
    expect(created!.name).toBe('Custom');
  });

  it('createPreset returns the new id', () => {
    const id = getState().createPreset('Test');
    expect(id).toMatch(/^preset_/);
  });

  it('createPreset uses draftConfig when no config provided', () => {
    const customConfig = {
      ...DEFAULT_LAYOUT_CONFIG,
      items: { ...DEFAULT_LAYOUT_CONFIG.items, columns: 4 },
    };
    getState().setDraftConfig(customConfig);
    const id = getState().createPreset('FromDraft');
    const preset = getState().presets.find((p) => p.id === id);
    expect(preset!.config.items.columns).toBe(4);
  });

  it('createPreset uses provided config', () => {
    const customConfig = {
      ...DEFAULT_LAYOUT_CONFIG,
      items: { ...DEFAULT_LAYOUT_CONFIG.items, columns: 3 },
    };
    const id = getState().createPreset('WithConfig', customConfig);
    const preset = getState().presets.find((p) => p.id === id);
    expect(preset!.config.items.columns).toBe(3);
  });

  it('updatePreset updates name', () => {
    getState().updatePreset('default', { name: 'Renamed' });
    expect(getState().presets[0].name).toBe('Renamed');
  });

  it('updatePreset updates config', () => {
    const newConfig = {
      ...DEFAULT_LAYOUT_CONFIG,
      items: { ...DEFAULT_LAYOUT_CONFIG.items, columns: 1 },
    };
    getState().updatePreset('default', { config: newConfig });
    expect(getState().presets[0].config.items.columns).toBe(1);
  });

  it('deletePreset removes a preset', () => {
    const id = getState().createPreset('ToDelete');
    getState().deletePreset(id);
    expect(getState().presets.find((p) => p.id === id)).toBeUndefined();
  });

  it('deletePreset of active switches to first remaining', () => {
    const id = getState().createPreset('Second');
    getState().activatePreset(id);
    getState().deletePreset(id);
    expect(getState().activePresetId).toBe('default');
  });

  it('activatePreset sets activePresetId and draftConfig', () => {
    const customConfig = {
      ...DEFAULT_LAYOUT_CONFIG,
      items: { ...DEFAULT_LAYOUT_CONFIG.items, columns: 5 },
    };
    const id = getState().createPreset('Activate', customConfig);
    getState().activatePreset(id);

    expect(getState().activePresetId).toBe(id);
    expect(getState().draftConfig.items.columns).toBe(5);
  });

  it('activatePreset with non-existent id does nothing', () => {
    const before = getState().activePresetId;
    getState().activatePreset('nonexistent');
    expect(getState().activePresetId).toBe(before);
  });

  it('setDraftConfig updates draft', () => {
    const newConfig = {
      ...DEFAULT_LAYOUT_CONFIG,
      items: { ...DEFAULT_LAYOUT_CONFIG.items, layout: 'list' as const },
    };
    getState().setDraftConfig(newConfig);
    expect(getState().draftConfig.items.layout).toBe('list');
  });

  it('patchDraftConfig merges partial config', () => {
    getState().patchDraftConfig({ items: { ...DEFAULT_LAYOUT_CONFIG.items, columns: 6 } });
    expect(getState().draftConfig.items.columns).toBe(6);
    // Other fields preserved
    expect(getState().draftConfig.header).toEqual(DEFAULT_LAYOUT_CONFIG.header);
  });

  it('saveActivePreset persists draft to active preset', () => {
    const newConfig = {
      ...DEFAULT_LAYOUT_CONFIG,
      items: { ...DEFAULT_LAYOUT_CONFIG.items, columns: 9 },
    };
    getState().setDraftConfig(newConfig);
    getState().saveActivePreset();

    const defaultPreset = getState().presets.find((p) => p.id === 'default');
    expect(defaultPreset!.config.items.columns).toBe(9);
  });

  it('saveActivePreset does nothing when no activePresetId', () => {
    useLayoutPresetsStore.setState({ activePresetId: null });
    const before = [...getState().presets];
    getState().saveActivePreset();
    expect(getState().presets).toEqual(before);
  });
});
