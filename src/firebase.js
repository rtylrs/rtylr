import firebase from 'firebase/compat/app';
import 'firebase/compat/auth';
import 'firebase/compat/firestore';
import 'firebase/compat/storage';


const firebaseConfig = {
    apiKey: "AIzaSyC_8k06RmjcWuQh3wm-N-02rEVnXjunlAs",
    authDomain: "sanity-3d4e1.firebaseapp.com",
    projectId: "sanity-3d4e1",
    storageBucket: "sanity-3d4e1.appspot.com",
    messagingSenderId: "512207225703",
    appId: "1:512207225703:web:6736e7c03a24d10879e5c2"
};

let app;
if (firebase.apps.length === 0) {
    app = firebase.initializeApp(firebaseConfig);
} else {
    // eslint-disable-next-line no-unused-vars
    app = firebase.app();
}

const auth = app.auth(); // for authentication
const db = app.firestore();    // for firestore
const storage = app.storage(); // for storage
db.settings({ merge: true });   // for firestore offline support


export { auth, db, storage, firebase };