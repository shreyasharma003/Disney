// Cartoon List Page JavaScript

// Global variables
let userFavorites = new Set();
let allCartoons = [];
let filteredCartoons = [];
let currentPage = 1;
const cartoonsPerPage = 5;
let searchTimeout;

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

  // Initialize page
  initializePage();
});

// ============================================
// INITIALIZATION
// ============================================

async function initializePage() {
  const user = getUser();

  // Set welcome message
  document.getElementById("welcomeMsg").textContent = `Welcome, ${user.name}`;

  // Load data
  await loadFavorites();
  await loadAllCartoons();

  // Setup event listeners
  setupSearchListeners();
  setupFilterListeners();
}

// ============================================
// LOAD FAVORITES
// ============================================

async function loadFavorites() {
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
  }
}

// ============================================
// LOAD ALL CARTOONS
// ============================================

async function loadAllCartoons() {
  const grid = document.getElementById("cartoonsGrid");

  // Fetch all cartoon names first
  const result = await apiRequest("/admin/cartoons/names");

  if (result && result.ok) {
    const cartoonNames = result.data.data || result.data || [];

    // Fetch full details for all cartoons
    allCartoons = [];
    for (const cartoon of cartoonNames) {
      const detailResult = await apiRequest(`/admin/cartoons/${cartoon.id}`);
      if (detailResult && detailResult.ok) {
        const cartoonData = detailResult.data.data || detailResult.data;
        allCartoons.push(cartoonData);
      }
    }

    filteredCartoons = [...allCartoons];
    displayCartoons();
  } else {
    grid.innerHTML = '<div class="empty-state">Failed to load cartoons</div>';
  }
}

// ============================================
// DISPLAY CARTOONS
// ============================================

function displayCartoons() {
  const grid = document.getElementById("cartoonsGrid");
  const resultTitle = document.getElementById("resultTitle");
  const resultCount = document.getElementById("resultCount");
  const pagination = document.getElementById("pagination");

  // Calculate pagination
  const totalPages = Math.ceil(filteredCartoons.length / cartoonsPerPage);
  const startIndex = (currentPage - 1) * cartoonsPerPage;
  const endIndex = startIndex + cartoonsPerPage;
  const cartoonsToDisplay = filteredCartoons.slice(startIndex, endIndex);

  // Clear grid
  grid.innerHTML = "";

  // Display result count
  resultCount.textContent = `${filteredCartoons.length} cartoon${
    filteredCartoons.length !== 1 ? "s" : ""
  }`;

  if (cartoonsToDisplay.length === 0) {
    grid.innerHTML = '<div class="empty-state">No cartoons found</div>';
    pagination.style.display = "none";
    return;
  }

  // Display cartoons
  cartoonsToDisplay.forEach((cartoon) => {
    const cartoonId = cartoon.id || cartoon.cartoon_id;
    const isFavorited = userFavorites.has(cartoonId);
    const card = createCartoonCard(cartoon, isFavorited);
    grid.appendChild(card);
  });

  // Update pagination
  if (totalPages > 1) {
    pagination.style.display = "flex";
    document.getElementById(
      "pageInfo"
    ).textContent = `Page ${currentPage} of ${totalPages}`;
    document.getElementById("prevBtn").disabled = currentPage === 1;
    document.getElementById("nextBtn").disabled = currentPage === totalPages;
  } else {
    pagination.style.display = "none";
  }
}

// ============================================
// CREATE CARTOON CARD
// ============================================

