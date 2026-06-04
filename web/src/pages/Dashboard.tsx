import React, { useState, useEffect } from "react";
import api from "../utils/api";
import { useTheme } from "../context/ThemeContext";
import {
  Server,
  CheckCircle2,
  XCircle,
  AlertCircle,
  Search,
  RotateCw,
  Cpu,
  FileText,
  Layers,
  ChevronDown,
  ChevronUp,
  Zap,
  ShieldCheck,
  ArrowRightLeft,
  Boxes,
  Code2,
  Rocket
} from "lucide-react";

// 与后端 ServerInfo 结构体对齐 (使用 PascalCase / camelCase 两种都兼容)
interface MCPTool {
  name?: string;
  Name?: string;
  description?: string;
  Description?: string;
  inputSchema?: any;
  InputSchema?: any;
}

interface MCPPrompt {
  name?: string;
  Name?: string;
  description?: string;
  Description?: string;
  arguments?: any[];
  Arguments?: any[];
}

interface MCPResource {
  uri?: string;
  URI?: string;
  name?: string;
  Name?: string;
  description?: string;
  Description?: string;
}

interface ServerInfo {
  Name: string;
  Route: string;
  Status: string;        // Connected / Failed / Unhealthy / Unknown
  Error: string;
  Description: string;
  Tools?: MCPTool[];
  Prompts?: MCPPrompt[];
  Resources?: MCPResource[];
}

// 工具函数：从两种命名风格字段中取一个
const pick = <T,>(a: T | undefined, b: T | undefined): T | undefined => a ?? b;

const toolName = (t: MCPTool) => pick(t.name, t.Name) ?? "";
const toolDesc = (t: MCPTool) => pick(t.description, t.Description) ?? "";
const toolSchema = (t: MCPTool) => pick(t.inputSchema, t.InputSchema);

const promptName = (p: MCPPrompt) => pick(p.name, p.Name) ?? "";
const promptDesc = (p: MCPPrompt) => pick(p.description, p.Description) ?? "";
const promptArgs = (p: MCPPrompt) => pick(p.arguments, p.Arguments) ?? [];

const resourceName = (r: MCPResource) => pick(r.name, r.Name) ?? "";
const resourceDesc = (r: MCPResource) => pick(r.description, r.Description) ?? "";
const resourceUri = (r: MCPResource) => pick(r.uri, r.URI) ?? "";

// 统计状态类型
type StatusKey = "connected" | "failed" | "unhealthy" | "unknown";

const normalizeStatus = (status: string): StatusKey => {
  const s = (status || "").toLowerCase();
  if (s === "connected") return "connected";
  if (s === "failed") return "failed";
  if (s === "unhealthy") return "unhealthy";
  return "unknown";
};

const statusMeta = (status: StatusKey) => {
  switch (status) {
    case "connected":
      return {
        label: "已连接",
        dotColor: "bg-emerald-500",
        ringColor: "bg-emerald-400",
        textColor: "text-emerald-400",
        bgColor: "bg-emerald-500/10",
        borderColor: "border-emerald-500/20"
      };
    case "failed":
      return {
        label: "连接失败",
        dotColor: "bg-rose-500",
        ringColor: "bg-rose-500",
        textColor: "text-rose-400",
        bgColor: "bg-rose-500/10",
        borderColor: "border-rose-500/20"
      };
    case "unhealthy":
      return {
        label: "不健康",
        dotColor: "bg-amber-500",
        ringColor: "bg-amber-400",
        textColor: "text-amber-400",
        bgColor: "bg-amber-500/10",
        borderColor: "border-amber-500/20"
      };
    default:
      return {
        label: "未知",
        dotColor: "bg-gray-500",
        ringColor: "bg-gray-500",
        textColor: "text-gray-400",
        bgColor: "bg-gray-500/10",
        borderColor: "border-gray-500/20"
      };
  }
};

