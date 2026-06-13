import React from 'react';
import type { LucideIcon } from 'lucide-react';
import { CakeSlice, CupSoda, Ellipsis, Salad, Utensils } from 'lucide-react';

export type Course = 'appetizer' | 'main' | 'dessert' | 'drink' | 'other';

interface CourseSelectorProps {
  value: Course;
  onChange: (c: Course) => void;
}

const COURSES: { id: Course; label: string; icon: LucideIcon }[] = [
  { id: 'appetizer', label: 'Appetizer', icon: Salad },
  { id: 'main', label: 'Main', icon: Utensils },
  { id: 'dessert', label: 'Dessert', icon: CakeSlice },
  { id: 'drink', label: 'Drink', icon: CupSoda },
  { id: 'other', label: 'Other', icon: Ellipsis },
];

const CourseSelector: React.FC<CourseSelectorProps> = ({ value, onChange }) => (
  <div className="grid grid-cols-5 gap-2">
    {COURSES.map((c) => {
      const Icon = c.icon;
      return (
        <button
          key={c.id}
          type="button"
          onClick={() => onChange(c.id)}
          className={`flex flex-col items-center gap-1 p-2 rounded-lg text-xs font-semibold transition min-h-[4rem] ${
            value === c.id
              ? 'bg-pos-text text-pos-bg'
              : 'bg-white/8 text-pos-text hover:bg-white/14'
          }`}
        >
          <Icon className="h-5 w-5" aria-hidden="true" strokeWidth={2} />
          {c.label}
        </button>
      );
    })}
  </div>
);

export default CourseSelector;
