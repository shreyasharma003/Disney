// User Dashboard JavaScript

// Global favorites set to track which cartoons are favorited
let userFavorites = new Set();

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

  // Load data (load favorites first to populate the set)
  loadGenres();
  loadFavoritesAndData();

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
  // For now, skip loading genres since backend doesn't have this endpoint
  // You'll need to add a genres endpoint in backend or hardcode genre options
  const genreSelect = document.getElementById("genreFilter");
  // Placeholder - add genres manually or create backend endpoint
}

// ============================================
// LOAD FAVORITES AND DATA
// ============================================

async function loadFavoritesAndData() {
  // Load favorites first to populate the global set
  await loadFavorites();
  // Then load other sections
  loadTrending();
  loadRecentlyViewed();
}

// ============================================
// LOAD TRENDING CARTOONS
// ============================================

async function loadTrending() {
  console.log("Loading trending...");
  const container = document.getElementById("trendingContainer");

  // Fetch trending cartoons from backend (top 5 by IMDb rating)
  const result = await apiRequest("/admin/cartoons/trending");

  console.log("Trending API response:", result);

  if (result && result.ok) {
    const cartoons = result.data.data || result.data;
    console.log("Trending cartoons:", cartoons);

    container.innerHTML = "";

    if (!cartoons || cartoons.length === 0) {
      container.innerHTML =
        '<div class="empty-state">No trending cartoons available</div>';
      return;
    }

    cartoons.forEach((cartoon) => {
      const cartoonId = cartoon.id || cartoon.cartoon_id;
      const isFavorited = userFavorites.has(cartoonId);
      const card = createCartoonCard(cartoon, true, isFavorited);
      container.appendChild(card);
    });
  } else {
    console.error("Failed to load trending:", result);
    container.innerHTML =
      '<div class="empty-state">Failed to load trending cartoons</div>';
  }
}

// ============================================
// LOAD FAVORITES
// ============================================

