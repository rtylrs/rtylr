import React from 'react';
import { Trans, useTranslation } from 'react-i18next';
import { useRouter } from 'next/router';
import { localeFromPath, localisePath } from '../../lib/locales';

interface LegalArticleProps {
  /** i18n namespace prefix for one legal document, e.g. 'legal.terms'. */
  i18nKey: string;
}

interface Subsection {
  heading: string;
  body: string;
}

/**
 * Every named tag any legal paragraph might use. Trans only renders the
 * ones a given string actually references, so building the full set once
 * here - shared by LegalArticle's body and LegalPageLayout's contact strip
 * - is what keeps every paragraph in every document in sync, instead of a
 * bespoke components map per call site.
 */
export function legalTransComponents(locale: string): Record<string, React.JSX.Element> {
  return {
    b: <strong />,
    secMail: <a href="mailto:security@voxire.com" />,
    infoMail: <a href="mailto:info@voxire.com" />,
    privacyMail: <a href="mailto:privacy@voxire.com" />,
    pricingLink: <a href={localisePath('/pricing/', locale)} />,
    privacyLink: <a href={localisePath('/legal/privacy-policy/', locale)} />,
    cookiesAnchor: <a href="#cookies" />,
  };
}

/**
 * Renders one legal document (Terms, Privacy Policy, Security) from its
 * i18n namespace, shared across all three rather than duplicated per page.
 *
 * A legal paragraph is not a flat string: "Terms" (bold, quoted),
 * "security@voxire.com" (a mailto link), "Privacy Policy" (a cross-page
 * link) all sit inside running prose. Storing that as raw HTML in the
 * locale file would mean a translator's typo can break markup, and no
 * gate catches it. Trans with a components map avoids that: the anchor
 * and <strong> elements live here, in code, and the locale string only
 * ever contains their tag *names* (<b>, <secMail>, ...). A translator can
 * move where a tag sits in a sentence but cannot inject an element that
 * isn't already wired below.
 *
 * The document's structure - which sections exist, which paragraphs and
 * lists each one holds - is data, not code, read from
 * `${i18nKey}.sections` and `${i18nKey}.body` with returnObjects. That is
 * what lets this one component serve three documents of different shapes
 * without a switch per page.
 */
export function LegalArticle({ i18nKey }: LegalArticleProps): React.JSX.Element {
  const { t } = useTranslation();
  const router = useRouter();
  const locale = localeFromPath(router.asPath).code;

  const sections = t(`${i18nKey}.sections`, { returnObjects: true }) as Record<string, string>;
  const body = t(`${i18nKey}.body`, { returnObjects: true }) as Record<
    string,
    Record<string, string | string[] | Subsection[]>
  >;
  const components = legalTransComponents(locale);

  return (
    <>
      {Object.entries(sections).map(([id, label]) => {
        const sectionBody = body[id];
        return (
          <React.Fragment key={id}>
            <h2 id={id}>{label}</h2>
            {sectionBody &&
              Object.entries(sectionBody).map(([blockKey, value]) => {
                if (blockKey === 'subsections' && Array.isArray(value)) {
                  return (value as Subsection[]).map((sub, i) => (
                    <React.Fragment key={`${id}-sub-${i}`}>
                      <h3>{sub.heading}</h3>
                      <p>
                        <Trans
                          i18nKey={`${i18nKey}.body.${id}.subsections.${i}.body`}
                          components={components}
                        />
                      </p>
                    </React.Fragment>
                  ));
                }
                if (blockKey.startsWith('list') && Array.isArray(value)) {
                  return (
                    <ul key={blockKey}>
                      {(value as string[]).map((_, i) => (
                        <li key={i}>
                          <Trans
                            i18nKey={`${i18nKey}.body.${id}.${blockKey}.${i}`}
                            components={components}
                          />
                        </li>
                      ))}
                    </ul>
                  );
                }
                return (
                  <p key={blockKey}>
                    <Trans i18nKey={`${i18nKey}.body.${id}.${blockKey}`} components={components} />
                  </p>
                );
              })}
          </React.Fragment>
        );
      })}
    </>
  );
}
