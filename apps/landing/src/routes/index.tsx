import { AnimatedRoutes } from '@voxire/shared';
import MainLayout from '../layouts/Main/MainLayout';
import React from 'react';

// Routes
const Home = React.lazy(() => import('../screens/Home/Home'));
const About = React.lazy(() => import('../screens/About'));
const Pricing = React.lazy(() => import('../screens/Pricing'));
const Contact = React.lazy(() => import('../screens/Contact'));
const Faq = React.lazy(() => import('../screens/FAQ'));
const News = React.lazy(() => import('../screens/News'));
const UseCases = React.lazy(() => import('../screens/UseCases'));
const Restaunrats = React.lazy(() => import('../screens/Restaunrats'));

const AppRoutes = [
  {
    path: '/',
    element: (
      <MainLayout>
        <AnimatedRoutes>
          <Home />
        </AnimatedRoutes>
      </MainLayout>
    ),
  },
  {
    path: '/about',
    element: (
      <MainLayout>
        <AnimatedRoutes>
          <About />
        </AnimatedRoutes>
      </MainLayout>
    ),
  },
  {
    path: '/pricing',
    element: (
      <MainLayout>
        <AnimatedRoutes>
          <Pricing />
        </AnimatedRoutes>
      </MainLayout>
    ),
  },
  {
    path: '/contact',
    element: (
      <MainLayout>
        <AnimatedRoutes>
          <Contact />
        </AnimatedRoutes>
      </MainLayout>
    ),
  },
  {
    path: '/faqs',
    element: (
      <MainLayout>
        <AnimatedRoutes>
          <Faq />
        </AnimatedRoutes>
      </MainLayout>
    ),
  },
  {
    path: '/news/:newsId',
    element: (
      <MainLayout>
        <AnimatedRoutes>
          <News />
        </AnimatedRoutes>
      </MainLayout>
    ),
  },
  {
    path: '/uses',
    element: (
      <MainLayout>
        <AnimatedRoutes>
          <UseCases />
        </AnimatedRoutes>
      </MainLayout>
    ),
  },
  {
    path: '/restaurants',
    element: (
      <MainLayout>
        <AnimatedRoutes>
          <Restaunrats />
        </AnimatedRoutes>
      </MainLayout>
    ),
  },
];

export default AppRoutes;
