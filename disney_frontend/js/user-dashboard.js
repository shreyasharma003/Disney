// User Dashboard JavaScript

// ============================================
// AUTHENTICATION CHECK
// ============================================

document.addEventListener("DOMContentLoaded", function () {
  // Check if user is logged in
  if (!isLoggedIn()) {
    window.location.href = "user-login.html";
    return;
  }

  // Check if user role is 'user'
  const user = getUser();
  if (!user || user.role !== "user") {
    window.location.href = "index.html";
    return;
  }

  // Initialize dashboard
  initializeDashboard();
});

// ============================================
// INITIALIZATION
// ============================================

function initializeDashboard() {
  const user = getUser();

  // Set welcome message
  document.getElementById("welcomeMsg").textContent = `Welcome, ${user.name}`;

  // Load data
  loadGenres();
  loadTrending();
  loadFavorites();
  loadRecentlyViewed();

  // Setup event listeners
  setupSearchListeners();
  setupFilterListeners();
}

// ============================================
// LOGOUT
// ============================================

function handleLogout() {
  if (confirm("Are you sure you want to logout?")) {
    clearAuth();
    window.location.href = "index.html";
  }
}

// ============================================
// API CALLS
// ============================================

async function apiRequest(endpoint, options = {}) {
  const token = getToken();
  const headers = {
    "Content-Type": "application/json",
    ...(token && { Authorization: `Bearer ${token}` }),
    ...options.headers,
  };

  try {
    const response = await fetch(`${API_BASE_URL}${endpoint}`, {
      ...options,
      headers,
    });

    if (response.status === 401) {
      // Token expired or invalid
      clearAuth();
      window.location.href = "user-login.html";
      return null;
    }

    const data = await response.json();
    return { ok: response.ok, data, status: response.status };
  } catch (error) {
    console.error("API Error:", error);
    showToast("Network error. Please try again.", "error");
    return null;
  }
}

// ============================================
// LOAD GENRES
// ============================================

async function loadGenres() {
  const result = await apiRequest("/cartoons/genres");
  if (result && result.ok) {
    const genreSelect = document.getElementById("genreFilter");
    const genres = result.data.data || result.data;

    genres.forEach((genre) => {
      const option = document.createElement("option");
      option.value = genre.id || genre.name;
      option.textContent = genre.name;
      genreSelect.appendChild(option);
    });
  }
}

// ============================================
// LOAD TRENDING CARTOONS
// ============================================

async function loadTrending() {
  const container = document.getElementById("trendingContainer");

  // For now, fetch regular cartoons and mark them as trending
  // You can modify this to use IMDB API or add a trending flag in backend
  const result = await apiRequest("/cartoons?limit=10");

  if (result && result.ok) {
    const cartoons = result.data.data || result.data;
    const trending = cartoons.slice(0, 5); // Get first 5

    container.innerHTML = "";

    if (trending.length === 0) {
      container.innerHTML =
        '<div class="empty-state">No trending cartoons available</div>';
      return;
    }

    trending.forEach((cartoon) => {
      const card = createCartoonCard(cartoon, true);
      container.appendChild(card);
    });
  } else {
    container.innerHTML =
      '<div class="empty-state">Failed to load trending cartoons</div>';
  }
}

// ============================================
// LOAD FAVORITES
// ============================================

async function loadFavorites() {
  const container = document.getElementById("favoritesContainer");
  const result = await apiRequest("/user/favorites");

  if (result && result.ok) {
    const favorites = result.data.data || result.data || [];

    container.innerHTML = "";

    if (favorites.length === 0) {
      container.innerHTML =
        '<div class="empty-state">No favorites yet ❤️</div>';
      return;
    }

    favorites.forEach((favorite) => {
      const cartoon = favorite.cartoon || favorite;
      const card = createCartoonCard(cartoon, true);
      container.appendChild(card);
    });
  } else {
    container.innerHTML =
      '<div class="empty-state">Failed to load favorites</div>';
  }
}

// ============================================
// LOAD RECENTLY VIEWED
// ============================================

async function loadRecentlyViewed() {
  const container = document.getElementById("recentlyViewedContainer");
  const result = await apiRequest("/user/recently-viewed");

  if (result && result.ok) {
    const recentlyViewed = result.data.data || result.data || [];

    container.innerHTML = "";

    if (recentlyViewed.length === 0) {
      container.innerHTML =
        '<div class="empty-state">No recently viewed cartoons</div>';
      return;
    }

    recentlyViewed.slice(0, 5).forEach((item) => {
      const cartoon = item.cartoon || item;
      const card = createCartoonCard(cartoon, false);
      container.appendChild(card);
    });
  } else {
    container.innerHTML =
      '<div class="empty-state">Failed to load recently viewed</div>';
  }
}

// ============================================
// CREATE CARTOON CARD
// ============================================

function createCartoonCard(cartoon, showFavorite = false) {
  const card = document.createElement("div");
  card.className = "cartoon-card";
  card.dataset.cartoonId = cartoon.id;

  const posterUrl =
    cartoon.poster_url ||
    cartoon.posterURL ||
    "https://via.placeholder.com/200x300?text=No+Poster";
  const title = cartoon.title;
  const year = cartoon.release_year || cartoon.releaseYear || "";
  const rating = cartoon.rating || "N/A";

  card.innerHTML = `
    ${
      showFavorite
        ? `<button class="favorite-btn" onclick="toggleFavorite(event, ${cartoon.id})">
             ❤️
           </button>`
        : ""
    }
    <img src="${posterUrl}" alt="${title}" class="card-poster" loading="lazy" />
    <div class="card-content">
      <h3 class="card-title">${title}</h3>
      <div class="card-meta">
        ${year ? `<span>${year}</span>` : ""}
        ${
          rating !== "N/A"
            ? `<span class="card-rating">⭐ ${rating}</span>`
            : ""
        }
      </div>
    </div>
  `;

  card.addEventListener("click", function (e) {
    if (!e.target.closest(".favorite-btn")) {
      viewCartoonDetails(cartoon.id);
    }
  });

  return card;
}

