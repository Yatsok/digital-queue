// Check if the browser supports service workers
if ("serviceWorker" in navigator) {
  navigator.serviceWorker
    .register("/js/service-worker.js")
    .then((registration) => {
      console.log("Service Worker registered with scope:", registration.scope);
    })
    .catch((error) => {
      console.error("Service Worker registration failed:", error);
    });
}

document.addEventListener("htmx:wsAfterMessage", (event) => {
  const data = event.detail;
  console.log("WebSocket message received:", data);

  const parsedMessage = JSON.parse(data.message);

  if (
    parsedMessage &&
    "event" in parsedMessage &&
    "notification" in parsedMessage
  ) {
    if ("serviceWorker" in navigator) {
      navigator.serviceWorker.getRegistrations().then((registrations) => {
        const registration = registrations[0];
        console.log("Service worker supported");
        registration.active.postMessage(parsedMessage.notification);
      });
    } else {
      console.log("Service worker not supported");
      showNotification(parsedMessage.notification);
    }
  } else {
    console.error("Invalid WebSocket message structure:", parsedMessage);
  }
});

function showNotification(message) {
  console.log("Received notification:", message);

  // Check if the Notification API is supported
  if ("Notification" in window) {
    // Request permission to show notifications
    Notification.requestPermission().then((permission) => {
      if (permission === "granted") {
        // Show the notification
        new Notification("Notification from Server", {
          body: message,
          icon: "/img/favicon.ico",
        });
      } else {
        console.warn("Notification permission denied");
      }
    });
  } else {
    console.warn("Notification API not supported");
  }
}