export const Dashboard: React.FC = () => {
  const { theme } = useTheme();
  const [servers, setServers] = useState<ServerInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [searchQuery, setSearchQuery] = useState("");
  const [statusFilter, setStatusFilter] = useState<"all" | StatusKey>("all");
  const [autoRefresh, setAutoRefresh] = useState(true);
  const [lastUpdated, setLastUpdated] = useState<Date | null>(null);
  const [expandedServer, setExpandedServer] = useState<string | null>(null);
  const [activeTabMap, setActiveTabMap] = useState<Record<string, "tools" | "prompts" | "resources">>({});

  const fetchServers = async () => {
    try {
      setLoading(true);
      const res = await api.get<ServerInfo[]>("/api/servers");
      const list = Array.isArray(res.data) ? res.data : [];
      setServers(list);
      setLastUpdated(new Date());
      setError("");
    } catch (err: any) {
      if (err?.response?.status === 401) {
        setError("认证已过期，请重新登录");
      } else {
        setError(err?.response?.data?.message || "拉取服务器状态失败，请检查服务连接");
      }
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchServers();
  }, []);

  useEffect(() => {
    if (!autoRefresh) return;
    const interval = setInterval(fetchServers, 15000); // 15 秒刷新
    return () => clearInterval(interval);
  }, [autoRefresh]);

  const toggleExpand = (name: string) => {
    setExpandedServer(expandedServer === name ? null : name);
    if (!activeTabMap[name]) {
      setActiveTabMap(prev => ({ ...prev, [name]: "tools" }));
    }
  };

  const handleTabChange = (serverName: string, tab: "tools" | "prompts" | "resources") => {
    setActiveTabMap(prev => ({ ...prev, [serverName]: tab }));
  };

  // 过滤后的服务器（搜索 + 状态筛选）
  const filteredServers = servers.filter(s => {
    const q = searchQuery.trim().toLowerCase();
    const matchesSearch = !q ||
      (s.Name || "").toLowerCase().includes(q) ||
      (s.Route || "").toLowerCase().includes(q) ||
      (s.Description || "").toLowerCase().includes(q) ||
      (s.Status || "").toLowerCase().includes(q);

    const statusKey = normalizeStatus(s.Status);
    const matchesStatus = statusFilter === "all" || statusKey === statusFilter;

    return matchesSearch && matchesStatus;
  });

  // 统计信息
  const totalServers = servers.length;
  const connectedCount = servers.filter(s => normalizeStatus(s.Status) === "connected").length;
  const failedCount = servers.filter(s => normalizeStatus(s.Status) === "failed").length;
  const unhealthyCount = servers.filter(s => normalizeStatus(s.Status) === "unhealthy").length;

  // 总能力统计
  const totalTools = servers.reduce((sum, s) => sum + (s.Tools?.length ?? 0), 0);
  const totalPrompts = servers.reduce((sum, s) => sum + (s.Prompts?.length ?? 0), 0);
  const totalResources = servers.reduce((sum, s) => sum + (s.Resources?.length ?? 0), 0);

  return (
    <div className="space-y-8 stagger-in">
      {/* ========== Hero 区域 ========== */}
      <section className="relative overflow-hidden rounded-3xl border border-white/10 glass-card p-8 md:p-10">
        {/* 背景装饰 */}
        <div className="absolute -top-20 -right-20 w-72 h-72 rounded-full bg-violet-600/10 blur-3xl pointer-events-none" />
        <div className="absolute -bottom-24 -left-10 w-72 h-72 rounded-full bg-indigo-600/10 blur-3xl pointer-events-none" />

        <div className="relative flex flex-col md:flex-row md:items-center justify-between gap-6">
          <div className="flex-1">
            <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-violet-500/10 border border-violet-500/20 text-violet-300 text-[11px] font-bold uppercase tracking-wider mb-4">
              <Rocket className="w-3.5 h-3.5" />
              <span>v1.0.0 · React SPA 控制台</span>
            </div>
            <h1 className="text-3xl md:text-4xl font-extrabold text-white tracking-tight leading-tight">
              实时监控你的 <span className="bg-clip-text text-transparent bg-gradient-to-r from-violet-400 via-fuchsia-400 to-indigo-400">MCP 网关</span>
            </h1>
            <p className={`text-sm md:text-base mt-3 max-w-2xl leading-relaxed ${theme === "dark" ? "text-gray-400" : "text-gray-600"}`}>
              本仪表盘将所有接入的 MCP Server 的状态、路由端点、工具 / 提示 / 资源清单集中呈现，
              并支持 15 秒自动刷新与搜索过滤，帮助你第一时间发现接入异常。
            </p>

            {/* 核心亮点 */}
            <div className="flex flex-wrap items-center gap-2 mt-5">
              {[
                { icon: <Zap className="w-3.5 h-3.5" />, label: "零延迟状态上报" },
                { icon: <ShieldCheck className="w-3.5 h-3.5" />, label: "Bearer Token 鉴权" },
                { icon: <ArrowRightLeft className="w-3.5 h-3.5" />, label: "Claude / Trae 格式互转" },
                { icon: <Boxes className="w-3.5 h-3.5" />, label: "Tools / Prompts / Resources" },
              ].map((item, idx) => (
                <div key={idx} className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg bg-white/5 border border-white/5 text-[11px] text-gray-400 font-medium">
                  <span className="text-violet-400">{item.icon}</span>
                  {item.label}
                </div>
              ))}
            </div>
          </div>

          {/* 右侧汇总数字卡片 */}
          <div className="grid grid-cols-2 gap-3 md:w-64 shrink-0">
            <StatMiniCard label="服务" value={totalServers} color="violet" />
            <StatMiniCard label="已连接" value={connectedCount} color="emerald" />
            <StatMiniCard label="异常" value={failedCount + unhealthyCount} color="rose" />
            <StatMiniCard label="能力" value={totalTools + totalPrompts + totalResources} color="indigo" />
          </div>
        </div>
      </section>

      {/* ========== 指标卡片行 ========== */}
      <section className="grid grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard
          label="托管服务"
          value={totalServers}
          icon={<Server className="w-5 h-5" />}
          color="violet"
          sub={`共 ${totalTools} 工具 / ${totalPrompts} 提示 / ${totalResources} 资源`}
        />
        <StatCard
          label="正常连接"
          value={connectedCount}
          icon={<CheckCircle2 className="w-5 h-5" />}
          color="emerald"
          sub={totalServers > 0 ? `可用率 ${Math.round((connectedCount / totalServers) * 100)}%` : "等待接入"}
        />
        <StatCard
          label="不健康"
          value={unhealthyCount}
          icon={<AlertCircle className="w-5 h-5" />}
          color="amber"
          sub={unhealthyCount > 0 ? "建议查看详情排查" : "一切正常"}
        />
        <StatCard
          label="失败"
          value={failedCount}
          icon={<XCircle className="w-5 h-5" />}
          color="rose"
          sub={failedCount > 0 ? "请查看错误日志" : "无失败服务"}
        />
      </section>

      {/* 错误提示 */}
      {error && (
        <div className="flex items-start gap-3 bg-rose-500/10 border border-rose-500/20 text-rose-300 p-4 rounded-2xl">
          <AlertCircle className="w-5 h-5 text-rose-400 shrink-0 mt-0.5" />
          <div className="flex-1 text-sm">
            <div className="font-bold text-rose-300 mb-0.5">接口请求异常</div>
            <div className="text-rose-400/90">{error}</div>
          </div>
        </div>
      )}

      {/* ========== 工具栏 ========== */}
      <section className="flex flex-col gap-3 glass-card rounded-2xl p-4 border border-white/10">
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-3">
          <div className="flex items-center gap-3 flex-1">
            <div className="relative flex-1 md:max-w-sm">
              <Search className="absolute left-3.5 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-500" />
              <input
                type="text"
                placeholder="搜索服务器名 / 路由 / 描述..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-full bg-white/5 border border-white/10 rounded-xl pl-10 pr-4 py-2.5 text-sm text-white placeholder-gray-500 focus:bg-white/10 focus:border-violet-500/50 focus:outline-none transition-all duration-300"
              />
            </div>
            <span className={`text-[11px] hidden md:inline ${theme === "dark" ? "text-gray-500" : "text-gray-500"}`}>
              {filteredServers.length} / {servers.length} 匹配
            </span>
          </div>

          <div className="flex items-center gap-2">
            <span className={`text-[11px] hidden md:inline ${theme === "dark" ? "text-gray-500" : "text-gray-500"}`}>
              {lastUpdated ? `最后更新 ${lastUpdated.toLocaleTimeString()}` : "加载中..."}
            </span>
            <button
              onClick={() => setAutoRefresh(!autoRefresh)}
              className={`flex items-center gap-1.5 px-3 py-2 rounded-xl border text-xs font-bold tracking-wider transition-all duration-300 cursor-pointer ${
                autoRefresh
                  ? "bg-violet-500/10 border-violet-500/30 text-violet-300 hover:bg-violet-500/20"
                  : "bg-white/5 border-white/10 text-gray-400 hover:bg-white/10"
              }`}
            >
              <span className={`w-1.5 h-1.5 rounded-full ${autoRefresh ? "bg-violet-400 animate-pulse" : "bg-gray-500"}`} />
              <span>自动 {autoRefresh ? "开" : "关"}</span>
            </button>
            <button
              onClick={fetchServers}
              disabled={loading}
              className="p-2 bg-white/5 hover:bg-white/10 border border-white/10 rounded-xl text-gray-300 hover:text-white transition-all duration-300 cursor-pointer disabled:opacity-50"
              title="手动刷新"
            >
              <RotateCw className={`w-4 h-4 ${loading ? "animate-spin" : ""}`} />
            </button>
          </div>
        </div>

        {/* 状态筛选器 */}
        <div className="flex flex-wrap items-center gap-2 pt-2 border-t border-white/5">
          <span className="text-[10px] font-bold uppercase tracking-wider text-gray-500 mr-1">状态筛选:</span>
          {[
            { key: "all", label: "全部", count: servers.length, color: "violet" },
            { key: "connected", label: "已连接", count: servers.filter(s => normalizeStatus(s.Status) === "connected").length, color: "emerald" },
            { key: "failed", label: "失败", count: servers.filter(s => normalizeStatus(s.Status) === "failed").length, color: "rose" },
            { key: "unhealthy", label: "不健康", count: servers.filter(s => normalizeStatus(s.Status) === "unhealthy").length, color: "amber" },
            { key: "unknown", label: "未知", count: servers.filter(s => normalizeStatus(s.Status) === "unknown").length, color: "gray" }
          ].map(item => {
            const isActive = statusFilter === item.key;
            const colorClasses: Record<string, { active: string; inactive: string; dot: string }> = {
              violet: { active: "bg-violet-500/15 border-violet-500/40 text-violet-300", inactive: "bg-white/5 border-white/10 text-gray-400 hover:bg-white/10", dot: "bg-violet-400" },
              emerald: { active: "bg-emerald-500/15 border-emerald-500/40 text-emerald-300", inactive: "bg-white/5 border-white/10 text-gray-400 hover:bg-white/10", dot: "bg-emerald-400" },
              rose: { active: "bg-rose-500/15 border-rose-500/40 text-rose-300", inactive: "bg-white/5 border-white/10 text-gray-400 hover:bg-white/10", dot: "bg-rose-400" },
              amber: { active: "bg-amber-500/15 border-amber-500/40 text-amber-300", inactive: "bg-white/5 border-white/10 text-gray-400 hover:bg-white/10", dot: "bg-amber-400" },
              gray: { active: "bg-gray-500/15 border-gray-500/40 text-gray-300", inactive: "bg-white/5 border-white/10 text-gray-400 hover:bg-white/10", dot: "bg-gray-400" }
            };
            const c = colorClasses[item.color];
            return (
              <button
                key={item.key}
                onClick={() => setStatusFilter(item.key as any)}
                className={`flex items-center gap-1.5 px-2.5 py-1.5 rounded-lg border text-[10px] font-bold uppercase tracking-wider transition-all duration-300 cursor-pointer ${isActive ? c.active : c.inactive}`}
              >
                <span className={`w-1.5 h-1.5 rounded-full ${c.dot}`} />
                <span>{item.label}</span>
                <span className={`px-1.5 py-0.5 rounded-md text-[9px] ${isActive ? "bg-black/30" : "bg-black/20"}`}>{item.count}</span>
              </button>
            );
          })}
        </div>
      </section>

      {/* ========== 服务器列表 ========== */}
      {filteredServers.length === 0 ? (
        <EmptyState hasServers={servers.length > 0} />
      ) : (
        <div className="grid grid-cols-1 gap-4">
          {filteredServers.map((server) => {
            const isExpanded = expandedServer === server.Name;
            const currentTab = activeTabMap[server.Name] || "tools";
            const meta = statusMeta(normalizeStatus(server.Status));
            const tools = server.Tools ?? [];
            const prompts = server.Prompts ?? [];
            const resources = server.Resources ?? [];

            return (
              <div
                key={server.Name}
                className={`glass-card rounded-3xl overflow-hidden transition-all duration-500 border ${
                  isExpanded
                    ? "border-violet-500/30 shadow-[0_4px_32px_rgba(99,102,241,0.08)]"
                    : "border-white/10 hover:border-white/20"
                }`}
              >
                {/* 头部 — 可点击展开 */}
                <div
                  onClick={() => toggleExpand(server.Name)}
                  className="p-6 flex items-center justify-between cursor-pointer select-none"
                >
                  <div className="flex items-center gap-4">
                    {/* 状态指示灯 */}
                    <div className="relative flex h-3 w-3 shrink-0">
                      {normalizeStatus(server.Status) === "connected" && (
                        <span className="animate-ping absolute inline-flex h-full w-full rounded-full opacity-75" style={{ background: "currentColor", color: "#10b981" }} />
                      )}
                      {normalizeStatus(server.Status) === "unhealthy" && (
                        <span className="animate-ping absolute inline-flex h-full w-full rounded-full opacity-75" style={{ background: "currentColor", color: "#f59e0b" }} />
                      )}
                      <span className={`relative inline-flex rounded-full h-3 w-3 ${meta.dotColor}`} />
                    </div>

                    {/* 名称 & 描述 */}
                    <div className="min-w-0">
                      <div className="flex flex-wrap items-center gap-2">
                        <h2 className="text-lg font-extrabold text-white tracking-tight truncate">
                          {server.Name}
                        </h2>
                        <span className={`text-[10px] px-2 py-0.5 rounded-full border font-bold uppercase tracking-wide ${meta.bgColor} ${meta.textColor} ${meta.borderColor}`}>
                          {meta.label}
                        </span>
                      </div>
                      {server.Description && (
                        <p className={`text-xs mt-1 leading-relaxed line-clamp-1 max-w-xl ${theme === "dark" ? "text-gray-400" : "text-gray-600"}`}>
                          {server.Description}
                        </p>
                      )}
                      <div className={`flex flex-wrap items-center gap-3 mt-2 text-[11px] ${theme === "dark" ? "text-gray-500" : "text-gray-500"}`}>
                        <span className={`flex items-center gap-1.5 px-2 py-0.5 rounded-md font-mono ${theme === "dark" ? "bg-black/30 text-gray-400" : "bg-gray-100 text-gray-600"}`}>
                          <Layers className="w-3 h-3" />
                          {server.Route || "/"}
                        </span>
                        <span className={`flex items-center gap-1 ${theme === "dark" ? "text-gray-400" : "text-gray-600"}`}>
                          <Cpu className="w-3 h-3" /> {tools.length} Tools
                        </span>
                        <span className={`flex items-center gap-1 ${theme === "dark" ? "text-gray-400" : "text-gray-600"}`}>
                          <Code2 className="w-3 h-3" /> {prompts.length} Prompts
                        </span>
                        <span className={`flex items-center gap-1 ${theme === "dark" ? "text-gray-400" : "text-gray-600"}`}>
                          <FileText className="w-3 h-3" /> {resources.length} Resources
                        </span>
                      </div>
                    </div>
                  </div>

                  <div className="flex items-center gap-2 shrink-0">
                    <a
                      href={server.Route || "/"}
                      onClick={(e) => e.stopPropagation()}
                      target="_blank"
                      rel="noreferrer"
                      className="hidden sm:inline-flex items-center gap-1.5 px-3 py-1.5 bg-white/5 hover:bg-violet-500/10 border border-white/10 hover:border-violet-500/30 rounded-xl text-xs font-bold text-gray-300 hover:text-violet-300 transition-all duration-300"
                    >
                      打开端点
                      <ArrowRightInline className="w-3.5 h-3.5" />
                    </a>
                    {isExpanded ? (
                      <ChevronUp className="w-5 h-5 text-gray-400" />
                    ) : (
                      <ChevronDown className="w-5 h-5 text-gray-400" />
                    )}
                  </div>
                </div>

                {/* 展开区域 */}
                {isExpanded && (
                  <div className={`border-t p-6 space-y-5 ${theme === "dark" ? "border-white/5 bg-black/20" : "border-gray-200 bg-gray-50/50"}`}>
                    {/* 错误信息 */}
                    {server.Error && (
                      <div className="p-4 bg-rose-500/10 border border-rose-500/20 rounded-2xl flex items-start gap-3">
                        <AlertCircle className="w-4 h-4 shrink-0 mt-0.5 text-rose-400" />
                        <div className="flex-1">
                          <div className="text-xs font-bold text-rose-300 mb-1 uppercase tracking-wider">
                            错误日志
                          </div>
                          <div className="text-xs text-rose-400/90 font-mono break-all leading-relaxed">
                            {server.Error}
                          </div>
                        </div>
                      </div>
                    )}

                    {/* 选项卡导航 */}
                    <div className={`flex border-b gap-1 overflow-x-auto ${theme === "dark" ? "border-white/5" : "border-gray-200"}`}>
                      {([
                        { id: "tools", label: `Tools`, count: tools.length, icon: <Cpu className="w-3.5 h-3.5" /> },
                        { id: "prompts", label: `Prompts`, count: prompts.length, icon: <Code2 className="w-3.5 h-3.5" /> },
                        { id: "resources", label: `Resources`, count: resources.length, icon: <FileText className="w-3.5 h-3.5" /> },
                      ] as const).map((tab) => {
                        const active = currentTab === tab.id;
                        return (
                          <button
                            key={tab.id}
                            onClick={(e) => {
                              e.stopPropagation();
                              handleTabChange(server.Name, tab.id);
                            }}
                            className={`flex items-center gap-2 pb-3 px-2 text-xs font-bold tracking-wider uppercase transition-all duration-300 cursor-pointer whitespace-nowrap border-b-2 -mb-px ${
                              active
                                ? "border-violet-500 text-violet-400"
                                : "border-transparent text-gray-500 hover:text-gray-300"
                            }`}
                          >
                            {tab.icon}
                            <span>{tab.label}</span>
                            <span className={`px-1.5 py-0.5 rounded-md text-[10px] ${active ? "bg-violet-500/15 text-violet-300" : "bg-white/5 text-gray-500"}`}>
                              {tab.count}
                            </span>
                          </button>
                        );
                      })}
                    </div>

                    {/* Tab 内容 */}
                    <div>
                      {currentTab === "tools" && (
                        <ToolList tools={tools} empty="此服务未对外暴露任何 Tool" />
                      )}
                      {currentTab === "prompts" && (
                        <PromptList prompts={prompts} empty="此服务未对外暴露任何 Prompt" />
                      )}
                      {currentTab === "resources" && (
                        <ResourceList resources={resources} empty="此服务未对外暴露任何 Resource" />
                      )}
                    </div>
                  </div>
                )}
              </div>
            );
          })}
        </div>
      )}

      {/* ========== 使用提示 ========== */}
      <section className="glass-card rounded-3xl p-6 md:p-8 border border-white/10">
        <div className="flex items-start gap-4">
          <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-violet-600/20 to-indigo-600/20 border border-violet-500/20 flex items-center justify-center shrink-0">
            <Zap className="w-5 h-5 text-violet-300" />
          </div>
          <div className="flex-1">
            <h3 className="font-bold text-white mb-1">快速接入你的客户端</h3>
            <p className="text-sm text-gray-400 leading-relaxed">
              复制上方任一端点 URL，粘贴到 Claude Desktop / Trae / Antigravity 等 MCP 客户端的配置文件即可直接使用。
              如果网关启用了鉴权，请记得在请求头附带 <code className="text-violet-300 font-mono text-[11px] bg-black/40 px-1.5 py-0.5 rounded">Authorization: Bearer &lt;token&gt;</code>。
              需要批量转换配置？前往 <span className="text-violet-300 font-bold">配置转换</span> 页面一键生成。
            </p>
          </div>
        </div>
      </section>
    </div>
  );
};

