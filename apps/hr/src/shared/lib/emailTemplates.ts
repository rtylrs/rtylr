/**
 * Reusable email templates for HR reminders and celebrations.
 * Used by notification/reminder features to compose outbound emails.
 */

export interface TemplateContext {
  name: string;
  date?: string;
  years?: number;
  documentName?: string;
  expiresOn?: string;
  company?: string;
}

export const emailTemplates = {
  birthday: ({ name, company = 'rtylr' }: TemplateContext) => ({
    subject: `Happy birthday, ${name}!`,
    body: `Hi ${name},\n\nWishing you a wonderful birthday from everyone at ${company}. Hope your day is full of great food, great company, and a little time to relax.\n\nThe ${company} team`,
  }),

  anniversary: ({ name, years = 1, company = 'rtylr' }: TemplateContext) => ({
    subject: `${years} year${years === 1 ? '' : 's'} at ${company}! Thank you, ${name}`,
    body: `Hi ${name},\n\nToday marks ${years} year${
      years === 1 ? '' : 's'
    } since you joined ${company}. Thank you for everything you bring to the team.\n\nHere's to many more.\n\nThe ${company} team`,
  }),

  documentExpiring: ({ name, documentName = 'a document', expiresOn = '' }: TemplateContext) => ({
    subject: `Action needed: ${documentName} expires soon`,
    body: `Hi ${name},\n\nThis is a friendly reminder that ${documentName} expires on ${expiresOn}. Please log in to the HR portal to upload a renewed copy before the expiration date.\n\nThank you,\nHR`,
  }),
};
