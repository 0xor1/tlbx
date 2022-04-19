importScripts("https://www.gstatic.com/firebasejs/8.2.5/firebase-app.js");
importScripts("https://www.gstatic.com/firebasejs/8.2.5/firebase-messaging.js");

// Initialize Firebase
firebase.initializeApp(
{
  apiKey: "AIzaSyDFoddqIYlq5p4gYenOS2o4lgNQV9aO364",
  projectId: "trees-b8fdf",
  messagingSenderId: "256753508336",
  appId: "1:256753508336:web:af836cd97251283faac735"
});
const fcm = firebase.messaging();

fcm.onBackgroundMessage(function(payload) {
  console.log('[firebase-messaging-sw.js] Received background message ', payload);
  //self.registration.showNotification("YOLO", {silent: true});
});