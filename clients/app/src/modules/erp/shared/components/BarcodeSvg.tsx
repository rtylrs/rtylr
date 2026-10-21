import React from 'react';

interface BarcodeSvgProps {
  value: string;
  width?: number;
  height?: number;
  showText?: boolean;
}

// Simple Code-128-ish visual barcode (not scannable but good for print layout previews)
const BarcodeSvg: React.FC<BarcodeSvgProps> = ({
  value,
  width = 180,
  height = 60,
  showText = true,
}) => {
  const bars: number[] = [];
  for (let i = 0; i < value.length; i++) {
    const code = value.charCodeAt(i);
    bars.push(1 + ((code >> 0) & 1));
    bars.push(1 + ((code >> 1) & 1));
    bars.push(2 + ((code >> 2) & 1));
    bars.push(1 + ((code >> 3) & 1));
    bars.push(1 + ((code >> 4) & 1));
    bars.push(2 + ((code >> 5) & 1));
  }
  const totalUnits = bars.reduce((a, b) => a + b, 0);
  const unitWidth = width / totalUnits;

  let x = 0;
  const rects: React.ReactElement[] = [];
  bars.forEach((b, i) => {
    if (i % 2 === 0) {
      rects.push(
        <rect
          key={i}
          x={x}
          y={0}
          width={b * unitWidth}
          height={height - (showText ? 12 : 0)}
          fill="#000"
        />
      );
    }
    x += b * unitWidth;
  });

  return (
    <svg width={width} height={height} style={{ background: '#fff' }}>
      {rects}
      {showText && (
        <text
          x={width / 2}
          y={height - 2}
          textAnchor="middle"
          fontSize="10"
          fontFamily="monospace"
          fill="#000"
        >
          {value}
        </text>
      )}
    </svg>
  );
};

export default BarcodeSvg;
