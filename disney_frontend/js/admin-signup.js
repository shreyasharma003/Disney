// Admin Signup JavaScript with Secret Key validation

document.addEventListener("DOMContentLoaded", function () {
  const form = document.getElementById("signupForm");
  const submitBtn = document.getElementById("submitBtn");
  const errorMessage = document.getElementById("errorMessage");
  const successMessage = document.getElementById("successMessage");

  form.addEventListener("submit", async function (e) {
    e.preventDefault();

    // Clear previous messages
    errorMessage.classList.remove("show");
    successMessage.classList.remove("show");
    errorMessage.textContent = "";
    successMessage.textContent = "";

    // Get form data
    const formData = {
      name: document.getElementById("name").value.trim(),
      email: document.getElementById("email").value.trim(),
      age: 25,
      password: document.getElementById("password").value,
      secret_key: document.getElementById("secretKey").value.trim(),
    };

    // Validate form data
    if (
      !formData.name ||
      !formData.email ||
      !formData.password ||
      !formData.secret_key
    ) {
      showError("Please fill in all fields including the admin secret key");
      return;
    }

    if (formData.password.length < 6) {
      showError("Password must be at least 6 characters long");
      return;
    }

    // Disable button and show loading
    submitBtn.disabled = true;
    submitBtn.classList.add("loading");
    submitBtn.textContent = "Creating Admin Account...";

    try {
      // Send admin signup request
      const response = await fetch(API_ENDPOINTS.ADMIN_SIGNUP, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(formData),
      });

      const data = await response.json();

      if (response.ok) {
        // Success
        showSuccess(data.message || "Admin account created successfully!");

        // Clear form
        form.reset();

        // Redirect to admin login page after 2 seconds
        setTimeout(() => {
          window.location.href = "admin-login.html";
        }, 2000);
      } else {
        // Error from server
        if (response.status === 403) {
          showError(
            "Invalid admin secret key. Please contact the system administrator."
          );
        } else {
          showError(data.error || data.message || "Registration failed");
        }
      }
    } catch (error) {
      console.error("Admin signup error:", error);
      showError("Network error. Please check your connection and try again.");
    } finally {
      // Re-enable button
      submitBtn.disabled = false;
      submitBtn.classList.remove("loading");
      submitBtn.textContent = "Create Admin Account";
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
