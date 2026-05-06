export interface FirebaseRuntimeEnv {
  PROD?: boolean;
  VITE_FIREBASE_API_KEY?: string;
  VITE_FIREBASE_AUTH_DOMAIN?: string;
  VITE_FIREBASE_PROJECT_ID?: string;
  VITE_FIREBASE_STORAGE_BUCKET?: string;
  VITE_FIREBASE_MESSAGING_SENDER_ID?: string;
  VITE_FIREBASE_APP_ID?: string;
  VITE_FIREBASE_MEASUREMENT_ID?: string;
}

export interface FirebaseConfig {
  apiKey: string;
  authDomain: string;
  projectId: string;
  storageBucket: string;
  messagingSenderId: string;
  appId: string;
  measurementId?: string;
}

const REQUIRED_KEYS = [
  'VITE_FIREBASE_API_KEY',
  'VITE_FIREBASE_AUTH_DOMAIN',
  'VITE_FIREBASE_PROJECT_ID',
  'VITE_FIREBASE_STORAGE_BUCKET',
  'VITE_FIREBASE_MESSAGING_SENDER_ID',
  'VITE_FIREBASE_APP_ID',
] as const;

function present(value: string | undefined): value is string {
  return typeof value === 'string' && value.trim().length > 0;
}

export function resolveFirebaseConfig(env: FirebaseRuntimeEnv): FirebaseConfig | null {
  const missing = REQUIRED_KEYS.filter((key) => !present(env[key]));

  if (missing.length > 0) {
    if (env.PROD) {
      throw new Error(`Missing required Firebase environment variables: ${missing.join(', ')}`);
    }
    return null;
  }

  const config: FirebaseConfig = {
    apiKey: env.VITE_FIREBASE_API_KEY!,
    authDomain: env.VITE_FIREBASE_AUTH_DOMAIN!,
    projectId: env.VITE_FIREBASE_PROJECT_ID!,
    storageBucket: env.VITE_FIREBASE_STORAGE_BUCKET!,
    messagingSenderId: env.VITE_FIREBASE_MESSAGING_SENDER_ID!,
    appId: env.VITE_FIREBASE_APP_ID!,
  };

  if (present(env.VITE_FIREBASE_MEASUREMENT_ID)) {
    config.measurementId = env.VITE_FIREBASE_MEASUREMENT_ID;
  }

  return config;
}
