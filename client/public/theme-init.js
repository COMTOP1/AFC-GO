// Applies the saved light/dark theme before first paint so there is no flash.
// Keep in step with client/theme/setting.ts (same key, same rules).
(function () {
  var setting = 'system';
  try {
    var saved = window.localStorage.getItem('afc-theme');
    if (saved === 'light' || saved === 'dark') {
      setting = saved;
    }
  } catch (e) {
    // Storage blocked (private mode, policy): follow the system.
  }
  var dark =
    setting === 'dark' ||
    (setting === 'system' &&
      typeof window.matchMedia === 'function' &&
      window.matchMedia('(prefers-color-scheme: dark)').matches);
  document.documentElement.dataset.theme = dark ? 'dark' : 'light';
})();
