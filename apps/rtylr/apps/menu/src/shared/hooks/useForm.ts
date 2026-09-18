import { useState, useCallback } from 'react';

type Errors<T> = Partial<Record<keyof T, string>>;
type Validator<T> = (values: T) => Errors<T>;

interface UseFormReturn<T> {
  values: T;
  errors: Errors<T>;
  touched: Partial<Record<keyof T, boolean>>;
  handleChange: (field: keyof T, value: T[keyof T]) => void;
  handleBlur: (field: keyof T) => void;
  handleSubmit: (onSubmit: (values: T) => void) => void;
  reset: () => void;
  isValid: boolean;
}

/**
 * Lightweight form state manager with validation.
 * Tracks values, touched fields, and per-field error messages.
 */
export function useForm<T extends object>(
  initialValues: T,
  validate?: Validator<T>
): UseFormReturn<T> {
  const [values, setValues] = useState<T>(initialValues);
  const [errors, setErrors] = useState<Errors<T>>({});
  const [touched, setTouched] = useState<Partial<Record<keyof T, boolean>>>({});

  const runValidation = useCallback(
    (current: T): Errors<T> => {
      return validate ? validate(current) : {};
    },
    [validate]
  );

  const handleChange = useCallback(
    (field: keyof T, value: T[keyof T]) => {
      setValues((prev) => {
        const next = { ...prev, [field]: value };
        setErrors(runValidation(next));
        return next;
      });
    },
    [runValidation]
  );

  const handleBlur = useCallback((field: keyof T) => {
    setTouched((prev) => ({ ...prev, [field]: true }));
  }, []);

  const handleSubmit = useCallback(
    (onSubmit: (values: T) => void) => {
      const errs = runValidation(values);
      const allTouched = Object.keys(values).reduce(
        (acc, k) => ({ ...acc, [k]: true }),
        {} as Partial<Record<keyof T, boolean>>
      );
      setTouched(allTouched);
      setErrors(errs);
      if (Object.keys(errs).length === 0) {
        onSubmit(values);
      }
    },
    [values, runValidation]
  );

  const reset = useCallback(() => {
    setValues(initialValues);
    setErrors({});
    setTouched({});
  }, [initialValues]);

  const isValid = Object.keys(runValidation(values)).length === 0;

  return { values, errors, touched, handleChange, handleBlur, handleSubmit, reset, isValid };
}
