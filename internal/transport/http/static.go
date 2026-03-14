package http

import (
	"net/http"
)

// UIHandler serves the web management interface
func UIHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(getManagementUI()))
	}
}

func getManagementUI() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Office Tracker - Management</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: #f5f5f5;
            padding: 20px;
        }
        .container { max-width: 1400px; margin: 0 auto; }
        header {
            background: white;
            padding: 20px;
            border-radius: 8px;
            margin-bottom: 20px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        h1 { color: #333; font-size: 24px; margin-bottom: 5px; }
        .subtitle { color: #666; font-size: 14px; }
        .tabs {
            display: flex;
            gap: 10px;
            margin-bottom: 20px;
        }
        .tab {
            padding: 10px 20px;
            background: white;
            border: 1px solid #ddd;
            border-radius: 6px;
            cursor: pointer;
            font-size: 14px;
            transition: all 0.2s;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
            color: #333;
        }
        .tab:hover { background: #f0f0f0; }
        .tab.active { background: #007bff; color: white; border-color: #007bff; }
        .panel {
            display: none;
            background: white;
            padding: 20px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .panel.active { display: block; }
        .form-group {
            margin-bottom: 15px;
        }
        label {
            display: block;
            margin-bottom: 5px;
            color: #333;
            font-weight: 500;
            font-size: 14px;
        }
        input, select {
            width: 100%;
            padding: 10px;
            border: 1px solid #ddd;
            border-radius: 4px;
            font-size: 14px;
        }
        input:focus, select:focus {
            outline: none;
            border-color: #007bff;
        }
        button {
            padding: 10px 20px;
            background: #007bff;
            color: white;
            border: none;
            border-radius: 4px;
            cursor: pointer;
            font-size: 14px;
            transition: background 0.2s;
        }
        button:hover { background: #0056b3; }
        button.secondary {
            background: #6c757d;
        }
        button.secondary:hover { background: #545b62; }
        button.danger { background: #dc3545; }
        button.danger:hover { background: #c82333; }
        table {
            width: 100%;
            border-collapse: collapse;
            margin-top: 20px;
        }
        th, td {
            padding: 12px;
            text-align: left;
            border-bottom: 1px solid #ddd;
        }
        th {
            background: #f8f9fa;
            font-weight: 600;
            color: #333;
            font-size: 14px;
        }
        td { font-size: 14px; }
        tr:hover { background: #f8f9fa; }
        .badge {
            display: inline-block;
            padding: 4px 8px;
            border-radius: 4px;
            font-size: 12px;
            font-weight: 500;
        }
        .badge.success { background: #d4edda; color: #155724; }
        .badge.danger { background: #f8d7da; color: #721c24; }
        .badge.info { background: #d1ecf1; color: #0c5460; }
        .actions {
            display: flex;
            gap: 5px;
        }
        .btn-small {
            padding: 5px 10px;
            font-size: 12px;
        }
        .message {
            padding: 12px;
            border-radius: 4px;
            margin-bottom: 15px;
            font-size: 14px;
        }
        .message.success { background: #d4edda; color: #155724; border: 1px solid #c3e6cb; }
        .message.error { background: #f8d7da; color: #721c24; border: 1px solid #f5c6cb; }
        .message.info { background: #d1ecf1; color: #0c5460; border: 1px solid #bee5eb; }
        .stats {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 15px;
            margin-bottom: 20px;
        }
        .stat-card {
            background: white;
            padding: 20px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .stat-value {
            font-size: 32px;
            font-weight: bold;
            color: #007bff;
        }
        .stat-label {
            color: #666;
            font-size: 14px;
            margin-top: 5px;
        }
        .search-box {
            margin-bottom: 15px;
        }
        .form-row {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 15px;
        }
        .form-row-3 {
            display: grid;
            grid-template-columns: repeat(3, 1fr);
            gap: 15px;
        }
        .section-header {
            display: flex;
            align-items: center;
            justify-content: space-between;
            gap: 10px;
            flex-wrap: wrap;
        }
        .inline-controls {
            display: flex;
            gap: 10px;
            align-items: center;
            flex-wrap: wrap;
        }
        .checkbox {
            display: inline-flex;
            align-items: center;
            gap: 6px;
            font-size: 14px;
            color: #333;
            font-weight: 500;
        }
        .muted {
            color: #666;
            font-size: 13px;
        }
    </style>
</head>
<body>
    <div class="container">
        <header>
            <h1>Office Tracker</h1>
            <div class="subtitle">System Management Dashboard</div>
        </header>

        <div class="tabs">
            <button class="tab active" onclick="switchTab('dashboard')">Dashboard</button>
            <button class="tab" onclick="switchTab('users')">Users</button>
            <button class="tab" onclick="switchTab('sessions')">Sessions</button>
            <button class="tab" onclick="switchTab('scans')">Scan History</button>
            <button class="tab" onclick="switchTab('reports')">Reports</button>
        </div>

        <div id="dashboard" class="panel active">
            <h2 style="margin-bottom: 20px;">Dashboard</h2>
            <div id="dashboardMessage"></div>
            <div style="margin-bottom: 15px; display: flex; gap: 10px; align-items: center;">
                <button onclick="loadDashboard()" class="secondary">Refresh</button>
                <button onclick="checkoutAll('dashboardMessage')" class="danger">Check out all</button>
                <label style="display: flex; align-items: center; gap: 6px; font-weight: 500;">
                    <input type="checkbox" id="autoRefreshToggle" onchange="toggleAutoRefresh()" />
                    Auto-refresh (30s)
                </label>
            </div>
            <div class="muted" id="environmentSummary" style="margin-bottom: 15px;"></div>
            <div class="stats">
                <div class="stat-card">
                    <div class="stat-value" id="totalUsers">0</div>
                    <div class="stat-label">Total Users</div>
                </div>
                <div class="stat-card">
                    <div class="stat-value" id="activeUsers">0</div>
                    <div class="stat-label">Currently Active</div>
                </div>
                <div class="stat-card">
                    <div class="stat-value" id="totalSessions">0</div>
                    <div class="stat-label">Total Sessions</div>
                </div>
                <div class="stat-card">
                    <div class="stat-value" id="recentScans">0</div>
                    <div class="stat-label">Recent Scans</div>
                </div>
            </div>
            <h3 style="margin-top: 30px; margin-bottom: 15px;">Currently Active Users</h3>
            <table id="dashboardPresenceTable">
                <thead>
                    <tr>
                        <th>User Name</th>
                        <th>Check In Time</th>
                        <th>Actions</th>
                    </tr>
                </thead>
                <tbody></tbody>
            </table>

            <h3 style="margin-top: 30px; margin-bottom: 15px;">ESP32 Health</h3>
            <table id="dashboardDeviceHealthTable">
                <thead>
                    <tr>
                        <th>Device</th>
                        <th>Status</th>
                        <th>Uptime</th>
                        <th>Free Heap</th>
                        <th>Wi-Fi</th>
                        <th>RSSI</th>
                        <th>IP</th>
                        <th>Firmware</th>
                        <th>Reset Reason</th>
                    </tr>
                </thead>
                <tbody></tbody>
            </table>
        </div>

        <div id="users" class="panel">
            <h2 style="margin-bottom: 20px;">User Management</h2>
            <div id="userMessage"></div>
            <div style="margin-bottom: 20px;">
                <button onclick="exportUsersCSV()" class="secondary">Export Users CSV</button>
                <input type="file" id="csvFileInput" accept=".csv" style="display:none" onchange="importUsersCSV()">
                <button onclick="document.getElementById('csvFileInput').click()" class="secondary">Import Users CSV</button>
            </div>
            <div id="editUserSection" style="display:none; margin-bottom: 30px; padding: 15px; background: #f8f9fa; border-radius: 8px; border-left: 4px solid #007bff;">
                <h3 style="margin-bottom: 15px;">Edit User</h3>
                <input type="hidden" id="editUserId">
                <div class="form-row">
                    <div class="form-group">
                        <label>Name *</label>
                        <input type="text" id="editUserName" placeholder="John Doe">
                    </div>
                    <div class="form-group">
                        <label>RFID UID *</label>
                        <input type="text" id="editUserRFID" placeholder="ABC123456">
                    </div>
                </div>
                <div class="form-group">
                    <label>Discord ID *</label>
                    <input type="text" id="editUserDiscord" placeholder="123456789012345678">
                </div>
                <div style="display: flex; gap: 10px;">
                    <button onclick="saveEditUser()">Save Changes</button>
                    <button onclick="cancelEditUser()" class="secondary">Cancel</button>
                </div>
            </div>
            <h3 style="margin-bottom: 15px;">Add New User</h3>
            <div class="form-row">
                <div class="form-group">
                    <label>Name *</label>
                    <input type="text" id="userName" placeholder="John Doe">
                </div>
                <div class="form-group">
                    <label>RFID UID *</label>
                    <input type="text" id="userRFID" placeholder="ABC123456">
                </div>
            </div>
            <div class="form-group">
                <label>Discord ID *</label>
                <input type="text" id="userDiscord" placeholder="123456789012345678">
            </div>
            <button onclick="createUser()">Add User</button>
            <h4 style="margin-top: 30px;">Advanced Filters</h4>
            <div class="form-row">
                <div class="form-group">
                    <label>Search Name</label>
                    <input type="text" id="userSearch" placeholder="Search users...">
                </div>
                <div class="form-group">
                    <label>Order By</label>
                    <select id="userOrderBy">
                        <option value="asc">Ascending</option>
                        <option value="desc">Descending</option>
                    </select>
                </div>
            </div>
            <div class="form-row">
                <div class="form-group">
                    <label>Sort By</label>
                    <select id="userSortBy">
                        <option value="name">Name</option>
                        <option value="created_at">Created At</option>
                    </select>
                </div>
                <div class="form-group">
                    <label>Limit</label>
                    <input type="number" id="userLimit" min="1" max="500" value="50">
                </div>
                <div class="form-group">
                    <label>Offset</label>
                    <input type="number" id="userOffset" min="0" value="0">
                </div>
            </div>
            <div style="margin-bottom: 15px;">
                <button onclick="filterUsers()" class="secondary">Apply Filters</button>
                <button onclick="resetUserFilters()" class="secondary">Reset Filters</button>
            </div>
            <table id="usersTable">
                <thead>
                    <tr>
                        <th>ID</th>
                        <th>Name</th>
                        <th>RFID UID</th>
                        <th>Discord ID</th>
                        <th>Status</th>
                        <th>Actions</th>
                    </tr>
                </thead>
                <tbody></tbody>
            </table>
        </div>

        <div id="sessions" class="panel">
            <h2 style="margin-bottom: 20px;">Session History</h2>
            <div id="sessionMessage"></div>
            <div id="editSessionSection" style="display:none; margin-bottom: 30px; padding: 15px; background: #f8f9fa; border-radius: 8px; border-left: 4px solid #007bff;">
                <h3 style="margin-bottom: 15px;">Edit Session</h3>
                <input type="hidden" id="editSessionId">
                <input type="hidden" id="editSessionHadCheckOut">
                <div class="form-row">
                    <div class="form-group">
                        <label>User</label>
                        <input type="text" id="editSessionUserName" readonly>
                    </div>
                    <div class="form-group">
                        <label>Check In *</label>
                        <input type="datetime-local" id="editSessionCheckIn">
                    </div>
                </div>
                <div class="form-row">
                    <div class="form-group">
                        <label>Check Out</label>
                        <input type="datetime-local" id="editSessionCheckOut">
                    </div>
                    <div class="form-group">
                        <label>Status</label>
                        <input type="text" id="editSessionStatus" readonly>
                    </div>
                </div>
                <div class="muted" style="margin-bottom: 15px;">Clearing an existing check-out time is not supported from the UI.</div>
                <div style="display: flex; gap: 10px;">
                    <button onclick="saveEditSession()">Save Changes</button>
                    <button onclick="cancelEditSession()" class="secondary">Cancel</button>
                </div>
            </div>
            <div style="margin-bottom: 15px;">
                <button onclick="loadSessions()" class="secondary">Refresh</button>
                <button onclick="downloadSessionsCSV()" class="secondary">Download CSV</button>
                <button onclick="downloadAllSessionsCSV()" class="secondary">Download All CSV</button>
                <button onclick="bulkDeleteSessions()" class="danger">Delete Visible Results</button>
            </div>
            <h4>Advanced Filters</h4>
            <div class="form-row">
                <div class="form-group">
                    <label>User Name</label>
                    <input type="text" id="sessionUserFilter" placeholder="Filter by user name...">
                </div>
                <div class="form-group">
                    <label>User ID</label>
                    <input type="number" id="sessionUserIDFilter" placeholder="User ID">
                </div>
                <div class="form-group">
                    <label>Discord ID</label>
                    <input type="text" id="sessionDiscordIDFilter" placeholder="Discord ID">
                </div>
            </div>
            <div class="form-row">
                <div class="form-group">
                    <label>From (Date)</label>
                    <input type="date" id="sessionFromFilter">
                </div>
                <div class="form-group">
                    <label>To (Date)</label>
                    <input type="date" id="sessionToFilter">
                </div>
                <div class="form-group">
                    <label>Status</label>
                    <select id="sessionStatusFilter">
                        <option value="all">All Sessions</option>
                        <option value="active">Active Only</option>
                        <option value="completed">Completed Only</option>
                    </select>
                </div>
            </div>
            <div class="form-row">
                <div class="form-group">
                    <label>Order By</label>
                    <select id="sessionOrderBy">
                        <option value="desc">Descending</option>
                        <option value="asc">Ascending</option>
                    </select>
                </div>
                <div class="form-group">
                    <label>Sort By</label>
                    <select id="sessionSortBy">
                        <option value="check_in">Check In</option>
                        <option value="check_out">Check Out</option>
                        <option value="user_name">User Name</option>
                    </select>
                </div>
                <div class="form-group">
                    <label>Limit</label>
                    <input type="number" id="sessionLimit" min="1" max="500" value="100">
                </div>
                <div class="form-group">
                    <label>Offset</label>
                    <input type="number" id="sessionOffset" min="0" value="0">
                </div>
            </div>
            <div style="margin-bottom: 15px;">
                <button onclick="filterSessions()" class="secondary">Apply Filters</button>
                <button onclick="resetSessionFilters()" class="secondary">Reset Filters</button>
            </div>
            <table id="sessionsTable">
                <thead>
                    <tr>
                        <th>User</th>
                        <th>Check In</th>
                        <th>Check Out</th>
                        <th>Duration</th>
                        <th>Status</th>
                        <th>Actions</th>
                    </tr>
                </thead>
                <tbody></tbody>
            </table>
        </div>

        <div id="scans" class="panel">
            <h2 style="margin-bottom: 20px;">Scan History</h2>
            <div style="margin-bottom: 15px;">
                <button onclick="loadScans()" class="secondary">Refresh</button>
                <button onclick="clearScans()" class="danger">Clear History</button>
            </div>
            <table id="scansTable">
                <thead>
                    <tr>
                        <th>UID</th>
                        <th>User</th>
                        <th>Status</th>
                        <th>Action</th>
                        <th>Time</th>
                    </tr>
                </thead>
                <tbody></tbody>
            </table>
        </div>

        <div id="reports" class="panel">
            <h2 style="margin-bottom: 20px;">Statistics</h2>
            <div id="reportsMessage"></div>

            <div class="section-header" style="margin-bottom: 10px;">
                <h3 style="margin: 0;">Office Statistics</h3>
                <div id="reportPeriodLabel" class="muted">Period: —</div>
            </div>

            <div class="form-row" style="margin-bottom: 15px;">
                <div class="form-group">
                    <label>Custom From</label>
                    <input type="date" id="reportFrom">
                </div>
                <div class="form-group">
                    <label>Custom To</label>
                    <input type="date" id="reportTo">
                </div>
            </div>
            <div style="margin-bottom: 15px; display: flex; gap: 10px; align-items: center; flex-wrap: wrap;">
                <button onclick="loadWeeklyReport()" class="secondary">Load Weekly</button>
                <button onclick="loadMonthlyReport()" class="secondary">Load Monthly</button>
                <button onclick="loadCustomReport()" class="secondary">Load Custom</button>
                <button onclick="resetReportFilters()" class="secondary">Reset Report Filters</button>
                <label class="checkbox">
                    <input type="checkbox" id="includeAutoCheckoutReports" />
                    Include non-RFID checkouts (including auto-checkout)
                </label>
            </div>

            <div class="stats">
                <div class="stat-card">
                    <div class="stat-value" id="reportTotalHours">0</div>
                    <div class="stat-label">Total Hours</div>
                </div>
                <div class="stat-card">
                    <div class="stat-value" id="reportTotalVisits">0</div>
                    <div class="stat-label">Total Visits</div>
                </div>
                <div class="stat-card">
                    <div class="stat-value" id="reportUniqueUsers">0</div>
                    <div class="stat-label">Unique Users</div>
                </div>
                <div class="stat-card">
                    <div class="stat-value" id="reportActiveDays">0</div>
                    <div class="stat-label">Active Days</div>
                </div>
                <div class="stat-card">
                    <div class="stat-value" id="reportAvgPerUser">0</div>
                    <div class="stat-label">Avg Hours / User</div>
                </div>
                <div class="stat-card">
                    <div class="stat-value" id="reportTopUser">—</div>
                    <div class="stat-label">Top User</div>
                </div>
                <div class="stat-card">
                    <div class="stat-value" id="reportTopUserHours">0</div>
                    <div class="stat-label">Top User Hours</div>
                </div>
            </div>

            <h3 style="margin-top: 20px; margin-bottom: 10px;">Leaderboard</h3>
            <div class="form-row-3" style="margin-bottom: 15px;">
                <div class="form-group">
                    <label>Rank By</label>
                    <select id="leaderboardMetric">
                        <option value="hours">Hours</option>
                        <option value="visits">Visits</option>
                    </select>
                </div>
                <div class="form-group">
                    <label>Range</label>
                    <select id="leaderboardRange">
                        <option value="weekly">Weekly</option>
                        <option value="monthly">Monthly</option>
                        <option value="custom">Custom</option>
                    </select>
                </div>
                <div class="form-group">
                    <label>Limit</label>
                    <input type="number" id="leaderboardLimit" value="10" min="1" max="50">
                </div>
            </div>
            <div class="form-row" style="margin-bottom: 15px;">
                <div class="form-group">
                    <label>From</label>
                    <input type="date" id="leaderboardFrom">
                </div>
                <div class="form-group">
                    <label>To</label>
                    <input type="date" id="leaderboardTo">
                </div>
            </div>
            <div style="margin-bottom: 15px;">
                <button onclick="loadLeaderboard()" class="secondary">Load Leaderboard</button>
            </div>
            <table id="leaderboardTable">
                <thead>
                    <tr>
                        <th>Rank</th>
                        <th>User</th>
                        <th>Total Hours</th>
                        <th>Visits</th>
                        <th>Active Days</th>
                        <th>Avg Session (hrs)</th>
                    </tr>
                </thead>
                <tbody></tbody>
            </table>

            <h3 style="margin-top: 30px; margin-bottom: 10px;">Member Stats</h3>
            <div class="form-row" style="margin-bottom: 15px;">
                <div class="form-group">
                    <label>Member</label>
                    <select id="memberStatsUser"></select>
                </div>
                <div class="form-group">
                    <label>Range</label>
                    <select id="memberStatsRange">
                        <option value="weekly">Weekly</option>
                        <option value="monthly">Monthly</option>
                        <option value="custom">Custom</option>
                    </select>
                </div>
            </div>
            <div class="form-row" style="margin-bottom: 15px;">
                <div class="form-group">
                    <label>From</label>
                    <input type="date" id="memberStatsFrom">
                </div>
                <div class="form-group">
                    <label>To</label>
                    <input type="date" id="memberStatsTo">
                </div>
            </div>
            <div class="inline-controls" style="margin-bottom: 15px;">
                <button onclick="loadMemberStats()" class="secondary">Load Member Stats</button>
                <label class="checkbox">
                    <input type="checkbox" id="memberIncludeAuto" />
                    Include non-RFID checkouts (including auto-checkout)
                </label>
                <div id="memberPeriodLabel" class="muted">Period: —</div>
            </div>

            <div class="stats">
                <div class="stat-card">
                    <div class="stat-value" id="memberTotalHours">0</div>
                    <div class="stat-label">Hours</div>
                </div>
                <div class="stat-card">
                    <div class="stat-value" id="memberTotalVisits">0</div>
                    <div class="stat-label">Visits</div>
                </div>
                <div class="stat-card">
                    <div class="stat-value" id="memberActiveDays">0</div>
                    <div class="stat-label">Active Days</div>
                </div>
                <div class="stat-card">
                    <div class="stat-value" id="memberAvgDuration">0</div>
                    <div class="stat-label">Avg Session (hrs)</div>
                </div>
                <div class="stat-card">
                    <div class="stat-value" id="memberFirstVisit">—</div>
                    <div class="stat-label">First Visit</div>
                </div>
                <div class="stat-card">
                    <div class="stat-value" id="memberLastVisit">—</div>
                    <div class="stat-label">Last Visit</div>
                </div>
                <div class="stat-card">
                    <div class="stat-value" id="memberHoursPct">0%</div>
                    <div class="stat-label">% of Hours</div>
                </div>
                <div class="stat-card">
                    <div class="stat-value" id="memberVisitsPct">0%</div>
                    <div class="stat-label">% of Visits</div>
                </div>
                <div class="stat-card">
                    <div class="stat-value" id="memberRankHours">—</div>
                    <div class="stat-label">Rank (Hours)</div>
                </div>
                <div class="stat-card">
                    <div class="stat-value" id="memberRankVisits">—</div>
                    <div class="stat-label">Rank (Visits)</div>
                </div>
            </div>

            <h4 style="margin-top: 20px; margin-bottom: 10px;">Member Session History</h4>
            <div class="inline-controls" style="margin-bottom: 10px;">
                <button onclick="loadMemberSessions()" class="secondary">Load Sessions</button>
                <span class="muted">Shows latest 20 sessions for the selected member</span>
            </div>
            <table id="memberSessionsTable">
                <thead>
                    <tr>
                        <th>Check In</th>
                        <th>Check Out</th>
                        <th>Duration (hrs)</th>
                    </tr>
                </thead>
                <tbody></tbody>
            </table>
        </div>

    </div>

    <script>
        var API_BASE = '';
        var allSessions = [];
        var API_KEY = localStorage.getItem('office_api_key') || '';
        var autoRefreshTimer = null;
        var sectionBusyCounts = {};

        function getHeaders() {
            var headers = { 'Content-Type': 'application/json' };
            if (API_KEY) headers['X-API-Key'] = API_KEY;
            return headers;
        }

        function checkAuth() {
            if (!API_KEY) {
                var key = prompt('Enter API Key (leave empty if not required):');
                if (key) {
                    API_KEY = key;
                    localStorage.setItem('office_api_key', key);
                }
            }
        }

        function switchTab(tab) {
            document.querySelectorAll('.tab').forEach(function(t) { t.classList.remove('active'); });
            document.querySelectorAll('.panel').forEach(function(p) { p.classList.remove('active'); });
            event.target.classList.add('active');
            document.getElementById(tab).classList.add('active');
            
            if (tab !== 'dashboard' && autoRefreshTimer) {
                clearInterval(autoRefreshTimer);
                autoRefreshTimer = null;
                var toggle = document.getElementById('autoRefreshToggle');
                if (toggle) toggle.checked = false;
            }

            if (tab === 'users') loadUsers();
            else if (tab === 'sessions') loadSessions();
            else if (tab === 'scans') loadScans();
            else if (tab === 'reports') loadReports();
            else if (tab === 'dashboard') loadDashboard();
        }

        function toggleAutoRefresh() {
            var toggle = document.getElementById('autoRefreshToggle');
            if (!toggle) return;

            if (toggle.checked) {
                loadDashboard();
                autoRefreshTimer = setInterval(loadDashboard, 30000);
            } else if (autoRefreshTimer) {
                clearInterval(autoRefreshTimer);
                autoRefreshTimer = null;
            }
        }

        function showMessage(elementId, message, type) {
            var el = document.getElementById(elementId);
            el.innerHTML = '<div class="message ' + type + '">' + message + '</div>';
            setTimeout(function() { el.innerHTML = ''; }, 5000);
        }

        function setSectionBusy(sectionSelector, isBusy) {
            var panel = document.querySelector(sectionSelector);
            if (!panel) return;

            var current = sectionBusyCounts[sectionSelector] || 0;
            current = isBusy ? (current + 1) : Math.max(0, current - 1);
            sectionBusyCounts[sectionSelector] = current;

            var disabled = current > 0;
            panel.querySelectorAll('button').forEach(function(btn) {
                btn.disabled = disabled;
                btn.style.opacity = disabled ? '0.65' : '';
                btn.style.cursor = disabled ? 'not-allowed' : '';
            });
        }

        function withSectionBusy(sectionSelector, messageElementId, loadingMessage, operation) {
            if (messageElementId && loadingMessage) {
                showMessage(messageElementId, loadingMessage, 'info');
            }

            setSectionBusy(sectionSelector, true);
            return Promise.resolve()
                .then(operation)
                .finally(function() {
                    setSectionBusy(sectionSelector, false);
                });
        }

        function refreshMainViews() {
            loadDashboard();
            loadUsers();
            loadSessions();
        }

        function arrayOrEmpty(value) {
            return Array.isArray(value) ? value : [];
        }

        function fetchJSON(url, fallbackMessage) {
            return fetch(url, { headers: getHeaders() }).then(function(response) {
                if (!response.ok) {
                    return response.text().then(function(error) {
                        throw new Error(error || fallbackMessage || 'Request failed');
                    });
                }
                return response.json();
            });
        }

        function formatSeconds(totalSeconds) {
            var seconds = Math.max(0, Number(totalSeconds || 0));
            var days = Math.floor(seconds / 86400);
            var hours = Math.floor((seconds % 86400) / 3600);
            var minutes = Math.floor((seconds % 3600) / 60);

            if (days > 0) return days + 'd ' + hours + 'h';
            if (hours > 0) return hours + 'h ' + minutes + 'm';
            return minutes + 'm';
        }

        function esc(value) {
            return String(value == null ? '' : value)
                .replace(/&/g, '&amp;')
                .replace(/</g, '&lt;')
                .replace(/>/g, '&gt;')
                .replace(/"/g, '&quot;')
                .replace(/'/g, '&#39;');
        }

        function formatDateTimeLocal(value) {
            if (!value) return '';

            var date = new Date(value);
            if (isNaN(date.getTime())) return '';

            var year = date.getFullYear();
            var month = String(date.getMonth() + 1).padStart(2, '0');
            var day = String(date.getDate()).padStart(2, '0');
            var hours = String(date.getHours()).padStart(2, '0');
            var minutes = String(date.getMinutes()).padStart(2, '0');

            return year + '-' + month + '-' + day + 'T' + hours + ':' + minutes;
        }

        function loadDashboard() {
            Promise.all([
                fetchJSON(API_BASE + '/api/users', 'Failed to load users'),
                fetchJSON(API_BASE + '/api/sessions/count', 'Failed to load session count'),
                fetchJSON(API_BASE + '/api/rfid/scans', 'Failed to load RFID scans'),
                fetchJSON(API_BASE + '/api/sessions/open', 'Failed to load active sessions'),
                fetchJSON(API_BASE + '/api/environment', 'Failed to load environmental data'),
                fetchJSON(API_BASE + '/api/devices/health', 'Failed to load device health')
            ]).then(function(results) {
                var users = arrayOrEmpty(results[0]);
                var sessionsCount = Number((results[1] || {}).total || 0);
                var scans = arrayOrEmpty(results[2]);
                var activeSessions = arrayOrEmpty(results[3]);
                var environment = results[4] || {};
                var deviceHealth = arrayOrEmpty(results[5]);

                document.getElementById('totalUsers').textContent = users.length;
                document.getElementById('activeUsers').textContent = activeSessions.length;
                document.getElementById('totalSessions').textContent = sessionsCount;
                document.getElementById('recentScans').textContent = scans.length;

                var environmentSummary = document.getElementById('environmentSummary');
                if (environment && environment.available && environment.fresh) {
                    environmentSummary.textContent = 'Environment: ' + Number(environment.temperature_c).toFixed(1) + '°C';
                } else if (environment && environment.available) {
                    environmentSummary.textContent = 'Environment: unavailable';
                } else {
                    environmentSummary.textContent = '';
                }
                
                var tbody = document.querySelector('#dashboardPresenceTable tbody');
                tbody.innerHTML = activeSessions.map(function(p) {
                    var userName = p.user_name || p.UserName || 'Unknown';
                    var userID = p.user_id || p.UserID;
                    return '<tr>' +
                        '<td>' + userName + '</td>' +
                        '<td>' + new Date(p.check_in).toLocaleString() + '</td>' +
                        '<td><button class="btn-small secondary" onclick="checkOutUser(' + userID + ', \'dashboardMessage\')">Check Out</button></td>' +
                        '</tr>';
                }).join('');

                var healthTbody = document.querySelector('#dashboardDeviceHealthTable tbody');
                if (deviceHealth.length === 0) {
                    healthTbody.innerHTML = '<tr><td colspan="9" class="muted">No device health data yet.</td></tr>';
                } else {
                    healthTbody.innerHTML = deviceHealth.map(function(d) {
                        var statusClass = d.fresh ? 'success' : 'danger';
                        var statusText = d.fresh ? 'Online' : 'Offline';
                        var wifiText = d.wifi_connected ? 'Connected' : 'Disconnected';

                        return '<tr>' +
                            '<td>' + esc(d.device_id || 'default') + '</td>' +
                            '<td><span class="badge ' + statusClass + '">' + statusText + '</span></td>' +
                            '<td>' + esc(formatSeconds(d.uptime_seconds)) + '</td>' +
                            '<td>' + esc((d.free_heap_bytes || 0).toLocaleString()) + ' B</td>' +
                            '<td>' + esc(wifiText) + '</td>' +
                            '<td>' + esc((d.rssi != null ? d.rssi : '--')) + '</td>' +
                            '<td>' + esc(d.ip || '-') + '</td>' +
                            '<td>' + esc(d.firmware_version || '-') + '</td>' +
                            '<td>' + esc(d.reset_reason || '-') + '</td>' +
                            '</tr>';
                    }).join('');
                }
            }).catch(function(error) {
                console.error('Error loading dashboard:', error);
                showMessage('dashboardMessage', 'Error loading dashboard: ' + error.message, 'error');
            });
        }

        function createUser() {
            checkAuth();
            var name = document.getElementById('userName').value;
            var rfid = document.getElementById('userRFID').value;
            var discord = document.getElementById('userDiscord').value;

            if (!name || !rfid || !discord) {
                showMessage('userMessage', 'Name, RFID UID, and Discord ID are required', 'error');
                return;
            }

            withSectionBusy('#users', 'userMessage', 'Creating user...', function() {
                return fetch(API_BASE + '/api/users', {
                    method: 'POST',
                    headers: getHeaders(),
                    body: JSON.stringify({ name: name, rfid_uid: rfid, discord_id: discord })
                }).then(function(response) {
                    if (response.ok) {
                        showMessage('userMessage', 'User created successfully!', 'success');
                        document.getElementById('userName').value = '';
                        document.getElementById('userRFID').value = '';
                        document.getElementById('userDiscord').value = '';
                        refreshMainViews();
                    } else {
                        return response.text().then(function(error) {
                            showMessage('userMessage', 'Error: ' + error, 'error');
                        });
                    }
                }).catch(function(error) {
                    showMessage('userMessage', 'Error creating user: ' + error.message, 'error');
                });
            });
        }

        function loadUsers() {
            filterUsers();
        }

        function buildUserFilterParams() {
            var params = new URLSearchParams();

            var search = document.getElementById('userSearch').value.trim();
            var order = document.getElementById('userOrderBy').value;
            var sortBy = document.getElementById('userSortBy').value;
            var limit = document.getElementById('userLimit').value;
            var offset = document.getElementById('userOffset').value;

            if (search) params.append('search', search);
            if (order) params.append('order', order);
            if (sortBy) params.append('sort_by', sortBy);
            if (limit) params.append('limit', limit);
            if (offset) params.append('offset', offset);

            return params;
        }

        function filterUsers() {
            var params = buildUserFilterParams();
            var query = params.toString();
            var usersUrl = API_BASE + '/api/users' + (query ? ('?' + query) : '');

            Promise.all([
                fetch(usersUrl, { headers: getHeaders() }).then(function(r) {
                    if (!r.ok) {
                        return r.text().then(function(error) {
                            throw new Error(error || 'Failed to load users');
                        });
                    }
                    return r.json();
                }),
                fetch(API_BASE + '/api/sessions/open', { headers: getHeaders() }).then(function(r) { return r.json(); })
            ]).then(function(results) {
                var users = results[0] || [];
                var openSessions = results[1] || [];
                var activeUserIDs = {};
                (openSessions || []).forEach(function(s) { activeUserIDs[s.user_id] = s; });

                var tbody = document.querySelector('#usersTable tbody');
                tbody.innerHTML = users.map(function(u) {
                    var isActive = !!activeUserIDs[u.id];
                    return '<tr>' +
                        '<td>' + u.id + '</td>' +
                        '<td>' + u.name + '</td>' +
                        '<td>' + (u.rfid_uid || '-') + '</td>' +
                        '<td>' + (u.discord_id || '-') + '</td>' +
                        '<td><span class="badge ' + (isActive ? 'success' : 'info') + '">' + (isActive ? 'Checked In' : 'Offline') + '</span></td>' +
                        '<td class="actions">' +
                            '<button class="btn-small" onclick="checkInUser(' + u.id + ')">Check In</button>' +
                            '<button class="btn-small secondary" onclick="checkOutUser(' + u.id + ')">Check Out</button>' +
                            '<button class="btn-small" style="background: #28a745;" onclick="editUser(' + u.id + ', \'' + u.name.replace(/'/g, "\\'") + '\', \'' + (u.rfid_uid || '').replace(/'/g, "\\'") + '\', \'' + (u.discord_id || '').replace(/'/g, "\\'") + '\')">Edit</button>' +
                            '<button class="btn-small danger" onclick="deleteUser(' + u.id + ')">Delete</button>' +
                        '</td>' +
                        '</tr>';
                }).join('');
            }).catch(function(err) {
                console.error('Error loading users:', err);
                showMessage('userMessage', 'Error loading users: ' + err.message, 'error');

                // Fallback: render users without status using users endpoint only
                fetch(usersUrl, { headers: getHeaders() }).then(function(r) {
                    if (!r.ok) {
                        return r.text().then(function(error) {
                            throw new Error(error || 'Failed to load users');
                        });
                    }
                    return r.json();
                }).then(function(users) {
                var tbody = document.querySelector('#usersTable tbody');
                tbody.innerHTML = (users || []).map(function(u) {
                    return '<tr>' +
                        '<td>' + u.id + '</td>' +
                        '<td>' + u.name + '</td>' +
                        '<td>' + (u.rfid_uid || '-') + '</td>' +
                        '<td>' + (u.discord_id || '-') + '</td>' +
                        '<td><span class="badge info">Unknown</span></td>' +
                        '<td class="actions">' +
                            '<button class="btn-small" onclick="checkInUser(' + u.id + ')">Check In</button>' +
                            '<button class="btn-small secondary" onclick="checkOutUser(' + u.id + ')">Check Out</button>' +
                            '<button class="btn-small" style="background: #28a745;" onclick="editUser(' + u.id + ', \'' + u.name.replace(/'/g, "\\'") + '\', \'' + (u.rfid_uid || '').replace(/'/g, "\\'") + '\', \'' + (u.discord_id || '').replace(/'/g, "\\'") + '\')">Edit</button>' +
                            '<button class="btn-small danger" onclick="deleteUser(' + u.id + ')">Delete</button>' +
                        '</td>' +
                        '</tr>';
                }).join('');
                }).catch(function(fallbackErr) {
                    showMessage('userMessage', 'Error loading users: ' + fallbackErr.message, 'error');
                });
            });
        }

        function resetUserFilters() {
            var search = document.getElementById('userSearch');
            var order = document.getElementById('userOrderBy');
            var sortBy = document.getElementById('userSortBy');
            var limit = document.getElementById('userLimit');
            var offset = document.getElementById('userOffset');

            if (search) search.value = '';
            if (order) order.value = 'asc';
            if (sortBy) sortBy.value = 'name';
            if (limit) limit.value = '50';
            if (offset) offset.value = '0';

            showMessage('userMessage', 'Filters reset. Click Apply Filters to refresh results.', 'info');
        }

        function editUser(id, name, rfidUID, discordID) {
            checkAuth();
            document.getElementById('editUserId').value = id;
            document.getElementById('editUserName').value = name;
            document.getElementById('editUserRFID').value = rfidUID;
            document.getElementById('editUserDiscord').value = discordID;
            document.getElementById('editUserSection').style.display = 'block';
            document.getElementById('editUserName').focus();
        }

        function cancelEditUser() {
            document.getElementById('editUserSection').style.display = 'none';
            document.getElementById('editUserId').value = '';
            document.getElementById('editUserName').value = '';
            document.getElementById('editUserRFID').value = '';
            document.getElementById('editUserDiscord').value = '';
        }

        function saveEditUser() {
            checkAuth();
            var id = document.getElementById('editUserId').value;
            var name = document.getElementById('editUserName').value;
            var rfid = document.getElementById('editUserRFID').value;
            var discord = document.getElementById('editUserDiscord').value;

            if (!name || !rfid || !discord) {
                showMessage('userMessage', 'Name, RFID UID, and Discord ID are required', 'error');
                return;
            }

            withSectionBusy('#users', 'userMessage', 'Saving user changes...', function() {
                return fetch(API_BASE + '/api/users/' + id, {
                    method: 'PUT',
                    headers: getHeaders(),
                    body: JSON.stringify({ name: name, rfid_uid: rfid, discord_id: discord })
                }).then(function(response) {
                    if (response.ok) {
                        showMessage('userMessage', 'User updated successfully!', 'success');
                        cancelEditUser();
                        refreshMainViews();
                    } else {
                        return response.text().then(function(error) {
                            showMessage('userMessage', 'Error: ' + error, 'error');
                        });
                    }
                }).catch(function(error) {
                    showMessage('userMessage', 'Error updating user: ' + error.message, 'error');
                });
            });
        }

        function deleteUser(id) {
            checkAuth();
            if (!confirm('Are you sure you want to delete this user?')) return;

            withSectionBusy('#users', 'userMessage', 'Deleting user...', function() {
                return fetch(API_BASE + '/api/users/' + id, {
                    method: 'DELETE',
                    headers: getHeaders()
                }).then(function(response) {
                    if (response.ok) {
                        showMessage('userMessage', 'User deleted successfully!', 'success');
                        refreshMainViews();
                    } else {
                        showMessage('userMessage', 'Error deleting user', 'error');
                    }
                }).catch(function(error) {
                    showMessage('userMessage', 'Error: ' + error.message, 'error');
                });
            });
        }

        function checkInUser(userID) {
            checkAuth();

            withSectionBusy('#users', 'userMessage', 'Checking in user...', function() {
                return fetch(API_BASE + '/api/sessions/checkin', {
                    method: 'POST',
                    headers: getHeaders(),
                    body: JSON.stringify({ user_id: userID })
                }).then(function(response) {
                    if (response.ok) {
                        showMessage('userMessage', 'User checked in successfully!', 'success');
                        refreshMainViews();
                    } else {
                        return response.text().then(function(error) {
                            showMessage('userMessage', 'Error: ' + error, 'error');
                        });
                    }
                }).catch(function(error) {
                    showMessage('userMessage', 'Error: ' + error.message, 'error');
                });
            });
        }

        function checkOutUser(userID, messageElementId) {
            checkAuth();
            var targetMessage = messageElementId || 'userMessage';
            var sectionSelector = targetMessage === 'dashboardMessage' ? '#dashboard' : '#users';

            withSectionBusy(sectionSelector, targetMessage, 'Checking out user...', function() {
                return fetch(API_BASE + '/api/sessions/checkout', {
                    method: 'POST',
                    headers: getHeaders(),
                    body: JSON.stringify({ user_id: userID })
                }).then(function(response) {
                    if (response.ok) {
                        showMessage(targetMessage, 'User checked out successfully!', 'success');
                        refreshMainViews();
                    } else {
                        return response.json().then(function(error) {
                            showMessage(targetMessage, 'Error: ' + (error.error || 'Failed to check out user'), 'error');
                        }).catch(function() {
                            return response.text().then(function(error) {
                                showMessage(targetMessage, 'Error: ' + error, 'error');
                            });
                        });
                    }
                }).catch(function(error) {
                    showMessage(targetMessage, 'Error: ' + error.message, 'error');
                });
            });
        }

        function checkoutAll(messageElementId) {
            checkAuth();
            var targetMessage = messageElementId || 'userMessage';
            if (!confirm('Are you sure you want to check out all active users?')) return;
            var sectionSelector = targetMessage === 'dashboardMessage' ? '#dashboard' : '#users';

            withSectionBusy(sectionSelector, targetMessage, 'Checking out all active users...', function() {
                return fetch(API_BASE + '/api/sessions/checkout-all', {
                    method: 'POST',
                    headers: getHeaders()
                }).then(function(response) {
                    if (response.ok) {
                        showMessage(targetMessage, 'All users checked out successfully!', 'success');
                        refreshMainViews();
                    } else {
                        return response.json().then(function(error) {
                            showMessage(targetMessage, 'Error: ' + (error.error || 'Failed to check out users'), 'error');
                        }).catch(function() {
                            return response.text().then(function(error) {
                                showMessage(targetMessage, 'Error: ' + error, 'error');
                            });
                        });
                    }
                }).catch(function(error) {
                    showMessage(targetMessage, 'Error: ' + error.message, 'error');
                });
            });
        }

        function searchUsers() {
            var search = document.getElementById('userSearch').value.toLowerCase();
            var rows = document.querySelectorAll('#usersTable tbody tr');
            rows.forEach(function(row) {
                var text = row.textContent.toLowerCase();
                row.style.display = text.indexOf(search) > -1 ? '' : 'none';
            });
        }

        function loadSessions() {
            filterSessions();
        }

        function buildSessionFilterParams(options) {
            options = options || {};
            var params = new URLSearchParams();

            var userName = document.getElementById('sessionUserFilter').value.trim();
            var userID = document.getElementById('sessionUserIDFilter').value.trim();
            var discordID = document.getElementById('sessionDiscordIDFilter').value.trim();
            var from = document.getElementById('sessionFromFilter').value;
            var to = document.getElementById('sessionToFilter').value;
            var status = document.getElementById('sessionStatusFilter').value;
            var order = document.getElementById('sessionOrderBy').value;
            var sortBy = document.getElementById('sessionSortBy').value;
            var limit = document.getElementById('sessionLimit').value;
            var offset = document.getElementById('sessionOffset').value;

            if (userName) params.append('name', userName);
            if (userID) params.append('user_id', userID);
            if (discordID) params.append('discord_id', discordID);
            if (from) params.append('from', from);
            if (to) params.append('to', to);
            if (status && status !== 'all') params.append('status', status);

            if (options.includeOrdering !== false) {
                if (order) params.append('order', order);
                if (sortBy) params.append('sort_by', sortBy);
            }

            if (options.includePagination !== false) {
                if (limit) params.append('limit', limit);
                if (offset) params.append('offset', offset);
            }

            return params;
        }

        function renderSessionsTable(sessions) {
            var tbody = document.querySelector('#sessionsTable tbody');
            if (!sessions || sessions.length === 0) {
                tbody.innerHTML = '<tr><td colspan="6" class="muted">No sessions match the current filters.</td></tr>';
                return;
            }

            tbody.innerHTML = sessions.map(function(s) {
                var checkIn = s.check_in ? new Date(s.check_in) : null;
                var checkOut = s.check_out ? new Date(s.check_out) : null;
                var isActive = s.active === true || !checkOut;
                var duration = '-';
                if (checkIn && checkOut) {
                    var completedMs = checkOut - checkIn;
                    var completedHours = Math.floor(completedMs / 36e5);
                    var completedMinutes = Math.floor((completedMs % 36e5) / 60000);
                    duration = completedHours + 'h ' + completedMinutes + 'm';
                } else if (checkIn && !checkOut) {
                    var activeMs = Date.now() - checkIn.getTime();
                    var activeHours = Math.floor(activeMs / 36e5);
                    var activeMinutes = Math.floor((activeMs % 36e5) / 60000);
                    duration = activeHours + 'h ' + activeMinutes + 'm (active)';
                }

                return '<tr>' +
                    '<td>' + esc(s.user_name || '-') + '</td>' +
                    '<td>' + esc(checkIn ? checkIn.toLocaleString() : '-') + '</td>' +
                    '<td>' + esc(checkOut ? checkOut.toLocaleString() : '-') + '</td>' +
                    '<td>' + esc(duration) + '</td>' +
                    '<td><span class="badge ' + (isActive ? 'success' : 'info') + '">' + (isActive ? 'Active' : 'Completed') + '</span></td>' +
                    '<td class="actions">' +
                        '<button class="btn-small" style="background: #28a745;" onclick="editSessionById(' + s.id + ')">Edit</button>' +
                        '<button class="btn-small danger" onclick="deleteSessionById(' + s.id + ')">Delete</button>' +
                    '</td>' +
                    '</tr>';
            }).join('');
        }

        function getSessionById(id) {
            for (var i = 0; i < allSessions.length; i++) {
                if (Number(allSessions[i].id) === Number(id)) {
                    return allSessions[i];
                }
            }
            return null;
        }

        function editSessionById(id) {
            var session = getSessionById(id);
            if (!session) {
                showMessage('sessionMessage', 'Could not find that session in the current table view', 'error');
                return;
            }
            editSession(session);
        }

        function editSession(session) {
            checkAuth();
            document.getElementById('editSessionId').value = session.id;
            document.getElementById('editSessionUserName').value = session.user_name || '';
            document.getElementById('editSessionCheckIn').value = formatDateTimeLocal(session.check_in);
            document.getElementById('editSessionCheckOut').value = formatDateTimeLocal(session.check_out);
            document.getElementById('editSessionStatus').value = session.check_out ? 'Completed' : 'Active';
            document.getElementById('editSessionHadCheckOut').value = session.check_out ? 'true' : 'false';
            document.getElementById('editSessionSection').style.display = 'block';
            document.getElementById('editSessionCheckIn').focus();
        }

        function cancelEditSession() {
            document.getElementById('editSessionSection').style.display = 'none';
            document.getElementById('editSessionId').value = '';
            document.getElementById('editSessionUserName').value = '';
            document.getElementById('editSessionCheckIn').value = '';
            document.getElementById('editSessionCheckOut').value = '';
            document.getElementById('editSessionStatus').value = '';
            document.getElementById('editSessionHadCheckOut').value = '';
        }

        function saveEditSession() {
            checkAuth();
            var id = document.getElementById('editSessionId').value;
            var checkInValue = document.getElementById('editSessionCheckIn').value;
            var checkOutValue = document.getElementById('editSessionCheckOut').value;
            var hadCheckOut = document.getElementById('editSessionHadCheckOut').value === 'true';

            if (!checkInValue) {
                showMessage('sessionMessage', 'Check-in time is required', 'error');
                return;
            }

            if (hadCheckOut && !checkOutValue) {
                showMessage('sessionMessage', 'Clearing an existing check-out time is not supported from the UI', 'error');
                return;
            }

            var checkInDate = new Date(checkInValue);
            if (isNaN(checkInDate.getTime())) {
                showMessage('sessionMessage', 'Invalid check-in time', 'error');
                return;
            }

            var payload = {
                check_in: checkInDate.toISOString()
            };

            if (checkOutValue) {
                var checkOutDate = new Date(checkOutValue);
                if (isNaN(checkOutDate.getTime())) {
                    showMessage('sessionMessage', 'Invalid check-out time', 'error');
                    return;
                }
                if (checkOutDate < checkInDate) {
                    showMessage('sessionMessage', 'Check-out time cannot be before check-in time', 'error');
                    return;
                }
                payload.check_out = checkOutDate.toISOString();
            }

            withSectionBusy('#sessions', 'sessionMessage', 'Saving session changes...', function() {
                return fetch(API_BASE + '/api/sessions/' + id, {
                    method: 'PUT',
                    headers: getHeaders(),
                    body: JSON.stringify(payload)
                }).then(function(response) {
                    if (response.ok) {
                        showMessage('sessionMessage', 'Session updated successfully!', 'success');
                        cancelEditSession();
                        refreshMainViews();
                    } else {
                        return response.text().then(function(error) {
                            showMessage('sessionMessage', 'Error: ' + error, 'error');
                        });
                    }
                }).catch(function(error) {
                    showMessage('sessionMessage', 'Error updating session: ' + error.message, 'error');
                });
            });
        }

        function deleteSessionById(id) {
            var session = getSessionById(id);
            if (!session) {
                showMessage('sessionMessage', 'Could not find that session in the current table view', 'error');
                return;
            }
            deleteSession(session);
        }

        function deleteSession(session) {
            checkAuth();
            var userName = session.user_name || 'Unknown';
            if (!confirm('Delete this session for ' + userName + '? This cannot be undone.')) return;

            withSectionBusy('#sessions', 'sessionMessage', 'Deleting session...', function() {
                return fetch(API_BASE + '/api/sessions/' + session.id, {
                    method: 'DELETE',
                    headers: getHeaders()
                }).then(function(response) {
                    if (response.ok) {
                        showMessage('sessionMessage', 'Session deleted successfully!', 'success');
                        cancelEditSession();
                        refreshMainViews();
                    } else {
                        return response.text().then(function(error) {
                            showMessage('sessionMessage', 'Error: ' + error, 'error');
                        });
                    }
                }).catch(function(error) {
                    showMessage('sessionMessage', 'Error deleting session: ' + error.message, 'error');
                });
            });
        }

        function bulkDeleteSessions() {
            checkAuth();

            var strictParams = buildSessionFilterParams({ includeOrdering: false, includePagination: false });
            if (!strictParams.toString()) {
                showMessage('sessionMessage', 'Apply at least one filter or status before bulk delete', 'error');
                return;
            }

            if (!allSessions || allSessions.length === 0) {
                showMessage('sessionMessage', 'No visible sessions to delete', 'error');
                return;
            }

            if (!confirm('Delete the ' + allSessions.length + ' session(s) currently shown in the table? This cannot be undone.')) return;

            var params = buildSessionFilterParams();
            var query = params.toString();

            withSectionBusy('#sessions', 'sessionMessage', 'Deleting visible sessions...', function() {
                return fetch(API_BASE + '/api/sessions' + (query ? ('?' + query) : ''), {
                    method: 'DELETE',
                    headers: getHeaders()
                }).then(function(response) {
                    if (!response.ok) {
                        return response.text().then(function(error) {
                            throw new Error(error || 'Failed to delete sessions');
                        });
                    }
                    return response.json();
                }).then(function(result) {
                    showMessage('sessionMessage', 'Deleted ' + (result.deleted || 0) + ' session(s).', 'success');
                    cancelEditSession();
                    refreshMainViews();
                }).catch(function(error) {
                    showMessage('sessionMessage', 'Error deleting sessions: ' + error.message, 'error');
                });
            });
        }

        function filterSessions() {
            var params = buildSessionFilterParams();
            var query = params.toString();
            var url = API_BASE + '/api/sessions' + (query ? ('?' + query) : '');

            fetch(url, { headers: getHeaders() }).then(function(response) {
                if (!response.ok) {
                    return response.text().then(function(error) {
                        throw new Error(error || 'Failed to load sessions');
                    });
                }
                return response.json();
            }).then(function(sessions) {
                allSessions = sessions;
                renderSessionsTable(sessions);
            }).catch(function(error) {
                console.error('Error loading sessions:', error);
                allSessions = [];
                renderSessionsTable([]);
                showMessage('sessionMessage', 'Error loading sessions: ' + error.message, 'error');
            });
        }

        function resetSessionFilters() {
            var userName = document.getElementById('sessionUserFilter');
            var userID = document.getElementById('sessionUserIDFilter');
            var discordID = document.getElementById('sessionDiscordIDFilter');
            var from = document.getElementById('sessionFromFilter');
            var to = document.getElementById('sessionToFilter');
            var status = document.getElementById('sessionStatusFilter');
            var order = document.getElementById('sessionOrderBy');
            var sortBy = document.getElementById('sessionSortBy');
            var limit = document.getElementById('sessionLimit');
            var offset = document.getElementById('sessionOffset');

            if (userName) userName.value = '';
            if (userID) userID.value = '';
            if (discordID) discordID.value = '';
            if (from) from.value = '';
            if (to) to.value = '';
            if (status) status.value = 'all';
            if (order) order.value = 'desc';
            if (sortBy) sortBy.value = 'check_in';
            if (limit) limit.value = '100';
            if (offset) offset.value = '0';

            cancelEditSession();
            showMessage('sessionMessage', 'Filters reset. Click Apply Filters to refresh results.', 'info');
        }

        function downloadSessionsCSV() {
            var params = buildSessionFilterParams();
            var query = params.toString();
            var url = API_BASE + '/api/sessions/export' + (query ? ('?' + query) : '');

            downloadSessionsCSVFromURL(url, 'sessions');
        }

        function downloadAllSessionsCSV() {
            var url = API_BASE + '/api/sessions/export?limit=0';

            downloadSessionsCSVFromURL(url, 'sessions_all');
        }

        function downloadSessionsCSVFromURL(url, filePrefix) {
            filePrefix = filePrefix || 'sessions';

            fetch(url, { headers: getHeaders() }).then(function(response) {
                if (!response.ok) {
                    return response.text().then(function(error) {
                        throw new Error(error || 'Failed to export sessions');
                    });
                }
                return response.blob();
            }).then(function(blob) {
                var downloadUrl = window.URL.createObjectURL(blob);
                var a = document.createElement('a');
                a.href = downloadUrl;
                a.download = filePrefix + '_' + new Date().toISOString().split('T')[0] + '.csv';
                document.body.appendChild(a);
                a.click();
                document.body.removeChild(a);
                window.URL.revokeObjectURL(downloadUrl);
            }).catch(function(error) {
                showMessage('sessionMessage', 'Error exporting sessions: ' + error.message, 'error');
            });
        }

        function loadScans() {
            fetch(API_BASE + '/api/rfid/scans', { headers: getHeaders() }).then(function(response) {
                return response.json();
            }).then(function(scans) {
                var tbody = document.querySelector('#scansTable tbody');
                tbody.innerHTML = scans.reverse().map(function(s) {
                    return '<tr>' +
                        '<td>' + s.uid + '</td>' +
                        '<td>' + (s.user_name || '-') + '</td>' +
                        '<td><span class="badge ' + (s.known ? 'success' : 'danger') + '">' + (s.known ? 'Known' : 'Unknown') + '</span></td>' +
                        '<td>' + (s.action || '-') + '</td>' +
                        '<td>' + new Date(s.timestamp).toLocaleString() + '</td>' +
                        '</tr>';
                }).join('');
            }).catch(function(error) {
                console.error('Error loading scans:', error);
            });
        }

        function clearScans() {
            checkAuth();
            if (!confirm('Are you sure you want to clear scan history?')) return;
            
            fetch(API_BASE + '/api/rfid/scans', { method: 'DELETE', headers: getHeaders() }).then(function() {
                loadScans();
            }).catch(function(error) {
                console.error('Error clearing scans:', error);
            });
        }

        function exportUsersCSV() {
            checkAuth();
            var params = buildUserFilterParams();
            var query = params.toString();
            var url = API_BASE + '/api/users/export' + (query ? ('?' + query) : '');

            fetch(url, {
                headers: getHeaders()
            }).then(function(response) {
                if (!response.ok) {
                    return response.text().then(function(error) {
                        throw new Error(error || 'Failed to export users');
                    });
                }
                return response.blob();
            }).then(function(blob) {
                var url = window.URL.createObjectURL(blob);
                var a = document.createElement('a');
                a.href = url;
                a.download = 'users_' + new Date().toISOString().split('T')[0] + '.csv';
                document.body.appendChild(a);
                a.click();
                document.body.removeChild(a);
                window.URL.revokeObjectURL(url);
                showMessage('userMessage', 'Users exported successfully!', 'success');
            }).catch(function(error) {
                showMessage('userMessage', 'Error exporting users: ' + error.message, 'error');
            });
        }

        function importUsersCSV() {
            checkAuth();
            var fileInput = document.getElementById('csvFileInput');
            var file = fileInput.files[0];
            if (!file) return;

            var formData = new FormData();
            formData.append('file', file);

            var headers = {};
            if (API_KEY) headers['X-API-Key'] = API_KEY;

            fetch(API_BASE + '/api/users/import', {
                method: 'POST',
                headers: headers,
                body: formData
            }).then(function(response) {
                if (response.ok) {
                    showMessage('userMessage', 'Users imported successfully!', 'success');
                    loadUsers();
                    fileInput.value = '';
                } else {
                    return response.text().then(function(error) {
                        showMessage('userMessage', 'Error importing: ' + error, 'error');
                    });
                }
            }).catch(function(error) {
                showMessage('userMessage', 'Error importing users: ' + error.message, 'error');
            });
        }

        function loadReports() {
            loadWeeklyReport();
            loadLeaderboard();
            loadMemberStatsUsers();
            updateRangeInputs();
        }

        function resetReportFilters() {
            var reportFrom = document.getElementById('reportFrom');
            var reportTo = document.getElementById('reportTo');
            var includeAuto = document.getElementById('includeAutoCheckoutReports');
            var leaderboardMetric = document.getElementById('leaderboardMetric');
            var leaderboardRange = document.getElementById('leaderboardRange');
            var leaderboardLimit = document.getElementById('leaderboardLimit');
            var leaderboardFrom = document.getElementById('leaderboardFrom');
            var leaderboardTo = document.getElementById('leaderboardTo');
            var memberStatsRange = document.getElementById('memberStatsRange');
            var memberStatsFrom = document.getElementById('memberStatsFrom');
            var memberStatsTo = document.getElementById('memberStatsTo');
            var memberIncludeAuto = document.getElementById('memberIncludeAuto');
            var memberStatsUser = document.getElementById('memberStatsUser');

            if (reportFrom) reportFrom.value = '';
            if (reportTo) reportTo.value = '';
            if (includeAuto) includeAuto.checked = false;

            if (leaderboardMetric) leaderboardMetric.value = 'hours';
            if (leaderboardRange) leaderboardRange.value = 'weekly';
            if (leaderboardLimit) leaderboardLimit.value = '10';
            if (leaderboardFrom) leaderboardFrom.value = '';
            if (leaderboardTo) leaderboardTo.value = '';

            if (memberStatsRange) memberStatsRange.value = 'weekly';
            if (memberStatsFrom) memberStatsFrom.value = '';
            if (memberStatsTo) memberStatsTo.value = '';
            if (memberIncludeAuto) memberIncludeAuto.checked = false;
            if (memberStatsUser && memberStatsUser.options && memberStatsUser.options.length > 0) {
                memberStatsUser.selectedIndex = 0;
            }

            updateRangeInputs();
            loadReports();
            showMessage('reportsMessage', 'Report filters reset', 'success');
        }

        function renderReportSummary(report) {
            document.getElementById('reportTotalHours').textContent = (report.total_hours || 0).toFixed(1);
            document.getElementById('reportTotalVisits').textContent = report.total_visits || 0;
            document.getElementById('reportUniqueUsers').textContent = report.unique_users || 0;
            document.getElementById('reportActiveDays').textContent = report.active_days || 0;
            document.getElementById('reportAvgPerUser').textContent = (report.average_per_user || 0).toFixed(2);
            var topUser = (report.top_users && report.top_users.length) ? report.top_users[0] : null;
            document.getElementById('reportTopUser').textContent = topUser ? topUser.name : '—';
            document.getElementById('reportTopUserHours').textContent = topUser ? (topUser.total_hours || 0).toFixed(1) : '0';
            var label = document.getElementById('reportPeriodLabel');
            if (label) {
                label.textContent = 'Period: ' + (report.period || '—');
            }
        }

        function getIncludeAutoCheckout(checkboxId) {
            var checkbox = document.getElementById(checkboxId);
            return checkbox && checkbox.checked;
        }

        function buildIncludeAutoCheckoutParam(checkboxId) {
            return getIncludeAutoCheckout(checkboxId) ? '&include_auto_checkout=true' : '';
        }

        function loadWeeklyReport() {
            var url = API_BASE + '/api/statistics/weekly' + (getIncludeAutoCheckout('includeAutoCheckoutReports') ? '?include_auto_checkout=true' : '');
            fetch(url, { headers: getHeaders() })
                .then(function(r) { return r.json(); })
                .then(function(report) { renderReportSummary(report); })
                .catch(function(error) {
                    showMessage('reportsMessage', 'Error loading weekly report: ' + error.message, 'error');
                });
        }

        function loadMonthlyReport() {
            var url = API_BASE + '/api/statistics/monthly' + (getIncludeAutoCheckout('includeAutoCheckoutReports') ? '?include_auto_checkout=true' : '');
            fetch(url, { headers: getHeaders() })
                .then(function(r) { return r.json(); })
                .then(function(report) { renderReportSummary(report); })
                .catch(function(error) {
                    showMessage('reportsMessage', 'Error loading monthly report: ' + error.message, 'error');
                });
        }

        function loadCustomReport() {
            var from = document.getElementById('reportFrom').value;
            var to = document.getElementById('reportTo').value;
            if (!from || !to) {
                showMessage('reportsMessage', 'Please select both From and To dates', 'error');
                return;
            }
            var url = API_BASE + '/api/statistics/report?from=' + from + '&to=' + to + '&limit=10' + buildIncludeAutoCheckoutParam('includeAutoCheckoutReports');
            fetch(url, { headers: getHeaders() })
                .then(function(r) { return r.json(); })
                .then(function(report) { renderReportSummary(report); })
                .catch(function(error) {
                    showMessage('reportsMessage', 'Error loading custom report: ' + error.message, 'error');
                });
        }

        function renderLeaderboard(users) {
            var tbody = document.querySelector('#leaderboardTable tbody');
            tbody.innerHTML = users.map(function(u, idx) {
                return '<tr>' +
                    '<td>' + (idx + 1) + '</td>' +
                    '<td>' + u.name + '</td>' +
                    '<td>' + (u.total_hours || 0).toFixed(1) + '</td>' +
                    '<td>' + (u.visit_count || 0) + '</td>' +
                    '<td>' + (u.active_days || 0) + '</td>' +
                    '<td>' + (u.avg_duration || 0).toFixed(2) + '</td>' +
                    '</tr>';
            }).join('');
        }

        function loadLeaderboard() {
            var metric = document.getElementById('leaderboardMetric').value || 'hours';
            var limit = document.getElementById('leaderboardLimit').value || 10;
            var range = document.getElementById('leaderboardRange').value || 'weekly';
            var from = document.getElementById('leaderboardFrom').value;
            var to = document.getElementById('leaderboardTo').value;
            var includeParam = buildIncludeAutoCheckoutParam('includeAutoCheckoutReports');

            var periodPromise;
            if (range === 'weekly') {
                periodPromise = fetch(API_BASE + '/api/statistics/weekly' + includeParam.replace('&', '?'), { headers: getHeaders() })
                    .then(function(r) { return r.json(); })
                    .then(function(report) {
                        return { from: toYMD(new Date(report.start_date)), to: toYMD(new Date(report.end_date)) };
                    });
            } else if (range === 'monthly') {
                periodPromise = fetch(API_BASE + '/api/statistics/monthly' + includeParam.replace('&', '?'), { headers: getHeaders() })
                    .then(function(r) { return r.json(); })
                    .then(function(report) {
                        return { from: toYMD(new Date(report.start_date)), to: toYMD(new Date(report.end_date)) };
                    });
            } else {
                if (!from || !to) {
                    showMessage('reportsMessage', 'Please select both From and To dates for custom leaderboard', 'error');
                    return;
                }
                periodPromise = Promise.resolve({ from: from, to: to });
            }

            periodPromise.then(function(period) {
                var url = API_BASE + '/api/statistics/leaderboard?rank_by=' + metric + '&limit=' + limit;
                url += '&from=' + period.from + '&to=' + period.to + includeParam;

                fetch(url, { headers: getHeaders() })
                    .then(function(r) { return r.json(); })
                    .then(function(data) { renderLeaderboard(data.users || []); })
                    .catch(function(error) {
                        showMessage('reportsMessage', 'Error loading leaderboard: ' + error.message, 'error');
                    });
            }).catch(function(error) {
                showMessage('reportsMessage', 'Error loading leaderboard period: ' + error.message, 'error');
            });
        }

        function renderMemberSessions(sessions) {
            var tbody = document.querySelector('#memberSessionsTable tbody');
            if (!tbody) return;
            var limited = sessions.slice(0, 20);
            tbody.innerHTML = limited.map(function(s) {
                var checkIn = s.check_in ? new Date(s.check_in) : null;
                var checkOut = s.check_out ? new Date(s.check_out) : null;
                var duration = 0;
                if (checkIn && checkOut) {
                    duration = (checkOut - checkIn) / 36e5;
                }
                return '<tr>' +
                    '<td>' + (checkIn ? checkIn.toLocaleString() : '—') + '</td>' +
                    '<td>' + (checkOut ? checkOut.toLocaleString() : '—') + '</td>' +
                    '<td>' + (duration ? duration.toFixed(2) : '—') + '</td>' +
                    '</tr>';
            }).join('');
        }

        function loadMemberSessions() {
            var userId = document.getElementById('memberStatsUser').value;
            if (!userId) {
                showMessage('reportsMessage', 'Please select a member', 'error');
                return;
            }

            fetch(API_BASE + '/api/sessions/user/' + userId, { headers: getHeaders() })
                .then(function(r) { return r.json(); })
                .then(function(sessions) {
                    var sorted = (sessions || []).sort(function(a, b) {
                        return new Date(b.check_in) - new Date(a.check_in);
                    });
                    renderMemberSessions(sorted);
                })
                .catch(function(error) {
                    showMessage('reportsMessage', 'Error loading member sessions: ' + error.message, 'error');
                });
        }

        function updateRangeInputs() {
            var leaderboardRange = document.getElementById('leaderboardRange');
            var leaderboardFrom = document.getElementById('leaderboardFrom');
            var leaderboardTo = document.getElementById('leaderboardTo');
            if (leaderboardRange && leaderboardFrom && leaderboardTo) {
                var leaderboardCustom = leaderboardRange.value === 'custom';
                leaderboardFrom.disabled = !leaderboardCustom;
                leaderboardTo.disabled = !leaderboardCustom;
            }

            var memberRange = document.getElementById('memberStatsRange');
            var memberFrom = document.getElementById('memberStatsFrom');
            var memberTo = document.getElementById('memberStatsTo');
            if (memberRange && memberFrom && memberTo) {
                var memberCustom = memberRange.value === 'custom';
                memberFrom.disabled = !memberCustom;
                memberTo.disabled = !memberCustom;
            }
        }

        function toYMD(date) {
            var year = date.getFullYear();
            var month = String(date.getMonth() + 1).padStart(2, '0');
            var day = String(date.getDate()).padStart(2, '0');
            return year + '-' + month + '-' + day;
        }

        function loadMemberStatsUsers() {
            fetch(API_BASE + '/api/users?limit=500', { headers: getHeaders() })
                .then(function(r) { return r.json(); })
                .then(function(users) {
                    var select = document.getElementById('memberStatsUser');
                    if (!select) return;
                    select.innerHTML = users.map(function(u) {
                        return '<option value="' + u.id + '">' + u.name + '</option>';
                    }).join('');
                })
                .catch(function(error) {
                    showMessage('reportsMessage', 'Error loading users for member stats: ' + error.message, 'error');
                });
        }

        function findRank(users, userId) {
            for (var i = 0; i < users.length; i++) {
                if (users[i].user_id === userId || users[i].user_id === Number(userId)) {
                    return i + 1;
                }
            }
            return 0;
        }

        function renderMemberStats(userStats, totals, hoursRank, visitsRank, participants, periodLabel) {
            document.getElementById('memberTotalHours').textContent = (userStats.total_hours || 0).toFixed(1);
            document.getElementById('memberTotalVisits').textContent = userStats.visit_count || 0;
            document.getElementById('memberActiveDays').textContent = userStats.active_days || 0;
            document.getElementById('memberAvgDuration').textContent = (userStats.avg_duration || 0).toFixed(2);

            var firstVisit = userStats.first_visit ? new Date(userStats.first_visit).toLocaleDateString() : '—';
            var lastVisit = userStats.last_visit ? new Date(userStats.last_visit).toLocaleDateString() : '—';
            document.getElementById('memberFirstVisit').textContent = firstVisit;
            document.getElementById('memberLastVisit').textContent = lastVisit;

            var hoursPct = totals.total_hours ? (userStats.total_hours / totals.total_hours) * 100 : 0;
            var visitsPct = totals.total_visits ? (userStats.visit_count / totals.total_visits) * 100 : 0;

            document.getElementById('memberHoursPct').textContent = hoursPct.toFixed(1) + '%';
            document.getElementById('memberVisitsPct').textContent = visitsPct.toFixed(1) + '%';

            document.getElementById('memberRankHours').textContent = hoursRank ? (hoursRank + ' / ' + participants) : 'N/A';
            document.getElementById('memberRankVisits').textContent = visitsRank ? (visitsRank + ' / ' + participants) : 'N/A';

            var label = document.getElementById('memberPeriodLabel');
            if (label) {
                label.textContent = 'Period: ' + (periodLabel || '—');
            }
        }

        function loadMemberStats() {
            var userId = document.getElementById('memberStatsUser').value;
            if (!userId) {
                showMessage('reportsMessage', 'Please select a member', 'error');
                return;
            }

            var range = document.getElementById('memberStatsRange').value || 'weekly';
            var includeAuto = getIncludeAutoCheckout('memberIncludeAuto');
            var includeParam = includeAuto ? '&include_auto_checkout=true' : '';

            var from = document.getElementById('memberStatsFrom').value;
            var to = document.getElementById('memberStatsTo').value;

            var periodPromise;
            if (range === 'weekly') {
                periodPromise = fetch(API_BASE + '/api/statistics/weekly' + (includeAuto ? '?include_auto_checkout=true' : ''), { headers: getHeaders() })
                    .then(function(r) { return r.json(); })
                    .then(function(report) {
                        return {
                            from: toYMD(new Date(report.start_date)),
                            to: toYMD(new Date(report.end_date)),
                            label: report.period
                        };
                    });
            } else if (range === 'monthly') {
                periodPromise = fetch(API_BASE + '/api/statistics/monthly' + (includeAuto ? '?include_auto_checkout=true' : ''), { headers: getHeaders() })
                    .then(function(r) { return r.json(); })
                    .then(function(report) {
                        return {
                            from: toYMD(new Date(report.start_date)),
                            to: toYMD(new Date(report.end_date)),
                            label: report.period
                        };
                    });
            } else {
                if (!from || !to) {
                    showMessage('reportsMessage', 'Please select both From and To dates for custom range', 'error');
                    return;
                }
                periodPromise = Promise.resolve({ from: from, to: to, label: from + ' to ' + to });
            }

            periodPromise.then(function(period) {
                var userUrl = API_BASE + '/api/statistics/users/' + userId + '?from=' + period.from + '&to=' + period.to + includeParam;
                var totalsUrl = API_BASE + '/api/statistics/report?from=' + period.from + '&to=' + period.to + '&limit=10' + includeParam;
                var hoursUrl = API_BASE + '/api/statistics/leaderboard?rank_by=hours&limit=100&from=' + period.from + '&to=' + period.to + includeParam;
                var visitsUrl = API_BASE + '/api/statistics/leaderboard?rank_by=visits&limit=100&from=' + period.from + '&to=' + period.to + includeParam;

                return Promise.all([
                    fetch(userUrl, { headers: getHeaders() }).then(function(r) { return r.json(); }),
                    fetch(totalsUrl, { headers: getHeaders() }).then(function(r) { return r.json(); }),
                    fetch(hoursUrl, { headers: getHeaders() }).then(function(r) { return r.json(); }),
                    fetch(visitsUrl, { headers: getHeaders() }).then(function(r) { return r.json(); })
                ]).then(function(results) {
                    var userStats = results[0];
                    var totals = results[1];
                    var hoursUsers = results[2].users || [];
                    var visitsUsers = results[3].users || [];
                    var hoursRank = findRank(hoursUsers, userId);
                    var visitsRank = findRank(visitsUsers, userId);
                    var participants = Math.max(hoursUsers.length, visitsUsers.length);

                    renderMemberStats(userStats, totals, hoursRank, visitsRank, participants, period.label);
                });
            }).catch(function(error) {
                showMessage('reportsMessage', 'Error loading member stats: ' + error.message, 'error');
            });
        }

        checkAuth();
        loadDashboard();
        document.addEventListener('change', function(event) {
            if (!event || !event.target) return;
            if (event.target.id === 'leaderboardRange' || event.target.id === 'memberStatsRange') {
                updateRangeInputs();
            }
        });
    </script>
</body>
</html>`
}
