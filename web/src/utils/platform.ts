export type Platform = 'windows' | 'macos' | 'linux' | 'unknown';

export interface ClientPaths {
  windows: string;
  macos: string;
  linux: string;
}

export const detectPlatform = (): Platform => {
  if (typeof window === 'undefined') {
    return 'unknown';
  }

  const userAgent = window.navigator.userAgent.toLowerCase();
  
  if (userAgent.includes('win')) {
    return 'windows';
  }
  
  if (userAgent.includes('mac') || userAgent.includes('darwin')) {
    return 'macos';
  }
  
  if (userAgent.includes('linux')) {
    return 'linux';
  }
  
  return 'unknown';
};

export const getPlatformLabel = (platform: Platform): string => {
  const labels: Record<Platform, string> = {
    windows: 'Windows',
    macos: 'macOS',
    linux: 'Linux',
    unknown: '未知系统'
  };
  return labels[platform];
};

export const getConfigPathForPlatform = (paths: ClientPaths, platform: Platform): string => {
  switch (platform) {
    case 'windows':
      return paths.windows;
    case 'macos':
      return paths.macos;
    case 'linux':
      return paths.linux;
    default:
      return paths.macos;
  }
};
