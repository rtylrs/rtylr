import firebase from 'firebase/compat/app';
import 'firebase/compat/auth';
import 'firebase/compat/analytics';
import { resolveFirebaseConfig } from './shared/config/firebaseEnv';

const firebaseConfig = resolveFirebaseConfig(import.meta.env);

const app = firebaseConfig
  ? firebase.apps.length === 0
    ? firebase.initializeApp(firebaseConfig)
    : firebase.app()
  : null;

const auth = app?.auth() ?? null;
const analytics = app && firebaseConfig?.measurementId ? app.analytics() : null;

export { auth, firebase, analytics };
