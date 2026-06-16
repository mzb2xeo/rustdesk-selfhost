const TokenKey = 'access_token'
const OidcCode = 'oidc_code'
const OidcCodeExpiry = 'oidc_code_expiry';

export function getToken () {
  return localStorage.getItem(TokenKey)
}

export function setToken (token) {
  localStorage.setItem(`wc-option:local:access_token`, token)
  return localStorage.setItem(TokenKey, token)
}

export function removeToken () {
  return localStorage.removeItem(TokenKey)
}

// Set code and store the current timestamp (unit: milliseconds)
export function setCode(code) {
  const now = Date.now(); // Current timestamp (milliseconds)
  const expiry = now + 60 * 1000; // Expires in 60 seconds

  localStorage.setItem(OidcCode, code); // store code
  localStorage.setItem(OidcCodeExpiry, expiry); // Store expiration timestamp
}

// Get the code, delete it and return null if it has expired
export function getCode() {
  const expiry = localStorage.getItem(OidcCodeExpiry); // Get expiration timestamp
  const now = Date.now(); // current timestamp

  if (expiry && now > parseInt(expiry)) {
    // If it has expired, delete the code and expiration time
    removeCode();
    return null;
  }
  return localStorage.getItem(OidcCode); // Return code (if not expired)
}

// Delete code and expiration time
export function removeCode() {
  localStorage.removeItem(OidcCode);
  localStorage.removeItem(OidcCodeExpiry);
}
