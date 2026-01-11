// User Signup JavaScript with JWT handling

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
      age: parseInt(document.getElementById("age").value),
      password: document.getElementById("password").value,
    };

    // Validate form data
    if (
      !formData.name ||
      !formData.email ||
      !formData.age ||
      !formData.password
    ) {
      showError("Please fill in all fields");
      return;
    }

    if (formData.password.length < 6) {
      showError("Password must be at least 6 characters long");
      return;
    }

    // Disable button and show loading
    submitBtn.disabled = true;
    submitBtn.classList.add("loading");
    submitBtn.textContent = "Creating Account...";

    try {
      // Send signup request
      const response = await fetch(API_ENDPOINTS.USER_SIGNUP, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(formData),
      });

      const data = await response.json();

      if (response.ok) {
        // Success
        showSuccess(data.message || "Account created successfully!");

        // Clear form
        form.reset();

        // Redirect to login page after 2 seconds
        setTimeout(() => {
          window.location.href = "user-login.html";
        }, 2000);
      } else {
        // Error from server
        showError(data.error || data.message || "Registration failed");
      }
    } catch (error) {
      console.error("Signup error:", error);
      showError("Network error. Please check your connection and try again.");
    } finally {
      // Re-enable button
      submitBtn.disabled = false;
      submitBtn.classList.remove("loading");
      submitBtn.textContent = "Sign Up";
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
