import { describe, it, expect, beforeEach } from 'vitest';
import { useConfigStore } from './configStore';

const { getState } = useConfigStore;

beforeEach(() => {
  useConfigStore.setState({
    websiteData: null,
    aboutContent: null,
    maintanance: false,
    businessCount: 0,
    contactInfo: { contactEmail: null, contactNumber: null },
    faqs: [],
    news: [],
  });
});

describe('configStore', () => {
  it('has correct initial state', () => {
    expect(getState().websiteData).toBeNull();
    expect(getState().aboutContent).toBeNull();
    expect(getState().maintanance).toBe(false);
    expect(getState().businessCount).toBe(0);
    expect(getState().contactInfo).toEqual({ contactEmail: null, contactNumber: null });
    expect(getState().faqs).toEqual([]);
    expect(getState().news).toEqual([]);
  });

  it('setWebsiteData updates websiteData', () => {
    const data = { title: 'My Site', url: 'example.com' };
    getState().setWebsiteData(data);
    expect(getState().websiteData).toEqual(data);
  });

  it('setWebsiteData preserves an existing slug when the next payload omits it', () => {
    getState().setWebsiteData({ slug: 'voxire', title: 'Old title' });
    getState().setWebsiteData({ title: 'New title' });

    expect(getState().websiteData).toEqual({
      slug: 'voxire',
      title: 'New title',
    });
  });

  it('setWebsiteData normalizes snake_case API config fields', () => {
    getState().setWebsiteData({
      site_name: 'Suplab Store',
      contact_email: 'suplabstore@gmail.com',
      contact_number: '+961000000',
      sender_email: 'orders@suplabstore.com',
      maps_url: 'https://maps.example.com',
      maintenance_mode: true,
      about_content: 'Retail supplements',
      footer_sub_text: 'Powered by rtylr',
      website: 'https://suplabstore.com/',
    });

    expect(getState().websiteData).toMatchObject({
      siteName: 'Suplab Store',
      contactEmail: 'suplabstore@gmail.com',
      contactNumber: '+961000000',
      senderEmail: 'orders@suplabstore.com',
      mapsUrl: 'https://maps.example.com',
      maintenanceMode: true,
      aboutContent: 'Retail supplements',
      footerSubText: 'Powered by rtylr',
      website: 'https://suplabstore.com/',
    });
  });

  it('setAboutContent updates aboutContent', () => {
    const content = { text: 'About us' };
    getState().setAboutContent(content);
    expect(getState().aboutContent).toEqual(content);
  });

  it('setAboutContent with null', () => {
    getState().setAboutContent({ text: 'something' });
    getState().setAboutContent(null);
    expect(getState().aboutContent).toBeNull();
  });

  it('setMaintanance enables maintenance mode', () => {
    getState().setMaintanance(true);
    expect(getState().maintanance).toBe(true);
  });

  it('setMaintanance disables maintenance mode', () => {
    getState().setMaintanance(true);
    getState().setMaintanance(false);
    expect(getState().maintanance).toBe(false);
  });

  it('setBusinessCount updates count', () => {
    getState().setBusinessCount(5);
    expect(getState().businessCount).toBe(5);
  });

  it('setContactInfo updates contact info', () => {
    const info = { contactEmail: 'a@b.com', contactNumber: '+123' };
    getState().setContactInfo(info);
    expect(getState().contactInfo).toEqual(info);
  });

  it('setFaqs updates faqs', () => {
    const faqs = [{ q: 'What?', a: 'This.' }];
    getState().setFaqs(faqs);
    expect(getState().faqs).toEqual(faqs);
  });

  it('setNews updates news', () => {
    const news = [{ title: 'Breaking', body: 'Content' }];
    getState().setNews(news);
    expect(getState().news).toEqual(news);
  });

  it('multiple updates are independent', () => {
    getState().setBusinessCount(10);
    getState().setMaintanance(true);
    getState().setFaqs([{ q: 'test' }]);

    expect(getState().businessCount).toBe(10);
    expect(getState().maintanance).toBe(true);
    expect(getState().faqs).toHaveLength(1);
    expect(getState().websiteData).toBeNull();
  });
});
