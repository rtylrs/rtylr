import React from 'react';
import {
  Card as MuiCard,
  CardContent,
  CardHeader,
  CardActions,
} from '@mui/material';
import classNames from 'classnames';
import { CardProps } from '../../types/common';
import './Card.scss';

const Card: React.FC<CardProps> = ({
  elevation = 1,
  padding = 'medium',
  header,
  footer,
  className,
  style,
  children,
  ...props
}) => {
  const cardClasses = classNames(
    'voxire-card',
    `voxire-card--elevation-${elevation}`,
    `voxire-card--padding-${padding}`,
    className
  );

  return (
    <MuiCard
      elevation={elevation}
      className={cardClasses}
      style={style}
      {...props}
    >
      {header && <CardHeader className="voxire-card__header" title={header} />}

      <CardContent className="voxire-card__content">{children}</CardContent>

      {footer && (
        <CardActions className="voxire-card__footer">{footer}</CardActions>
      )}
    </MuiCard>
  );
};

export default Card;
