self.addEventListener("message", (event) => {
  const notificationMessage = event.data;
  console.log("Service Worker received message:", notificationMessage);

  self.registration.showNotification("DigiQue: Your Queue is Coming", {
    body: notificationMessage,
    icon: "/img/favicon.ico",
  });
});

self.addEventListener("notificationclick", (event) => {
  event.notification.close();
});

self.addEventListener("install", function (event) {
  event.waitUntil(self.skipWaiting());
});

self.addEventListener("activate", function (event) {
  console.log("Service Worker activated");
  event.waitUntil(self.clients.claim());
});

self.addEventListener("error", (error) => {
  console.error("Service Worker error:", error);
});
