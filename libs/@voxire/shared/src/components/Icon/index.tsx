import React, { CSSProperties, useEffect, useState } from "react";

interface IconProps {
  icon: string;
  className?: string;
  color?: string;
  size?: string | number;
  style?: CSSProperties;
  alt?: string;
  title?: string;
  titleId?: string;
}

const Icon: React.FC<IconProps> = ({
  icon,
  className,
  color,
  size = "1em",
  style,
  alt,
  title,
  titleId,
  ...props
}) => {
  const [svgContent, setSvgContent] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    setIsLoading(true);
    fetch(`/icons/${icon}.svg`)
      .then((response) => response.text())
      .then((content) => {
        setSvgContent(content);
        setIsLoading(false);
      })
      .catch((error) => {
        console.error('Failed to load icon:', error);
        setIsLoading(false);
      });
  }, [icon]);

  if (isLoading) {
    return (
      <div
        style={{
          width: size,
          height: size,
          display: 'inline-block',
          backgroundColor: '#f0f0f0',
          borderRadius: '50%',
          animation: 'pulse 1.5s ease-in-out infinite',
        }}
      />
    );
  }

  if (!svgContent) {
    return null;
  }

  return (
    <span
      className={className}
      style={{ color, width: size, height: size, display: 'inline-block', ...style }}
      title={title}
      aria-labelledby={titleId}
      dangerouslySetInnerHTML={{ __html: svgContent }}
      {...props}
    />
  );
};

export default Icon;
