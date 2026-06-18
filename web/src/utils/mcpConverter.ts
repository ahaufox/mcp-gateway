import React from "react";
import {
  Monitor,
  Puzzle,
  Sparkles,
  Bot,
  Braces,
  Terminal,
  Globe,
  Zap
} from "lucide-react";
import { type Platform, type ClientPaths } from "./platform";

export interface ClientDef {
  id: string;
  name: string;
  icon: React.ComponentType<{ className?: string }>;
  desc: string;
  category: "ide" | "terminal" | "platform";
  color: "emerald" | "violet" | "indigo" | "amber" | "teal" | "blue" | "green" | "purple";
  fmtType: "generic";
  configPaths: ClientPaths;
  keywords: string[];
  configFormat: {
    rootKey: "mcpServers" | "servers";
    httpField?: "url" | "httpUrl";
    requireType?: boolean;
    useStdioBridge?: boolean;
    platformOverrides?: Partial<Record<Platform, {
      rootKey?: "mcpServers" | "servers";
      httpField?: "url" | "httpUrl";
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
    desc: "Anthropic 官方桌面客户端 — 仅支持 stdio 传输", 
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
    desc: "AI-first 代码编辑器 — 支持全局和项目级配置", 
    category: "ide", 
    color: "indigo", 
    fmtType: "generic", 
    configPaths: {
      windows: "%USERPROFILE%\\.cursor\\mcp.json",
      macos: "~/.cursor/mcp.json",
      linux: "~/.cursor/mcp.json"
    }, 
    keywords: ["cursor", "ai editor"],
    configFormat: { rootKey: "mcpServers", httpField: "url", requireType: false }
  },
  { 
    id: "trae", 
    name: "Trae IDE", 
    icon: Sparkles, 
    desc: "字节跳动 AI 开发环境 — 项目级 .trae/mcp.json", 
    category: "ide", 
    color: "violet", 
    fmtType: "generic", 
    configPaths: {
      windows: ".trae\\mcp.json",
      macos: ".trae/mcp.json",
      linux: ".trae/mcp.json"
    }, 
    keywords: ["trae", "字节跳动", "bytedance"],
    configFormat: { rootKey: "mcpServers", httpField: "url", requireType: true }
  },
  { 
    id: "vscode", 
    name: "VS Code", 
    icon: Braces, 
    desc: "Visual Studio Code — GitHub Copilot Agent 模式，根键为 servers", 
    category: "ide", 
    color: "blue", 
    fmtType: "generic", 
    configPaths: {
      windows: ".vscode\\mcp.json",
      macos: ".vscode/mcp.json",
      linux: ".vscode/mcp.json"
    }, 
    keywords: ["vscode", "visual studio", "microsoft"],
    configFormat: { rootKey: "servers", httpField: "url", requireType: false }
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
    id: "claude-code", 
    name: "Claude Code", 
    icon: Bot, 
    desc: "Anthropic Claude Code CLI — 支持全部传输方式", 
    category: "terminal", 
    color: "green", 
    fmtType: "generic", 
    configPaths: {
      windows: "%USERPROFILE%\\.claude.json",
      macos: "~/.claude.json",
      linux: "~/.claude.json"
    }, 
    keywords: ["claude code", "cli", "anthropic"],
    configFormat: { rootKey: "mcpServers", httpField: "url", requireType: false }
  },
  { 
    id: "gemini", 
    name: "Gemini CLI", 
    icon: Zap, 
    desc: "Google Gemini CLI — 终端 AI 助手", 
    category: "terminal", 
    color: "amber", 
    fmtType: "generic", 
    configPaths: {
      windows: "%USERPROFILE%\\.gemini\\settings.json",
      macos: "~/.gemini/settings.json",
      linux: "~/.gemini/settings.json"
    }, 
    keywords: ["gemini", "google", "cli"],
    configFormat: { rootKey: "mcpServers", httpField: "httpUrl", requireType: false }
  },

  // AI 平台
  { 
    id: "antigravity", 
    name: "Antigravity", 
    icon: Globe, 
    desc: "Google Antigravity AI IDE — 基于 Gemini 生态", 
    category: "platform", 
    color: "purple", 
    fmtType: "generic", 
    configPaths: {
      windows: "%USERPROFILE%\\.gemini\\antigravity\\mcp_config.json",
      macos: "~/.gemini/antigravity/mcp_config.json",
      linux: "~/.gemini/antigravity/mcp_config.json"
    }, 
    keywords: ["antigravity", "gemini", "google"],
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

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export const convertToProxy = (fromConfig: any, tokenOverride: string, keys: Set<string>) => {
  if (!fromConfig) return null;
  const config = JSON.parse(JSON.stringify(fromConfig));
  
  if (config.mcpServers) {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const filteredServers: any = {};
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

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export const convertToFormat = (fromConfig: any, options: ConvertOptions) => {
  const { tokenOverride, selectedServers, clientConfig, platform } = options;
  const { configFormat } = clientConfig;
  
  const platformOverride = configFormat.platformOverrides?.[platform];
  const rootKey = platformOverride?.rootKey ?? configFormat.rootKey;
  const httpField = platformOverride?.httpField ?? configFormat.httpField ?? "url";
  const requireType = platformOverride?.requireType ?? configFormat.requireType ?? false;
  const useStdioBridge = platformOverride?.useStdioBridge ?? configFormat.useStdioBridge ?? false;
  
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const config: any = { [rootKey]: {} };
  
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
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const server: any = { [httpField]: serverUrl };
      
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