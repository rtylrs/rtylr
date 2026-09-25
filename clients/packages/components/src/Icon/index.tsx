import React, { CSSProperties, useEffect, useState } from 'react';

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

/**
 * @param icon - Name of the icon file (without .svg extension)
 * @param className - Additional CSS classes for the icon container
 * @param color - Color of the icon (CSS color value)
 * @param size - Size of the icon in pixels or CSS units (default: 1em)
 * @param style - Additional inline styles for the icon container
 * @param alt - Alternative text for accessibility
 * @param title - Title attribute for accessibility and tooltips
 * @param titleId - ID for the title element (for accessibility)
 */
const Icon: React.FC<IconProps> = ({
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
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    let isMounted = true;
    setIsLoading(true);
    setSvgContent(null);

    // Sanitize icon name to prevent path traversal
    const safeIcon = icon.replace(/[^a-zA-Z0-9_-]/g, '');
    if (!safeIcon) {
      setIsLoading(false);
      return;
    }

    fetch(`/icons/${safeIcon}.svg`)
      .then((response) => {
        if (!response.ok) {
          throw new Error(`HTTP error ${response.status}`);
        }
        return response.text();
      })
      .then((content) => {
        if (!isMounted) return;
        // Basic check to ensure fetched content is actually an SVG element
        if (content.trim().toLowerCase().includes('<svg')) {
          setSvgContent(content);
        } else {
          console.error(`Loaded content for icon '${icon}' is not a valid SVG.`);
        }
        setIsLoading(false);
      })
      .catch((error) => {
        if (!isMounted) return;
        console.error('Failed to load icon:', error);
        setIsLoading(false);
      });

    return () => {
      isMounted = false;
    };
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
      style={{
        color,
        width: size,
        height: size,
        display: 'inline-block',
        ...style,
      }}
      title={title}
      aria-labelledby={titleId}
      dangerouslySetInnerHTML={{ __html: svgContent }}
      {...props}
    />
  );
};

export default Icon;
