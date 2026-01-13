// ============================================
// ADMIN DASHBOARD - JAVASCRIPT
// ============================================

// Ensure API_BASE_URL is available
if (typeof API_BASE_URL === 'undefined') {
  // Fallback if config.js hasn't loaded yet or API_BASE isn't available
  const API_BASE_URL = typeof API_BASE !== 'undefined' ? `${API_BASE}/api` : "https://disney-79c7.onrender.com/api";
  window.API_BASE_URL = API_BASE_URL;
}

// State Management
const state = {
  cartoons: [],
  filteredCartoons: [],
  characters: [],
  genres: [],
  ageGroups: [],
  editingCartoonId: null,
  deleteTarget: null,
  token: null,
  charactersToSave: [], // Track characters being added/edited

  // Pagination (5 columns × 2 rows = 10 items per page)
  currentPage: 1,
  itemsPerPage: 10,
  totalPages: 1,
};

// ============================================
// INITIALIZATION
// ============================================

window.addEventListener("DOMContentLoaded", () => {
  // Check authentication - use the key from config.js (disney_auth_token)
  const token = localStorage.getItem("disney_auth_token");
  if (!token) {
    window.location.href = "./admin-login.html";
    return;
  }

  state.token = token;

  // Display user email from user data
  const userData = localStorage.getItem("disney_user_data");
  if (userData) {
    try {
      const user = JSON.parse(userData);
      document.getElementById("user-email").textContent = user.email || "User";

      // Check if user is admin
      if (user.role !== "admin") {
        alert("Access denied. Admin privileges required.");
        window.location.href = "./homepage.html";
        return;
      }
    } catch (e) {
      console.error("Error parsing user data:", e);
    }
  }

  // Initialize dashboard
  loadDashboardData();

  // Close character image modal when clicking outside
  document
    .getElementById("character-image-modal")
    .addEventListener("click", function (e) {
      if (e.target === this) {
        closeCharacterImageModal();
      }
    });
});

/**
 * Load all dashboard data
 */
async function loadDashboardData() {
  try {
    // Load cartoons first
    await loadCartoons();

    // Extract unique genres and age groups from cartoons
    extractGenresAndAgeGroups();

    // Populate filter dropdowns
    populateFilterDropdowns();
  } catch (error) {
    console.error("Failed to load dashboard data:", error);
    showError("Failed to load dashboard data");
  }
}

/**
 * Extract unique genres and age groups from cartoons data
 */
function extractGenresAndAgeGroups() {
  const genresMap = {};
  const ageGroupsMap = {};

  state.cartoons.forEach((cartoon) => {
    if (cartoon.genre) {
      genresMap[cartoon.genre.id] = cartoon.genre;
    }
    if (cartoon.age_group) {
      ageGroupsMap[cartoon.age_group.id] = cartoon.age_group;
    }
  });

  state.genres = Object.values(genresMap);
  state.ageGroups = Object.values(ageGroupsMap);
}

/**
 * Populate filter dropdowns
 */
function populateFilterDropdowns() {
  // Genres
  const genreFilter = document.getElementById("genre-filter");
  genreFilter.innerHTML = '<option value="">All Genres</option>';
  state.genres.forEach((genre) => {
    const option = document.createElement("option");
    option.value = genre.id;
    option.textContent = genre.name;
    genreFilter.appendChild(option);
  });

  // Age Groups
  const ageGroupFilter = document.getElementById("age-group-filter");
  ageGroupFilter.innerHTML = '<option value="">All Age Groups</option>';
  state.ageGroups.forEach((ageGroup) => {
    const option = document.createElement("option");
    option.value = ageGroup.id;
    option.textContent = ageGroup.label || ageGroup.name;
    ageGroupFilter.appendChild(option);
  });

  // Character form selects
  const genreSelect = document.getElementById("cartoon-genre");
  if (genreSelect) {
    genreSelect.innerHTML = '<option value="">Select Genre</option>';
    state.genres.forEach((genre) => {
      const option = document.createElement("option");
      option.value = genre.id;
      option.textContent = genre.name;
      genreSelect.appendChild(option);
    });
  }

  const ageGroupSelect = document.getElementById("cartoon-age-group");
  if (ageGroupSelect) {
    ageGroupSelect.innerHTML = '<option value="">Select Age Group</option>';
    state.ageGroups.forEach((ageGroup) => {
      const option = document.createElement("option");
      option.value = ageGroup.id;
      option.textContent = ageGroup.label || ageGroup.name;
      ageGroupSelect.appendChild(option);
    });
  }
}

