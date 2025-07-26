# @voxire/utils

A utility library for Voxire applications, providing common helper functions for validation, formatting, and general utilities.

## Features

- 🔍 **Validation**: Email, phone, password, and form validation
- 📅 **Formatting**: Date, number, currency, and text formatting
- 🛠️ **Helpers**: Common utility functions for everyday development
- 🚀 **TypeScript Support**: Full TypeScript support with proper type definitions
- 📦 **Tree Shaking**: Optimized bundle size with tree shaking support

## Installation

```bash
npm install @voxire/utils
# or
yarn add @voxire/utils
# or
pnpm add @voxire/utils
```

## Usage

### Basic Import

```tsx
import { validation, formatting, helpers } from '@voxire/utils';

// Or import specific functions
import { validation } from '@voxire/utils';
```

### Validation

```tsx
import { validation } from '@voxire/utils';

// Email validation
const isValidEmail = validation.email('user@example.com'); // true

// Phone validation
const isValidPhone = validation.phone('+1234567890'); // true

// Password validation
const passwordCheck = validation.password('MyPassword123!');
console.log(passwordCheck.isValid); // true/false
console.log(passwordCheck.errors); // array of error messages

// Required field validation
const isRequired = validation.required('some value'); // true

// Length validation
const isMinLength = validation.minLength('hello', 3); // true
const isMaxLength = validation.maxLength('hello', 10); // true

// URL validation
const isValidUrl = validation.url('https://example.com'); // true

// Numeric validation
const isNumeric = validation.numeric('12345'); // true
const isAlphanumeric = validation.alphanumeric('abc123'); // true
```

### Formatting

```tsx
import { formatting } from '@voxire/utils';

// Date formatting
const formattedDate = formatting.formatDate(new Date(), 'MMM dd, yyyy'); // "Jan 15, 2024"
const relativeDate = formatting.formatRelativeDate(new Date()); // "today at 2:30 PM"
const distanceDate = formatting.formatDistance(new Date()); // "2 hours ago"

// Number formatting
const currency = formatting.formatCurrency(1234.56, 'USD'); // "$1,234.56"
const number = formatting.formatNumber(1234567); // "1,234,567"
const percentage = formatting.formatPercentage(0.1234); // "12.34%"

// Text formatting
const capitalized = formatting.capitalize('hello world'); // "Hello world"
const capitalizedWords = formatting.capitalizeWords('hello world'); // "Hello World"
const truncated = formatting.truncate('This is a long text', 10); // "This is a..."
const slug = formatting.slugify('Hello World!'); // "hello-world"
```

### Helpers

```tsx
import { helpers } from '@voxire/utils';

// Debounce function calls
const debouncedSearch = helpers.debounce((query: string) => {
  // Perform search
}, 300);

// Throttle function calls
const throttledScroll = helpers.throttle(() => {
  // Handle scroll
}, 100);

// Check if value is empty
const isEmpty = helpers.isEmpty(''); // true
const isEmptyObj = helpers.isEmpty({}); // true

// Deep equality check
const isEqual = helpers.isEqual({ a: 1 }, { a: 1 }); // true

// Generate random ID
const id = helpers.generateId(10); // "aB3x9mK2pQ"

// Deep clone object
const original = { a: 1, b: { c: 2 } };
const cloned = helpers.deepClone(original);

// Retry function with exponential backoff
const result = await helpers.retry(
  async () => {
    // Some async operation that might fail
    return await fetchData();
  },
  3,
  1000
);

// Sleep utility
await helpers.sleep(1000); // Wait for 1 second

// Safe object property access
const value = helpers.get(obj, 'user.profile.name', 'default'); // Safe access with default
helpers.set(obj, 'user.profile.name', 'John'); // Safe property setting
```

## API Reference

### Validation

| Function                | Description                      | Returns                                  |
| ----------------------- | -------------------------------- | ---------------------------------------- |
| `email(email)`          | Validates email format           | `boolean`                                |
| `phone(phone)`          | Validates phone number format    | `boolean`                                |
| `password(password)`    | Validates password strength      | `{ isValid: boolean, errors: string[] }` |
| `required(value)`       | Checks if value is not empty     | `boolean`                                |
| `minLength(value, min)` | Checks minimum length            | `boolean`                                |
| `maxLength(value, max)` | Checks maximum length            | `boolean`                                |
| `url(url)`              | Validates URL format             | `boolean`                                |
| `numeric(value)`        | Checks if string is numeric      | `boolean`                                |
| `alphanumeric(value)`   | Checks if string is alphanumeric | `boolean`                                |

### Formatting

| Function                                   | Description                           | Returns  |
| ------------------------------------------ | ------------------------------------- | -------- |
| `formatDate(date, format)`                 | Formats date with custom format       | `string` |
| `formatRelativeDate(date)`                 | Formats date as relative time         | `string` |
| `formatDistance(date)`                     | Formats date as distance from now     | `string` |
| `formatCurrency(amount, currency, locale)` | Formats number as currency            | `string` |
| `formatNumber(number, locale)`             | Formats number with locale            | `string` |
| `formatPercentage(value, decimals)`        | Formats number as percentage          | `string` |
| `capitalize(text)`                         | Capitalizes first letter              | `string` |
| `capitalizeWords(text)`                    | Capitalizes first letter of each word | `string` |
| `truncate(text, maxLength, suffix)`        | Truncates text with suffix            | `string` |
| `slugify(text)`                            | Converts text to URL-friendly slug    | `string` |

### Helpers

| Function                        | Description                   | Returns         |
| ------------------------------- | ----------------------------- | --------------- |
| `debounce(func, wait)`          | Debounces function calls      | `Function`      |
| `throttle(func, wait)`          | Throttles function calls      | `Function`      |
| `isEmpty(value)`                | Checks if value is empty      | `boolean`       |
| `isEqual(value1, value2)`       | Deep equality check           | `boolean`       |
| `generateId(length)`            | Generates random ID           | `string`        |
| `deepClone(obj)`                | Deep clones object            | `T`             |
| `retry(fn, maxAttempts, delay)` | Retries function with backoff | `Promise<T>`    |
| `sleep(ms)`                     | Sleeps for specified time     | `Promise<void>` |
| `get(obj, path, defaultValue)`  | Safe object property access   | `any`           |
| `set(obj, path, value)`         | Safe object property setting  | `void`          |

## Development

### Building the Library

```bash
# Install dependencies
pnpm install

# Build the library
pnpm build

# Development mode with watch
pnpm dev

# Type checking
pnpm typecheck
```

### Project Structure

```
src/
├── validation/     # Validation utilities
├── formatting/     # Formatting utilities
├── helpers/        # Helper utilities
└── index.ts       # Main entry point
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

MIT
