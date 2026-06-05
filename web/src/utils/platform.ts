/* eslint-disable @typescript-eslint/no-explicit-any */
export type Platform = 'windows' | 'macos' | 'linux' | 'unknown';

export interface ClientPaths {
  windows: string;
  macos: string;
  linux: string;
}

export interface PlatformConfig {
  rootKey: "mcpServers" | "servers";
  httpField: "url" | "serverUrl";
  requireType: boolean;
  pathEnvVar: string;
  pathSeparator: string;
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

export const getPlatformConfig = (platform: Platform): PlatformConfig => {
  const configs: Record<Platform, PlatformConfig> = {
    windows: {
      rootKey: "mcpServers",
      httpField: "url",
      requireType: false,
      pathEnvVar: "%APPDATA%",
      pathSeparator: "\\"
    },
    macos: {
      rootKey: "mcpServers",
      httpField: "url",
      requireType: false,
      pathEnvVar: "~/Library/Application Support",
      pathSeparator: "/"
    },
    linux: {
      rootKey: "mcpServers",
      httpField: "url",
      requireType: false,
      pathEnvVar: "~/.config",
      pathSeparator: "/"
    },
    unknown: {
      rootKey: "mcpServers",
      httpField: "url",
      requireType: false,
      pathEnvVar: "~",
      pathSeparator: "/"
    }
  };
  return configs[platform];
};

export const getPlatformSpecificFormat = (platform: Platform): Record<string, any> => {
  const formats: Record<Platform, Record<string, any>> = {
    windows: {
      pathFormat: "windows",
      envVarPrefix: "%",
      envVarSuffix: "%"
    },
    macos: {
      pathFormat: "unix",
      envVarPrefix: "${",
      envVarSuffix: "}"
    },
    linux: {
      pathFormat: "unix",
      envVarPrefix: "${",
      envVarSuffix: "}"
    },
    unknown: {
      pathFormat: "unix",
      envVarPrefix: "${",
      envVarSuffix: "}"
    }
  };
  return formats[platform];
};