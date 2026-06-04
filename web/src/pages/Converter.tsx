import React, { useState, useEffect, useMemo, useCallback } from "react";
import api from "../utils/api";
import { useTheme } from "../context/ThemeContext";
import {
  ArrowRightLeft,
  Copy,
  Check,
  AlertCircle,
  Info,
  Server,
  Settings,
  Monitor,
  Sparkles,
  Box,
  Globe,
  Terminal,
  Cpu,
  Cloud,
  Bot,
  Zap,
  Layers,
  Play,
  Puzzle,
  SquareTerminal,
  Braces,
  HelpCircle,
  X,
} from "lucide-react";

// ─── 客户端定义 ────────────────────────────────────────

interface ClientDef {
  id: string;
  name: string;
  icon: React.ComponentType<{ className?: string }>;
  desc: string;
  category: "ide" | "terminal" | "assistant" | "platform" | "native";
  color: "emerald" | "violet" | "indigo" | "amber" | "rose" | "cyan" | "orange" | "teal" | "blue" | "green" | "purple" | "slate";
  fmtType: "claude" | "trae" | "proxy";
  configPath: string;
  keywords: string[]; // 模糊搜索关键词
}

const CLIENTS: ClientDef[] = [
  // IDE / 编辑器
  { id: "claude", name: "Claude Desktop", icon: Monitor, desc: "Anthropic 官方桌面客户端", category: "ide", color: "emerald", fmtType: "claude", configPath: "~/.config/Claude/claude_desktop_config.json", keywords: ["claude", "anthropic", "desktop"] },
  { id: "cursor", name: "Cursor", icon: Puzzle, desc: "AI-first 代码编辑器", category: "ide", color: "indigo", fmtType: "claude", configPath: "~/.cursor/mcp.json", keywords: ["cursor", "ai editor"] },
  { id: "windsurf", name: "Windsurf", icon: Cloud, desc: "Codeium 流式 AI IDE", category: "ide", color: "cyan", fmtType: "claude", configPath: "~/.codeium/windsurf/mcp_config.json", keywords: ["windsurf", "codeium"] },
  { id: "trae", name: "Trae IDE", icon: Sparkles, desc: "字节跳动 AI 开发环境", category: "ide", color: "violet", fmtType: "trae", configPath: "~/.trae/mcp_config.json", keywords: ["trae", "字节跳动", "bytedance"] },
  { id: "cline", name: "Cline", icon: Bot, desc: "VS Code 全能 AI 助手", category: "ide", color: "rose", fmtType: "claude", configPath: "~/.cline/mcp_settings.json", keywords: ["cline", "vscode extension"] },
  { id: "roocode", name: "Roo Code", icon: Cpu, desc: "VS Code 多模型 AI 编程", category: "ide", color: "orange", fmtType: "claude", configPath: "~/.roo-code/mcp_settings.json", keywords: ["roo", "roo code", "vscode"] },
  { id: "vscode", name: "VS Code", icon: Braces, desc: "微软编辑器 MCP 扩展", category: "ide", color: "blue", fmtType: "claude", configPath: "~/.vscode/mcp.json", keywords: ["vscode", "visual studio", "microsoft"] },
  { id: "zed", name: "Zed Editor", icon: Layers, desc: "高性能协作编辑器", category: "ide", color: "slate", fmtType: "claude", configPath: "~/.zed/mcp.json", keywords: ["zed", "editor"] },
  { id: "continue", name: "Continue", icon: Play, desc: "开源 AI 代码助手", category: "ide", color: "purple", fmtType: "claude", configPath: "~/.continue/config.json", keywords: ["continue", "continue.dev"] },

  // 终端 / CLI
  { id: "codex", name: "Codex CLI", icon: Terminal, desc: "OpenAI 终端编程助手", category: "terminal", color: "teal", fmtType: "claude", configPath: "~/.codex/mcp.json", keywords: ["codex", "openai", "cli"] },
  { id: "warp", name: "Warp Terminal", icon: SquareTerminal, desc: "Rust 重写智能终端", category: "terminal", color: "amber", fmtType: "claude", configPath: "~/.warp/mcp.json", keywords: ["warp", "terminal", "rust"] },

  // AI 平台 / 助手
  { id: "antigravity", name: "Antigravity", icon: Globe, desc: "Gemini 生态 AI 扩展", category: "platform", color: "indigo", fmtType: "trae", configPath: "~/.gemini/antigravity/mcp_config.json", keywords: ["antigravity", "gemini", "google"] },
  { id: "openinterpreter", name: "Open Interpreter", icon: Zap, desc: "自然语言操控计算机", category: "assistant", color: "green", fmtType: "claude", configPath: "~/.open-interpreter/mcp.json", keywords: ["open interpreter", "interpreter"] },

  // mcp-proxy 原生
  { id: "proxy", name: "mcp-proxy", icon: Box, desc: "本网关代理原生配置", category: "native", color: "amber", fmtType: "proxy", configPath: "config.json", keywords: ["mcp-proxy", "proxy", "gateway"] },
];

