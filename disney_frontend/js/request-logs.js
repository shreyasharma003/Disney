// Request Logs JavaScript

// API_BASE_URL is defined in config.js which is loaded before this script

// State Management
const state = {
  logs: [],
  filteredLogs: [],
  token: null,
  userId: null,
  stats: {
    total: 0,
    success: 0,
    clientError: 0,
    serverError: 0,
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
  loadRequestLogs();

  // Set today's date as default in date filter
  const today = new Date().toISOString().split("T")[0];
  document.getElementById("date-to").value = today;

  // Auto-refresh logs every 5 seconds
  setInterval(() => {
    loadRequestLogs();
  }, 5000);
});

// ============================================
// LOAD REQUEST LOGS
// ============================================

async function loadRequestLogs() {
  const logsLoading = document.getElementById("logs-loading");
  const logsError = document.getElementById("logs-error");
  const logsTable = document.getElementById("logs-tbody");
  const noLogs = document.getElementById("no-logs");

  logsLoading.classList.remove("hidden");
  logsError.classList.add("hidden");
  noLogs.classList.add("hidden");

  try {
    // Fetch logs from backend
    const response = await fetch(`${API_BASE_URL}/admin/request-logs`, {
      method: "GET",
      headers: {
        Authorization: `Bearer ${state.token}`,
        "Content-Type": "application/json",
      },
    });

    let logs = [];

    if (response.ok) {
      const data = await response.json();
      logs = data.data || data.logs || [];
      console.log("Fetched request logs:", logs);
    } else {
      console.error("Failed to fetch request logs:", response.status);
    }

    state.logs = logs;
    state.filteredLogs = logs;

    // Update statistics
    updateStats();

    // Apply any active filters
    applyFilters();

    logsLoading.classList.add("hidden");
  } catch (error) {
    console.error("Error loading request logs:", error);
    logsLoading.classList.add("hidden");
    logsError.textContent = "Failed to load request logs. Please try again.";
    logsError.classList.remove("hidden");
  }
}

// ============================================
// UPDATE STATISTICS
// ============================================

function updateStats() {
  const total = state.logs.length;
  const success = state.logs.filter(
    (log) => log.status_code >= 200 && log.status_code < 300
  ).length;
  const clientError = state.logs.filter(
    (log) => log.status_code >= 400 && log.status_code < 500
  ).length;
  const serverError = state.logs.filter((log) => log.status_code >= 500).length;

  state.stats = { total, success, clientError, serverError };

  document.getElementById("total-requests").textContent = total;
  document.getElementById("success-count").textContent = success;
  document.getElementById("client-error-count").textContent = clientError;
  document.getElementById("server-error-count").textContent = serverError;
}

// ============================================
// FILTERS
// ============================================

function applyFilters() {
  const methodFilter = document.getElementById("method-filter").value;
  const statusFilter = document.getElementById("status-filter").value;
  const dateFrom = document.getElementById("date-from").value;
  const dateTo = document.getElementById("date-to").value;

  let filtered = [...state.logs];

  // Filter by method
  if (methodFilter) {
    filtered = filtered.filter((log) => log.method === methodFilter);
  }

  // Filter by status code range
  if (statusFilter === "2xx") {
    filtered = filtered.filter(
      (log) => log.status_code >= 200 && log.status_code < 300
    );
  } else if (statusFilter === "4xx") {
    filtered = filtered.filter(
      (log) => log.status_code >= 400 && log.status_code < 500
    );
  } else if (statusFilter === "5xx") {
    filtered = filtered.filter((log) => log.status_code >= 500);
  }

  // Filter by date range
  if (dateFrom) {
    filtered = filtered.filter((log) => {
      const logDate = new Date(log.created_at).toISOString().split("T")[0];
      return logDate >= dateFrom;
    });
  }

  if (dateTo) {
    filtered = filtered.filter((log) => {
      const logDate = new Date(log.created_at).toISOString().split("T")[0];
      return logDate <= dateTo;
    });
  }

  state.filteredLogs = filtered;
  renderLogs();
}

function clearFilters() {
  document.getElementById("method-filter").value = "";
  document.getElementById("status-filter").value = "";
  document.getElementById("date-from").value = "";
  document.getElementById("date-to").value = "";
  applyFilters();
}

// ============================================
// RENDER LOGS
// ============================================

function renderLogs() {
  const tbody = document.getElementById("logs-tbody");
  const noLogs = document.getElementById("no-logs");

  tbody.innerHTML = "";

  if (state.filteredLogs.length === 0) {
    noLogs.classList.remove("hidden");
    return;
  }

  noLogs.classList.add("hidden");

  state.filteredLogs.forEach((log) => {
    const tr = document.createElement("tr");

    // ID
    const tdId = document.createElement("td");
    tdId.textContent = log.id;
    tr.appendChild(tdId);

    // Method
    const tdMethod = document.createElement("td");
    const methodBadge = document.createElement("span");
    methodBadge.className = `method-badge method-${log.method.toLowerCase()}`;
    methodBadge.textContent = log.method;
    tdMethod.appendChild(methodBadge);
    tr.appendChild(tdMethod);

    // Endpoint
    const tdEndpoint = document.createElement("td");
    tdEndpoint.textContent = log.endpoint || log.path || "-";
    tdEndpoint.style.maxWidth = "300px";
    tdEndpoint.style.overflow = "hidden";
    tdEndpoint.style.textOverflow = "ellipsis";
    tdEndpoint.style.whiteSpace = "nowrap";
    tr.appendChild(tdEndpoint);

    // User
    const tdUser = document.createElement("td");
    tdUser.textContent = log.user_email || log.user?.email || "Anonymous";
    tr.appendChild(tdUser);

    // Status
    const tdStatus = document.createElement("td");
    const statusBadge = document.createElement("span");
    statusBadge.className = `status-badge status-${getStatusClass(
      log.status_code
    )}`;
    statusBadge.textContent = log.status_code;
    tdStatus.appendChild(statusBadge);
    tr.appendChild(tdStatus);

    // Date & Time
    const tdDate = document.createElement("td");
    const date = new Date(log.created_at);
    tdDate.textContent = date.toLocaleString();
    tr.appendChild(tdDate);

    tbody.appendChild(tr);
  });
}

function getStatusClass(statusCode) {
  if (statusCode >= 200 && statusCode < 300) return "success";
  if (statusCode >= 400 && statusCode < 500) return "client-error";
  if (statusCode >= 500) return "server-error";
  return "info";
}

// ============================================
// UTILITY FUNCTIONS
// ============================================

function goBack() {
  window.location.href = "./admin-dashboard.html";
}

function handleLogout() {
  localStorage.removeItem("token");
  localStorage.removeItem("user-email");
  localStorage.removeItem("disney_auth_token");
  localStorage.removeItem("disney_user_data");
  window.location.href = "./index.html";
}
