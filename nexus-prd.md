# 📄 Product Requirements Document (PRD)
**Product Name:** Nexus Integration Platform (v1.0 - Enterprise Readiness)
**Status:** Draft / Proposed
**Date:** December 11, 2025
**Author:** Product Manager (AI Partner)

---

## 1. Product Vision & Scope
### 1.1 Vision Statement
To provide a **cloud-native, high-performance integration platform** that bridges the gap between modern API ecosystems and legacy enterprise systems. Unlike UNCAL or MuleSoft which are heavy and complex, Nexus uses a modern stack (Golang + Redis) to deliver **Zero-Data-Loss reliability** with **1/10th the infrastructure footprint**.

### 1.2 Target Audience
* **Primary:** Mid-to-Large Enterprises (Fintech, Retail, Logistics) looking to modernize legacy ESBs.
* **User Persona:** "DevOps Engineer" or "Backend Developer" who values speed, CLI/API-first configuration, and observability over complex GUI wizards.

### 1.3 Strategic Goals (OKRs)
* **Reliability:** Achieve "Zero Data Loss" guarantee using Redis Streams persistence.
* **Performance:** Maintain <10ms latency at 10,000 RPS (Requests Per Second) on standard hardware.
* **Connectivity:** Reach parity with UNCAL’s core connectors (DB, REST, File) by Q1.

---

## 2. Competitive Gap Analysis Summary
*Comparison against UNCAL (Local), MuleSoft (Global), Seeburger (Legacy).*

| Feature Area | Competitor Status (UNCAL/Mule) | Nexus Target State (The Gap Filler) |
| :--- | :--- | :--- |
| **Architecture** | Heavy Java-based; High RAM usage; Complex install. | **Lightweight Golang Binary**; Container-native; Low footprint. |
| **Reliability** | Standard Message Queues (Disk/DB). | **Redis Streams (In-Memory + AOF Persistence)** for instant replayability. |
| **Observability** | "Black Box" logs or expensive proprietary dashboards. | **OpenTelemetry Support** (Trace ID per request) + Real-time Metrics. |
| **Large Files** | Struggles with 1GB+ files (Memory crashes). | **Streaming ETL**: Process 10GB+ files line-by-line with constant RAM. |
| **Pricing** | Expensive (Core-based) or Complex Modules. | **Transparent**, Node-based pricing. |

---

## 3. Functional Requirements (Features)

### 3.1 Core Engine: Reliability & Processing (Priority: P0 - Critical)
*Ref: "The Iron-Clad Reliability Upgrade"*

| ID | Feature Name | Description | Acceptance Criteria |
| :--- | :--- | :--- | :--- |
| **BE-01** | **Redis Streams Ingest** | Replace memory caching with Redis Streams. All incoming data is written to a Stream before processing. | System must recover all pending messages after a hard restart. |
| **BE-02** | **Consumer Groups** | Implement Golang Consumer Groups to allow parallel processing of messages. | Multiple worker nodes can consume from the same stream without duplication. |
| **BE-03** | **Dead Letter Queue (DLQ)** | Move messages that fail processing 5x to a separate "Error Stream" for manual review. | Failed messages appear in the "Error" tab in UI, not lost. |
| **BE-04** | **Graceful Shutdown** | Workers must finish current job before terminating during deployments. | No "interrupted" logs during service restart. |

### 3.2 Connectivity & Adapters (Priority: P0 & P1)
*Ref: "Any-to-Any Strategy"*

| ID | Feature Name | Description | Acceptance Criteria |
| :--- | :--- | :--- | :--- |
| **CON-01** | **Generic REST Client** | Configurable HTTP Client (Method, Headers, Body, Auth) to call 3rd party APIs. | Can successfully post JSON to Slack/Salesforce API. |
| **CON-02** | **Database CDC (Change Data Capture)** | Listen to DB `INSERT`/`UPDATE` events to trigger workflows instantly (Polling fallback if CDC complex). | New row in PostgreSQL triggers Nexus workflow within 1 second. |
| **CON-03** | **File Watcher (CSV/JSON)** | Monitor a local folder or S3 bucket for new files. Stream process them. | 100MB CSV file is parsed and inserted to DB without OOM. |
| **CON-04** | **ISO 8583 Parser (P2)** | Golang library to parse raw ISO 8583 banking messages. | Can decode a standard ATM request message into JSON. |

