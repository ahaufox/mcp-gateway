import React from "react";
import {
  Monitor,
  Puzzle,
  Cloud,
  Sparkles,
  Bot,
  Cpu,
  Braces,
  Layers,
  Play,
  Terminal,
  SquareTerminal,
  Globe,
  Zap,
  Box
} from "lucide-react";
import { type Platform, type ClientPaths } from "./platform";

export interface ClientDef {
  id: string;
  name: string;
  icon: React.ComponentType<{ className?: string }>;
  desc: string;
  category: "ide" | "terminal" | "assistant" | "platform" | "native";
  color: "emerald" | "violet" | "indigo" | "amber" | "rose" | "cyan" | "orange" | "teal" | "blue" | "green" | "purple" | "slate";
  fmtType: "generic" | "proxy";
  configPaths: ClientPaths;
  keywords: string[];
  configFormat: {
    rootKey: "mcpServers" | "servers" | "context_servers";
    httpField?: "url" | "serverUrl";
    requireType?: boolean;
    useStdioBridge?: boolean;
    platformOverrides?: Partial<Record<Platform, {
      rootKey?: "mcpServers" | "servers" | "context_servers";
      httpField?: "url" | "serverUrl";
      requireType?: boolean;
      useStdioBridge?: boolean;
    }>>;
  };
}