function createCartoonCard(cartoon, isFavorited = false) {
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
    <button class="favorite-btn ${
      isFavorited ? "active" : ""
    }" onclick="toggleFavorite(event, ${cartoonId})">
      ❤️
    </button>
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
      userFavorites.delete(cartoonId);
      updateAllFavoriteButtons(cartoonId, false);
      showToast("Removed from favorites", "success");
    }
  } else {
    // Add to favorites
    const result = await apiRequest("/user/favourites", {
      method: "POST",
      body: JSON.stringify({ cartoon_id: cartoonId }),
    });

    if (result && (result.ok || result.status === 202)) {
      userFavorites.add(cartoonId);
      updateAllFavoriteButtons(cartoonId, true);
      showToast("Added to favorites", "success");
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

function setupSearchListeners() {
  const searchInput = document.getElementById("searchInput");
  const clearBtn = document.getElementById("clearSearchBtn");

  searchInput.addEventListener("input", function () {
    clearTimeout(searchTimeout);

    if (searchInput.value.trim()) {
      clearBtn.style.display = "block";
      searchTimeout = setTimeout(() => {
        performSearch();
      }, 500);
    } else {
      clearBtn.style.display = "none";
      const genre = document.getElementById("genreFilter").value;
      const year = document.getElementById("yearFilter").value;
      if (!genre && !year) {
        resetToAllCartoons();
      } else {
        performSearch();
      }
    }
  });

  clearBtn.addEventListener("click", function () {
    searchInput.value = "";
    clearBtn.style.display = "none";
    const genre = document.getElementById("genreFilter").value;
    const year = document.getElementById("yearFilter").value;
    if (!genre && !year) {
      resetToAllCartoons();
    } else {
      performSearch();
    }
  });
}

async function performSearch() {
  const query = document.getElementById("searchInput").value.trim();
  const genre = document.getElementById("genreFilter").value;
  const year = document.getElementById("yearFilter").value;

  const resultTitle = document.getElementById("resultTitle");
  const grid = document.getElementById("cartoonsGrid");

  grid.innerHTML = '<div class="skeleton-card"></div>'.repeat(5);

  let cartoons = [];

  // If there's a text query, search by cartoon title AND character name
  if (query) {
    // Search by cartoon title
    const titleMatches = allCartoons.filter((c) =>
      c.title.toLowerCase().includes(query.toLowerCase())
    );

    // Also search by character name
    const charResult = await apiRequest(
      `/admin/cartoons/by-character?name=${encodeURIComponent(query)}`
    );
    if (charResult && charResult.ok) {
      const charCartoons = charResult.data.data || charResult.data || [];

      // Merge and deduplicate
      charCartoons.forEach((cartoon) => {
        if (!titleMatches.find((c) => c.id === cartoon.id)) {
          titleMatches.push(cartoon);
        }
      });
    }

    cartoons = titleMatches;
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
      cartoons = cartoons.filter((c) => c.release_year == year);
    }
  } else {
    cartoons = [...allCartoons];
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

  filteredCartoons = cartoons;
  currentPage = 1;

  // Update title
  if (query || genre || year) {
    resultTitle.textContent = "Search Results";
  } else {
    resultTitle.textContent = "All Disney Cartoons";
  }

  displayCartoons();
}

function resetToAllCartoons() {
  filteredCartoons = [...allCartoons];
  currentPage = 1;
  document.getElementById("resultTitle").textContent = "All Disney Cartoons";
  displayCartoons();
}

// ============================================
// FILTER FUNCTIONALITY
// ============================================

function setupFilterListeners() {
  document
    .getElementById("genreFilter")
    .addEventListener("change", function () {
      performSearch();
    });

  document.getElementById("yearFilter").addEventListener("change", function () {
    performSearch();
  });
}

function clearAllFilters() {
  document.getElementById("searchInput").value = "";
  document.getElementById("genreFilter").value = "";
  document.getElementById("yearFilter").value = "";
  document.getElementById("clearSearchBtn").style.display = "none";
  resetToAllCartoons();
}

// ============================================
// PAGINATION
// ============================================

function previousPage() {
  if (currentPage > 1) {
    currentPage--;
    displayCartoons();
    window.scrollTo({ top: 0, behavior: "smooth" });
  }
}

function nextPage() {
  const totalPages = Math.ceil(filteredCartoons.length / cartoonsPerPage);
  if (currentPage < totalPages) {
    currentPage++;
    displayCartoons();
    window.scrollTo({ top: 0, behavior: "smooth" });
  }
}

// ============================================
// NAVIGATION
// ============================================

function goBack() {
  window.location.href = "user-dashboard.html";
}

async function viewCartoonDetails(cartoonId) {
  // Navigate to cartoon details page
  window.location.href = `cartoon-details.html?id=${cartoonId}`;
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