// ============================================
// CARTOONS MANAGEMENT
// ============================================

/**
 * Load cartoons from API
 */
async function loadCartoons() {
  showLoading("cartoons-loading");

  try {
    // Fetch all cartoon names first
    const namesResponse = await fetch(`${API_BASE_URL}/admin/cartoons/names`, {
      headers: {
        Authorization: `Bearer ${state.token}`,
        "Content-Type": "application/json",
      },
    });

    if (namesResponse.ok) {
      const namesData = await namesResponse.json();
      const allNames = namesData.data || [];

      // Fetch full details for each cartoon
      const cartoonPromises = allNames.map((cartoon) =>
        fetch(`${API_BASE_URL}/admin/cartoons/${cartoon.id}`, {
          headers: {
            Authorization: `Bearer ${state.token}`,
            "Content-Type": "application/json",
          },
        })
          .then((response) => {
            if (response.ok) {
              return response.json().then((data) => {
                console.log(`Fetched cartoon ${cartoon.id}:`, data.data);
                return data.data;
              });
            } else {
              console.warn(
                `Failed to fetch cartoon ${cartoon.id}: ${response.status}`
              );
              return cartoon;
            }
          })
          .catch((err) => {
            console.error(`Error fetching cartoon ${cartoon.id}:`, err);
            return cartoon;
          })
      );

      const allCartoons = await Promise.all(cartoonPromises);

      console.log("All cartoons loaded:", allCartoons);
      state.cartoons = allCartoons;
      state.filteredCartoons = [...state.cartoons];

      // Extract unique genres and age groups from cartoons
      extractGenresAndAgeGroups();

      // Populate filter dropdowns
      populateFilterDropdowns();

      // Render cartoons
      renderCartoons();

      // Hide loading
      hideLoading("cartoons-loading");
    } else {
      throw new Error("Failed to fetch cartoon names");
    }
  } catch (error) {
    console.error("Cartoons error:", error);
    hideLoading("cartoons-loading");
    showErrorMessage(
      "Failed to load cartoons: " + error.message,
      "cartoons-error"
    );
  }
}

/**
 * Update dashboard statistics
 */
function updateDashboardStats() {
  document.getElementById("total-cartoons").textContent = state.cartoons.length;
  // Note: Characters count would need a separate API call or calculation
  document.getElementById("total-characters").textContent = "0";
}

/**
 * Render cartoons as cards with pagination
 */
function renderCartoons() {
  const cardsContainer = document.getElementById("cartoons-cards");
  const noCartoons = document.getElementById("no-cartoons");
  const paginationSection = document.getElementById("pagination-section");
  const resultsCount = document.getElementById("results-count");

  // Update results count
  resultsCount.textContent = state.filteredCartoons.length;

  // Show/hide no results
  if (state.filteredCartoons.length === 0) {
    cardsContainer.innerHTML = "";
    noCartoons.classList.remove("hidden");
    paginationSection.classList.add("hidden");
    return;
  }

  noCartoons.classList.add("hidden");

  // Calculate pagination
  state.totalPages = Math.ceil(
    state.filteredCartoons.length / state.itemsPerPage
  );
  state.currentPage = Math.max(
    1,
    Math.min(state.currentPage, state.totalPages)
  );

  // Get paginated cartoons
  const startIdx = (state.currentPage - 1) * state.itemsPerPage;
  const endIdx = startIdx + state.itemsPerPage;
  const paginatedCartoons = state.filteredCartoons.slice(startIdx, endIdx);

  // Render cards
  cardsContainer.innerHTML = paginatedCartoons
    .map(
      (cartoon) => `
        <div class="cartoon-card" onclick="openDetailModal(${cartoon.id})">
            <div class="cartoon-card-image">
                <img src="${
                  cartoon.poster_url ||
                  "https://via.placeholder.com/300x400?text=No+Image"
                }" alt="${cartoon.title}">
            </div>
            <div class="cartoon-card-info">
                <h3>${cartoon.title}</h3>
            </div>
        </div>
    `
    )
    .join("");

  // Update pagination controls
  updatePaginationControls();
}

