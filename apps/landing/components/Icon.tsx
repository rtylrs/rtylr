import React, { CSSProperties, useEffect, useState } from 'react';
import { LoadingSpinner } from '@voxire/components';

interface SvgIconProps {
  icon: string;
  className?: string;
  color?: string;
  size?: string | number;
  style?: CSSProperties;
  alt?: string;
  title?: string;
  titleId?: string;
}

const SvgIcon: React.FC<SvgIconProps> = ({
  icon,
  className,
  color,
  size = '1em',
  style,
  alt,
  title,
  titleId,
  ...props
}) => {
  const [svgContent, setSvgContent] = useState<string | null>(null);

  useEffect(() => {
    fetch(`/icons/${icon}.svg`)
      .then((response) => response.text())
      .then(setSvgContent)
      .catch(() => {});
  }, [icon]);

  if (!svgContent) {
    return <LoadingSpinner size="small" color="primary" />;
  }

  return (
    <span
      className={className}
      style={{ color, width: size, height: size, ...style }}
      title={title}
      aria-labelledby={titleId}
      dangerouslySetInnerHTML={{ __html: svgContent }}
      {...props}
    />
  );
};

export default SvgIcon;