// ============================================
// FAVORITE MANAGEMENT
// ============================================

async function toggleFavorite(event, cartoonId) {
  event.stopPropagation();
  const btn = event.currentTarget;
  const isActive = btn.classList.contains("active");

  if (isActive) {
    // Remove from favorites
    const result = await apiRequest(`/user/favorites/${cartoonId}`, {
      method: "DELETE",
    });

    if (result && result.ok) {
      btn.classList.remove("active");
      showToast("Removed from favorites", "success");
      loadFavorites(); // Refresh favorites section
    }
  } else {
    // Add to favorites
    const result = await apiRequest("/user/favorites", {
      method: "POST",
      body: JSON.stringify({ cartoon_id: cartoonId }),
    });

    if (result && result.ok) {
      btn.classList.add("active");
      showToast("Added to favorites", "success");
      loadFavorites(); // Refresh favorites section
    }
  }
}

// ============================================
// SEARCH FUNCTIONALITY
// ============================================

let searchTimeout;

function setupSearchListeners() {
  const searchInput = document.getElementById("searchInput");
  const clearBtn = document.getElementById("clearSearchBtn");

  searchInput.addEventListener("input", function () {
    clearTimeout(searchTimeout);

    if (searchInput.value.trim()) {
      clearBtn.style.display = "block";
      searchTimeout = setTimeout(() => {
        performSearch();
      }, 500); // Debounce 500ms
    } else {
      clearBtn.style.display = "none";
      hideSearchResults();
    }
  });

  clearBtn.addEventListener("click", function () {
    searchInput.value = "";
    clearBtn.style.display = "none";
    hideSearchResults();
  });
}

async function performSearch() {
  const query = document.getElementById("searchInput").value.trim();
  const genre = document.getElementById("genreFilter").value;
  const year = document.getElementById("yearFilter").value;

  if (!query) return;

  const resultsContainer = document.getElementById("searchResults");
  resultsContainer.style.display = "grid";
  resultsContainer.innerHTML = '<div class="skeleton-card"></div>'.repeat(4);

  // Build query params
  let endpoint = `/cartoons/search?query=${encodeURIComponent(query)}`;
  if (genre) endpoint += `&genre_id=${genre}`;
  if (year) endpoint += `&year=${year}`;

  const result = await apiRequest(endpoint);

  if (result && result.ok) {
    const cartoons = result.data.data || result.data || [];

    resultsContainer.innerHTML = "";

    if (cartoons.length === 0) {
      resultsContainer.innerHTML =
        '<div class="empty-state">No results found</div>';
      return;
    }

    cartoons.forEach((cartoon) => {
      const card = createCartoonCard(cartoon, true);
      resultsContainer.appendChild(card);
    });
  } else {
    resultsContainer.innerHTML =
      '<div class="empty-state">Failed to load search results</div>';
  }
}

function hideSearchResults() {
  const resultsContainer = document.getElementById("searchResults");
  resultsContainer.style.display = "none";
  resultsContainer.innerHTML = "";
}

// ============================================
// FILTER FUNCTIONALITY
// ============================================

function setupFilterListeners() {
  document
    .getElementById("genreFilter")
    .addEventListener("change", function () {
      if (document.getElementById("searchInput").value.trim()) {
        performSearch();
      }
    });

  document.getElementById("yearFilter").addEventListener("change", function () {
    if (document.getElementById("searchInput").value.trim()) {
      performSearch();
    }
  });
}

// ============================================
// VIEW CARTOON DETAILS
// ============================================

function viewCartoonDetails(cartoonId) {
  // For now, just show an alert
  // You can create a details page later
  showToast("Cartoon details page coming soon!", "success");
  console.log("View cartoon:", cartoonId);

  // TODO: Navigate to cartoon-details.html?id=${cartoonId}
  // window.location.href = `cartoon-details.html?id=${cartoonId}`;
}

// ============================================
// TOAST NOTIFICATION
// ============================================

function showToast(message, type = "success") {
  const toast = document.getElementById("toast");
  toast.textContent = message;
  toast.className = `toast ${type} show`;

  setTimeout(() => {
    toast.classList.remove("show");
  }, 3000);
}

// ============================================
// BACKGROUND CUTOUTS
// ============================================

// Placeholder cutout URLs - will be replaced with user-provided URLs
const cutoutUrls = [
  // Add URLs here when provided by user
];

function addBackgroundCutouts() {
  const container = document.getElementById("cutoutsContainer");

  cutoutUrls.forEach((url, index) => {
    const cutout = document.createElement("img");
    cutout.src = url;
    cutout.className = "cutout";
    cutout.style.width = `${100 + Math.random() * 150}px`;
    cutout.style.top = `${Math.random() * 80}%`;
    cutout.style.left = `${Math.random() * 90}%`;
    cutout.style.transform = `rotate(${Math.random() * 360}deg)`;
    container.appendChild(cutout);
  });
}

// Call when cutout URLs are available
if (cutoutUrls.length > 0) {
  addBackgroundCutouts();
}
