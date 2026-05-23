import React from 'react';
import { AppIcon } from '../AppIcon';
import { WindowType } from '../../stores/useDesktopStore';

interface TitleBarProps {
  type: WindowType;
  title: string;
  onClose: () => void;
  onMinimize: () => void;
  onMaximize?: () => void;
}

export const TitleBar: React.FC<TitleBarProps> = ({
  type,
  title,
  onClose,
  onMinimize,
  onMaximize,
}) => {
  return (
    <div
      className="window-titlebar"
      style={{
        display: 'flex',
        alignItems: 'center',
        height: 38,
        padding: '0 12px',
        background: 'rgba(30,30,32,0.98)',
        borderBottom: '1px solid rgba(255,255,255,0.08)',
        borderRadius: '10px 10px 0 0',
        userSelect: 'none',
        cursor: 'grab',
        gap: 8,
        flexShrink: 0,
      }}
    >
      {/* Traffic lights */}
      <div style={{ display: 'flex', gap: 6, marginRight: 4 }}>
        <button
          onClick={(e) => {
            e.stopPropagation();
            onClose();
          }}
          style={{
            width: 12,
            height: 12,
            borderRadius: '50%',
            background: '#ff5f57',
            border: 'none',
            cursor: 'pointer',
            padding: 0,
          }}
          aria-label="Close"
        />
        <button
          onClick={(e) => {
            e.stopPropagation();
            onMinimize();
          }}
          style={{
            width: 12,
            height: 12,
            borderRadius: '50%',
            background: '#febc2e',
            border: 'none',
            cursor: 'pointer',
            padding: 0,
          }}
          aria-label="Minimize"
        />
        <button
          onClick={(e) => {
            e.stopPropagation();
            onMaximize?.();
          }}
          style={{
            width: 12,
            height: 12,
            borderRadius: '50%',
            background: '#28c840',
            border: 'none',
            cursor: 'pointer',
            padding: 0,
          }}
          aria-label="Maximize"
        />
      </div>

      <AppIcon type={type} size={13} color="rgba(255,255,255,0.55)" strokeWidth={1.75} />
      <span style={{ fontSize: 12, fontWeight: 500, color: 'rgba(255,255,255,0.7)', flex: 1 }}>
        {title}
      </span>
    </div>
  );
};
