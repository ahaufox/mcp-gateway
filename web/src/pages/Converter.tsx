import React, { useState, useEffect, useMemo, useCallback } from "react";
import api from "../utils/api";
import { useTheme } from "../context/ThemeContext";
import {
  ArrowRightLeft,
  Copy,
  Check,
  AlertCircle,
  HelpCircle,
  X,
  ChevronDown,
  Download,
} from "lucide-react";
import { detectPlatform, getPlatformLabel, getConfigPathForPlatform, type Platform } from "../utils/platform";
import {
  type ClientDef,
  CLIENTS,
  convertToFormat
} from "../utils/mcpConverter";

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
  platform: "AI 平台",
};

// ─── 组件 ──────────────────────────────────────────────

export const Converter: React.FC = () => {
  const { theme } = useTheme();
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const [proxyConfig, setProxyConfig] = useState<any>(null);
  const [overrideToken, setOverrideToken] = useState("");
  const [availableServers, setAvailableServers] = useState<string[]>([]);
  const [selectedServers, setSelectedServers] = useState<Set<string>>(new Set());
  const [selectedClient, setSelectedClient] = useState<string>("claude");
  const [copiedType, setCopiedType] = useState<string | null>(null);
  const [error, setError] = useState("");
  const [showGuide, setShowGuide] = useState(false);
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedPlatform, setSelectedPlatform] = useState<Platform>("macos");
  const [showPlatformSelector, setShowPlatformSelector] = useState(false);

  useEffect(() => {
    const detected = detectPlatform();
    if (detected !== "unknown") {
      Promise.resolve().then(() => {
        setSelectedPlatform(detected);
      });
    }
  }, []);

  // 点击外部关闭平台选择器
  useEffect(() => {
    const handleClickOutside = () => {
      if (showPlatformSelector) {
        setShowPlatformSelector(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [showPlatformSelector]);

  // ── 加载配置 (在初始化时调用一次) ──

  const loadConfig = useCallback(async () => {
    try {
      const res = await api.get("/api/config");
      setProxyConfig(res.data);
      const servers = Object.keys(res.data?.mcpServers || {});
      setAvailableServers(servers);
      setSelectedServers(new Set(servers));
      setError("");
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    } catch (err: any) {
      if (err.response?.status === 401) {
        setError("认证过期，请重新登录");
      } else {
        setError("获取 mcp-proxy 配置文件失败，您也可以在下方手动粘贴 JSON 进行转换");
      }
    }
  }, []);

  useEffect(() => {
    Promise.resolve().then(() => {
      loadConfig();
    });
  }, [loadConfig]);

  // ── 转换输出 (自动响应所有输入变化) ──

  const formattedOutput = useMemo(() => {
    if (!proxyConfig) return "";
    
    // 根据选中的客户端配置格式和平台生成正确的输出
    const clientDef = CLIENTS.find(c => c.id === selectedClient)!;
    const formatted = convertToFormat(proxyConfig, {
      tokenOverride: overrideToken,
      selectedServers,
      clientConfig: clientDef,
      platform: selectedPlatform
    });
    
    return JSON.stringify(formatted, null, 2);
  }, [proxyConfig, overrideToken, selectedServers, selectedClient, selectedPlatform]);

  const selectedClientDef = useMemo(() => CLIENTS.find(c => c.id === selectedClient)!, [selectedClient]);
  
  const currentConfigPath = useMemo(() => {
    return getConfigPathForPlatform(selectedClientDef.configPaths, selectedPlatform);
  }, [selectedClientDef, selectedPlatform]);

  const currentOutput = useMemo(() => {
    return formattedOutput;
  }, [formattedOutput]);

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
    navigator.clipboard.writeText(text);
    setCopiedType(type);
    setTimeout(() => setCopiedType(null), 2000);
  };

  const toggleServer = (name: string) => {
    setSelectedServers(prev => {
      const next = new Set(prev);
      if (next.has(name)) {
        next.delete(name);
      } else {
        next.add(name);
      }
      return next;
    });
  };

  return (
    <div className="space-y-6">
      {/* ── 头部 ── */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className={`text-3xl font-bold tracking-tight flex items-center gap-2 text-gradient`}>
            <ArrowRightLeft className="w-7 h-7 text-violet-500" />
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
          <div className={`glass-strong rounded-2xl p-5 border ${theme === "dark" ? "border-violet-500/20" : "border-violet-200"}`}>
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
            className={`input-vp w-full rounded-xl px-3 py-2 text-sm transition-all focus:outline-none ${theme === "dark" ? "bg-white/5 text-white placeholder-gray-500" : "bg-white text-gray-800 placeholder-gray-400"}`}
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
                    className={`flex items-center gap-2 px-4 py-2.5 rounded-2xl transition-all duration-300 cursor-pointer group ${
                      isActive
                        ? `pill-active ${c.bg}`
                        : `pill ${theme === "dark" ? "bg-white/[0.02]" : "bg-gray-50"}`
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
          <div className="glass rounded-2xl p-5">
            <div className="flex items-center justify-between mb-4">
              <h3 className={`text-[11px] font-semibold uppercase tracking-[0.15em] ${theme === "dark" ? "text-violet-300/70" : "text-violet-500/70"}`}>导出服务器</h3>
              <div className="flex items-center gap-3">
                <button onClick={() => setSelectedServers(new Set(availableServers))} className={`text-[10px] uppercase tracking-[0.15em] font-semibold cursor-pointer ${theme === "dark" ? "text-violet-300 hover:text-white" : "text-violet-500 hover:text-violet-800"}`}>全选</button>
                <button onClick={() => setSelectedServers(new Set())} className={`text-[10px] uppercase tracking-[0.15em] font-semibold cursor-pointer ${theme === "dark" ? "text-violet-300/50 hover:text-violet-200" : "text-violet-500/50 hover:text-violet-700"}`}>清空</button>
              </div>
            </div>

            {availableServers.length === 0 ? (
              <p className={`text-[12px] italic ${theme === "dark" ? "text-violet-300/40" : "text-violet-500/50"}`}>暂无可用服务器配置</p>
            ) : (
              <div className="space-y-3">
                <div className="flex flex-wrap gap-1.5">
                  {availableServers.map(server => (
                    <button
                      key={server}
                      onClick={() => toggleServer(server)}
                      className={`pill h-7 px-3 border rounded-lg text-[11px] cursor-pointer ${
                        selectedServers.has(server)
                          ? "pill-active"
                          : ""
                      }`}
                    >
                      {server}
                    </button>
                  ))}
                </div>

                <div className="flex items-center justify-between pt-3 border-t border-white/5">
                  <span className={`text-[10px] mono ${theme === "dark" ? "text-violet-300/50" : "text-violet-500/60"}`}>
                    已选 {selectedServers.size} / 共 {availableServers.length}
                  </span>
                  <div className="flex-1 mx-3 h-1 bg-white/5 rounded-full overflow-hidden">
                    <div
                      className="h-full bg-gradient-to-r from-violet-500 to-indigo-500 rounded-full transition-all duration-300"
                      style={{ width: `${availableServers.length > 0 ? (selectedServers.size / availableServers.length) * 100 : 0}%`, boxShadow: "0 0 10px #8B5CF6" }}
                    />
                  </div>
                </div>
              </div>
            )}
          </div>

          {/* Token 重写 */}
          <div className="glass rounded-2xl p-5">
            <h3 className={`text-[11px] font-semibold uppercase tracking-[0.15em] mb-3 ${theme === "dark" ? "text-violet-300/70" : "text-violet-500/70"}`}>覆盖 Token（可选）</h3>
            <input
              type="text"
              placeholder="Bearer token..."
              value={overrideToken}
              onChange={(e) => setOverrideToken(e.target.value)}
              className={`input-vp w-full h-9 px-3 rounded-lg text-[12px] mono ${theme === "dark" ? "text-white placeholder-violet-300/30" : "text-gray-800 placeholder-gray-400"}`}
            />
            <p className={`text-[10px] mt-2 leading-relaxed ${theme === "dark" ? "text-violet-300/40" : "text-violet-500/50"}`}>填入后所有导出服务器将统一使用此 Token</p>
          </div>
        </div>

        {/* 右侧输出栏 — 单客户端输出 */}
        <div className="lg:col-span-7 space-y-3">
          {/* 元数据栏 */}
          <div className="glass rounded-2xl p-4 flex items-center justify-between">
            <div className="flex items-center gap-5 text-[11px]">
              <div className="flex items-center gap-1.5">
                <span className={theme === "dark" ? "text-violet-300/50" : "text-violet-500/60"}>输出格式</span>
                <span className={`mono font-medium ${theme === "dark" ? "text-white" : "text-gray-800"}`}>{selectedClientDef.configPaths && selectedClientDef.configPaths[selectedPlatform] ? selectedClientDef.configPaths[selectedPlatform].split("/").pop() : `${selectedClient}.json`}</span>
              </div>
              <div className="w-px h-4 bg-white/10"></div>
              <div className="flex items-center gap-1.5">
                <span className={theme === "dark" ? "text-violet-300/50" : "text-violet-500/60"}>平台</span>
                <div className="relative">
                  <button
                    onClick={(e) => { e.stopPropagation(); setShowPlatformSelector(!showPlatformSelector); }}
                    className={`mono font-medium flex items-center gap-1 cursor-pointer ${theme === "dark" ? "text-white hover:text-violet-300" : "text-gray-800 hover:text-violet-600"}`}
                  >
                    {getPlatformLabel(selectedPlatform)}
                    <ChevronDown className={`w-3 h-3 transition-transform duration-200 ${showPlatformSelector ? "rotate-180" : ""}`} />
                  </button>
                  {showPlatformSelector && (
                    <div className={`absolute top-full left-0 mt-1 rounded-lg shadow-lg z-50 border min-w-[90px] ${theme === "dark" ? "bg-gray-900 border-white/10" : "bg-white border-gray-200"}`}>
                      {(['windows', 'macos', 'linux'] as Platform[]).map((platform) => (
                        <button
                          key={platform}
                          onClick={() => { setSelectedPlatform(platform); setShowPlatformSelector(false); }}
                          className={`w-full text-left px-3 py-1.5 text-[11px] font-medium transition-colors cursor-pointer ${selectedPlatform === platform ? (theme === "dark" ? "bg-violet-500/20 text-violet-300" : "bg-violet-50 text-violet-700") : (theme === "dark" ? "text-gray-400 hover:bg-white/5" : "text-gray-600 hover:bg-gray-50")}`}
                        >
                          {getPlatformLabel(platform)}
                        </button>
                      ))}
                    </div>
                  )}
                </div>
              </div>
              <div className="w-px h-4 bg-white/10"></div>
              <div className="flex items-center gap-1.5">
                <span className={theme === "dark" ? "text-violet-300/50" : "text-violet-500/60"}>路径</span>
                <span className={`mono text-[10px] font-medium ${theme === "dark" ? "text-white" : "text-gray-800"}`}>{currentConfigPath}</span>
              </div>
            </div>
            <div className="flex items-center gap-2">
              <button className={`h-8 px-3 rounded-lg text-[12px] flex items-center gap-1.5 cursor-pointer transition-all ${theme === "dark" ? "glass-strong text-violet-200/80 hover:text-white" : "bg-gray-100 text-gray-600 hover:text-gray-800"}`}>
                <Download className="w-3.5 h-3.5" />
                <span>下载</span>
              </button>
              <button
                onClick={() => handleCopy(currentOutput, selectedClient)}
                className="btn-glow h-8 px-4 rounded-lg text-white text-[12px] font-semibold flex items-center gap-1.5 cursor-pointer"
              >
                {copiedType === selectedClient ? (
                  <>
                    <Check className="w-3.5 h-3.5" />
                    <span>已复制</span>
                  </>
                ) : (
                  <>
                    <Copy className="w-3.5 h-3.5" />
                    <span>复制</span>
                  </>
                )}
              </button>
            </div>
          </div>

          {/* 代码区 */}
          <div className="glass rounded-2xl overflow-hidden code-glow">
            <div className="flex items-center justify-between px-5 h-11 border-b border-white/5">
              <div className="flex items-center gap-2.5">
                <div className="flex gap-1.5">
                  <div className="w-2.5 h-2.5 rounded-full bg-rose-400/60"></div>
                  <div className="w-2.5 h-2.5 rounded-full bg-amber-400/60"></div>
                  <div className="w-2.5 h-2.5 rounded-full bg-emerald-400/60"></div>
                </div>
                <div className="w-px h-4 bg-white/10 mx-1"></div>
                <span className={`mono text-[12px] ${theme === "dark" ? "text-white" : "text-gray-800"}`}>
                  {selectedClientDef.configPaths && selectedClientDef.configPaths[selectedPlatform] ? selectedClientDef.configPaths[selectedPlatform].split("/").pop() : `${selectedClient}.json`}
                </span>
              </div>
              <div className={`flex items-center gap-3 text-[10px] mono ${theme === "dark" ? "text-violet-300/50" : "text-violet-500/60"}`}>
                <span>{formattedOutput ? `${formattedOutput.split("\n").length} 行` : "0 行"}</span>
                <span>·</span>
                <span>{formattedOutput ? `${(new Blob([formattedOutput]).size / 1024).toFixed(1)} KB` : "0 KB"}</span>
                <span>·</span>
                <span className="text-emerald-400">JSON valid</span>
              </div>
            </div>

            <textarea
              readOnly
              value={currentOutput}
              className={`w-full p-6 text-[12.5px] leading-[1.8] font-mono resize-none overflow-y-auto focus:outline-none border-0 ${theme === "dark" ? "bg-transparent text-gray-300" : "bg-transparent text-gray-800"}`}
              placeholder="勾选左侧服务器并选择上方客户端选项卡，自动生成输出..."
              style={{ height: "400px" }}
            />
          </div>

          {/* 操作指引 */}
          <div className={`glass rounded-2xl p-4 flex items-start gap-3 relative overflow-hidden ${theme === "dark" ? "" : "border-gray-200"}`}>
            <div className="absolute top-0 left-0 w-1 h-full bg-gradient-to-b from-violet-400 to-indigo-500"></div>
            <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-violet-500/20 to-indigo-500/20 border border-violet-400/30 flex items-center justify-center shrink-0">
              <HelpCircle className="w-4 h-4 text-violet-400" />
            </div>
            <div className={`text-[12px] leading-relaxed flex-1 ${theme === "dark" ? "text-violet-200/70" : "text-violet-700/70"}`}>
              <div className={`font-medium mb-1 ${theme === "dark" ? "text-white" : "text-gray-800"}`}>下一步操作</div>
              将上述 JSON 复制到 <span className={`mono px-1.5 py-0.5 rounded ${theme === "dark" ? "text-violet-300 bg-white/5" : "text-violet-600 bg-gray-100"}`}>{currentConfigPath}</span> 后重启 {selectedClientDef.name} 即可生效。
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};