/**
 * Update pagination controls
 */
function updatePaginationControls() {
  const paginationSection = document.getElementById("pagination-section");
  const currentPageSpan = document.getElementById("current-page");
  const totalPagesSpan = document.getElementById("total-pages");
  const prevBtn = document.getElementById("prev-btn");
  const nextBtn = document.getElementById("next-btn");

  if (state.totalPages <= 1) {
    paginationSection.classList.add("hidden");
    return;
  }

  paginationSection.classList.remove("hidden");
  currentPageSpan.textContent = state.currentPage;
  totalPagesSpan.textContent = state.totalPages;

  prevBtn.disabled = state.currentPage <= 1;
  nextBtn.disabled = state.currentPage >= state.totalPages;
}

/**
 * Go to next page
 */
function nextPage() {
  if (state.currentPage < state.totalPages) {
    state.currentPage++;
    renderCartoons();
    window.scrollTo({ top: 0, behavior: "smooth" });
  }
}

/**
 * Go to previous page
 */
function previousPage() {
  if (state.currentPage > 1) {
    state.currentPage--;
    renderCartoons();
    window.scrollTo({ top: 0, behavior: "smooth" });
  }
}

/**
 * Apply filters to cartoons
 */
function applyFilters() {
  const searchText = document
    .getElementById("search-input")
    .value.toLowerCase();
  const genreId = document.getElementById("genre-filter").value;
  const ageGroupId = document.getElementById("age-group-filter").value;
  const featured = document.getElementById("featured-filter").value;

  state.filteredCartoons = state.cartoons.filter((cartoon) => {
    // Search by title
    if (searchText && !cartoon.title.toLowerCase().includes(searchText)) {
      return false;
    }

    // Filter by genre
    if (genreId && cartoon.genre_id != genreId) {
      return false;
    }

    // Filter by age group
    if (ageGroupId && cartoon.age_group_id != ageGroupId) {
      return false;
    }

    // Filter by featured status
    if (featured !== "" && String(cartoon.is_featured) !== featured) {
      return false;
    }

    return true;
  });

  // Reset to first page when filters are applied
  state.currentPage = 1;
  renderCartoons();
}

/**
 * Clear all filters
 */
function clearFilters() {
  document.getElementById("search-input").value = "";
  document.getElementById("genre-filter").value = "";
  document.getElementById("age-group-filter").value = "";
  document.getElementById("featured-filter").value = "";
  applyFilters();
}

// ============================================
// MODALS
// ============================================

/**
 * Open detail modal for a cartoon
 */
function openDetailModal(cartoonId) {
  const cartoon = state.cartoons.find((c) => c.id === cartoonId);
  if (!cartoon) {
    alert("Cartoon not found");
    return;
  }

  // Populate detail modal with cartoon data
  document.getElementById("detail-id").textContent = cartoon.id;
  document.getElementById("detail-title-text").textContent = cartoon.title;
  document.getElementById("detail-year").textContent =
    cartoon.release_year || "-";
  document.getElementById("detail-genre").textContent = cartoon.genre
    ? cartoon.genre.name
    : "-";
  document.getElementById("detail-age-group").textContent = cartoon.age_group
    ? cartoon.age_group.label || cartoon.age_group.name
    : "-";
  document.getElementById("detail-featured").textContent = cartoon.is_featured
    ? "Yes"
    : "No";
  document.getElementById("detail-description").textContent =
    cartoon.description || "No description available";

  // Set poster image if available
  if (cartoon.poster_url) {
    document.getElementById("detail-poster").src = cartoon.poster_url;
    document.getElementById("detail-poster").style.display = "block";
  }

  // Display characters
  const charactersContainer = document.getElementById("detail-characters");
  if (cartoon.characters && cartoon.characters.length > 0) {
    charactersContainer.innerHTML = cartoon.characters
      .map(
        (char) => `
                <div class="character-tag" onclick="showCharacterImage('${
                  char.image_url
                }', '${char.name}')">
                    ${
                      char.image_url
                        ? `<img src="${char.image_url}" alt="${char.name}" onerror="this.style.display='none'">`
                        : ""
                    }
                    <span>${char.name}</span>
                </div>
            `
      )
      .join("");
  } else {
    charactersContainer.innerHTML = "<p>No characters found</p>";
  }

  // Store current cartoon for editing/deleting
  state.detailCartoonId = cartoonId;

  // Open the modal
  openModal("detail-modal");
}

