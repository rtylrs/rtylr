import { Helmet } from 'react-helmet-async';
import { useTranslation } from 'react-i18next';

interface HelmetContainerProps {
  title?: string;
  description?: string;
  image?: string;
  url?: string;
  author?: string;
  keywords?: string;
  robots?: string;
  ogType?: string;
  ogLocale?: string;
  ogSiteName?: string;
  ogDescription?: string;
  twitterCard?: string;
  twitterSite?: string;
  twitterCreator?: string;
  customMetaTags?: Record<string, string>[];
  favicon?: string;
}

function HelmetContainer({
  title,
  description = '',
  image = '',
  url = '',
  author = '',
  keywords = '',
  robots = '',
  ogType = '',
  ogLocale = '',
  ogSiteName = '',
  ogDescription = '',
  twitterCard = '',
  twitterSite = '',
  twitterCreator = '',
  customMetaTags = [],
  favicon = '',
}: HelmetContainerProps) {
  const { t } = useTranslation();
  // Default resolved here, not in the parameter list: a default there is a
  // literal and cannot reach a hook.
  const pageTitle = title ?? t('common.defaultPageTitle');
  return (
    <Helmet>
      <meta charSet="utf-8" />
      <title>{pageTitle}</title>
      {favicon && <link rel="icon" href={favicon} type="image/x-icon" />}
      <meta name="description" content={description} />
      {author && <meta name="author" content={author} />}
      {keywords && <meta name="keywords" content={keywords} />}
      {robots && <meta name="robots" content={robots} />}

      <meta property="og:title" content={pageTitle} />
      <meta property="og:image" content={image} />

      <meta property="og:url" content={url} />
      {ogType && <meta property="og:type" content={ogType} />}
      {ogLocale && <meta property="og:locale" content={ogLocale} />}
      {ogSiteName && <meta property="og:site_name" content={ogSiteName} />}
      {ogDescription && <meta property="og:description" content={ogDescription} />}

      {twitterCard && <meta name="twitter:card" content={twitterCard} />}
      {twitterSite && <meta name="twitter:site" content={twitterSite} />}
      {twitterCreator && <meta name="twitter:creator" content={twitterCreator} />}

      {customMetaTags && customMetaTags.map((tag, index) => <meta key={index} {...tag} />)}
    </Helmet>
  );
}

export default HelmetContainer;