// —— 小组件 ——
const colorMap: Record<string, { bg: string; text: string; ring: string; border: string }> = {
  violet: { bg: "bg-violet-500/15", text: "text-violet-400", ring: "text-violet-300", border: "border-violet-500/20" },
  emerald: { bg: "bg-emerald-500/15", text: "text-emerald-400", ring: "text-emerald-300", border: "border-emerald-500/20" },
  amber: { bg: "bg-amber-500/15", text: "text-amber-400", ring: "text-amber-300", border: "border-amber-500/20" },
  rose: { bg: "bg-rose-500/15", text: "text-rose-400", ring: "text-rose-300", border: "border-rose-500/20" },
  indigo: { bg: "bg-indigo-500/15", text: "text-indigo-400", ring: "text-indigo-300", border: "border-indigo-500/20" },
};

const StatCard: React.FC<{ label: string; value: number; icon: React.ReactNode; color: keyof typeof colorMap; sub?: string }> = ({
  label, value, icon, color, sub
}) => {
  const c = colorMap[color];
  return (
    <div className={`glass-card rounded-2xl p-5 border transition-all duration-300 ${theme === "dark" ? "border-white/10 hover:border-white/15" : "border-gray-200 hover:border-gray-300"}`}>
      <div className="flex items-start justify-between mb-3">
        <p className={`text-[11px] font-bold uppercase tracking-wider ${theme === "dark" ? "text-gray-500" : "text-gray-500"}`}>{label}</p>
        <div className={`p-2 rounded-xl ${c.bg} ${c.text}`}>{icon}</div>
      </div>
      <h3 className={`text-3xl font-extrabold leading-none ${theme === "dark" ? "text-white" : "text-gray-900"}`}>{value}</h3>
      {sub && <p className={`text-[11px] mt-2 ${c.text}`}>{sub}</p>}
    </div>
  );
};

