export type CaretModel =
  | 'auto'
  | 'gpt-4o-mini'
  | 'gpt-4o'
  | 'gemini-2.5-flash'
  | 'gemini-2.5-pro'
  | 'claude-haiku-4-5'
  | 'claude-sonnet-4-6'
  | 'claude-opus-4-7';

export interface CaretModelDef {
  id: CaretModel;
  label: string;
  provider: 'openai' | 'gemini' | 'anthropic' | 'auto';
  description: string;
  contextWindow: number;
}

export const CARET_MODELS: Record<CaretModel, CaretModelDef> = {
  auto: {
    id: 'auto',
    label: 'Auto',
    provider: 'auto',
    description: 'Pick the best available provider for your account.',
    contextWindow: 0,
  },
  'gpt-4o-mini': {
    id: 'gpt-4o-mini',
    label: 'GPT-4o mini',
    provider: 'openai',
    description: 'Fast and affordable. Good for everyday questions.',
    contextWindow: 128_000,
  },
  'gpt-4o': {
    id: 'gpt-4o',
    label: 'GPT-4o',
    provider: 'openai',
    description: 'Stronger reasoning, slower.',
    contextWindow: 128_000,
  },
  'gemini-2.5-flash': {
    id: 'gemini-2.5-flash',
    label: 'Gemini 2.5 Flash',
    provider: 'gemini',
    description: 'Fast Google model with long context.',
    contextWindow: 1_000_000,
  },
  'gemini-2.5-pro': {
    id: 'gemini-2.5-pro',
    label: 'Gemini 2.5 Pro',
    provider: 'gemini',
    description: "Google's frontier model.",
    contextWindow: 1_000_000,
  },
  'claude-haiku-4-5': {
    id: 'claude-haiku-4-5',
    label: 'Claude Haiku 4.5',
    provider: 'anthropic',
    description: 'Anthropic small model. Fast.',
    contextWindow: 200_000,
  },
  'claude-sonnet-4-6': {
    id: 'claude-sonnet-4-6',
    label: 'Claude Sonnet 4.6',
    provider: 'anthropic',
    description: 'Balanced Anthropic model.',
    contextWindow: 200_000,
  },
  'claude-opus-4-7': {
    id: 'claude-opus-4-7',
    label: 'Claude Opus 4.7',
    provider: 'anthropic',
    description: "Anthropic's frontier model.",
    contextWindow: 200_000,
  },
};

export const DEFAULT_MODEL: CaretModel = 'auto';
