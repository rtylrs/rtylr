import firebase from 'firebase/compat/app';
import 'firebase/compat/auth';
import 'firebase/compat/analytics';

const firebaseConfig = {
  apiKey: 'AIzaSyAu1ZafTsGoCio-amq8ziFHw-zy6b7Nd24',
  authDomain: 'vbitez-d12a8.firebaseapp.com',
  projectId: 'vbitez-d12a8',
  storageBucket: 'vbitez-d12a8.appspot.com',
  messagingSenderId: '46577622358',
  appId: '1:46577622358:web:8b078611a13a8fcd3b8af7',
  measurementId: 'G-SFJPPQ6RRT',
};

let app;
if (firebase.apps.length === 0) {
  app = firebase.initializeApp(firebaseConfig);
} else {
  app = firebase.app();
}

const auth = app.auth(); // for authentication
const analytics = app.analytics(); // for analytics

export { auth, firebase, analytics };
