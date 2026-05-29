/** Version and changelog helpers for the docs app. */

export interface VersionEntry {
  version: string;
  date?: string;
  breaking?: boolean;
}

/**
 * Parses a semantic version string into its components.
 * Returns null for invalid version strings.
 */
export function parseSemver(
  version: string
): { major: number; minor: number; patch: number } | null {
  const match = version.replace(/^v/, '').match(/^(\d+)\.(\d+)\.(\d+)$/);
  if (!match) return null;
  return { major: parseInt(match[1]), minor: parseInt(match[2]), patch: parseInt(match[3]) };
}

/**
 * Compares two semver strings. Returns -1, 0, or 1.
 */
export function compareSemver(a: string, b: string): number {
  const pa = parseSemver(a);
  const pb = parseSemver(b);
  if (!pa || !pb) return 0;
  if (pa.major !== pb.major) return pa.major > pb.major ? 1 : -1;
  if (pa.minor !== pb.minor) return pa.minor > pb.minor ? 1 : -1;
  if (pa.patch !== pb.patch) return pa.patch > pb.patch ? 1 : -1;
  return 0;
}

/**
 * Returns the latest version from a list of version entries (by semver comparison).
 */
export function latestVersion(entries: VersionEntry[]): VersionEntry | undefined {
  if (entries.length === 0) return undefined;
  return entries.reduce((latest, entry) =>
    compareSemver(entry.version, latest.version) > 0 ? entry : latest
  );
}

/**
 * Returns true if the version is a major version (x.0.0).
 */
export function isMajorVersion(version: string): boolean {
  const parsed = parseSemver(version);
  return parsed !== null && parsed.minor === 0 && parsed.patch === 0;
}

/**
 * Returns version entries that contain breaking changes.
 */
export function breakingVersions(entries: VersionEntry[]): VersionEntry[] {
  return entries.filter((e) => e.breaking === true);
}
