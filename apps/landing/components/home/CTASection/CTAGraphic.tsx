import React, { useState } from 'react';
import {
  ShoppingCart,
  Package,
  Users,
  Heart,
  DollarSign,
  Workflow,
  BarChart2,
  Share2,
} from 'lucide-react';
import styles from './CTAGraphic.module.scss';

interface OrbitNode {
  id: string;
  label: string;
  Icon: React.ComponentType<{ size?: number; strokeWidth?: number }>;
}

// The suite capabilities orbit the shared data model - deliberately not
// POS-centric, since the whole suite is the point, not any one workflow.
const NODES: OrbitNode[] = [
  { id: 'pos', label: 'POS', Icon: ShoppingCart },
  { id: 'erp', label: 'ERP', Icon: Package },
  { id: 'hr', label: 'HR', Icon: Users },
  { id: 'crm', label: 'CRM', Icon: Heart },
  { id: 'finance', label: 'Finance', Icon: DollarSign },
  { id: 'flow', label: 'Flow', Icon: Workflow },
  { id: 'insights', label: 'Insights', Icon: BarChart2 },
];

const RADIUS = 38;
const CENTER = 50;

function nodePosition(index: number, total: number): { x: number; y: number } {
  const angle = (index / total) * 2 * Math.PI - Math.PI / 2;
  return {
    x: CENTER + RADIUS * Math.cos(angle),
    y: CENTER + RADIUS * Math.sin(angle),
  };
}

const CTAGraphic: React.FC = (): React.JSX.Element => {
  const [hovered, setHovered] = useState<number | null>(null);

  return (
    <div className={styles.graphic}>
      <svg
        className={styles.lines}
        viewBox="0 0 100 100"
        preserveAspectRatio="xMidYMid meet"
        aria-hidden="true"
      >
        {NODES.map((node, i) => {
          const { x, y } = nodePosition(i, NODES.length);
          return (
            <line
              key={node.id}
              x1={CENTER}
              y1={CENTER}
              x2={x}
              y2={y}
              className={hovered === i ? styles.lineActive : styles.line}
            />
          );
        })}
        <circle cx={CENTER} cy={CENTER} r={9} className={styles.hubRing} />
      </svg>

      <div className={styles.hub}>
        <Share2 size={20} strokeWidth={1.6} />
      </div>

      {NODES.map((node, i) => {
        const { x, y } = nodePosition(i, NODES.length);
        const { Icon } = node;
        return (
          <button
            type="button"
            key={node.id}
            className={node.id === 'pos' ? `${styles.node} ${styles.nodePos}` : styles.node}
            style={{ left: `${x}%`, top: `${y}%` }}
            onMouseEnter={() => setHovered(i)}
            onMouseLeave={() => setHovered(null)}
            onFocus={() => setHovered(i)}
            onBlur={() => setHovered(null)}
          >
            <span className={hovered === i ? styles.nodeCircleActive : styles.nodeCircle}>
              <Icon size={18} strokeWidth={1.6} />
            </span>
            <span className={hovered === i ? styles.nodeLabelActive : styles.nodeLabel}>
              {node.label}
            </span>
          </button>
        );
      })}
    </div>
  );
};

export default CTAGraphic;