const StatMiniCard: React.FC<{ label: string; value: number; color: keyof typeof colorMap }> = ({ label, value, color }) => {
  const c = colorMap[color];
  return (
    <div className={`p-3 rounded-2xl border ${c.border} ${c.bg} text-center`}>
      <div className={`text-2xl font-extrabold ${c.text}`}>{value}</div>
      <div className={`text-[10px] mt-0.5 font-bold uppercase tracking-wider ${theme === "dark" ? "text-gray-500" : "text-gray-500"}`}>{label}</div>
    </div>
  );
};

const ArrowRightInline: React.FC<{ className?: string }> = ({ className }) => (
  <svg className={className} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <path d="M5 12h14M13 5l7 7-7 7" />
  </svg>
);

const EmptyState: React.FC<{ hasServers: boolean }> = ({ hasServers }) => (
  <div className={`glass-card rounded-3xl p-12 text-center border ${theme === "dark" ? "border-white/10" : "border-gray-200"}`}>
    <div className="w-14 h-14 rounded-2xl bg-violet-500/10 border border-violet-500/20 flex items-center justify-center mx-auto mb-4">
      <Server className="w-7 h-7 text-violet-400" />
    </div>
    <h3 className={`text-lg font-extrabold mb-1 ${theme === "dark" ? "text-white" : "text-gray-900"}`}>
      {hasServers ? "没有匹配的服务器" : "暂无已注册的 MCP 服务器"}
    </h3>
    <p className={`text-sm max-w-md mx-auto leading-relaxed ${theme === "dark" ? "text-gray-400" : "text-gray-600"}`}>
      {hasServers
        ? "请尝试修改搜索关键词或清空筛选条件。"
        : "请先在后端 config.json 中配置一个或多个 MCP 服务，重启网关后刷新本页面即可看到实时状态。"}
    </p>
  </div>
);

