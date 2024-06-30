import { Helmet } from "react-helmet";
import PropTypes from "prop-types";

function HelmetContainer({
  title = "Vbitez - Contactless QR Menu",
  description,
  image,
  url,
  author,
  keywords,
  robots,
  ogType,
  ogLocale,
  ogSiteName,
  ogDescription,
  twitterCard,
  twitterSite,
  twitterCreator,
  customMetaTags,
  favicon,
}) {
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
      {ogDescription && (
        <meta property="og:description" content={ogDescription} />
      )}

      {twitterCard && <meta name="twitter:card" content={twitterCard} />}
      {twitterSite && <meta name="twitter:site" content={twitterSite} />}
      {twitterCreator && (
        <meta name="twitter:creator" content={twitterCreator} />
      )}

      {customMetaTags &&
        customMetaTags.map((tag, index) => <meta key={index} {...tag} />)}
    </Helmet>
  );
}

export default HelmetContainer;

HelmetContainer.propTypes = {
  title: PropTypes.string,
  description: PropTypes.string,
  image: PropTypes.string,
  url: PropTypes.string,
  author: PropTypes.string,
  keywords: PropTypes.string,
  robots: PropTypes.string,
  ogType: PropTypes.string,
  ogLocale: PropTypes.string,
  ogSiteName: PropTypes.string,
  ogDescription: PropTypes.string,
  twitterCard: PropTypes.string,
  twitterSite: PropTypes.string,
  twitterCreator: PropTypes.string,
  customMetaTags: PropTypes.arrayOf(PropTypes.object),
  favicon: PropTypes.string,
};
