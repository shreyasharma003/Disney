// Admin Logs & Analytics JavaScript

const API_BASE_URL = "http://localhost:8080/api";

// State Management
const state = {
  logs: [],
  filteredLogs: [],
  token: null,
  userId: null,
  stats: {
    total: 0,
    create: 0,
    update: 0,
    delete: 0,
  },
};

// ============================================
// INITIALIZATION
// ============================================

window.addEventListener("DOMContentLoaded", () => {
  // Check authentication
  const token = localStorage.getItem("disney_auth_token");
  if (!token) {
    window.location.href = "./admin-login.html";
    return;
  }

  state.token = token;

  // Display user email
  const userData = localStorage.getItem("disney_user_data");
  if (userData) {
    try {
      const user = JSON.parse(userData);
      document.getElementById("user-email").textContent = user.email || "User";
      state.userId = user.id;

      // Check if user is admin
      if (user.role !== "admin") {
        alert("Access denied. Admin privileges required.");
        window.location.href = "./admin-dashboard.html";
        return;
      }
    } catch (error) {
      console.error("Error parsing user data:", error);
    }
  }

  // Load logs
  loadAdminLogs();

  // Set today's date as default in date filter
  const today = new Date().toISOString().split("T")[0];
  document.getElementById("date-to").value = today;

  // Auto-refresh logs every 3 seconds to show new entries
  setInterval(() => {
    loadAdminLogs();
  }, 3000);
});

// ============================================
// LOAD ADMIN LOGS
// ============================================

/**
 * Load all admin logs from backend
 * Note: Backend endpoint would be /api/admin/logs
 * Since it doesn't exist yet, we'll make a POST request to log actions
 */
async function loadAdminLogs() {
  const logsLoading = document.getElementById("logs-loading");
  const logsError = document.getElementById("logs-error");
  const logsTable = document.getElementById("logs-tbody");
  const noLogs = document.getElementById("no-logs");

  logsLoading.classList.remove("hidden");
  logsError.classList.add("hidden");
  noLogs.classList.add("hidden");

  try {
    // Fetch logs from backend
    const response = await fetch(`${API_BASE_URL}/admin/logs`, {
      method: "GET",
      headers: {
        Authorization: `Bearer ${state.token}`,
        "Content-Type": "application/json",
      },
    });

    console.log("Admin logs response status:", response.status);
    console.log("Admin logs response:", response);

    let logs = [];

    if (response.ok) {
      const data = await response.json();
      console.log("Admin logs data:", data);
      logs = data.data || [];
    } else if (response.status === 404) {
      // If endpoint doesn't exist, load from localStorage as fallback
      console.log("Logs endpoint returned 404, falling back to localStorage");
      const storedLogs = localStorage.getItem("admin_logs");
      logs = storedLogs ? JSON.parse(storedLogs) : [];
    } else {
      throw new Error(`Failed to fetch logs: ${response.statusText}`);
    }

    console.log("Logs loaded:", logs);
    state.logs = logs;
    state.filteredLogs = logs;

    // Calculate statistics
    calculateStats();

    // Render logs
    renderLogs();

    logsLoading.classList.add("hidden");

    if (logs.length === 0) {
      noLogs.classList.remove("hidden");
    }
  } catch (error) {
    console.error("Error loading logs:", error);
    logsLoading.classList.add("hidden");

    // Fallback to localStorage
    const storedLogs = localStorage.getItem("admin_logs");
    state.logs = storedLogs ? JSON.parse(storedLogs) : [];
    state.filteredLogs = state.logs;

    if (state.logs.length > 0) {
      console.log("Using localStorage logs:", state.logs);
      calculateStats();
      renderLogs();
    } else {
      logsError.textContent = "Failed to load admin logs. No data available.";
      logsError.classList.remove("hidden");
    }
  }
}

/**
 * Log an admin action
 */
