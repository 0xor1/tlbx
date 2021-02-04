importScripts("https://www.gstatic.com/firebasejs/8.2.5/firebase-app.js");
importScripts("https://www.gstatic.com/firebasejs/8.2.5/firebase-messaging.js");

// Initialize Firebase
firebase.initializeApp({
  apiKey: "AIzaSyAg43CfgwC2HLC9x582IMq2UwM6NQ3FRCc",
  projectId: "trees-82a30",
  messagingSenderId: "69294578877",
  appId: "1:69294578877:web:1edb203c55b78f43956bd4",
});
const fcm = firebase.messaging();

fcm.onBackgroundMessage(function(payload) {
  console.log('[firebase-messaging-sw.js] Received background message ', payload);
});