const ToolList: React.FC<{ tools: MCPTool[]; empty: string }> = ({ tools, empty }) => {
  if (tools.length === 0) {
    return <p className={`text-xs italic py-2 ${theme === "dark" ? "text-gray-500" : "text-gray-500"}`}>{empty}</p>;
  }
  return (
    <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
      {tools.map((t, idx) => (
        <div className={`p-4 rounded-2xl border transition-all duration-300 ${theme === "dark" ? "bg-white/5 border-white/10 hover:border-violet-500/20" : "bg-white border-gray-200 hover:border-violet-300"}`}>
          <div className="flex items-start justify-between gap-3 mb-2">
            <div className="flex items-center gap-2 min-w-0">
              <div className="p-1.5 rounded-lg bg-violet-500/10 text-violet-400 shrink-0">
                <Cpu className="w-3.5 h-3.5" />
              </div>
              <h4 className={`font-bold text-sm truncate ${theme === "dark" ? "text-white" : "text-gray-900"}`}>{toolName(t)}</h4>
            </div>
          </div>
          {toolDesc(t) && (
            <p className={`text-xs leading-relaxed pl-9 ${theme === "dark" ? "text-gray-400" : "text-gray-600"}`}>{toolDesc(t)}</p>
          )}
          {toolSchema(t) && (
            <div className="mt-3 pl-9">
              <div className={`text-[10px] font-bold uppercase tracking-wider mb-1.5 ${theme === "dark" ? "text-gray-500" : "text-gray-500"}`}>参数 Schema</div>
              <pre className={`text-[10px] border p-3 rounded-xl overflow-x-auto font-mono leading-relaxed max-h-32 ${theme === "dark" ? "bg-black/40 border-white/5 text-gray-400" : "bg-gray-100 border-gray-200 text-gray-700"}`}>
                {JSON.stringify(toolSchema(t), null, 2)}
              </pre>
            </div>
          )}
        </div>
      ))}
    </div>
  );
};