### 3.3 Data Transformation (Priority: P1)
*Ref: UNCAL's "UDM" Gap*

| ID | Feature Name | Description | Acceptance Criteria |
| :--- | :--- | :--- | :--- |
| **TR-01** | **Visual Data Mapper** | UI to draw lines between Source JSON fields and Destination JSON fields. | User can map `user_name` -> `fullName` without writing code. |
| **TR-02** | **JSON Logic Engine** | Apply basic logic (If/Else, Concatenate) during mapping. | Can combine `first` + `last` names during transfer. |

### 3.4 Management Dashboard (Next.js) (Priority: P1)
*Ref: Enterprise UX*

| ID | Feature Name | Description | Acceptance Criteria |
| :--- | :--- | :--- | :--- |
| **UI-01** | **Real-Time Traffic Monitor** | Charts showing RPS (Requests Per Second) and Error Rates. | Charts update every 5-10 seconds. |
| **UI-02** | **Message Trace View** | Search for a specific transaction ID and see its status (Pending, Success, Failed). | "Where is my Order #123?" query returns instant status. |
| **UI-03** | **RBAC (Role Based Access)** | "Admin" (Read/Write) vs "Viewer" (Read Only) roles. | Viewer cannot delete a workflow. |

---

## 4. Non-Functional Requirements (NFRs)
*The "Enterprise Grade" constraints.*

1.  **Scalability:** The system must handle a sudden spike of **50,000 requests/minute** without crashing (Queued in Redis).
2.  **Latency:** 95% of standard HTTP-to-DB transactions must complete in **< 50ms**.
3.  **Security:**
    * All "Secrets" (DB Passwords, API Keys) must be encrypted at rest (AES-256).
    * API Endpoints must be secured via Bearer Token / API Key.
4.  **Auditability:** Every configuration change (e.g., changing a target DB host) must be logged in a permanent Audit Log (ClickHouse/File).

---

## 5. Technical Architecture & Stack
* **Backend:** Golang (1.21+)
* **Queue/Broker:** Redis (Streams + Consumer Groups)
* **Frontend:** Next.js (React) + Tailwind CSS
* **Long-Term Storage (Logs):** ClickHouse (Recommended for V1.1) or PostgreSQL.
* **Deployment:** Docker Compose (MVP) -> Kubernetes (Enterprise).

---

## 6. Roadmap / Phasing

### **Phase 1: The "Trustable Core" (MVP - Month 1-2)**
* **Focus:** Reliability. Replace memory cache with Redis Streams.
* **Deliverables:**
    * Golang Worker Pool implementation.
    * Generic REST & DB Connectors.
    * Basic Dashboard (Status only).
    * Logs table in UI to show wether data is successfully sent, sent from retry, or data is deleted from redis because the destination is down.
    * *Result:* A product that never loses data.

### **Phase 2: The "Integrator" (Month 3-4)**
* **Focus:** Usability & Formats.
* **Deliverables:**
    * Visual Data Mapper (Drag & Drop).
    * CSV/Excel File Streaming.
    * RBAC & Audit Logs.
    * *Result:* A product business users can actually configure.

### **Phase 3: The "Enterprise Slayer" (Month 5+)**
* **Focus:** Legacy Protocols & Scale.
* **Deliverables:**
    * ISO 8583 (Banking) Adapter.
    * SAP RFC Connector.
    * ClickHouse integration for massive log retention.
    * *Result:* Competitor to UNCAL/MuleSoft in Banking/Retail.

---

## 7. Risks & Mitigation

| Risk | Impact | Mitigation Strategy |
| :--- | :--- | :--- |
| **Redis Memory Full** | Critical (System Stop) | Implement "Max Length" on Streams (e.g., keep last 1M messages) + Offload old data to Disk/ClickHouse. |
| **Complex Mapping Logic** | High (User Frustration) | Keep UI simple. Allow "Code Mode" (write raw Lua/JS) for complex logic where UI fails. |
| **Legacy Protocol Complexity** | Medium (Dev Delays) | Don't build ISO 8583 from scratch. Use established Go libraries (e.g., `moov-io/iso8583`). |