async function loadFavorites() {
  const container = document.getElementById("favoritesContainer");
  const result = await apiRequest("/user/favourites");

  if (result && result.ok) {
    const favorites = result.data.data || result.data || [];

    // Populate global favorites set
    userFavorites.clear();
    favorites.forEach((favorite) => {
      const cartoonId =
        favorite.cartoon_id || favorite.cartoon?.id || favorite.id;
      userFavorites.add(cartoonId);
    });

    container.innerHTML = "";

    if (!favorites || favorites.length === 0) {
      container.innerHTML =
        '<div class="empty-state">No favorite cartoons yet ❤️</div>';
      return;
    }

    favorites.forEach((favorite) => {
      // Backend returns {id, user_id, cartoon_id, cartoon: {...}}
      const cartoon = favorite.cartoon || favorite;
      const card = createCartoonCard(cartoon, true, true); // Show as favorited
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
  console.log("Loading recently viewed...");
  const container = document.getElementById("recentlyViewedContainer");
  const result = await apiRequest("/admin/recently-viewed");

  console.log("Recently viewed API response:", result);

  if (result && result.ok) {
    const recentlyViewed = result.data.data || result.data || [];
    console.log("Recently viewed cartoons:", recentlyViewed);

    container.innerHTML = "";

    if (!recentlyViewed || recentlyViewed.length === 0) {
      container.innerHTML =
        '<div class="empty-state">No recently viewed cartoons</div>';
      return;
    }

    recentlyViewed.forEach((cartoon) => {
      const cartoonId = cartoon.id || cartoon.cartoon_id;
      const isFavorited = userFavorites.has(cartoonId);
      const card = createCartoonCard(cartoon, true, isFavorited);
      container.appendChild(card);
    });
  } else {
    console.error("Failed to load recently viewed:", result);
    container.innerHTML =
      '<div class="empty-state">Failed to load recently viewed</div>';
  }
}

// ============================================
// CREATE CARTOON CARD
// ============================================

function createCartoonCard(cartoon, showFavorite = false, isFavorited = false) {
  const card = document.createElement("div");
  card.className = "cartoon-card";
  const cartoonId = cartoon.id || cartoon.cartoon_id;
  card.dataset.cartoonId = cartoonId;

  const posterUrl =
    cartoon.poster_url ||
    cartoon.posterURL ||
    "data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='200' height='300'%3E%3Crect width='200' height='300' fill='%23333'/%3E%3Ctext x='50%25' y='50%25' dominant-baseline='middle' text-anchor='middle' font-family='Arial' font-size='16' fill='%23999'%3ENo Poster%3C/text%3E%3C/svg%3E";
  const title = cartoon.title;
  const year = cartoon.release_year || cartoon.releaseYear || "";
  const rating = cartoon.imdb_rating || cartoon.rating || "N/A";

  card.innerHTML = `
    ${
      showFavorite
        ? `<button class="favorite-btn ${
            isFavorited ? "active" : ""
          }" onclick="toggleFavorite(event, ${cartoonId})">
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
      viewCartoonDetails(cartoonId);
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
    const result = await apiRequest(`/user/favourites/${cartoonId}`, {
      method: "DELETE",
    });

    if (result && (result.ok || result.status === 202)) {
      userFavorites.delete(cartoonId); // Update global set

      // Update ALL favorite buttons for this cartoon across all sections
      updateAllFavoriteButtons(cartoonId, false);

      showToast("Removed from favorites", "success");
      loadFavorites(); // Refresh favorites section
    }
  } else {
    // Add to favorites
    const result = await apiRequest("/user/favourites", {
      method: "POST",
      body: JSON.stringify({ cartoon_id: cartoonId }),
    });

    if (result && (result.ok || result.status === 202)) {
      userFavorites.add(cartoonId); // Update global set

      // Update ALL favorite buttons for this cartoon across all sections
      updateAllFavoriteButtons(cartoonId, true);

      showToast("Added to favorites", "success");
      loadFavorites(); // Refresh favorites section
    }
  }
}

// Helper function to update all favorite buttons for a specific cartoon
function updateAllFavoriteButtons(cartoonId, isActive) {
  const allCards = document.querySelectorAll(
    `.cartoon-card[data-cartoon-id="${cartoonId}"]`
  );
  allCards.forEach((card) => {
    const btn = card.querySelector(".favorite-btn");
    if (btn) {
      if (isActive) {
        btn.classList.add("active");
      } else {
        btn.classList.remove("active");
      }
    }
  });
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
      // Don't hide results immediately - user might have filters selected
      const genre = document.getElementById("genreFilter").value;
      const year = document.getElementById("yearFilter").value;
      if (!genre && !year) {
        hideSearchResults();
      }
    }
  });

  clearBtn.addEventListener("click", function () {
    searchInput.value = "";
    clearBtn.style.display = "none";
    const genre = document.getElementById("genreFilter").value;
    const year = document.getElementById("yearFilter").value;
    if (!genre && !year) {
      hideSearchResults();
    }
  });
}

