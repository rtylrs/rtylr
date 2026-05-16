import React from 'react';
import { visualForTool } from '../../../shared/lib/caret/toolIcons';

export interface ToolBadgeProps {
  toolName: string;
}

/** Small pill showing the module + icon for a tool name. */
export const ToolBadge: React.FC<ToolBadgeProps> = ({ toolName }) => {
  const v = visualForTool(toolName);
  const Icon = v.icon;
  return (
    <span
      style={{
        display: 'inline-flex',
        alignItems: 'center',
        gap: 4,
        padding: '2px 6px',
        background: v.bg,
        color: v.color,
        borderRadius: 6,
        fontSize: 11,
        fontWeight: 500,
        fontFamily: 'SFMono-Regular, Menlo, Consolas, monospace',
      }}
    >
      <Icon size={11} />
      <span>{toolName}</span>
    </span>
  );
};

export default ToolBadge;
