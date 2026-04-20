import React from 'react';

import MinimalIndex from './minimal/Index';
import MinimalJob from './minimal/Job';
import EditorialIndex from './editorial/Index';
import EditorialJob from './editorial/Job';
import BoldIndex from './bold/Index';
import BoldJob from './bold/Job';
import type { Customization, RecruitCareer } from '../shared/lib/types';

// The template a business picks in their Customization record resolves
// to a component tree here. Each template implements the same two
// entry points: an Index (careers list) and a Job (single role).
// Adding a new template means dropping a folder under templates/ with
// those two components and adding a case here.

export interface IndexProps {
  careers: RecruitCareer[];
  customization: Customization;
}

export interface JobProps {
  career: RecruitCareer;
  customization: Customization;
  slug: string;
}

const INDEX: Record<string, React.FC<IndexProps>> = {
  minimal: MinimalIndex,
  editorial: EditorialIndex,
  bold: BoldIndex,
};

const JOB: Record<string, React.FC<JobProps>> = {
  minimal: MinimalJob,
  editorial: EditorialJob,
  bold: BoldJob,
};

export function resolveIndex(template: string): React.FC<IndexProps> {
  return INDEX[template] ?? MinimalIndex;
}

export function resolveJob(template: string): React.FC<JobProps> {
  return JOB[template] ?? MinimalJob;
}

// applyBrand applies the brand color to the CSS custom property that
// every template reads. Called once at the top of each public page so
// the rest of the tree doesn't have to think about it.
export function applyBrand(primaryColor: string | undefined): void {
  if (typeof document === 'undefined') return;
  document.documentElement.style.setProperty(
    '--recruit-accent',
    primaryColor && primaryColor.trim() !== '' ? primaryColor : '#111111'
  );
}
