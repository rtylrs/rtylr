import type { GetStaticPaths, GetStaticProps } from 'next';
import React from 'react';

import { baseRouteOf, localeRouteParams } from '../../lib/locales';
import AboutlessNotFound from '../404';
import ContactPage from '../contact';
import HomePage from '../index';
import PricingPage from '../pricing';
import TermsPage from '../legal/terms';
import PrivacyPage from '../legal/privacy-policy';
import SecurityPage from '../legal/security';
import TrustPage from '../trust';

/**
 * One catch-all that serves every localised route.
 *
 * The alternative shapes were worse. Duplicating each page under
 * `pages/ar/` and `pages/fr/` means three copies of every page to keep in
 * step, which is exactly the drift this codebase has spent the week
 * removing. Moving to the app router for its `[locale]` segment is the
 * better long-term shape and much too large to do as a side effect of
 * shipping two languages.
 *
 * So the page components stay where they are and this file maps a route to
 * one of them. There is no locale-specific rendering here at all: the copy
 * comes from i18next, which takes its language from the URL, and `lang` and
 * `dir` come from `_document`, which reads the same URL. A localised page is
 * the same component with a different active language, which is the only
 * arrangement where the two cannot disagree.
 *
 * `getStaticPaths` returns nothing while only the default locale ships, so
 * no `/ar/` or `/fr/` route reaches the export until its copy is reviewed.
 * `fallback: false` means a path outside the table 404s rather than being
 * rendered empty.
 */

const PAGES: Record<string, React.ComponentType> = {
  '/': HomePage,
  '/pricing/': PricingPage,
  '/contact/': ContactPage,
  '/legal/terms/': TermsPage,
  '/legal/privacy-policy/': PrivacyPage,
  '/legal/security/': SecurityPage,
  '/trust/': TrustPage,
};

interface Props {
  route: string;
}

const LocalisedPage = ({ route }: Props): React.JSX.Element => {
  const Page = PAGES[route];
  // Unreachable through the generated paths, since both come from
  // LOCALISED_ROUTES. Handled rather than asserted so a future route added
  // to the registry and not to PAGES renders the 404 instead of throwing
  // during the export.
  if (!Page) return <AboutlessNotFound />;
  return <Page />;
};

export const getStaticPaths: GetStaticPaths = () => ({
  // The cast is the shape Next expects for a catch-all: ParsedUrlQuery
  // values are string | string[], and `slug` is genuinely a string[]. The
  // params themselves come from the registry, so nothing here is guessed.
  paths: localeRouteParams().map(({ locale, slug }) => ({ params: { locale, slug } })),
  fallback: false,
});

export const getStaticProps: GetStaticProps<Props> = ({ params }) => ({
  props: { route: baseRouteOf(params?.slug as string[] | undefined) },
});

export default LocalisedPage;
