// Cartoon Details Page JavaScript

let cartoonId = null;
let cartoonData = null;
let isFavorited = false;

// ============================================
// AUTHENTICATION CHECK & INITIALIZATION
// ============================================

document.addEventListener("DOMContentLoaded", function () {
  // Check if user is logged in
  if (!isLoggedIn()) {
    window.location.href = "user-login.html";
    return;
  }

  // Get cartoon ID from URL
  const urlParams = new URLSearchParams(window.location.search);
  cartoonId = urlParams.get("id");

  if (!cartoonId) {
    showToast("No cartoon ID provided", "error");
    setTimeout(() => {
      window.location.href = "user-dashboard.html";
    }, 2000);
    return;
  }

  // Load data
  initializePage();
});

// ============================================
// INITIALIZATION
// ============================================

async function initializePage() {
  await checkIfFavorited();
  await loadCartoonDetails();
  await recordView();
}

// ============================================
// CHECK IF FAVORITED
// ============================================

async function checkIfFavorited() {
  const result = await apiRequest("/user/favourites");

  if (result && result.ok) {
    const favorites = result.data.data || result.data || [];
    const favorite = favorites.find(
      (f) =>
        f.cartoon_id == cartoonId ||
        f.cartoon?.id == cartoonId ||
        f.id == cartoonId
    );
    isFavorited = !!favorite;
    updateFavoriteButton();
  }
}

// ============================================
// LOAD CARTOON DETAILS
// ============================================

async function loadCartoonDetails() {
  const result = await apiRequest(`/admin/cartoons/${cartoonId}`);

  if (result && result.ok) {
    cartoonData = result.data.data || result.data;
    displayCartoonDetails();
  } else {
    showToast("Failed to load cartoon details", "error");
    setTimeout(() => {
      window.location.href = "user-dashboard.html";
    }, 2000);
  }
}

// ============================================
// DISPLAY CARTOON DETAILS
// ============================================

function displayCartoonDetails() {
  const {
    title,
    description,
    poster_url,
    release_year,
    genre,
    age_group,
    imdb_rating,
    characters,
  } = cartoonData;

  // Set page title
  document.title = `${title} - Disney`;

  // Hero section background (use poster as background)
  const heroSection = document.getElementById("heroSection");
  heroSection.style.backgroundImage = `url(${poster_url || ""})`;

  // Hero poster
  document.getElementById("heroPoster").src =
    poster_url ||
    "data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='300' height='450'%3E%3Crect width='300' height='450' fill='%23333'/%3E%3Ctext x='50%25' y='50%25' dominant-baseline='middle' text-anchor='middle' font-family='Arial' font-size='20' fill='%23999'%3ENo Poster%3C/text%3E%3C/svg%3E";
  document.getElementById("heroPoster").alt = title;

  // Title and description
  document.getElementById("cartoonTitle").textContent = title;
  document.getElementById("cartoonDescription").textContent =
    description || "No description available.";

  // Meta tags
  if (release_year) {
    document.getElementById("releaseYear").textContent = release_year;
  }
  if (genre && genre.name) {
    document.getElementById("genre").textContent = genre.name;
  }
  if (age_group && age_group.label) {
    document.getElementById("ageGroup").textContent = age_group.label;
  }

  const ratingElement = document.getElementById("rating");
  if (imdb_rating) {
    ratingElement.textContent = `⭐ ${imdb_rating}`;
    ratingElement.style.display = "inline-block";
  } else {
    ratingElement.style.display = "none";
  }

  // Info cards
  document.getElementById("genreInfo").textContent =
    genre?.name || "Not specified";
  document.getElementById("yearInfo").textContent =
    release_year || "Not specified";
  document.getElementById("ageInfo").textContent =
    age_group?.label || "Not specified";
  document.getElementById("ratingInfo").textContent =
    imdb_rating || "Not rated";

  // Characters
  displayCharacters(characters || []);

  // Initialize star rating
  initializeStarRating();
  loadUserRating();
}

// ============================================
// DISPLAY CHARACTERS
// ============================================

function displayCharacters(characters) {
  const grid = document.getElementById("charactersGrid");
  grid.innerHTML = "";

  if (characters.length === 0) {
    grid.innerHTML =
      '<div class="no-characters">No character information available</div>';
    return;
  }

  characters.forEach((character) => {
    const card = document.createElement("div");
    card.className = "character-card";

    const imageUrl =
      character.image_url ||
      "data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='150' height='150'%3E%3Crect width='150' height='150' fill='%23444'/%3E%3Ctext x='50%25' y='50%25' dominant-baseline='middle' text-anchor='middle' font-family='Arial' font-size='14' fill='%23999'%3ENo Image%3C/text%3E%3C/svg%3E";

    card.innerHTML = `
      <img src="${imageUrl}" alt="${character.name}" class="character-image" loading="lazy" />
      <div class="character-name">${character.name}</div>
    `;

    grid.appendChild(card);
  });
}

// ============================================
// FAVORITE MANAGEMENT
// ============================================

