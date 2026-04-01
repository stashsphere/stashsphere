export interface VersionInfo {
  version: string;
  gitRevision: string;
}

const DEFAULT_VERSION: VersionInfo = {
  version: 'dev',
  gitRevision: 'unknown',
};

const versionModules = import.meta.glob<{ default: VersionInfo }>('../version.json', {
  eager: false,
});

export async function getVersionInfo(): Promise<VersionInfo> {
  try {
    const versionPath = '../version.json';
    const loader = versionModules[versionPath];

    if (loader) {
      const module = await loader();
      return module.default;
    } else {
      throw new Error('version.json not found in glob');
    }
  } catch {
    console.warn('version.json not found, using default version info');
    return DEFAULT_VERSION;
  }
}
