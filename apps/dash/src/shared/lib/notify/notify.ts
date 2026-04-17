// Sonner.js

import { toast } from 'sonner';

const message = (message: string) => {
  toast(message);
};

const success = (message: string) => {
  toast.success(message);
};

const error = (message: string) => {
  toast.error(message);
};

const info = (message: string) => {
  toast.info(message);
};

const warning = (message: string) => {
  toast.warning(message);
};

const description = (message: string, description: string) => {
  toast.message(message, {
    description: description,
  });
};

const action = (message: string, action: () => void, label: string) => {
  toast.message(message, {
    action: {
      label: label,
      onClick: action,
    },
  });
};

const notify = {
  success,
  error,
  info,
  warning,
  description,
  action,
  message,
};

export default notify;
