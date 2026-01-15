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
  // Only load trending on initial page load
  loadTrending();
  // Favourites and Recently Viewed will be loaded when user clicks their tabs
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
  const container = document.getElementById("favoritesTabContainer");
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
  const container = document.getElementById("recentlyViewedTabContainer");
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
      
      // Refresh favorites tab if currently active (with delay for async processing)
      const favoritesTab = document.querySelector('[data-tab="favourites"]');
      if (favoritesTab && favoritesTab.classList.contains('active')) {
        // Wait a bit for the async worker to process the request
        setTimeout(() => {
          loadFavoritesTab();
        }, 1000);
      }
    }
  } else {
    // Add to favorites
    console.log("Adding to favorites, cartoon ID:", cartoonId);
    const result = await apiRequest("/user/favourites", {
      method: "POST",
      body: JSON.stringify({ cartoon_id: cartoonId }),
    });

    console.log("Add favorite result:", result);

    if (result && (result.ok || result.status === 202)) {
      userFavorites.add(cartoonId); // Update global set
      console.log("Updated userFavorites set:", userFavorites);

      // Update ALL favorite buttons for this cartoon across all sections
      updateAllFavoriteButtons(cartoonId, true);

      showToast("Added to favorites", "success");
      
      // Refresh favorites tab if currently active (with retry logic for async processing)
      const favoritesTab = document.querySelector('[data-tab="favourites"]');
      if (favoritesTab && favoritesTab.classList.contains('active')) {
        // Try to reload favorites with retries
        retryLoadFavorites(cartoonId, 3);
      }
    } else {
      console.error("Failed to add favorite:", result);
      showToast("Failed to add to favorites", "error");
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
    // Filter by year range
    const [startYear, endYear] = year.split("-").map(Number);
    const yearResults = [];

    // Fetch cartoons for each year in the range
    for (let y = startYear; y <= endYear; y++) {
      const result = await apiRequest(`/admin/cartoons/by-year?year=${y}`);
      if (result && result.ok) {
        const yearCartoons = result.data.data || result.data || [];
        yearResults.push(...yearCartoons);
      }
    }

    // Remove duplicates by cartoon ID
    const uniqueCartoons = new Map();
    yearResults.forEach((cartoon) => {
      if (!uniqueCartoons.has(cartoon.id)) {
        uniqueCartoons.set(cartoon.id, cartoon);
      }
    });
    cartoons = Array.from(uniqueCartoons.values());
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
      // Filter by year range client-side
      const [startYear, endYear] = year.split("-").map(Number);
      cartoons = cartoons.filter(
        (c) => c.release_year >= startYear && c.release_year <= endYear
      );
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
      const [startYear, endYear] = year.split("-").map(Number);
      cartoons = cartoons.filter(
        (c) => c.release_year >= startYear && c.release_year <= endYear
      );
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

function clearAllFilters() {
  // Clear search input
  const searchInput = document.getElementById("searchInput");
  const clearBtn = document.getElementById("clearSearchBtn");
  searchInput.value = "";
  clearBtn.style.display = "none";

  // Reset genre filter
  document.getElementById("genreFilter").value = "";

  // Reset year filter
  document.getElementById("yearFilter").value = "";

  // Hide search results
  hideSearchResults();
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

// ============================================
// TAB SWITCHING FUNCTIONALITY
// ============================================

// Pagination state
let currentPage = 1;
let allCartoonsData = [];
const cartoonsPerPage = 5;

function switchTab(tabName) {
  // Update active tab button
  document.querySelectorAll(".nav-tab").forEach((tab) => {
    tab.classList.remove("active");
  });
  document.querySelector(`[data-tab="${tabName}"]`).classList.add("active");

  // Update active content
  document.querySelectorAll(".tab-content").forEach((content) => {
    content.classList.remove("active");
  });
  document.getElementById(`${tabName}-content`).classList.add("active");

  // Load data for the selected tab if not already loaded
  if (tabName === "all-cartoons") {
    loadAllCartoonsTab();
  } else if (tabName === "favourites") {
    loadFavoritesTab();
  } else if (tabName === "recently-viewed") {
    loadRecentlyViewedTab();
  }
}

// ============================================
// LOAD ALL CARTOONS TAB WITH PAGINATION
// ============================================

async function loadAllCartoonsTab() {
  console.log("Loading all cartoons...");
  const container = document.getElementById("allCartoonsContainer");

  // Show loading state
  container.innerHTML = `
    <div class="skeleton-card"></div>
    <div class="skeleton-card"></div>
    <div class="skeleton-card"></div>
    <div class="skeleton-card"></div>
  `;

  const result = await apiRequest("/admin/cartoons/names");
  console.log("All cartoons API response:", result);

  if (result && result.ok) {
    allCartoonsData = result.data.data || result.data || [];
    console.log("All cartoons data:", allCartoonsData);
    currentPage = 1;
    renderAllCartoonsPage();
  } else {
    console.error("Failed to load all cartoons:", result);
    container.innerHTML =
      '<div class="empty-state">Failed to load cartoons</div>';
  }
}

function renderAllCartoonsPage(filteredData = null) {
  console.log("Rendering all cartoons page, current page:", currentPage);
  const container = document.getElementById("allCartoonsContainer");
  const dataToRender = filteredData || allCartoonsData;
  console.log("Data to render:", dataToRender.length, "items");

  const startIndex = (currentPage - 1) * cartoonsPerPage;
  const endIndex = startIndex + cartoonsPerPage;
  const pageCartoons = dataToRender.slice(startIndex, endIndex);
  const totalPages = Math.ceil(dataToRender.length / cartoonsPerPage);

  console.log("Page cartoons:", pageCartoons);

  container.innerHTML = "";

  if (pageCartoons.length === 0) {
    container.innerHTML = '<div class="empty-state">No cartoons found</div>';
    document.getElementById("paginationControls").style.display = "none";
    return;
  }

  // Create cards directly from basic cartoon data (no need to fetch full details for display)
  pageCartoons.forEach((cartoon) => {
    if (cartoon) {
      console.log("Creating card for cartoon:", cartoon);
      const cartoonId = cartoon.id;
      const isFavorited = userFavorites.has(cartoonId);
      
      // Create a basic cartoon object for display
      const displayCartoon = {
        id: cartoon.id,
        title: cartoon.title,
        description: "Click to view details", // Basic description
        poster_url: "", // Will use placeholder
        release_year: "", // Will be empty for basic view
        genre: null,
        age_group: null
      };
      
      const card = createCartoonCard(displayCartoon, true, isFavorited);
      container.appendChild(card);
    }
  });

  // Update pagination controls
  document.getElementById("paginationControls").style.display = "flex";
  document.getElementById(
    "pageInfo"
  ).textContent = `Page ${currentPage} of ${totalPages}`;
  document.getElementById("prevPageBtn").disabled = currentPage === 1;
  document.getElementById("nextPageBtn").disabled = currentPage >= totalPages;
}

function changePage(direction) {
  currentPage += direction;
  renderAllCartoonsPage();
  // Scroll to top of content
  document.querySelector(".dashboard-content").scrollTop = 0;
}

// ============================================
// LOAD FAVOURITES TAB
// ============================================

async function loadFavoritesTab() {
  console.log("Loading favorites tab...");
  const container = document.getElementById("favoritesTabContainer");

  // Show loading state
  container.innerHTML = `
    <div class="skeleton-card"></div>
    <div class="skeleton-card"></div>
    <div class="skeleton-card"></div>
  `;

  const result = await apiRequest("/user/favourites");
  console.log("Favorites API response:", result);

  if (result && result.ok) {
    const favorites = result.data.data || result.data || [];
    console.log("Favorites data:", favorites);

    // Update global favorites set
    userFavorites.clear();
    favorites.forEach((favorite) => {
      const cartoonId =
        favorite.cartoon_id || favorite.cartoon?.id || favorite.id;
      userFavorites.add(cartoonId);
    });

    container.innerHTML = "";

    if (!favorites || favorites.length === 0) {
      container.innerHTML =
        '<div class="empty-state">No favorite cartoons yet ❤️<br><small>Browse cartoons and add your favorites!</small></div>';
      return;
    }

    favorites.forEach((favorite) => {
      const cartoon = favorite.cartoon || favorite;
      console.log("Processing favorite:", favorite);
      console.log("Cartoon data:", cartoon);
      
      if (cartoon && cartoon.id) {
        const card = createCartoonCard(cartoon, true, true);
        container.appendChild(card);
      } else {
        console.error("Invalid cartoon data:", cartoon);
      }
    });
    
    console.log("Favorites tab populated with", favorites.length, "items");
  } else {
    container.innerHTML =
      '<div class="empty-state">Failed to load favorites</div>';
  }
}

// ============================================
// DEBUG FUNCTIONS (for testing)
// ============================================

// Test function to manually check favorites API
async function testFavoritesAPI() {
  console.log("=== TESTING FAVORITES API ===");
  
  // Test GET favorites
  const getFavs = await apiRequest("/user/favourites");
  console.log("GET /user/favourites response:", getFavs);
  
  // Test POST favorite (add cartoon ID 1)
  const addFav = await apiRequest("/user/favourites", {
    method: "POST",
    body: JSON.stringify({ cartoon_id: 1 }),
  });
  console.log("POST /user/favourites response:", addFav);
  
  // Wait 2 seconds and check again
  setTimeout(async () => {
    const getFavsAgain = await apiRequest("/user/favourites");
    console.log("GET /user/favourites after 2s:", getFavsAgain);
  }, 2000);
}

// Make it available globally for console testing
window.testFavoritesAPI = testFavoritesAPI;

// Retry loading favorites until the added item appears
async function retryLoadFavorites(expectedCartoonId, maxRetries = 3) {
  for (let i = 0; i < maxRetries; i++) {
    console.log(`Retry ${i + 1}: Loading favorites...`);
    
    // Wait before checking
    await new Promise(resolve => setTimeout(resolve, 1000 + (i * 500)));
    
    const result = await apiRequest("/user/favourites");
    if (result && result.ok) {
      const favorites = result.data.data || result.data || [];
      console.log(`Retry ${i + 1}: Found ${favorites.length} favorites`);
      
      // Check if our expected cartoon is in the list
      const found = favorites.some(fav => {
        const cartoonId = fav.cartoon_id || fav.cartoon?.id || fav.id;
        return cartoonId == expectedCartoonId;
      });
      
      if (found || favorites.length > 0) {
        console.log("Favorites found! Refreshing tab...");
        loadFavoritesTab();
        return;
      }
    }
  }
  
  console.log("Max retries reached, forcing refresh anyway...");
  loadFavoritesTab();
}

// ============================================
// LOAD RECENTLY VIEWED TAB
// ============================================

async function loadRecentlyViewedTab() {
  const container = document.getElementById("recentlyViewedTabContainer");

  // Show loading state
  container.innerHTML = `
    <div class="skeleton-card"></div>
    <div class="skeleton-card"></div>
    <div class="skeleton-card"></div>
  `;

  const result = await apiRequest("/admin/recently-viewed");

  if (result && result.ok) {
    const recentlyViewed = result.data.data || result.data || [];

    container.innerHTML = "";

    if (!recentlyViewed || recentlyViewed.length === 0) {
      container.innerHTML =
        '<div class="empty-state">No recently viewed cartoons<br><small>Start watching cartoons to see them here!</small></div>';
      return;
    }

    recentlyViewed.forEach((cartoon) => {
      const cartoonId = cartoon.id || cartoon.cartoon_id;
      const isFavorited = userFavorites.has(cartoonId);
      const card = createCartoonCard(cartoon, true, isFavorited);
      container.appendChild(card);
    });
  } else {
    container.innerHTML =
      '<div class="empty-state">Failed to load recently viewed</div>';
  }
}
