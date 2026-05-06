import React from 'react';

export type Course = 'appetizer' | 'main' | 'dessert' | 'drink' | 'other';

interface CourseSelectorProps {
  value: Course;
  onChange: (c: Course) => void;
}

const COURSES: { id: Course; label: string; icon: string }[] = [
  { id: 'appetizer', label: 'Appetizer', icon: '🥗' },
  { id: 'main', label: 'Main', icon: '🍽' },
  { id: 'dessert', label: 'Dessert', icon: '🍰' },
  { id: 'drink', label: 'Drink', icon: '🥤' },
  { id: 'other', label: 'Other', icon: '⋯' },
];

const CourseSelector: React.FC<CourseSelectorProps> = ({ value, onChange }) => (
  <div className="grid grid-cols-5 gap-2">
    {COURSES.map((c) => (
      <button
        key={c.id}
        type="button"
        onClick={() => onChange(c.id)}
        className={`flex flex-col items-center gap-1 p-2 rounded-lg text-xs font-semibold transition min-h-[4rem] ${
          value === c.id ? 'bg-pos-text text-pos-bg' : 'bg-white/8 text-pos-text hover:bg-white/14'
        }`}
      >
        <span className="text-lg">{c.icon}</span>
        {c.label}
      </button>
    ))}
  </div>
);

export default CourseSelector;
