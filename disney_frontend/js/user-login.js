// User Login JavaScript with JWT handling

document.addEventListener("DOMContentLoaded", function () {
  const form = document.getElementById("loginForm");
  const submitBtn = document.getElementById("submitBtn");
  const errorMessage = document.getElementById("errorMessage");
  const successMessage = document.getElementById("successMessage");

  // Check if already logged in
  /* Temporarily disabled to prevent redirect loop
  if (typeof isLoggedIn === 'function' && typeof getUser === 'function') {
    if (isLoggedIn()) {
      const user = getUser();
      if (user && user.role === "user") {
        window.location.href = "index.html";
      }
    }
  }
  */

  form.addEventListener("submit", async function (e) {
    e.preventDefault();

    // Clear previous messages
    errorMessage.classList.remove("show");
    successMessage.classList.remove("show");
    errorMessage.textContent = "";
    successMessage.textContent = "";

    // Get form data
    const formData = {
      email: document.getElementById("email").value.trim(),
      password: document.getElementById("password").value,
    };

    // Validate form data
    if (!formData.email || !formData.password) {
      showError("Please fill in all fields");
      return;
    }

    // Disable button and show loading
    submitBtn.disabled = true;
    submitBtn.classList.add("loading");
    submitBtn.textContent = "Logging in...";

    try {
      // Send login request
      const response = await fetch(API_ENDPOINTS.USER_LOGIN, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(formData),
      });

      const data = await response.json();

      if (response.ok) {
        // Check if user role is 'user'
        if (data.data && data.data.user && data.data.user.role !== "user") {
          showError("This is a user login page. Please use admin login.");
          return;
        }

        // Save token and user data
        if (data.token) {
          saveToken(data.token);
        }
        if (data.data && data.data.user) {
          saveUser(data.data.user);
        }

        // Success
        showSuccess(data.message || "Login successful!");

        // Redirect to homepage after 1 second
        setTimeout(() => {
          window.location.href = "index.html";
        }, 1000);
      } else {
        // Error from server
        showError(data.error || data.message || "Login failed");
      }
    } catch (error) {
      console.error("Login error:", error);
      showError("Network error. Please check your connection and try again.");
    } finally {
      // Re-enable button
      submitBtn.disabled = false;
      submitBtn.classList.remove("loading");
      submitBtn.textContent = "Login";
    }
  });

  function showError(message) {
    errorMessage.textContent = message;
    errorMessage.classList.add("show");
  }

  function showSuccess(message) {
    successMessage.textContent = message;
    successMessage.classList.add("show");
  }
});
