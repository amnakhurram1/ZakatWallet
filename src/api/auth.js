import { post } from './client.js';

/**
 * Request a one‑time password for the specified email address.
 * Wraps the `POST /auth/request-otp` endpoint.
 *
 * @param {string} email The user's email address
 * @returns {Promise<{email:string, otp:string}>}
 */
export function requestOtp(email) {
  return post('/auth/request-otp', { email });
}

/**
 * Verify the one‑time password for the specified email address.
 * Wraps the `POST /auth/verify-otp` endpoint.
 *
 * @param {string} email The user's email address
 * @param {string} otp The one‑time password provided by the server
 * @returns {Promise<{success:boolean, message:string}>}
 */
export function verifyOtp(email, otp) {
  return post('/auth/verify-otp', { email, otp });
}

/**
 * Register a new user.  Creates a Supabase user record and a
 * blockchain wallet.  The API expects snake‑case keys (full_name).
 *
 * @param {{fullName:string, email:string, cnic:string}} data Registration details
 * @returns {Promise<{
 *   user_id:string,
 *   full_name:string,
 *   email:string,
 *   cnic:string,
 *   wallet_address:string,
 *   private_key:string
 * }>} The created user and wallet details
 */
export function registerUser({ fullName, email, cnic }) {
  // Map camelCase to the snake_case expected by the backend
  return post('/register', {
    full_name: fullName,
    email,
    cnic,
  });
}