const PromptList: React.FC<{ prompts: MCPPrompt[]; empty: string }> = ({ prompts, empty }) => {
  if (prompts.length === 0) {
    return <p className={`text-xs italic py-2 ${theme === "dark" ? "text-gray-500" : "text-gray-500"}`}>{empty}</p>;
  }
  return (
    <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
      {prompts.map((p, idx) => (
        <div className={`p-4 rounded-2xl border transition-all duration-300 ${theme === "dark" ? "bg-white/5 border-white/10 hover:border-indigo-500/20" : "bg-white border-gray-200 hover:border-indigo-300"}`}>
          <div className="flex items-start gap-2 mb-2">
            <div className="p-1.5 rounded-lg bg-indigo-500/10 text-indigo-400 shrink-0">
              <Code2 className="w-3.5 h-3.5" />
            </div>
            <h4 className={`font-bold text-sm ${theme === "dark" ? "text-white" : "text-gray-900"}`}>{promptName(p)}</h4>
          </div>
          {promptDesc(p) && (
            <p className={`text-xs leading-relaxed pl-9 ${theme === "dark" ? "text-gray-400" : "text-gray-600"}`}>{promptDesc(p)}</p>
          )}
          {promptArgs(p).length > 0 && (
            <div className="mt-3 pl-9">
              <div className={`text-[10px] font-bold uppercase tracking-wider mb-1.5 ${theme === "dark" ? "text-gray-500" : "text-gray-500"}`}>参数</div>
              <div className="flex flex-wrap gap-1.5">
                {promptArgs(p).map((arg, aidx) => {
                  const name: string = (arg as any)?.name || (typeof arg === "string" ? arg : `arg${aidx + 1}`);
                  return (
                    <span key={aidx} className={`text-[10px] px-2 py-0.5 rounded-md border font-mono ${theme === "dark" ? "bg-white/5 border-white/5 text-gray-400" : "bg-gray-100 border-gray-200 text-gray-600"}`}>
                      {name}
                    </span>
                  );
                })}
              </div>
            </div>
          )}
        </div>
      ))}
    </div>
  );
};

