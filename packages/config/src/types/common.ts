export interface BaseComponentProps {
  className?: string;
  style?: React.CSSProperties;
  children?: React.ReactNode;
}

export interface ButtonProps extends Omit<
  React.ButtonHTMLAttributes<HTMLButtonElement>,
  'onClick' | 'prefix' | 'suffix'
> {
  children: React.ReactNode;
  onClick?: (event: React.MouseEvent<HTMLButtonElement>) => void;
  variant?: 'primary' | 'secondary' | 'outline' | 'ghost';
  size?: 'sm' | 'md' | 'lg';
  color?: string;
  secondaryColor?: string;
  hidden?: boolean;
  loading?: boolean;
  fullWidth?: boolean;
  styles?: React.CSSProperties;
  prefix?: React.ReactNode;
  suffix?: React.ReactNode;
}

export interface InputProps extends BaseComponentProps {
  type?: 'text' | 'email' | 'password' | 'number' | 'tel' | 'url';
  placeholder?: string;
  value?: string;
  defaultValue?: string;
  disabled?: boolean;
  required?: boolean;
  error?: boolean;
  helperText?: string;
  onChange?: (event: React.ChangeEvent<HTMLInputElement>) => void;
  onBlur?: (event: React.FocusEvent<HTMLInputElement>) => void;
  onFocus?: (event: React.FocusEvent<HTMLInputElement>) => void;
  id?: string;
  name?: string;
  autoComplete?: string;
  label?: string;
  hasLabel?: boolean;
  labelProps?: React.LabelHTMLAttributes<HTMLLabelElement>;
  autoFocus?: boolean;
}

export interface ModalProps extends BaseComponentProps {
  open: boolean;
  onClose: () => void;
  title?: string;
  maxWidth?: 'xs' | 'sm' | 'md' | 'lg' | 'xl';
  fullWidth?: boolean;
}

export interface CardProps extends BaseComponentProps {
  elevation?: number;
  padding?: 'none' | 'small' | 'medium' | 'large';
  header?: React.ReactNode;
  footer?: React.ReactNode;
}

export interface LoadingSpinnerProps extends BaseComponentProps {
  size?: 'small' | 'medium' | 'large';
  color?: 'primary' | 'secondary' | 'inherit';
}

export interface AnimatedRoutesProps extends BaseComponentProps {
  animationType?: 'slide' | 'fade' | 'scale' | 'slideUp' | 'slideDown' | 'none';
}

export interface SelectOption {
  id: string | number;
  title: string;
  description?: string;
  value: string;
}

export interface SelectProps extends BaseComponentProps {
  options: SelectOption[];
  value?: string;
  placeholder?: string;
  disabled?: boolean;
  error?: boolean;
  helperText?: string;
  onChange?: (value: string, option: SelectOption) => void;
  onBlur?: () => void;
  onFocus?: () => void;
  id?: string;
  name?: string;
  label?: string;
  hasLabel?: boolean;
  labelProps?: React.LabelHTMLAttributes<HTMLLabelElement>;
  required?: boolean;
}
