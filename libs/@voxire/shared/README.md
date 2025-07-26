# @voxire/shared

A shared UI component library for Voxire applications, built with React, TypeScript, and Material-UI.

## Features

- 🎨 **Consistent Design System**: Pre-built components with consistent styling
- 🚀 **TypeScript Support**: Full TypeScript support with proper type definitions
- 🎯 **Material-UI Integration**: Built on top of Material-UI for reliability
- 📦 **Tree Shaking**: Optimized bundle size with tree shaking support
- 🎭 **Customizable**: Easy to customize with SCSS variables and theming
- 📱 **Responsive**: Mobile-first responsive design

## Installation

```bash
npm install @voxire/shared
# or
yarn add @voxire/shared
# or
pnpm add @voxire/shared
```

## Peer Dependencies

This library requires the following peer dependencies:

```json
{
  "react": ">=18.0.0",
  "react-dom": ">=18.0.0"
}
```

## Usage

### Basic Import

```tsx
import { Button, Card, Input, Modal, LoadingSpinner } from '@voxire/shared';

function App() {
  return (
    <div>
      <Button variant="contained" color="primary">
        Click me
      </Button>

      <Card header="My Card" padding="medium">
        <p>Card content goes here</p>
      </Card>

      <Input placeholder="Enter your name" />
    </div>
  );
}
```

### Components

#### Button

```tsx
import { Button } from '@voxire/shared';

// Basic usage
<Button variant="contained" color="primary">
  Primary Button
</Button>

// With loading state
<Button loading={true} disabled={true}>
  Loading...
</Button>

// Different variants
<Button variant="outlined" color="secondary">
  Outlined Button
</Button>

<Button variant="text" color="success">
  Text Button
</Button>
```

#### Card

```tsx
import { Card } from '@voxire/shared';

// Basic card
<Card>
  <p>Simple card content</p>
</Card>

// Card with header and footer
<Card
  header="Card Title"
  footer={<Button>Action</Button>}
  padding="large"
>
  <p>Card content with header and footer</p>
</Card>

// Card with custom elevation
<Card elevation={3}>
  <p>Card with higher elevation</p>
</Card>
```

#### Input

```tsx
import { Input } from '@voxire/shared';

// Basic input
<Input placeholder="Enter text" />

// Input with validation
<Input
  type="email"
  placeholder="Enter email"
  error={true}
  helperText="Please enter a valid email"
/>

// Required input
<Input
  placeholder="Required field"
  required={true}
/>
```

#### Modal

```tsx
import { Modal } from '@voxire/shared';

function MyComponent() {
  const [open, setOpen] = useState(false);

  return (
    <>
      <Button onClick={() => setOpen(true)}>Open Modal</Button>

      <Modal
        open={open}
        onClose={() => setOpen(false)}
        title="My Modal"
        maxWidth="md"
      >
        <p>Modal content goes here</p>
        <Button onClick={() => setOpen(false)}>Close</Button>
      </Modal>
    </>
  );
}
```

#### LoadingSpinner

```tsx
import { LoadingSpinner } from '@voxire/shared';

// Basic spinner
<LoadingSpinner />

// Spinner with text
<LoadingSpinner size="large">
  Loading data...
</LoadingSpinner>

// Different colors
<LoadingSpinner color="secondary" />
```

### Utilities

#### Colors

```tsx
import { getColor, getGreyColor } from '@voxire/shared';

const primaryColor = getColor('primary');
const grey500 = getGreyColor(500);
```

#### Validation

```tsx
import { validation } from '@voxire/shared';

// Email validation
const isValidEmail = validation.email('user@example.com');

// Password validation
const passwordCheck = validation.password('MyPassword123!');
console.log(passwordCheck.isValid); // true/false
console.log(passwordCheck.errors); // array of error messages

// Required field validation
const isRequired = validation.required('some value');
```

### Theming

The library includes a Material-UI theme that you can use in your application:

```tsx
import { ThemeProvider } from '@mui/material/styles';
import { theme } from '@voxire/shared';

function App() {
  return <ThemeProvider theme={theme}>{/* Your app content */}</ThemeProvider>;
}
```

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
├── components/          # React components
│   ├── Button/
│   ├── Card/
│   ├── Input/
│   ├── Modal/
│   └── LoadingSpinner/
├── styles/             # SCSS styles and variables
├── utils/              # Utility functions
├── types/              # TypeScript type definitions
└── index.ts           # Main entry point
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

MIT