const ResourceList: React.FC<{ resources: MCPResource[]; empty: string }> = ({ resources, empty }) => {
  if (resources.length === 0) {
    return <p className={`text-xs italic py-2 ${theme === "dark" ? "text-gray-500" : "text-gray-500"}`}>{empty}</p>;
  }
  return (
    <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
      {resources.map((r, idx) => (
        <div className={`p-4 rounded-2xl border transition-all duration-300 ${theme === "dark" ? "bg-white/5 border-white/10 hover:border-emerald-500/20" : "bg-white border-gray-200 hover:border-emerald-300"}`}>
          <div className="flex items-start gap-2 mb-2">
            <div className="p-1.5 rounded-lg bg-emerald-500/10 text-emerald-400 shrink-0">
              <FileText className="w-3.5 h-3.5" />
            </div>
            <h4 className={`font-bold text-sm truncate ${theme === "dark" ? "text-white" : "text-gray-900"}`}>{resourceName(r)}</h4>
          </div>
          {resourceDesc(r) && (
            <p className={`text-xs leading-relaxed pl-9 ${theme === "dark" ? "text-gray-400" : "text-gray-600"}`}>{resourceDesc(r)}</p>
          )}
          {resourceUri(r) && (
            <div className="mt-3 pl-9">
              <div className={`text-[10px] font-bold uppercase tracking-wider mb-1.5 ${theme === "dark" ? "text-gray-500" : "text-gray-500"}`}>URI</div>
              <div className={`text-[11px] border px-2.5 py-1.5 rounded-lg font-mono truncate select-all ${theme === "dark" ? "bg-black/40 border-white/5 text-emerald-300/90" : "bg-gray-100 border-gray-200 text-emerald-600"}`}>
                {resourceUri(r)}
              </div>
            </div>
          )}
        </div>
      ))}
    </div>
  );
};