export const CLIENTS: ClientDef[] = [
  // IDE / 编辑器
  { 
    id: "claude", 
    name: "Claude Desktop", 
    icon: Monitor, 
    desc: "Anthropic 官方桌面客户端", 
    category: "ide", 
    color: "emerald", 
    fmtType: "generic", 
    configPaths: {
      windows: "%APPDATA%\\Claude\\claude_desktop_config.json",
      macos: "~/Library/Application Support/Claude/claude_desktop_config.json",
      linux: "~/.config/Claude/claude_desktop_config.json"
    }, 
    keywords: ["claude", "anthropic", "desktop"],
    configFormat: { rootKey: "mcpServers", useStdioBridge: true }
  },
  { 
    id: "cursor", 
    name: "Cursor", 
    icon: Puzzle, 
    desc: "AI-first 代码编辑器", 
    category: "ide", 
    color: "indigo", 
    fmtType: "generic", 
    configPaths: {
      windows: "%APPDATA%\\Cursor\\mcp.json",
      macos: "~/Library/Application Support/Cursor/mcp.json",
      linux: "~/.config/Cursor/mcp.json"
    }, 
    keywords: ["cursor", "ai editor"],
    configFormat: { rootKey: "mcpServers", httpField: "url", requireType: false }
  },
  { 
    id: "windsurf", 
    name: "Windsurf", 
    icon: Cloud, 
    desc: "Codeium 流式 AI IDE", 
    category: "ide", 
    color: "cyan", 
    fmtType: "generic", 
    configPaths: {
      windows: "%APPDATA%\\Codeium\\Windsurf\\mcp_config.json",
      macos: "~/Library/Application Support/Codeium/Windsurf/mcp_config.json",
      linux: "~/.codeium/windsurf/mcp_config.json"
    }, 
    keywords: ["windsurf", "codeium"],
    configFormat: { rootKey: "mcpServers", httpField: "serverUrl", requireType: false }
  },
  { 
    id: "trae", 
    name: "Trae IDE", 
    icon: Sparkles, 
    desc: "字节跳动 AI 开发环境", 
    category: "ide", 
    color: "violet", 
    fmtType: "generic", 
    configPaths: {
      windows: "%APPDATA%\\Trae\\mcp_config.json",
      macos: "~/Library/Application Support/Trae/mcp_config.json",
      linux: "~/.trae/mcp_config.json"
    }, 
    keywords: ["trae", "字节跳动", "bytedance"],
    configFormat: { rootKey: "mcpServers", httpField: "url", requireType: true }
  },
  { 
    id: "cline", 
    name: "Cline", 
    icon: Bot, 
    desc: "VS Code 全能 AI 助手", 
    category: "ide", 
    color: "rose", 
    fmtType: "generic", 
    configPaths: {
      windows: "%APPDATA%\\.cline\\mcp_settings.json",
      macos: "~/.cline/mcp_settings.json",
      linux: "~/.cline/mcp_settings.json"
    }, 
    keywords: ["cline", "vscode extension"],
    configFormat: { rootKey: "mcpServers", httpField: "url", requireType: false }
  },
  { 
    id: "roocode", 
    name: "Roo Code", 
    icon: Cpu, 
    desc: "VS Code 多模型 AI 编程", 
    category: "ide", 
    color: "orange", 
    fmtType: "generic", 
    configPaths: {
      windows: "%APPDATA%\\.roo-code\\mcp_settings.json",
      macos: "~/.roo-code/mcp_settings.json",
      linux: "~/.roo-code/mcp_settings.json"
    }, 
    keywords: ["roo", "roo code", "vscode"],
    configFormat: { rootKey: "mcpServers", httpField: "url", requireType: false }
  },
  { 
    id: "vscode", 
    name: "VS Code", 
    icon: Braces, 
    desc: "微软编辑器 MCP 扩展", 
    category: "ide", 
    color: "blue", 
    fmtType: "generic", 
    configPaths: {
      windows: "%APPDATA%\\Code\\User\\mcp.json",
      macos: "~/Library/Application Support/Code/User/mcp.json",
      linux: "~/.config/Code/User/mcp.json"
    }, 
    keywords: ["vscode", "visual studio", "microsoft"],
    configFormat: { rootKey: "servers", httpField: "url", requireType: false }
  },
  { 
    id: "zed", 
    name: "Zed Editor", 
    icon: Layers, 
    desc: "高性能协作编辑器", 
    category: "ide", 
    color: "slate", 
    fmtType: "generic", 
    configPaths: {
      windows: "%APPDATA%\\Zed\\settings.json",
      macos: "~/.zed/settings.json",
      linux: "~/.config/zed/settings.json"
    }, 
    keywords: ["zed", "editor"],
    configFormat: { rootKey: "context_servers", useStdioBridge: true }
  },
  { 
    id: "continue", 
    name: "Continue", 
    icon: Play, 
    desc: "开源 AI 代码助手", 
    category: "ide", 
    color: "purple", 
    fmtType: "generic", 
    configPaths: {
      windows: "%APPDATA%\\Continue\\config.json",
      macos: "~/.continue/config.json",
      linux: "~/.continue/config.json"
    }, 
    keywords: ["continue", "continue.dev"],
    configFormat: { rootKey: "mcpServers", httpField: "url", requireType: true }
  },

  // 终端 / CLI
  { 
    id: "codex", 
    name: "Codex CLI", 
    icon: Terminal, 
    desc: "OpenAI 终端编程助手", 
    category: "terminal", 
    color: "teal", 
    fmtType: "generic", 
    configPaths: {
      windows: "%USERPROFILE%\\.codex\\mcp.json",
      macos: "~/.codex/mcp.json",
      linux: "~/.codex/mcp.json"
    }, 
    keywords: ["codex", "openai", "cli"],
    configFormat: { rootKey: "mcpServers", httpField: "url", requireType: false }
  },
  { 
    id: "warp", 
    name: "Warp Terminal", 
    icon: SquareTerminal, 
    desc: "Rust 重写智能终端", 
    category: "terminal", 
    color: "amber", 
    fmtType: "generic", 
    configPaths: {
      windows: "%APPDATA%\\Warp\\mcp.json",
      macos: "~/.warp/mcp.json",
      linux: "~/.config/warp/mcp.json"
    }, 
    keywords: ["warp", "terminal", "rust"],
    configFormat: { rootKey: "mcpServers", httpField: "url", requireType: false }
  },

  // AI 平台 / 助手
  { 
    id: "antigravity", 
    name: "Antigravity", 
    icon: Globe, 
    desc: "Gemini 生态 AI 扩展", 
    category: "platform", 
    color: "indigo", 
    fmtType: "generic", 
    configPaths: {
      windows: "%APPDATA%\\Gemini\\Antigravity\\mcp_config.json",
      macos: "~/.gemini/antigravity/mcp_config.json",
      linux: "~/.config/gemini/antigravity/mcp_config.json"
    }, 
    keywords: ["antigravity", "gemini", "google"],
    configFormat: { rootKey: "mcpServers", httpField: "serverUrl", requireType: true }
  },
  { 
    id: "openinterpreter", 
    name: "Open Interpreter", 
    icon: Zap, 
    desc: "自然语言操控计算机", 
    category: "assistant", 
    color: "green", 
    fmtType: "generic", 
    configPaths: {
      windows: "%APPDATA%\\Open Interpreter\\mcp.json",
      macos: "~/.open-interpreter/mcp.json",
      linux: "~/.open-interpreter/mcp.json"
    }, 
    keywords: ["open interpreter", "interpreter"],
    configFormat: { rootKey: "mcpServers", httpField: "url", requireType: false }
  },

  // mcp-proxy 原生
  { 
    id: "proxy", 
    name: "mcp-proxy", 
    icon: Box, 
    desc: "本网关代理原生配置", 
    category: "native", 
    color: "amber", 
    fmtType: "proxy", 
    configPaths: {
      windows: "config.json",
      macos: "config.json",
      linux: "config.json"
    }, 
    keywords: ["mcp-proxy", "proxy", "gateway"],
    configFormat: { rootKey: "mcpServers", httpField: "url", requireType: false }
  },
];

