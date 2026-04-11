export type LayoutType = 'default' | 'no-sidebar' | 'custom-sidebar';

export interface LayoutConfig {
  type: LayoutType;
  showNavbar?: boolean;
}
