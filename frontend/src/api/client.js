// API client helper functions.
//
// The backend base URL is fixed based on the Go server definition in
// cmd/server/main.go.  All API requests should prefix their
// endpoints with this value.  The request helper wraps the Fetch
// API and automatically parses JSON responses; non‑2xx responses
// throw the parsed JSON (if any) to simplify error handling.

export const API_BASE_URL = 'http://localhost:8080/api/v1';

// Internal helper to parse responses and throw for non‑2xx status codes.
async function handleResponse(res) {
  // Attempt to parse JSON if the response declares a JSON content type
  const contentType = res.headers.get('content-type') || '';
  let data = null;
  if (contentType.includes('application/json')) {
    try {
      data = await res.json();
    } catch (_err) {
      data = null;
    }
  } else {
    // Fallback to text for non‑JSON responses
    data = await res.text().catch(() => null);
  }
  if (!res.ok) {
    // Prefer server supplied error details if available
    throw data || { error: 'Request failed' };
  }
  return data;
}

/**
 * Perform a GET request against the backend.
 *
 * @param {string} endpoint Relative API path starting with '/'
 * @returns {Promise<any>} Parsed JSON response
 */
export async function get(endpoint) {
  const res = await fetch(`${API_BASE_URL}${endpoint}`, {
    method: 'GET',
    // GET requests typically have no body and don't need content type headers
  });
  return handleResponse(res);
}

/**
 * Perform a POST request against the backend.  Automatically
 * stringifies the supplied body and sets the JSON content type.
 *
 * @param {string} endpoint Relative API path starting with '/'
 * @param {object} body JSON‑serialisable payload
 * @returns {Promise<any>} Parsed JSON response
 */
export async function post(endpoint, body = {}) {
  const res = await fetch(`${API_BASE_URL}${endpoint}`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(body),
  });
  return handleResponse(res);
}

// Legacy request helper retained for compatibility.  This uses the
// default JSON content type and may be removed when all code
// switches to get() and post().
export async function request(endpoint, options = {}) {
  const res = await fetch(`${API_BASE_URL}${endpoint}`, {
    headers: {
      'Content-Type': 'application/json',
      ...(options.headers || {}),
    },
    ...options,
  });
  return handleResponse(res);
}
