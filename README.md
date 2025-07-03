# Vbitez

A modern React-based web application built with Vite, TypeScript, and Firebase. This project includes features for restaurant management, QR code generation, and user dashboards.

## 🚀 Quick Start

### Prerequisites

- **Node.js** (v18 or higher)
- **pnpm** (recommended) or npm/yarn
- **Git**

### Installation

1. **Clone the repository**

   ```bash
   git clone <your-repo-url>
   cd vbitez
   ```

2. **Install dependencies**

   ```bash
   pnpm install
   ```

3. **Set up environment variables**
   Create a `.env` file in the root directory with your configuration:
   ```env
   VITE_API_URL=your_api_url
   VITE_FIREBASE_API_KEY=your_firebase_api_key
   VITE_FIREBASE_AUTH_DOMAIN=your_firebase_auth_domain
   VITE_FIREBASE_PROJECT_ID=your_firebase_project_id
   ```

## 🛠️ Development

### Running the Development Server

```bash
pnpm start
```

The application will be available at `http://localhost:5173`

### Available Scripts

- `pnpm start` - Start development server
- `pnpm build` - Build for production
- `pnpm preview` - Preview production build locally
- `pnpm lint` - Run ESLint
- `pnpm deploy` - Build and deploy to Firebase

### Project Structure

```
src/
├── api/           # API configurations and routes
├── components/    # Reusable React components
├── hooks/         # Custom React hooks
├── layouts/       # Layout components (Admin, Client, Vendor)
├── redux/         # Redux store and slices
├── routes/        # Route configurations
├── screens/       # Page components
├── services/      # External services (i18n, etc.)
├── styles/        # Global styles and SCSS
└── utils/         # Utility functions
```

## 🚀 Deployment

### Automatic Deployment (Recommended)

This project uses GitHub Actions for automatic deployment. The workflow is configured to deploy only when the version in `package.json` is updated.

#### How to Deploy:

1. **Update the version** in `package.json`:

   ```json
   {
     "version": "0.0.10"
   }
   ```

2. **Commit and push** your changes:

   ```bash
   git add package.json
   git commit -m "Bump version to 0.0.10"
   git push origin main
   ```

3. **The deployment will automatically trigger** when:
   - Changes are pushed to the `main` branch
   - The `package.json` file is modified
   - The version number has changed from the previous commit

#### Deployment Process:

The GitHub Actions workflow will:

1. Set up Node.js and pnpm
2. Check if the version has changed
3. Install dependencies
4. Build the project
5. Deploy to HostGator via FTP

### Manual Deployment

If you need to deploy manually:

```bash
# Build the project
pnpm build

# Deploy to Firebase (if using Firebase hosting)
pnpm deploy

# Or manually upload the dist/ folder to your hosting provider
```

## 🔧 Configuration

### Firebase Configuration

The project is configured for Firebase hosting. Update `firebase.json` if needed:

```json
{
  "hosting": {
    "public": "dist",
    "ignore": ["firebase.json", "**/.*", "**/node_modules/**"],
    "rewrites": [
      {
        "source": "**",
        "destination": "/index.html"
      }
    ]
  }
}
```

### Vite Configuration

The project uses Vite with custom aliases and optimizations. Key configurations in `vite.config.ts`:

- **Aliases**: `@app`, `@components`, `@hooks`, etc.
- **SCSS**: Global styles integration
- **Build optimization**: Code splitting and chunk optimization
- **Development server**: Port 5173 with HMR

## 📦 Dependencies

### Key Technologies

- **React 18** - UI framework
- **TypeScript** - Type safety
- **Vite** - Build tool and dev server
- **Redux Toolkit** - State management
- **React Router** - Client-side routing
- **Material-UI** - UI components
- **SCSS** - Styling
- **Firebase** - Backend services

### Package Manager

This project uses **pnpm** as the package manager. If you prefer npm or yarn:

```bash
# Using npm
npm install
npm start

# Using yarn
yarn install
yarn start
```

## 🐛 Troubleshooting

### Common Issues

1. **Port already in use**

   - Change the port in `vite.config.ts` or kill the process using port 5173

2. **Build errors**

   - Clear node_modules and reinstall: `rm -rf node_modules && pnpm install`

3. **Deployment fails**
   - Check GitHub Actions logs for specific error messages
   - Verify environment variables are set correctly

### Getting Help

- Check the GitHub Actions logs for deployment issues
- Review the console for development errors
- Ensure all environment variables are properly configured

## 📝 Version Management

The project uses semantic versioning. To update the version:

1. Edit `package.json` and update the version field
2. Commit the change
3. Push to trigger automatic deployment

Example version formats:

- `0.0.9` - Patch version
- `0.1.0` - Minor version
- `1.0.0` - Major version

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

---

**Note**: This project is configured for automatic deployment via GitHub Actions. Only version changes in `package.json` will trigger a deployment to production.
