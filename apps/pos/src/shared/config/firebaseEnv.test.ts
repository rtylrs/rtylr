import { describe, expect, it } from 'vitest';
import { resolveFirebaseConfig } from './firebaseEnv';

describe('resolveFirebaseConfig', () => {
  it('returns null outside production when Firebase env is absent', () => {
    expect(resolveFirebaseConfig({ PROD: false })).toBeNull();
  });

  it('throws in production when required Firebase env is missing', () => {
    expect(() => resolveFirebaseConfig({ PROD: true })).toThrow(
      'Missing required Firebase environment variables: VITE_FIREBASE_API_KEY'
    );
  });

  it('maps complete Firebase env to Firebase config', () => {
    expect(
      resolveFirebaseConfig({
        PROD: true,
        VITE_FIREBASE_API_KEY: 'api-key',
        VITE_FIREBASE_AUTH_DOMAIN: 'example.firebaseapp.com',
        VITE_FIREBASE_PROJECT_ID: 'example',
        VITE_FIREBASE_STORAGE_BUCKET: 'example.appspot.com',
        VITE_FIREBASE_MESSAGING_SENDER_ID: 'sender',
        VITE_FIREBASE_APP_ID: 'app-id',
        VITE_FIREBASE_MEASUREMENT_ID: 'measurement',
      })
    ).toEqual({
      apiKey: 'api-key',
      authDomain: 'example.firebaseapp.com',
      projectId: 'example',
      storageBucket: 'example.appspot.com',
      messagingSenderId: 'sender',
      appId: 'app-id',
      measurementId: 'measurement',
    });
  });

  it('omits analytics measurement id when not configured', () => {
    expect(
      resolveFirebaseConfig({
        PROD: true,
        VITE_FIREBASE_API_KEY: 'api-key',
        VITE_FIREBASE_AUTH_DOMAIN: 'example.firebaseapp.com',
        VITE_FIREBASE_PROJECT_ID: 'example',
        VITE_FIREBASE_STORAGE_BUCKET: 'example.appspot.com',
        VITE_FIREBASE_MESSAGING_SENDER_ID: 'sender',
        VITE_FIREBASE_APP_ID: 'app-id',
      })
    ).toEqual({
      apiKey: 'api-key',
      authDomain: 'example.firebaseapp.com',
      projectId: 'example',
      storageBucket: 'example.appspot.com',
      messagingSenderId: 'sender',
      appId: 'app-id',
    });
  });
});