const COLOR_MAP: Record<string, { bg: string; border: string; text: string; ring: string; bgDark: string; borderDark: string; textDark: string }> = {
  emerald:  { bg: "bg-emerald-100 dark:bg-emerald-500/10", border: "border-emerald-300 dark:border-emerald-500/20", text: "text-emerald-700 dark:text-emerald-400", ring: "ring-emerald-400/30", bgDark: "bg-emerald-500/10", borderDark: "border-emerald-500/20", textDark: "text-emerald-400" },
  violet:   { bg: "bg-violet-100 dark:bg-violet-500/10", border: "border-violet-300 dark:border-violet-500/20", text: "text-violet-700 dark:text-violet-400", ring: "ring-violet-400/30", bgDark: "bg-violet-500/10", borderDark: "border-violet-500/20", textDark: "text-violet-400" },
  indigo:   { bg: "bg-indigo-100 dark:bg-indigo-500/10", border: "border-indigo-300 dark:border-indigo-500/20", text: "text-indigo-700 dark:text-indigo-400", ring: "ring-indigo-400/30", bgDark: "bg-indigo-500/10", borderDark: "border-indigo-500/20", textDark: "text-indigo-400" },
  amber:    { bg: "bg-amber-100 dark:bg-amber-500/10", border: "border-amber-300 dark:border-amber-500/20", text: "text-amber-700 dark:text-amber-400", ring: "ring-amber-400/30", bgDark: "bg-amber-500/10", borderDark: "border-amber-500/20", textDark: "text-amber-400" },
  rose:     { bg: "bg-rose-100 dark:bg-rose-500/10", border: "border-rose-300 dark:border-rose-500/20", text: "text-rose-700 dark:text-rose-400", ring: "ring-rose-400/30", bgDark: "bg-rose-500/10", borderDark: "border-rose-500/20", textDark: "text-rose-400" },
  cyan:     { bg: "bg-cyan-100 dark:bg-cyan-500/10", border: "border-cyan-300 dark:border-cyan-500/20", text: "text-cyan-700 dark:text-cyan-400", ring: "ring-cyan-400/30", bgDark: "bg-cyan-500/10", borderDark: "border-cyan-500/20", textDark: "text-cyan-400" },
  orange:   { bg: "bg-orange-100 dark:bg-orange-500/10", border: "border-orange-300 dark:border-orange-500/20", text: "text-orange-700 dark:text-orange-400", ring: "ring-orange-400/30", bgDark: "bg-orange-500/10", borderDark: "border-orange-500/20", textDark: "text-orange-400" },
  teal:     { bg: "bg-teal-100 dark:bg-teal-500/10", border: "border-teal-300 dark:border-teal-500/20", text: "text-teal-700 dark:text-teal-400", ring: "ring-teal-400/30", bgDark: "bg-teal-500/10", borderDark: "border-teal-500/20", textDark: "text-teal-400" },
  blue:     { bg: "bg-blue-100 dark:bg-blue-500/10", border: "border-blue-300 dark:border-blue-500/20", text: "text-blue-700 dark:text-blue-400", ring: "ring-blue-400/30", bgDark: "bg-blue-500/10", borderDark: "border-blue-500/20", textDark: "text-blue-400" },
  green:    { bg: "bg-green-100 dark:bg-green-500/10", border: "border-green-300 dark:border-green-500/20", text: "text-green-700 dark:text-green-400", ring: "ring-green-400/30", bgDark: "bg-green-500/10", borderDark: "border-green-500/20", textDark: "text-green-400" },
  purple:   { bg: "bg-purple-100 dark:bg-purple-500/10", border: "border-purple-300 dark:border-purple-500/20", text: "text-purple-700 dark:text-purple-400", ring: "ring-purple-400/30", bgDark: "bg-purple-500/10", borderDark: "border-purple-500/20", textDark: "text-purple-400" },
  slate:    { bg: "bg-slate-100 dark:bg-slate-500/10", border: "border-slate-300 dark:border-slate-500/20", text: "text-slate-700 dark:text-slate-400", ring: "ring-slate-400/30", bgDark: "bg-slate-500/10", borderDark: "border-slate-500/20", textDark: "text-slate-400" },
};