async function performSearch() {
  const query = document.getElementById("searchInput").value.trim();
  const genre = document.getElementById("genreFilter").value;
  const year = document.getElementById("yearFilter").value;

  if (!query && !genre && !year) return;

  const resultsContainer = document.getElementById("searchResults");
  resultsContainer.style.display = "grid";
  resultsContainer.innerHTML = '<div class="skeleton-card"></div>'.repeat(4);

  let cartoons = [];

  // If there's a text query, search by cartoon title AND character name
  if (query) {
    // Search by cartoon title (fetch all names and filter client-side)
    const namesResult = await apiRequest("/admin/cartoons/names");
    if (namesResult && namesResult.ok) {
      const allCartoonNames = namesResult.data.data || namesResult.data || [];

      // Filter cartoon names that match the query (case-insensitive)
      const matchingIds = allCartoonNames
        .filter((c) => c.title.toLowerCase().includes(query.toLowerCase()))
        .map((c) => c.id);

      // Fetch full details for matching cartoons
      if (matchingIds.length > 0) {
        for (const id of matchingIds) {
          const detailResult = await apiRequest(`/admin/cartoons/${id}`);
          if (detailResult && detailResult.ok) {
            const cartoon = detailResult.data.data || detailResult.data;
            cartoons.push(cartoon);
          }
        }
      }
    }

    // Also search by character name and merge results
    const charResult = await apiRequest(
      `/admin/cartoons/by-character?name=${encodeURIComponent(query)}`
    );
    if (charResult && charResult.ok) {
      const charCartoons = charResult.data.data || charResult.data || [];

      // Merge and deduplicate by cartoon ID
      charCartoons.forEach((cartoon) => {
        if (!cartoons.find((c) => c.id === cartoon.id)) {
          cartoons.push(cartoon);
        }
      });
    }
  } else if (genre && !year) {
    // Filter by genre only
    const genreMap = {
      1: "Action",
      2: "Comedy",
      3: "Drama",
      4: "Romance",
      5: "Thriller",
      6: "Horror",
      7: "Sci-Fi",
      8: "Fantasy",
    };
    const result = await apiRequest(
      `/admin/cartoons/by-genre?genre=${genreMap[genre]}`
    );
    if (result && result.ok) {
      cartoons = result.data.data || result.data || [];
    }
  } else if (year && !genre) {
    // Filter by year only
    const result = await apiRequest(`/admin/cartoons/by-year?year=${year}`);
    if (result && result.ok) {
      cartoons = result.data.data || result.data || [];
    }
  } else if (genre && year) {
    // Both genre and year
    const genreMap = {
      1: "Action",
      2: "Comedy",
      3: "Drama",
      4: "Romance",
      5: "Thriller",
      6: "Horror",
      7: "Sci-Fi",
      8: "Fantasy",
    };
    const result = await apiRequest(
      `/admin/cartoons/by-genre?genre=${genreMap[genre]}`
    );
    if (result && result.ok) {
      cartoons = result.data.data || result.data || [];
      // Filter by year client-side
      cartoons = cartoons.filter((c) => c.release_year == year);
    }
  }

  // Apply additional filters if query was used with genre/year
  if (query && (genre || year)) {
    if (genre) {
      const genreMap = {
        1: "Action",
        2: "Comedy",
        3: "Drama",
        4: "Romance",
        5: "Thriller",
        6: "Horror",
        7: "Sci-Fi",
        8: "Fantasy",
      };
      cartoons = cartoons.filter(
        (c) => c.genre && c.genre.name === genreMap[genre]
      );
    }
    if (year) {
      cartoons = cartoons.filter((c) => c.release_year == year);
    }
  }

  resultsContainer.innerHTML = "";

  if (cartoons.length === 0) {
    resultsContainer.innerHTML =
      '<div class="empty-state">No results found</div>';
    return;
  }

  cartoons.forEach((cartoon) => {
    const cartoonId = cartoon.id || cartoon.cartoon_id;
    const isFavorited = userFavorites.has(cartoonId);
    const card = createCartoonCard(cartoon, true, isFavorited);
    resultsContainer.appendChild(card);
  });
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
      const query = document.getElementById("searchInput").value.trim();
      const genre = this.value;
      const year = document.getElementById("yearFilter").value;

      // Perform search if any filter is active
      if (query || genre || year) {
        performSearch();
      }
    });

  document.getElementById("yearFilter").addEventListener("change", function () {
    const query = document.getElementById("searchInput").value.trim();
    const genre = document.getElementById("genreFilter").value;
    const year = this.value;

    // Perform search if any filter is active
    if (query || genre || year) {
      performSearch();
    }
  });
}

// ============================================
// VIEW CARTOON DETAILS
// ============================================

async function viewCartoonDetails(cartoonId) {
  // Navigate to cartoon details page
  window.location.href = `cartoon-details.html?id=${cartoonId}`;
}

// ============================================
// NAVIGATION
// ============================================

function navigateToCartoonList() {
  window.location.href = "cartoon-list.html";
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