function updateFavoriteButton() {
  const btn = document.getElementById("favoriteBtn");
  if (isFavorited) {
    btn.classList.add("active");
  } else {
    btn.classList.remove("active");
  }
}

async function toggleFavorite() {
  const btn = document.getElementById("favoriteBtn");

  if (isFavorited) {
    // Remove from favorites
    const result = await apiRequest(`/user/favourites/${cartoonId}`, {
      method: "DELETE",
    });

    if (result && (result.ok || result.status === 202)) {
      isFavorited = false;
      btn.classList.remove("active");
      showToast("Removed from favorites", "success");
    } else {
      showToast("Failed to remove from favorites", "error");
    }
  } else {
    // Add to favorites
    const result = await apiRequest("/user/favourites", {
      method: "POST",
      body: JSON.stringify({ cartoon_id: parseInt(cartoonId) }),
    });

    if (result && (result.ok || result.status === 202)) {
      isFavorited = true;
      btn.classList.add("active");
      showToast("Added to favorites", "success");
    } else {
      showToast("Failed to add to favorites", "error");
    }
  }
}

// ============================================
// RECORD VIEW
// ============================================

async function recordView() {
  await apiRequest("/user/views", {
    method: "POST",
    body: JSON.stringify({ cartoon_id: parseInt(cartoonId) }),
  });
}

// ============================================
// NAVIGATION
// ============================================

function goBack() {
  // Check if there's history to go back to
  if (document.referrer && document.referrer.includes(window.location.host)) {
    window.history.back();
  } else {
    window.location.href = "user-dashboard.html";
  }
}

// ============================================
// API REQUEST HELPER
// ============================================

async function apiRequest(endpoint, options = {}) {
  try {
    const token = getToken();
    const defaultOptions = {
      headers: {
        "Content-Type": "application/json",
        ...(token && { Authorization: `Bearer ${token}` }),
      },
    };

    const response = await fetch(`${API_BASE_URL}${endpoint}`, {
      ...defaultOptions,
      ...options,
      headers: {
        ...defaultOptions.headers,
        ...options.headers,
      },
    });

    const data = await response.json();

    if (response.status === 401) {
      clearAuth();
      window.location.href = "user-login.html";
      return null;
    }

    return {
      ok: response.ok,
      status: response.status,
      data: data,
    };
  } catch (error) {
    console.error("API Error:", error);
    return null;
  }
}

// ============================================
// TOAST NOTIFICATION
// ============================================

function showToast(message, type = "success") {
  const toast = document.getElementById("toast");
  toast.textContent = message;
  toast.className = `toast show ${type}`;

  setTimeout(() => {
    toast.className = "toast";
  }, 3000);
}

// ============================================
// STAR RATING FUNCTIONALITY
// ============================================

let userRating = 0;
let hasRated = false;

function initializeStarRating() {
  const stars = document.querySelectorAll(".star");
  const ratingText = document.getElementById("ratingText");

  stars.forEach((star, index) => {
    // Hover effect
    star.addEventListener("mouseenter", () => {
      if (!hasRated) {
        highlightStars(index + 1, "hover");
      }
    });

    // Click to rate
    star.addEventListener("click", () => {
      if (!hasRated) {
        submitRating(index + 1);
      }
    });
  });

  // Reset hover effect when mouse leaves rating area
  const ratingContainer = document.getElementById("starRating");
  ratingContainer.addEventListener("mouseleave", () => {
    if (!hasRated) {
      resetStars();
    } else {
      highlightStars(userRating, "active");
    }
  });
}

function highlightStars(count, className = "active") {
  const stars = document.querySelectorAll(".star");
  stars.forEach((star, index) => {
    star.classList.remove("active", "hover");
    if (index < count) {
      star.classList.add(className);
    }
  });
}

function resetStars() {
  const stars = document.querySelectorAll(".star");
  stars.forEach((star) => {
    star.classList.remove("hover");
  });
}

async function loadUserRating() {
  const result = await apiRequest(`/user/ratings/${cartoonId}`);

  if (result && result.ok && result.data) {
    const rating = result.data.data || result.data;
    if (rating && rating.rating) {
      userRating = rating.rating;
      hasRated = true;
      highlightStars(userRating, "active");
      document.getElementById(
        "ratingText"
      ).textContent = `You rated: ${userRating}/10`;
      document.getElementById("ratingText").classList.add("submitted");
    }
  }
}

async function submitRating(rating) {
  const result = await apiRequest("/user/ratings", {
    method: "POST",
    body: JSON.stringify({
      cartoon_id: parseInt(cartoonId),
      rating: rating,
    }),
  });

  if (result && result.ok) {
    userRating = rating;
    hasRated = true;
    highlightStars(rating, "active");
    document.getElementById(
      "ratingText"
    ).textContent = `You rated: ${rating}/10`;
    document.getElementById("ratingText").classList.add("submitted");
    showToast("Rating submitted successfully!", "success");
  } else {
    const errorMsg = result?.data?.error || "Failed to submit rating";
    showToast(errorMsg, "error");
    resetStars();
  }
}