async function logAdminAction(action, entity, details = "") {
  try {
    const logData = {
      admin_id: state.userId,
      action: action,
      entity: entity,
      details: details,
      timestamp: new Date().toISOString(),
    };

    // Try to send to backend
    // This assumes backend has POST /api/admin/logs endpoint
    try {
      await fetch(`${API_BASE_URL}/admin/logs`, {
        method: "POST",
        headers: {
          Authorization: `Bearer ${state.token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify(logData),
      });
    } catch (error) {
      console.warn("Could not send log to backend:", error);
    }

    // Also save to localStorage as fallback
    const logs = state.logs || [];
    logs.unshift(logData); // Add to beginning (newest first)
    state.logs = logs;
    localStorage.setItem("admin_logs", JSON.stringify(logs));
  } catch (error) {
    console.error("Error logging action:", error);
  }
}

// ============================================
// RENDER LOGS
// ============================================

/**
 * Render logs table
 */
function renderLogs() {
  const logsTable = document.getElementById("logs-tbody");
  const noLogs = document.getElementById("no-logs");

  if (state.filteredLogs.length === 0) {
    logsTable.innerHTML =
      '<tr><td colspan="5" class="no-data">No logs found</td></tr>';
    noLogs.classList.remove("hidden");
    return;
  }

  noLogs.classList.add("hidden");

  logsTable.innerHTML = state.filteredLogs
    .map((log) => {
      const date = new Date(log.created_at || log.timestamp);
      const formattedDate = date.toLocaleString("en-US", {
        year: "numeric",
        month: "short",
        day: "2-digit",
        hour: "2-digit",
        minute: "2-digit",
        second: "2-digit",
      });

      const actionClass = `action-${log.action.toLowerCase()}`;

      // Handle both backend response format and localStorage format
      let adminName = "Unknown Admin";
      if (log.admin) {
        if (typeof log.admin === "object" && log.admin.email) {
          adminName = log.admin.email;
        } else if (typeof log.admin === "string") {
          adminName = log.admin;
        }
      }

      return `
            <tr>
                <td>${log.id}</td>
                <td>${adminName}</td>
                <td><span class="action-badge ${actionClass}">${log.action}</span></td>
                <td><span class="entity-badge">${log.entity}</span></td>
                <td>${formattedDate}</td>
            </tr>
        `;
    })
    .join("");
}

// ============================================
// FILTERS
// ============================================

/**
 * Apply filters
 */
function applyFilters() {
  const actionFilter = document.getElementById("action-filter").value;
  const entityFilter = document.getElementById("entity-filter").value;
  const dateFrom = document.getElementById("date-from").value;
  const dateTo = document.getElementById("date-to").value;

  state.filteredLogs = state.logs.filter((log) => {
    // Filter by action
    if (actionFilter && log.action !== actionFilter) {
      return false;
    }

    // Filter by entity
    if (entityFilter && log.entity !== entityFilter) {
      return false;
    }

    // Filter by date range
    if (dateFrom || dateTo) {
      const logDate = new Date(log.created_at || log.timestamp)
        .toISOString()
        .split("T")[0];

      if (dateFrom && logDate < dateFrom) {
        return false;
      }

      if (dateTo && logDate > dateTo) {
        return false;
      }
    }

    return true;
  });

  calculateStats();
  renderLogs();
}

/**
 * Clear all filters
 */
function clearFilters() {
  document.getElementById("action-filter").value = "";
  document.getElementById("entity-filter").value = "";
  document.getElementById("date-from").value = "";
  document.getElementById("date-to").value = "";

  state.filteredLogs = [...state.logs];
  calculateStats();
  renderLogs();
}

// ============================================
// STATISTICS
// ============================================

/**
 * Calculate statistics from filtered logs
 */
function calculateStats() {
  const logs = state.filteredLogs;

  state.stats = {
    total: logs.length,
    create: logs.filter((log) => log.action === "CREATE").length,
    update: logs.filter((log) => log.action === "UPDATE").length,
    delete: logs.filter((log) => log.action === "DELETE").length,
  };

  // Update UI
  document.getElementById("total-logs").textContent = state.stats.total;
  document.getElementById("create-count").textContent = state.stats.create;
  document.getElementById("update-count").textContent = state.stats.update;
  document.getElementById("delete-count").textContent = state.stats.delete;
}

// ============================================
// NAVIGATION
// ============================================

/**
 * Go back to admin dashboard
 */
function goBack() {
  window.location.href = "./admin-dashboard.html";
}

/**
 * Handle logout
 */
function handleLogout() {
  localStorage.removeItem("disney_auth_token");
  localStorage.removeItem("disney_user_data");
  window.location.href = "./admin-login.html";
}

// Export function to be used from admin-dashboard.js
function logToAnalytics(action, entity, details = "") {
  logAdminAction(action, entity, details);
}