/**
 * Close detail modal
 */
function closeDetailModal() {
  closeModal("detail-modal");
  state.detailCartoonId = null;
}

/**
 * Edit cartoon from detail modal
 */
function editDetailCartoon() {
  if (state.detailCartoonId) {
    // Don't close the detail modal, just open the edit modal on top
    // This will stack the modals
    openEditCartoonModal(state.detailCartoonId);
  }
}

/**
 * Delete cartoon from detail modal
 */
function deleteDetailCartoon() {
  // Get the ID and title BEFORE closing the modal
  const cartoonId = state.detailCartoonId;
  const cartoon = state.cartoons.find((c) => c.id === cartoonId);

  if (cartoon) {
    // Close the detail modal but don't reset state yet
    closeModal("detail-modal");
    // Reset state after saving the ID
    state.detailCartoonId = null;
    // Now open delete modal with the saved info
    openDeleteModal("cartoon", cartoonId, cartoon.title);
  }
}

/**
 * Open add cartoon modal
 */
function openAddCartoonModal() {
  resetCartoonForm();
  state.editingCartoonId = null;
  state.charactersToSave = [];
  document.getElementById("modal-title").textContent = "Add Cartoon";
  document.getElementById("characters-container").innerHTML = ""; // Clear characters
  openModal("cartoon-modal");
}

/**
 * Open edit cartoon modal
 */
async function openEditCartoonModal(cartoonId) {
  const cartoon = state.cartoons.find((c) => c.id === cartoonId);
  if (!cartoon) return;

  // Close the modal first if it's already open to prevent duplication
  const modal = document.getElementById("cartoon-modal");
  if (!modal.classList.contains("hidden")) {
    closeModal("cartoon-modal");
  }

  state.editingCartoonId = cartoonId;
  state.charactersToSave = [];
  document.getElementById("modal-title").textContent = "Edit Cartoon";

  // Populate form with existing cartoon data
  document.getElementById("cartoon-title").value = cartoon.title;
  document.getElementById("cartoon-year").value = cartoon.release_year;
  document.getElementById("cartoon-description").value = cartoon.description;
  document.getElementById("cartoon-poster").value = cartoon.poster_url;
  document.getElementById("cartoon-genre").value = cartoon.genre_id || "";
  document.getElementById("cartoon-age-group").value =
    cartoon.age_group_id || "";
  document.getElementById("cartoon-featured").checked =
    cartoon.is_featured || false;

  // Ensure characters container is completely cleared before adding new ones
  const container = document.getElementById("characters-container");
  container.innerHTML = "";

  // Small delay to ensure DOM is cleared before adding new elements
  setTimeout(() => {
    if (cartoon.characters && cartoon.characters.length > 0) {
      cartoon.characters.forEach((char) => {
        addCharacterField(char);
      });
    }
    openModal("cartoon-modal");
  }, 10);
}

/**
 * Reset cartoon form
 */
function resetCartoonForm() {
  document.getElementById("cartoon-form").reset();
  document.getElementById("cartoon-error").classList.add("hidden");
}

/**
 * Open add character modal
 */
function openAddCharacterModal() {
  document.getElementById("character-form").reset();
  document.getElementById("character-error").classList.add("hidden");

  // Populate cartoon select
  const cartoonSelect = document.getElementById("character-cartoon");
  cartoonSelect.innerHTML = '<option value="">Select Cartoon</option>';
  state.cartoons.forEach((cartoon) => {
    const option = document.createElement("option");
    option.value = cartoon.id;
    option.textContent = cartoon.title;
    cartoonSelect.appendChild(option);
  });

  openModal("character-modal");
}

