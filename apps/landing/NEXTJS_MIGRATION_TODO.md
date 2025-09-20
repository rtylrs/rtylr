# Next.js Migration Todo

## Migration Strategy
Migrate the Vite + React Router landing app to Next.js while preserving all functionality and improving performance with SSR/SSG capabilities.

## Pre-Migration Setup
- [x] Create backup copy as `landing-old`
- [x] Install Next.js dependencies
- [x] Set up Next.js configuration
- [x] Configure Nx to work with Next.js
- [x] Basic Next.js build and serve working

## Core Migration Steps

### 1. Project Structure & Configuration
- [ ] Replace `vite.config.ts` with `next.config.js`
- [ ] Update `project.json` for Nx Next.js integration
- [ ] Migrate path aliases from Vite to Next.js
- [ ] Set up SCSS configuration for Next.js
- [ ] Configure asset handling (public folder structure)

### 2. Routing Migration
- [ ] Remove React Router dependencies (`react-router-dom`)
- [ ] Convert route structure from `src/routes/index.tsx` to Next.js pages/app directory
- [ ] Migrate route components:
  - [ ] `/` (Home) → `pages/index.tsx` or `app/page.tsx`
  - [ ] `/about` → `pages/about.tsx` or `app/about/page.tsx`
  - [ ] `/pricing` → `pages/pricing.tsx` or `app/pricing/page.tsx`
  - [ ] `/contact` → `pages/contact.tsx` or `app/contact/page.tsx`
  - [ ] `/faqs` → `pages/faqs.tsx` or `app/faqs/page.tsx`
  - [ ] `/news/[newsId]` → `pages/news/[newsId].tsx` or `app/news/[newsId]/page.tsx`
- [ ] Update internal navigation from React Router `Link` to Next.js `Link`
- [ ] Handle 404 page (`NoMatch` component)

### 3. Layout System
- [ ] Convert `MainLayout` to Next.js layout pattern
- [ ] Implement `_app.tsx` (Pages Router) or root layout (App Router)
- [ ] Migrate layout providers and wrapper components
- [ ] Handle scroll restoration and top button functionality

### 4. State Management
- [x] ~~Adapt Redux store for SSR compatibility~~ **MIGRATED TO ZUSTAND + REACT QUERY**
- [x] Create Zustand stores (user, config, theme, menu)
- [x] Set up TanStack React Query for API calls
- [x] Update _app.tsx to use React Query provider
- [x] Remove Redux dependencies
- [ ] Fix API hooks for SSR compatibility
- [ ] Re-enable API calls in components

### 5. Styling & Assets
- [ ] Configure SCSS modules or global styles
- [ ] Migrate style imports and path references
- [ ] Update font loading strategy (next/font)
- [ ] Optimize image loading with next/image
- [ ] Configure SVG handling (next-svgr or similar)

### 6. API Integration
- [ ] Convert client-side API calls to work with SSR
- [ ] Implement API routes if needed (`pages/api/` or `app/api/`)
- [ ] Handle Firebase integration with SSR
- [ ] Update data fetching patterns (getStaticProps, getServerSideProps, or App Router equivalents)

### 7. Third-Party Libraries
- [ ] Verify Material-UI SSR compatibility
- [ ] Configure Framer Motion for Next.js
- [ ] Update i18next for Next.js (next-i18next)
- [ ] Migrate Chart.js setup
- [ ] Handle react-modal SSR issues
- [ ] Update other component libraries for SSR

### 8. Performance Optimizations
- [ ] Implement code splitting with Next.js dynamic imports
- [ ] Configure image optimization
- [ ] Set up font optimization
- [ ] Implement static generation where appropriate
- [ ] Configure caching strategies

### 9. Build & Development
- [ ] Update package.json scripts for Next.js
- [ ] Configure development server settings
- [ ] Set up production build configuration
- [ ] Update deployment configuration
- [ ] Configure environment variables handling

### 10. Testing & Validation
- [ ] Test all routes and navigation
- [ ] Verify SSR/hydration works correctly
- [ ] Test Redux state persistence
- [ ] Validate theme switching functionality
- [ ] Test internationalization
- [ ] Verify all components render correctly
- [ ] Test form submissions and API calls
- [ ] Validate SEO improvements (meta tags, structured data)

### 11. Cleanup
- [ ] Remove unused Vite dependencies
- [ ] Remove React Router dependencies
- [ ] Clean up unused configuration files
- [ ] Update documentation
- [ ] Remove old build artifacts

## Decision Points
- **Router Choice**: App Router (recommended) vs Pages Router
- **Data Fetching**: SSR, SSG, or ISR for different pages
- **Styling**: CSS Modules, styled-components, or global SCSS
- **Deployment**: Vercel, Netlify, or custom hosting

## Potential Challenges
- Redux persist hydration mismatches
- i18n routing and locale handling
- Third-party component SSR compatibility
- Asset path references in SCSS files
- Development workflow changes in Nx monorepo

## Testing Checklist
- [ ] All pages load correctly
- [ ] Navigation works between pages
- [ ] Redux state persists across page refreshes
- [ ] Theme switching works
- [ ] Language switching works
- [ ] Forms submit correctly
- [ ] Images and assets load properly
- [ ] Mobile responsiveness maintained
- [ ] Performance metrics improved
- [ ] SEO meta tags working

## Notes
- Keep backup in `landing-old` until migration is complete and tested
- Consider progressive migration if needed (hybrid approach)
- Monitor bundle size and performance throughout migration