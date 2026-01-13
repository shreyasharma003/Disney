// Homepage Navigation Logic
console.log("Homepage.js loaded successfully!");

window.navigateToSignup = function (role) {
  console.log("Navigate to signup:", role);
  if (role === "user") {
    window.location.href = "user-signup.html";
  } else if (role === "admin") {
    window.location.href = "admin-signup.html";
  }
};

window.navigateToLogin = function (role) {
  console.log("Navigate to login:", role);
  if (role === "user") {
    window.location.href = "user-login.html";
  } else if (role === "admin") {
    window.location.href = "admin-login.html";
  }
};

window.navigateToAdminLogin = function (event) {
  event.preventDefault();
  console.log("Navigate to admin login");
  window.location.href = "admin-login.html";
};
