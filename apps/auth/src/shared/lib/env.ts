/**
 * Environment detection helpers for the auth app.
 */

/**
 * Returns true when running in a browser environment.
 */
export function isBrowserEnv(): boolean {
  return typeof window !== 'undefined' && typeof document !== 'undefined';
}

/**
 * Returns true when running in a Node.js environment.
 */
export function isNodeEnv(): boolean {
  return (
    typeof process !== 'undefined' && process.versions != null && process.versions.node != null
  );
}

/**
 * Reads an environment variable by name.
 * Falls back to the provided default when the variable is absent or empty.
 */
export function getEnvVar(name: string, fallback = ''): string {
  if (typeof process !== 'undefined' && process.env) {
    return process.env[name] ?? fallback;
  }
  return fallback;
}

/**
 * Returns the current app version from the VITE_APP_VERSION or npm_package_version env var.
 */
export function getAppVersion(): string {
  return getEnvVar('VITE_APP_VERSION') || getEnvVar('npm_package_version') || '0.0.0';
}