const CATEGORY_LABELS: Record<string, string> = {
  ide: "IDE / 编辑器",
  terminal: "终端 / CLI",
  assistant: "AI 助手",
  platform: "AI 平台",
  native: "本代理",
};

// ─── 组件 ──────────────────────────────────────────────

export const Converter: React.FC = () => {
  const { theme } = useTheme();
  const [proxyConfig, setProxyConfig] = useState<any>(null);
  const [overrideToken, setOverrideToken] = useState("");
  const [availableServers, setAvailableServers] = useState<string[]>([]);
  const [selectedServers, setSelectedServers] = useState<Set<string>>(new Set());
  const [selectedClient, setSelectedClient] = useState<string>("claude");
  const [copiedType, setCopiedType] = useState<string | null>(null);
  const [error, setError] = useState("");
  const [showGuide, setShowGuide] = useState(false);
  const [searchQuery, setSearchQuery] = useState("");

  // ── 加载配置 (在初始化时调用一次) ──

  const loadConfig = useCallback(async () => {
    try {
      const res = await api.get("/api/config");
      setProxyConfig(res.data);
      const servers = Object.keys(res.data?.mcpServers || {});
      setAvailableServers(servers);
      setSelectedServers(new Set(servers));
      setError("");
    } catch (err: any) {
      if (err.response?.status === 401) {
        setError("认证过期，请重新登录");
      } else {
        setError("获取 mcp-proxy 配置文件失败，您也可以在下方手动粘贴 JSON 进行转换");
      }
    }
  }, []);

  useEffect(() => { loadConfig(); }, [loadConfig]);

  // ── 转换输出 (自动响应所有输入变化) ──

  const { claudeOutput, traeOutput } = useMemo(() => {
    if (!proxyConfig) return { claudeOutput: "", traeOutput: "" };
    const claude = convertToClaude(proxyConfig, overrideToken, selectedServers);
    const trae = convertToTrae(proxyConfig, overrideToken, selectedServers);
    return {
      claudeOutput: JSON.stringify(claude, null, 2),
      traeOutput: JSON.stringify(trae, null, 2),
    };
  }, [proxyConfig, overrideToken, selectedServers]);

  const selectedClientDef = useMemo(() => CLIENTS.find(c => c.id === selectedClient)!, [selectedClient]);

  const currentOutput = useMemo(() => {
    if (selectedClientDef.fmtType === "claude") return claudeOutput;
    if (selectedClientDef.fmtType === "trae") return traeOutput;
    if (selectedClientDef.fmtType === "proxy") return JSON.stringify(proxyConfig, null, 2);
    return "";
  }, [selectedClientDef, claudeOutput, traeOutput, proxyConfig]);

  // ── 模糊搜索客户端 ──

  const filteredClients = useMemo(() => {
    const q = searchQuery.toLowerCase().trim();
    if (!q) return CLIENTS;
    return CLIENTS.filter(c =>
      c.name.toLowerCase().includes(q) ||
      c.desc.toLowerCase().includes(q) ||
      (c.keywords && c.keywords.some(k => k.includes(q)))
    );
  }, [searchQuery]);

  // ── 操作函数 ──

  const handleCopy = (text: string, type: string) => {
    if (!text) return;

    const fallbackCopy = (str: string) => {
      const textarea = document.createElement("textarea");
      textarea.value = str;
      textarea.style.position = "fixed";
      textarea.style.left = "-9999px";
      document.body.appendChild(textarea);
      textarea.select();
      try {
        document.execCommand("copy");
      } catch (e) {
        console.error("copy failed", e);
      }
      document.body.removeChild(textarea);
    };

    if (navigator.clipboard && navigator.clipboard.writeText) {
      navigator.clipboard.writeText(text);
    } else {
      fallbackCopy(text);
    }
    setCopiedType(type);
    setTimeout(() => setCopiedType(null), 2000);
  };

  const toggleServer = (name: string) => {
    setSelectedServers(prev => {
      const next = new Set(prev);
      next.has(name) ? next.delete(name) : next.add(name);
      return next;
    });
  };

  return (
    <div className="space-y-6">
      {/* ── 头部 ── */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className={`text-3xl font-bold tracking-tight flex items-center gap-2 ${theme === "dark" ? "text-white" : "text-gray-900"}`}>
            <ArrowRightLeft className="w-7 h-7 text-violet-600 dark:text-violet-500" />
            <span>配置格式转换器</span>
          </h1>
          <p className={`text-sm mt-1.5 ${theme === "dark" ? "text-gray-400" : "text-gray-600"}`}>
            选择目标客户端，一键生成兼容的 MCP 配置文件 — 支持 {CLIENTS.length} 种主流 IDE、终端与 AI 平台
          </p>
        </div>
        {/* 帮助按钮 */}
        <button
          onClick={() => setShowGuide(!showGuide)}
          className="relative flex items-center justify-center w-9 h-9 rounded-full bg-gray-100 dark:bg-white/10 border border-gray-200 dark:border-white/10 text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-white hover:bg-gray-200 dark:hover:bg-white/15 transition-all cursor-pointer"
          title="使用指南"
        >
          <HelpCircle className="w-5 h-5" />
        </button>
      </div>

      {/* ── 使用指南 弹出框 ── */}
      {showGuide && (
        <div className="relative">
          <div className={`glass-card rounded-2xl p-5 border ${theme === "dark" ? "border-violet-500/20" : "border-violet-200"} ${theme === "dark" ? "bg-violet-500/5" : "bg-violet-50/80"}`}>
            <div className="flex items-start justify-between mb-3">
              <h3 className={`text-base font-bold ${theme === "dark" ? "text-white" : "text-gray-900"}`}>快速使用指南</h3>
              <button onClick={() => setShowGuide(false)} className={`cursor-pointer ${theme === "dark" ? "text-gray-400 hover:text-white" : "text-gray-400 hover:text-gray-600"}`}>
                <X className="w-4 h-4" />
              </button>
            </div>
            <div className={`grid grid-cols-1 md:grid-cols-2 gap-3 text-sm leading-relaxed ${theme === "dark" ? "text-gray-400" : "text-gray-600"}`}>
              <div className="flex items-start gap-2">
                <span className="text-violet-600 dark:text-violet-400 font-bold mt-0.5 shrink-0">①</span>
                <span><strong className={theme === "dark" ? "text-gray-200" : "text-gray-800"}>选择目标客户端</strong>：点击下方选项卡，切换要生成的配置格式，支持模糊搜索</span>
              </div>
              <div className="flex items-start gap-2">
                <span className="text-violet-600 dark:text-violet-400 font-bold mt-0.5 shrink-0">②</span>
                <span><strong className={theme === "dark" ? "text-gray-200" : "text-gray-800"}>选择服务器</strong>：勾选要导出的 MCP 服务器，配置自动实时生成</span>
              </div>
              <div className="flex items-start gap-2">
                <span className="text-violet-600 dark:text-violet-400 font-bold mt-0.5 shrink-0">③</span>
                <span><strong className={theme === "dark" ? "text-gray-200" : "text-gray-800"}>Token 管理</strong>：可统一覆盖所有服务器的 Authorization Bearer Token</span>
              </div>
              <div className="flex items-start gap-2">
                <span className="text-violet-600 dark:text-violet-400 font-bold mt-0.5 shrink-0">④</span>
                <span><strong className={theme === "dark" ? "text-gray-200" : "text-gray-800"}>部署路径</strong>：将生成的 JSON 复制到对应客户端配置目录即可生效</span>
              </div>
            </div>
          </div>
        </div>
      )}

      {error && (
        <div className="flex items-center gap-3 bg-rose-50 dark:bg-rose-500/10 border border-rose-200 dark:border-rose-500/20 text-rose-700 dark:text-rose-300 p-4 rounded-2xl">
          <AlertCircle className="w-5 h-5 text-rose-500 dark:text-rose-400 shrink-0" />
          <span className="text-sm">{error}</span>
        </div>
      )}

      {/* ── 目标客户端选项卡 (换行) ── */}
      <section>
        <div className="flex items-center gap-2 mb-3">
          <span className={`text-xs font-bold uppercase tracking-widest ${theme === "dark" ? "text-gray-500" : "text-gray-400"}`}>选择目标客户端</span>
          <span className="text-xs text-violet-600 dark:text-violet-400">— 点击选项卡直接切换输出格式</span>
        </div>

        {/* 搜索框 */}
        <div className="mb-3">
          <div className="relative max-w-xs">
            <input
            type="text"
            placeholder="搜索客户端..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className={`w-full border rounded-xl px-3 py-2 text-sm transition-all focus:outline-none focus:border-violet-400 ${theme === "dark" ? "bg-white/5 border-white/10 text-white placeholder-gray-500 focus:border-violet-500/50" : "bg-white border-gray-200 text-gray-800 placeholder-gray-400"}`}
          />
            {searchQuery && (
              <button
                onClick={() => setSearchQuery("")}
                className={`absolute right-2 top-1/2 -translate-y-1/2 cursor-pointer ${theme === "dark" ? "text-gray-500 hover:text-white" : "text-gray-400 hover:text-gray-600"}`}
              >
                <X className="w-4 h-4" />
              </button>
            )}
          </div>
        </div>

        {/* 客户端选项卡 — 换行显示 */}
        <div className="flex flex-wrap gap-2">
          {Object.entries(
            filteredClients.reduce((acc, c) => {
              if (!acc[c.category]) acc[c.category] = [];
              acc[c.category].push(c);
              return acc;
            }, {} as Record<string, ClientDef[]>)
          ).map(([category, clients]) => (
            <React.Fragment key={category}>
              <div className="flex items-center w-full mt-1 first:mt-0">
                <span className={`text-[11px] font-bold uppercase tracking-widest ${theme === "dark" ? "text-gray-600" : "text-gray-300"}`}>
                  {CATEGORY_LABELS[category]}
                </span>
              </div>
              {clients.map(client => {
                const c = COLOR_MAP[client.color];
                const isActive = selectedClient === client.id;
                return (
                  <button
                    key={client.id}
                    onClick={() => setSelectedClient(client.id)}
                    className={`flex items-center gap-2 px-4 py-2.5 rounded-2xl border transition-all duration-300 cursor-pointer group ${
                      isActive
                        ? `${c.bg} ${c.border} ${c.text} ring-1 ${c.ring} shadow-sm`
                        : theme === "dark"
                          ? "bg-white/[0.02] border-white/5 text-gray-500 hover:text-gray-300 hover:border-white/10 hover:bg-white/[0.04]"
                          : "bg-gray-50 border-gray-200 text-gray-600 hover:text-gray-800 hover:border-gray-300 hover:bg-gray-100"
                    }`}
                    title={`${client.name}: ${client.desc}`}
                  >
                    <client.icon className={`w-4 h-4 ${isActive ? c.text : theme === "dark" ? "text-gray-500 group-hover:text-gray-400" : "text-gray-400 group-hover:text-gray-600"} transition-colors`} />
                    <span className="text-sm font-semibold whitespace-nowrap">{client.name}</span>
                  </button>
                );
              })}
            </React.Fragment>
          ))}
        </div>
      </section>

      {/* ── 主体两栏 ── */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-8">
        {/* 左侧控制栏 */}
        <div className="lg:col-span-5 space-y-6">
          {/* 服务器选择 */}
          <div className="glass-card rounded-3xl p-6">
            <h3 className={`text-base font-bold mb-4 flex items-center gap-2 ${theme === "dark" ? "text-white" : "text-gray-900"}`}>
            <Server className="w-4 h-4 text-violet-600 dark:text-violet-400" />
            <span>选择要导出的服务器</span>
          </h3>

          {availableServers.length === 0 ? (
            <p className={`text-sm italic ${theme === "dark" ? "text-gray-500" : "text-gray-500"}`}>暂无可用服务器配置</p>
          ) : (
              <div className="space-y-4">
                <div className="flex flex-wrap gap-2 max-h-52 overflow-y-auto pr-1">
                  {availableServers.map(server => (
                    <button
                    key={server}
                    onClick={() => toggleServer(server)}
                    className={`px-3 py-1.5 rounded-xl text-sm font-medium border transition-all duration-300 cursor-pointer ${
                      selectedServers.has(server)
                        ? theme === "dark"
                          ? "bg-violet-500/10 border-violet-500/30 text-violet-300 hover:bg-violet-500/20"
                          : "bg-violet-50 border-violet-300 text-violet-700 hover:bg-violet-100"
                        : theme === "dark"
                          ? "bg-white/5 border-white/5 text-gray-500 hover:text-gray-400 hover:border-white/10"
                          : "bg-gray-50 border-gray-200 text-gray-600 hover:text-gray-800 hover:border-gray-300"
                    }`}
                  >
                    {server}
                  </button>
                  ))}
                </div>

                <div className={`flex gap-3 border-t pt-3 ${theme === "dark" ? "border-white/5" : "border-gray-200"}`}>
                  <button onClick={() => setSelectedServers(new Set(availableServers))} className={`text-xs font-bold cursor-pointer uppercase tracking-wider ${theme === "dark" ? "text-violet-400 hover:text-violet-300" : "text-violet-600 hover:text-violet-500"}`}>全选</button>
                  <button onClick={() => setSelectedServers(new Set())} className={`text-xs font-bold cursor-pointer uppercase tracking-wider ${theme === "dark" ? "text-gray-500 hover:text-gray-400" : "text-gray-400 hover:text-gray-600"}`}>清空</button>
                </div>
              </div>
            )}
          </div>

          {/* Token 重写 */}
          <div className="glass-card rounded-3xl p-6">
            <h3 className={`text-base font-bold mb-4 flex items-center gap-2 ${theme === "dark" ? "text-white" : "text-gray-900"}`}>
            <Settings className="w-4 h-4 text-violet-600 dark:text-violet-400" />
            <span>覆盖 Authorization Token (可选)</span>
          </h3>
          <input
            type="text"
            placeholder="若填入，导出的所有服务器都将强行使用该 Token"
            value={overrideToken}
            onChange={(e) => setOverrideToken(e.target.value)}
            className={`w-full border rounded-2xl px-4 py-3 text-sm transition-all focus:outline-none focus:border-violet-400 ${theme === "dark" ? "bg-white/5 border-white/10 text-white placeholder-gray-500 focus:border-violet-500/50" : "bg-white border-gray-200 text-gray-800 placeholder-gray-400"}`}
          />
            <div className="flex gap-2 items-start mt-3 text-xs text-gray-500 dark:text-gray-500 leading-relaxed">
              <Info className="w-3.5 h-3.5 shrink-0 text-violet-500 mt-0.5" />
              <span>如果不填写，转换器将默认读取 <code className="text-gray-700 dark:text-gray-300 bg-gray-100 dark:bg-white/5 px-1 rounded">mcpServers.&lt;name&gt;.options.authTokens[0]</code>，或使用全局 <code className="text-gray-700 dark:text-gray-300 bg-gray-100 dark:bg-white/5 px-1 rounded">mcpProxy.options.authTokens[0]</code> 作为备用。</span>
            </div>
          </div>
        </div>

        {/* 右侧输出栏 — 单客户端输出 */}
        <div className="lg:col-span-7">
          <div className="glass-card rounded-3xl p-6 flex flex-col h-[calc(280px+280px+1.5rem)]">
            <div className="flex justify-between items-center mb-4">
              <div className="flex items-center gap-3">
                <div className={`p-2 rounded-xl ${COLOR_MAP[selectedClientDef.color].bg} border ${COLOR_MAP[selectedClientDef.color].border}`}>
                  <selectedClientDef.icon className={`w-5 h-5 ${COLOR_MAP[selectedClientDef.color].text}`} />
                </div>
                <div>
                  <h3 className={`text-base font-bold ${theme === "dark" ? "text-white" : "text-gray-900"}`}>
                    {selectedClientDef.name} 配置
                    <span className={`ml-2 text-xs font-normal ${theme === "dark" ? "text-gray-500" : "text-gray-400"}`}>
                      {selectedClientDef.fmtType === "proxy" ? "原生格式" : selectedClientDef.fmtType === "trae" ? "Trae 格式" : "Claude 兼容格式"}
                    </span>
                  </h3>
                  <p className={`text-xs mt-0.5 ${theme === "dark" ? "text-gray-500" : "text-gray-500"}`}>
                    {selectedClientDef.fmtType === "proxy"
                      ? "直接输出 mcp-proxy 原始配置"
                      : `目标存放路径：${selectedClientDef.configPath}`}
                  </p>
                </div>
              </div>
              <button
                onClick={() => handleCopy(currentOutput, selectedClient)}
                className={`flex items-center gap-1.5 px-3 py-1.5 border rounded-xl text-sm font-semibold cursor-pointer transition-all duration-300 ${theme === "dark" ? "bg-white/5 border-white/10 text-gray-300 hover:text-white" : "bg-gray-100 border-gray-200 text-gray-600 hover:text-gray-800"}`}
              >
                {copiedType === selectedClient ? (
                  <>
                    <Check className="w-3.5 h-3.5 text-emerald-500" />
                    <span>已复制</span>
                  </>
                ) : (
                  <>
                    <Copy className="w-3.5 h-3.5" />
                    <span>复制配置</span>
                  </>
                )}
              </button>
            </div>
            <textarea
              readOnly
              value={currentOutput}
              className={`flex-1 border rounded-2xl p-4 text-sm font-mono resize-none overflow-y-auto leading-relaxed focus:outline-none ${theme === "dark" ? "bg-black/40 border-white/5 text-gray-300 placeholder-gray-500" : "bg-gray-50 border-gray-200 text-gray-800 placeholder-gray-400"}`}
              placeholder="勾选左侧服务器并选择上方客户端选项卡，自动生成输出..."
            />
          </div>
        </div>
      </div>
    </div>
  );
};

// ─── 转换逻辑 ──────────────────────────────────────────

const convertToClaude = (fromConfig: any, tokenOverride: string, keys: Set<string>) => {
  const claudeConfig: any = { mcpServers: {} };
  const options = fromConfig?.mcpProxy?.options ?? {};
  let baseURL = fromConfig?.mcpProxy?.baseURL || "";
  const suffix = fromConfig?.mcpProxy?.type === "streamable-http" ? "mcp" : "sse";

  if (!baseURL || baseURL.includes("localhost") || baseURL.includes("${")) {
    baseURL = typeof window !== "undefined" ? window.location.origin : "";
  }

  const mcpServers = fromConfig?.mcpServers ?? {};
  for (const key of keys) {
    const serverConfig = mcpServers[key];
    if (!serverConfig) continue;

    const cleanBase = baseURL.replace(/\/+$/, "");
    const serverUrl = `${cleanBase}/${key}/${suffix}`.replace(/\/+/g, "/").replace(":/", "://");

    const server: any = { url: serverUrl };
    const token = tokenOverride || serverConfig?.options?.authTokens?.[0] || options?.authTokens?.[0];
    if (token) {
      server.headers = { Authorization: token };
    }
    claudeConfig.mcpServers[key] = server;
  }
  return claudeConfig;
};

const convertToTrae = (fromConfig: any, tokenOverride: string, keys: Set<string>) => {
  const agConfig: any = { mcpServers: {} };
  const options = fromConfig?.mcpProxy?.options ?? {};
  let baseURL = fromConfig?.mcpProxy?.baseURL || "";
  const suffix = fromConfig?.mcpProxy?.type === "streamable-http" ? "mcp" : "sse";

  if (!baseURL || baseURL.includes("localhost") || baseURL.includes("${")) {
    baseURL = typeof window !== "undefined" ? window.location.origin : "";
  }

  const mcpServers = fromConfig?.mcpServers ?? {};
  for (const key of keys) {
    const serverConfig = mcpServers[key];
    if (!serverConfig) continue;

    const cleanBase = baseURL.replace(/\/+$/, "");
    const serverUrl = `${cleanBase}/${key}/${suffix}`.replace(/\/+/g, "/").replace(":/", "://");

    const server: any = { url: serverUrl };
    const token = tokenOverride || serverConfig?.options?.authTokens?.[0] || options?.authTokens?.[0];
    if (token) {
      server.headers = {
        Authorization: token,
        "Content-Type": "application/json"
      };
    }
    agConfig.mcpServers[key] = server;
  }
  return agConfig;
};