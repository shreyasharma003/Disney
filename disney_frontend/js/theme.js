// Theme Toggle Functionality

// Get saved theme from localStorage or default to 'light'
function getSavedTheme() {
  return localStorage.getItem("theme") || "light";
}

// Set theme and save to localStorage
function setTheme(theme) {
  document.documentElement.setAttribute("data-theme", theme);
  localStorage.setItem("theme", theme);
  updateThemeIcon(theme);
}

// Update the toggle button icon based on current theme
function updateThemeIcon(theme) {
  const themeIcon = document.getElementById("theme-icon");
  if (themeIcon) {
    themeIcon.textContent = theme === "dark" ? "☀️" : "🌙";
  }
}

// Toggle between light and dark themes
function toggleTheme() {
  const currentTheme = document.documentElement.getAttribute("data-theme");
  const newTheme = currentTheme === "dark" ? "light" : "dark";
  setTheme(newTheme);
}

// Initialize theme on page load
window.addEventListener("DOMContentLoaded", () => {
  const savedTheme = getSavedTheme();
  setTheme(savedTheme);
});