export interface ConvertOptions {
  tokenOverride: string;
  selectedServers: Set<string>;
  clientConfig: ClientDef;
  platform: Platform;
}

export const formatAuthHeader = (token: string): string => {
  const trimmed = token.trim();
  if (!trimmed) return "";
  return trimmed.toLowerCase().startsWith("bearer ") ? trimmed : `Bearer ${trimmed}`;
};

export const convertToProxy = (fromConfig: any /* eslint-disable-line @typescript-eslint/no-explicit-any */, tokenOverride: string, keys: Set<string>) => {
  if (!fromConfig) return null;
  const config = JSON.parse(JSON.stringify(fromConfig));
  
  if (config.mcpServers) {
    const filteredServers: any = {}; // eslint-disable-line @typescript-eslint/no-explicit-any
    for (const key of keys) {
      if (config.mcpServers[key]) {
        filteredServers[key] = config.mcpServers[key];
      }
    }
    config.mcpServers = filteredServers;
  }
  
  if (tokenOverride) {
    if (!config.mcpProxy) {
      config.mcpProxy = {};
    }
    if (!config.mcpProxy.options) {
      config.mcpProxy.options = {};
    }
    config.mcpProxy.options.authTokens = [tokenOverride];
  }
  
  return config;
};

export const convertToFormat = (fromConfig: any /* eslint-disable-line @typescript-eslint/no-explicit-any */, options: ConvertOptions) => {
  const { tokenOverride, selectedServers, clientConfig, platform } = options;
  const { configFormat } = clientConfig;
  
  const platformOverride = configFormat.platformOverrides?.[platform];
  const rootKey = platformOverride?.rootKey ?? configFormat.rootKey;
  const httpField = platformOverride?.httpField ?? configFormat.httpField ?? "url";
  const requireType = platformOverride?.requireType ?? configFormat.requireType ?? false;
  const useStdioBridge = platformOverride?.useStdioBridge ?? configFormat.useStdioBridge ?? false;
  
  const config: any = { [rootKey]: {} }; // eslint-disable-line @typescript-eslint/no-explicit-any
  
  const options_ = fromConfig?.mcpProxy?.options ?? {};
  let baseURL = fromConfig?.mcpProxy?.baseURL || "";
  const suffix = fromConfig?.mcpProxy?.type === "streamable-http" ? "mcp" : "sse";

  if (!baseURL || baseURL.includes("localhost") || baseURL.includes("${")) {
    baseURL = typeof window !== "undefined" ? window.location.origin : "";
  }

  const mcpServers = fromConfig?.mcpServers ?? {};
  
  for (const key of selectedServers) {
    const serverConfig = mcpServers[key];
    if (!serverConfig) continue;

    const cleanBase = baseURL.replace(/\/+$/, "");
    const serverUrl = `${cleanBase}/${key}/${suffix}`.replace(/\/+/g, "/").replace(":/", "://");

    const token = tokenOverride || serverConfig?.options?.authTokens?.[0] || options_.authTokens?.[0];

    if (useStdioBridge) {
      const isWindows = platform === "windows";
      const cmd = isWindows ? "cmd" : "npx";
      const baseArgs = isWindows ? ["/c", "npx", "-y", "mcp-remote", serverUrl] : ["-y", "mcp-remote", serverUrl];
      
      if (token) {
        const formattedToken = formatAuthHeader(token);
        baseArgs.push("--header", `Authorization: ${formattedToken}`);
      }
      
      config[rootKey][key] = {
        command: cmd,
        args: baseArgs
      };
    } else {
      const server: any = { [httpField]: serverUrl }; // eslint-disable-line @typescript-eslint/no-explicit-any
      
      if (requireType) {
        server.type = "sse";
      }
      
      if (token) {
        const formattedToken = formatAuthHeader(token);
        server.headers = { Authorization: formattedToken };
      }
      
      config[rootKey][key] = server;
    }
  }
  
  return config;
};