/**
 * Open modal
 */
function openModal(modalId) {
  document.getElementById(modalId).classList.remove("hidden");
}

/**
 * Close modal
 */
function closeModal(modalId) {
  document.getElementById(modalId).classList.add("hidden");

  // Clear characters container when closing cartoon modal to prevent duplication
  if (modalId === "cartoon-modal") {
    document.getElementById("characters-container").innerHTML = "";
  }
}

// ============================================
// CHARACTER MANAGEMENT
// ============================================

/**
 * Show character image in modal
 */
function showCharacterImage(imageUrl, characterName) {
  document.getElementById("character-modal-image").src = imageUrl;
  document.getElementById("character-modal-name").textContent = characterName;
  document.getElementById("character-image-modal").classList.add("show");
}

/**
 * Close character image modal
 */
function closeCharacterImageModal() {
  document.getElementById("character-image-modal").classList.remove("show");
}

/**
 * Add character input field
 */
function addCharacterField(existingCharacter = null) {
  const container = document.getElementById("characters-container");
  const fieldId = `character-field-${Date.now()}`;

  const div = document.createElement("div");
  div.className = "character-field";
  div.id = fieldId;

  const nameGroup = document.createElement("div");
  nameGroup.className = "form-group";
  nameGroup.innerHTML = `
        <label>Character Name</label>
        <input type="text" class="character-name" placeholder="e.g., Mickey Mouse" value="${
          existingCharacter?.name || ""
        }" required>
    `;

  const imageGroup = document.createElement("div");
  imageGroup.className = "form-group";
  imageGroup.innerHTML = `
        <label>Image URL</label>
        <input type="url" class="character-image" placeholder="https://example.com/image.jpg" value="${
          existingCharacter?.image_url || ""
        }" required>
    `;

  const removeBtn = document.createElement("button");
  removeBtn.type = "button";
  removeBtn.className = "btn btn-danger btn-small btn-remove";
  removeBtn.textContent = "Remove";
  removeBtn.onclick = (e) => {
    e.preventDefault();
    e.stopPropagation();
    removeCharacterField(fieldId);
  };

  div.appendChild(nameGroup);
  div.appendChild(imageGroup);
  div.appendChild(removeBtn);

  container.appendChild(div);
}

/**
 * Remove character input field
 */
function removeCharacterField(fieldId) {
  const field = document.getElementById(fieldId);
  if (field) {
    field.remove();
  }
}

/**
 * Get all character data from form
 */
function getCharactersFromForm() {
  const characters = [];
  const container = document.getElementById("characters-container");
  const fields = container.querySelectorAll(".character-field");

  fields.forEach((field) => {
    const name = field.querySelector(".character-name").value.trim();
    const imageUrl = field.querySelector(".character-image").value.trim();

    if (name && imageUrl) {
      characters.push({
        name: name,
        image_url: imageUrl,
      });
    }
  });

  return characters;
}

// ============================================
// FORM SUBMISSIONS
// ============================================

/**
 * Handle cartoon form submission
 */
