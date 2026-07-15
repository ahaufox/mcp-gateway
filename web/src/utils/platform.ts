export type Platform = 'windows' | 'macos' | 'linux' | 'unknown';

export interface ClientPaths {
  windows: string;
  macos: string;
  linux: string;
  // unknown 是 Platform 类型中的兜底值（detectPlatform 在浏览器 UA 无法
  // 识别时返回）。平台配置 JSON 不必提供该键；下标访问在 unknown 上会
  // 返回 undefined，由调用方决定如何降级（例如显示 "${client}.json"）。
  unknown?: string;
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

export interface PlatformFormat {
  pathFormat: string;
  envVarPrefix: string;
  envVarSuffix: string;
}

export const getPlatformSpecificFormat = (platform: Platform): PlatformFormat => {
  const formats: Record<Platform, PlatformFormat> = {
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