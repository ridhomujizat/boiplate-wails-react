# Activity Monitoring Feature Prompt for Wails App

## Context

You are a senior desktop application engineer with deep experience in **Wails (Go + Web Frontend)**, **system-level activity tracking**, and **analytics dashboards**.

I already have an **existing Wails application**.
Your task is to **extend the application** by adding an **Activity Monitoring feature similar to ActivityWatch**.

---

## Goals

Add a feature to **record user activity on the operating system** and visualize it in a **dashboard**.

The feature must:

* Track **active applications**
* Track **active window / tab**
* Detect **AFK (idle) vs active**
* Aggregate activity data
* Display insights in a dashboard

---

## Functional Requirements

### 1 Activity Tracking (Background Service)

Implement a background activity tracker that runs while the Wails app is active.

Track the following events:

#### Application Activity

* Application name
* Process name / bundle id
* Window title (if available)
* Start time & end time
* Total active duration

#### Tab / Window Activity (if supported by OS)

* Browser name (Chrome, Edge, Firefox)
* Active tab title
* Domain / URL (optional & configurable)

#### AFK Detection

* Detect idle state based on:

  * No mouse movement
  * No keyboard input
* Configurable threshold (default: 3–5 minutes)
* Mark activity as:

  * `active`
  * `afk`

---

## 2️ Data Model

Design a local data schema (SQLite / embedded DB):

### ActivityEvent

* id
* app_name
* window_title
* tab_title (optional)
* start_time
* end_time
* duration
* status (`active | afk`)
* os (`windows | darwin | linux`)

### AggregatedStats

* app_name
* total_active_time
* date

---

## 3 Backend (Go – Wails)

* Use Go to:

  * Detect active window (Windows & macOS)
  * Capture idle time
  * Store data locally
* Expose APIs via Wails bindings:

  * `GetTimeline(dateRange)`
  * `GetTopApplications(limit)`
  * `GetActivitySummaryByApp()`

---

## Dashboard Requirements (Frontend)

Create a **Dashboard Page** with the following components:

### 1 Timeline View

* Horizontal timeline per day
* Color-coded:

  * Active
  * AFK
* Grouped by application
* Zoomable by time range (hour / day)

### 2 Top Applications Chart

* Bar / Pie chart
* Top applications by total active time
* Configurable date range

### 3 Activity Time per App

* Table or stacked bar chart
* Columns:

  * Application
  * Active Time
  * AFK Time
  * Percentage

---

## UI / UX Guidelines

* Clean analytics-style dashboard
* Similar UX to ActivityWatch
* Responsive layout
* Dark mode friendly
* Real-time update (polling or event-based)

---

## Technical Constraints

* Must work inside **existing Wails project**
* Cross-platform:

  * macOS (darwin)
  * Windows
* No cloud dependency (local-first)
* User privacy:

  * All data stored locally
  * No external tracking

---

## Deliverables

1. Architecture explanation
2. Go backend implementation strategy
3. Frontend dashboard component structure
4. Suggested libraries for:

   * Active window detection
   * Idle detection
   * Chart visualization
5. Example API contracts between Go & frontend

---

##  Optional Enhancements

* Daily / weekly reports
* Export to JSON / CSV
* Ignore application list
* Pause / resume tracking
* Tray indicator for tracking status