async function handleCartoonSubmit(event) {
  event.preventDefault();

  const cartoonData = {
    title: document.getElementById("cartoon-title").value,
    release_year: parseInt(document.getElementById("cartoon-year").value),
    description: document.getElementById("cartoon-description").value,
    poster_url: document.getElementById("cartoon-poster").value,
    genre_id: parseInt(document.getElementById("cartoon-genre").value),
    age_group_id: parseInt(document.getElementById("cartoon-age-group").value),
    is_featured: document.getElementById("cartoon-featured").checked,
  };

  // Get characters from form
  const characters = getCharactersFromForm();

  try {
    // Ensure token is available
    if (!state.token) {
      throw new Error("Not authenticated. Please login again.");
    }

    console.log(
      "Submitting cartoon with token:",
      state.token.substring(0, 20) + "..."
    );
    console.log("Editing cartoon ID:", state.editingCartoonId);

    let cartoonId;
    let response;

    if (state.editingCartoonId) {
      // Update existing cartoon
      cartoonId = state.editingCartoonId;
      response = await fetch(
        `${API_BASE_URL}/admin/cartoons/${state.editingCartoonId}`,
        {
          method: "PUT",
          headers: {
            Authorization: `Bearer ${state.token}`,
            "Content-Type": "application/json",
          },
          body: JSON.stringify(cartoonData),
        }
      );
    } else {
      // Create new cartoon
      response = await fetch(`${API_BASE_URL}/admin/cartoons`, {
        method: "POST",
        headers: {
          Authorization: `Bearer ${state.token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify(cartoonData),
      });
    }

    const data = await response.json();

    if (!response.ok) {
      if (response.status === 403) {
        throw new Error("Access denied. Admin privileges required.");
      }
      throw new Error(data.message || "Failed to save cartoon");
    }

    // Get the cartoon ID (for new cartoons, it's in the response)
    if (!state.editingCartoonId) {
      cartoonId = data.id || data.data?.id;
    }

    // If editing, delete all existing characters first to avoid duplicates
    if (state.editingCartoonId && cartoonId) {
      try {
        console.log(`Deleting existing characters for cartoon ID ${cartoonId}`);
        const existingCharsResponse = await fetch(
          `${API_BASE_URL}/admin/characters/cartoon/${cartoonId}`,
          {
            headers: {
              Authorization: `Bearer ${state.token}`,
            },
          }
        );

        if (existingCharsResponse.ok) {
          const existingCharsData = await existingCharsResponse.json();
          const existingCharacters = existingCharsData.characters || [];

          // Delete each existing character
          for (const existingChar of existingCharacters) {
            try {
              await fetch(
                `${API_BASE_URL}/admin/characters/${existingChar.id}`,
                {
                  method: "DELETE",
                  headers: {
                    Authorization: `Bearer ${state.token}`,
                  },
                }
              );
              console.log(`Deleted character: ${existingChar.name}`);
            } catch (deleteError) {
              console.warn(
                `Error deleting character ${existingChar.name}:`,
                deleteError
              );
            }
          }
        }
      } catch (error) {
        console.warn("Error fetching/deleting existing characters:", error);
      }
    }

    // Save characters if any
    if (characters.length > 0 && cartoonId) {
      console.log(
        `Saving ${characters.length} characters for cartoon ID ${cartoonId}`
      );
      for (const character of characters) {
        try {
          console.log(`Saving character: ${character.name}`);
          const charResponse = await fetch(`${API_BASE_URL}/admin/characters`, {
            method: "POST",
            headers: {
              Authorization: `Bearer ${state.token}`,
              "Content-Type": "application/json",
            },
            body: JSON.stringify({
              name: character.name,
              image_url: character.image_url,
              cartoon_id: cartoonId,
            }),
          });

          const charData = await charResponse.json();

          if (!charResponse.ok) {
            console.error(
              `Failed to save character ${character.name}:`,
              charResponse.status,
              charData
            );
          } else {
            console.log(
              `Successfully saved character ${character.name}:`,
              charData
            );
          }
        } catch (charError) {
          console.warn(`Error saving character ${character.name}:`, charError);
        }
      }
    }

    const isEditing = state.editingCartoonId !== null;
    state.editingCartoonId = null;
    closeModal("cartoon-modal");
    resetCartoonForm();
    showSuccess(`Cartoon ${isEditing ? "updated" : "created"} successfully!`);
    await loadCartoons();
  } catch (error) {
    console.error("Cartoon submission error:", error);
    document.getElementById("cartoon-error").textContent = error.message;
    document.getElementById("cartoon-error").classList.remove("hidden");
  }
}

/**
 * Handle character form submission
 */
async function handleCharacterSubmit(event) {
  event.preventDefault();

  const characterData = {
    name: document.getElementById("character-name").value,
    cartoon_id: parseInt(document.getElementById("character-cartoon").value),
  };

  try {
    const response = await fetch(`${API_BASE_URL}/admin/characters`, {
      method: "POST",
      headers: {
        Authorization: `Bearer ${state.token}`,
        "Content-Type": "application/json",
      },
      body: JSON.stringify(characterData),
    });

    const data = await response.json();

    if (!response.ok) {
      throw new Error(data.message || "Failed to add character");
    }

    closeModal("character-modal");
    showSuccess("Character added successfully!");
  } catch (error) {
    console.error("Character submission error:", error);
    document.getElementById("character-error").textContent = error.message;
    document.getElementById("character-error").classList.remove("hidden");
  }
}

// ============================================
// DELETE OPERATIONS
// ============================================

/**
 * Open delete confirmation modal
 */
function openDeleteModal(type, id, name) {
  state.deleteTarget = { type, id, name };
  const message = `Are you sure you want to delete "${name}"? This action cannot be undone.`;
  document.getElementById("delete-message").textContent = message;
  openModal("delete-modal");
}

/**
 * Confirm and execute delete
 */
async function confirmDelete() {
  if (!state.deleteTarget) return;

  const { type, id, name } = state.deleteTarget;

  try {
    // Ensure token is available
    if (!state.token) {
      throw new Error("Not authenticated. Please login again.");
    }

    console.log(`Deleting ${type} with ID: ${id}`);

    // For cartoons, use query parameter
    const url =
      type === "cartoon"
        ? `${API_BASE_URL}/admin/cartoons?id=${id}`
        : `${API_BASE_URL}/admin/${type}s/${id}`;

    const response = await fetch(url, {
      method: "DELETE",
      headers: {
        Authorization: `Bearer ${state.token}`,
        "Content-Type": "application/json",
      },
    });

    const data = await response.json();

    if (!response.ok) {
      if (response.status === 403) {
        throw new Error("Access denied. Admin privileges required.");
      }
      throw new Error(data.message || `Failed to delete ${type}`);
    }

    closeModal("delete-modal");
    showSuccess(
      `${type.charAt(0).toUpperCase() + type.slice(1)} deleted successfully!`
    );
    state.deleteTarget = null;

    // Reload data based on type
    if (type === "cartoon") {
      await loadCartoons();
      closeDetailModal();
    }
  } catch (error) {
    console.error("Delete error:", error);
    document.getElementById(
      "delete-message"
    ).textContent = `Error: ${error.message}`;
  }
}

// ============================================
// UTILITY FUNCTIONS
// ============================================

/**
 * Handle logout
 */
function handleLogout() {
  localStorage.removeItem("token");
  localStorage.removeItem("user-email");
  localStorage.removeItem("disney_auth_token");
  localStorage.removeItem("disney_user_data");
  window.location.href = "./index.html";
}

// ============================================
// UI HELPERS
// ============================================

/**
 * Show loading indicator
 */
function showLoading(elementId) {
  const element = document.getElementById(elementId);
  if (element) {
    element.classList.remove("hidden");
  }
}

/**
 * Hide loading indicator
 */
function hideLoading(elementId) {
  const element = document.getElementById(elementId);
  if (element) {
    element.classList.add("hidden");
  }
}

/**
 * Show success message
 */
function showSuccess(message) {
  const element = document.getElementById("success-message");
  document.getElementById("success-text").textContent = message;
  element.classList.remove("hidden");

  setTimeout(() => {
    element.classList.add("hidden");
  }, 3000);
}

/**
 * Show error message
 */
function showError(message) {
  const element = document.getElementById("error-alert");
  const textEl = document.getElementById("error-text");

  if (!element || !textEl) return;

  textEl.textContent = message;
  element.classList.remove("hidden");

  setTimeout(() => {
    element.classList.add("hidden");
  }, 4000);
}


/**
 * Show error in specific section
 */
function showErrorMessage(message, elementId) {
  const element = document.getElementById(elementId);
  if (element) {
    element.textContent = message;
    element.classList.remove("hidden");
  }
}

// ============================================
// DEBUG
// ============================================

console.log("Admin Dashboard Loaded");
console.log("API Base URL:", typeof API_BASE_URL !== 'undefined' ? API_BASE_URL : 'Not defined yet');
