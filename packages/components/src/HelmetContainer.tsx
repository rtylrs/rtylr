import { Helmet } from 'react-helmet-async';

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
  customMetaTags?: any[];
  favicon?: string;
}

export const HelmetContainer = ({
  title = '',
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
}: HelmetContainerProps) => {
  return (
    <Helmet>
      <meta charSet="utf-8" />
      <title>{title}</title>
      {favicon && <link rel="icon" href={favicon} type="image/x-icon" />}
      <meta name="description" content={description} />
      {author && <meta name="author" content={author} />}
      {keywords && <meta name="keywords" content={keywords} />}
      {robots && <meta name="robots" content={robots} />}

      <meta property="og:title" content={title} />
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
};
