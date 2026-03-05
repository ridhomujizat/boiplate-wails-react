# Upload API Documentation

Screen recording upload uses a **two-step signed-URL flow**. The client first requests a signed PUT URL, uploads the file directly to Google Cloud Storage, then confirms completion.

> **Authentication**: Both endpoints require `Authorization: Bearer <token>` header.

---

## Flow Overview

```
Client                          Backend                         GCS
  │                                │                              │
  │  POST /api/record/upload/init  │                              │
  │  ─────────────────────────────>│                              │
  │                                │  Generate signed PUT URL     │
  │  { upload_url, upload_id }     │                              │
  │  <─────────────────────────────│                              │
  │                                │                              │
  │  PUT upload_url (binary body)  │                              │
  │  ─────────────────────────────────────────────────────────────>│
  │                                │                              │
  │  POST /api/record/upload/done  │                              │
  │  ─────────────────────────────>│  Verify file exists in GCS   │
  │                                │  ─────────────────────────-->│
  │  { record, download_url }      │                              │
  │  <─────────────────────────────│                              │
```

---

## 1. Initiate Upload

Generates a signed PUT URL so the client can upload directly to GCS.

### Request

```
POST /api/record/upload/init
Content-Type: application/json
Authorization: Bearer <token>
```

**Body:**

| Field          | Type     | Required | Description                                    |
|----------------|----------|----------|------------------------------------------------|
| `session_id`   | `string` | ✅       | Recording session ID (must exist and belong to the current user) |
| `filename`     | `string` | ✅       | Original file name (e.g. `recording.webm`)     |
| `content_type` | `string` | ✅       | MIME type, must start with `video/` (e.g. `video/webm`, `video/mp4`) |
| `file_size`    | `int`    | ✅       | File size in bytes. Must be > 0, max **1 GB** (1073741824 bytes) |

**Example:**

```json
{
  "session_id": "abc12345-def6-7890-ghij-klmnopqrstuv",
  "filename": "screen-recording-2026-03-05.webm",
  "content_type": "video/webm",
  "file_size": 52428800
}
```

### Response (200 OK)

```json
{
  "code": 200,
  "message": "Signed upload URL generated",
  "data": {
    "upload_url": "https://storage.googleapis.com/bucket/records/abc12345/uploadid_screen-recording.webm?X-Goog-Signature=...",
    "upload_id": "a1b2c3d4e5f6g7h8",
    "record_id": 42,
    "expires_at": "2026-03-05T11:00:00Z"
  }
}
```

| Field        | Type       | Description                                      |
|--------------|------------|--------------------------------------------------|
| `upload_url` | `string`   | Signed PUT URL — use this to upload the file directly |
| `upload_id`  | `string`   | Unique upload identifier (needed for `/upload/done`) |
| `record_id`  | `uint`     | Internal record ID                               |
| `expires_at` | `datetime` | URL expiration time (1 hour from creation)       |

### Uploading the File

After receiving the `upload_url`, upload the file with a **PUT** request directly to GCS:

```bash
curl -X PUT \
  -H "Content-Type: video/webm" \
  --data-binary @screen-recording.webm \
  "<upload_url>"
```

### Error Responses

| Code | Message                              | Cause                              |
|------|--------------------------------------|------------------------------------|
| 400  | `content_type must be a video/* MIME type` | Non-video content type         |
| 400  | `file_size exceeds maximum allowed size`  | File > 1 GB                   |
| 400  | `Validation error`                   | Missing required fields            |
| 401  | `Unauthorized`                       | Missing or invalid Bearer token    |
| 403  | `Session does not belong to current user` | Session owned by another user |
| 404  | `Session not found`                  | Invalid `session_id`               |
| 500  | `Cloud storage not configured`       | GCS not set up on server           |

---

## 2. Complete Upload

Verifies the uploaded file exists in GCS and marks the recording as successful.

### Request

```
POST /api/record/upload/done
Content-Type: application/json
Authorization: Bearer <token>
```

**Body:**

| Field        | Type     | Required | Description                                   |
|--------------|----------|----------|-----------------------------------------------|
| `session_id` | `string` | ✅       | Same session ID used in `/upload/init`        |
| `upload_id`  | `string` | ✅       | The `upload_id` returned from `/upload/init`  |

**Example:**

```json
{
  "session_id": "abc12345-def6-7890-ghij-klmnopqrstuv",
  "upload_id": "a1b2c3d4e5f6g7h8"
}
```

### Response (200 OK)

```json
{
  "code": 200,
  "message": "Upload completed successfully",
  "data": {
    "record": {
      "id": 42,
      "session_id": "abc12345-def6-7890-ghij-klmnopqrstuv",
      "user_id": 2,
      "last_action": "upload_done",
      "status": "success",
      "error_message": null,
      "created_at": "2026-03-05T10:00:00Z",
      "updated_at": "2026-03-05T10:05:00Z"
    },
    "download_url": "https://storage.googleapis.com/bucket/records/abc12345/uploadid_screen-recording.webm"
  }
}
```

| Field          | Type     | Description                                 |
|----------------|----------|---------------------------------------------|
| `record`       | `object` | Updated record with `status: "success"`     |
| `download_url` | `string` | Public download URL for the uploaded file   |

### Error Responses

| Code | Message                                          | Cause                                     |
|------|--------------------------------------------------|-------------------------------------------|
| 400  | `upload_id does not match`                       | `upload_id` doesn't match the session's   |
| 400  | `File not found in storage — upload may not have completed` | File not yet in GCS        |
| 400  | `Validation error`                               | Missing required fields                   |
| 401  | `Unauthorized`                                   | Missing or invalid Bearer token           |
| 403  | `Session does not belong to current user`        | Session owned by another user             |
| 404  | `Session not found`                              | Invalid `session_id`                      |
| 500  | `Cloud storage not configured`                   | GCS not set up on server                  |
