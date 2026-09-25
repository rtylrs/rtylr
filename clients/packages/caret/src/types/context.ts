export type CaretPageContextValue =
  | string
  | number
  | boolean
  | null
  | CaretPageContextValue[]
  | { [key: string]: CaretPageContextValue };

export interface CaretPageContext {
  app?: string;
  appName?: string;
  module?: string;
  route?: string;
  page?: string;
  businessName?: string;
  role?: string;
  selected?: Record<string, CaretPageContextValue>;
  filters?: Record<string, CaretPageContextValue>;
  visibleRows?: CaretPageContextValue[];
  businessMemory?: Record<string, CaretPageContextValue>;
  cacheContext?: Record<string, CaretPageContextValue>;
  dataFreshness?: CaretPageContextValue;
  hints?: string[];
  [key: string]: CaretPageContextValue | undefined;
}
