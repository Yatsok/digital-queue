let userTimezone = Intl.DateTimeFormat().resolvedOptions().timeZone;

document.addEventListener("DOMContentLoaded", function () {
  document.getElementById("timezoneInput").value = userTimezone;
});
